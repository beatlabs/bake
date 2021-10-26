// Package docker contains Docker-related helpers for component tests.
package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"golang.org/x/sync/errgroup"
)

// DefaultSessionFile is the file name used for storing sessions.
const DefaultSessionFile = ".bakesession"

// Component is a logical service, it groups together several containers.
type Component interface {
	Start(*Session) error
}

// Session is the docker session, used to manage the lifecycle of components.
type Session struct {
	id                         string
	networkID                  string
	inDocker                   bool
	mu                         sync.Mutex
	serviceAddresses           map[string]string
	hostMappedServiceAddresses map[string]string
}

// NewSession prepares a new Docker session.
func NewSession(id, networkID string) (*Session, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	if networkID == "" || networkID == "bridge" {
		return nil, errors.New("networkID is required, bridge network not supported")
	}

	return &Session{
		id:                         id,
		networkID:                  networkID,
		inDocker:                   inDocker(),
		serviceAddresses:           map[string]string{},
		hostMappedServiceAddresses: map[string]string{},
	}, nil
}

// ID returns the Session ID.
func (s *Session) ID() string {
	return s.id
}

// IsInDocker indicates whether this session was started from inside a Docker container.
func (s *Session) IsInDocker() bool {
	return s.inDocker
}

// StartComponents starts the provided components.
func (s *Session) StartComponents(cs ...Component) error {
	g := errgroup.Group{}
	for _, c := range cs {
		c := c
		g.Go(func() error {
			return c.Start(s)
		})
	}
	return g.Wait()
}

// RegisterInternalDockerService registers an internal endpoint against the service name.
func (s *Session) RegisterInternalDockerService(serviceName, endpoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.serviceAddresses[serviceName]
	if ok {
		return fmt.Errorf("service %q which already exists with value: %q", serviceName, v)
	}

	s.serviceAddresses[serviceName] = endpoint
	return nil
}

// RegisterHostMappedDockerService registers a host mapped endpoint against the service name.
func (s *Session) RegisterHostMappedDockerService(serviceName, endpoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.hostMappedServiceAddresses[serviceName]
	if ok {
		return fmt.Errorf("service %q which already exists with value: %q", serviceName, v)
	}

	s.hostMappedServiceAddresses[serviceName] = endpoint
	return nil
}

// DockerToDockerServiceAddress retrieves an internal endpoint for a service name.
func (s *Session) DockerToDockerServiceAddress(serviceName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr, ok := s.serviceAddresses[serviceName]
	if !ok {
		return "", fmt.Errorf("internal service address not registered for %q", serviceName)
	}

	return addr, nil
}

// HostToDockerServiceAddress retrieves a host mapped endpoint for a service name.
func (s *Session) HostToDockerServiceAddress(serviceName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr, ok := s.hostMappedServiceAddresses[serviceName]
	if !ok {
		return "", fmt.Errorf("external service address not registered for %q", serviceName)
	}

	return addr, nil
}

// AutoServiceAddress retrieves an endpoint for a service name, appropriate for the running code.
func (s *Session) AutoServiceAddress(serviceName string) (string, error) {
	if s.inDocker {
		return s.DockerToDockerServiceAddress(serviceName)
	}
	return s.HostToDockerServiceAddress(serviceName)
}

// PersistToFile serializes a session and writes it to a file.
func (s *Session) PersistToFile(fpath string) error {
	if s.inDocker {
		return nil
	}

	b, err := json.MarshalIndent(sessionDump{
		ID:                         s.id,
		NetworkID:                  s.networkID,
		ServiceAddresses:           s.serviceAddresses,
		HostMappedServiceAddresses: s.hostMappedServiceAddresses,
	}, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path.Clean(fpath), b, 0600)
}

// Persist stores the session data in the default store.
func (s *Session) Persist() error {
	return s.PersistToFile(DefaultSessionFile)
}

// GetEnv retrieves bake related env vars, with defaults.
func GetEnv() (sessionID, networkID string, err error) {
	sessionID = os.Getenv("BAKE_SESSION_ID")
	if sessionID == "" {
		sessionID = "000"
	}

	networkID = os.Getenv("BAKE_NETWORK_ID")
	if networkID == "" {
		networkID, err = createNetwork(sessionID)
	}

	return
}

type sessionDump struct {
	ID                         string
	NetworkID                  string
	ServiceAddresses           map[string]string
	HostMappedServiceAddresses map[string]string
}

// LoadSession attempts to load a Session from the default file location.
func LoadSession() (*Session, error) {
	return LoadSessionFromFile(inDocker(), DefaultSessionFile)
}

// LoadSessionFromFile attempts to load a session from a file.
func LoadSessionFromFile(inDocker bool, fpath string) (*Session, error) {
	if inDocker {
		return nil, errors.New("not supported inside of docker")
	}

	data, err := ioutil.ReadFile(path.Clean(fpath))
	if err != nil {
		return nil, err
	}

	var d sessionDump
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}

	return &Session{
		id:                         d.ID,
		networkID:                  d.NetworkID,
		serviceAddresses:           d.ServiceAddresses,
		hostMappedServiceAddresses: d.HostMappedServiceAddresses,
	}, nil
}

// CleanupResources finds all session files and prunes Docker resources associated with them.
func CleanupResources() error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == DefaultSessionFile {
			if err := cleanupSessionResources(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func cleanupSessionResources(fname string) error {
	fmt.Println(fname)

	session, err := LoadSessionFromFile(inDocker(), fname)
	if err != nil {
		return err
	}

	err = os.Remove(fname)
	if err != nil {
		return err
	}

	var pool *dockertest.Pool
	pool, err = dockertest.NewPool("")
	if err != nil {
		return err
	}

	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.HasPrefix(name, "/"+session.id) {
				fmt.Println(name)
				err := pool.RemoveContainerByName(name)
				if err != nil {
					return err
				}
			}
		}
	}

	err = deleteNetwork(session.networkID)
	if err != nil {
		return err
	}

	return nil
}

func deleteNetwork(id string) error {
	var pool *dockertest.Pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		return err
	}

	return pool.Client.RemoveNetwork(id)
}

func createNetwork(id string) (string, error) {
	var pool *dockertest.Pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", err
	}

	var net *dockertest.Network
	net, err = pool.CreateNetwork(id)
	if err != nil {
		return "", err
	}
	return net.Network.ID, nil
}

func inDocker() bool {
	_, staterr := os.Stat("/.dockerenv")
	return !os.IsNotExist(staterr)
}
