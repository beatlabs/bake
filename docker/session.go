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
	"regexp"
	"strings"
	"sync"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"golang.org/x/sync/errgroup"
)

const SessionFile = ".bakesession"

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

func (s *Session) ID() string {
	return s.id
}

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

func (s *Session) RegisterInternalDockerSevice(serviceName, endpoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.serviceAddresses[serviceName]
	if ok {
		return fmt.Errorf("service %q which already exists with value: %q", serviceName, v)
	}

	s.serviceAddresses[serviceName] = endpoint
	return nil
}

func (s *Session) RegisterHostMappedDockerSevice(serviceName, endpoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.hostMappedServiceAddresses[serviceName]
	if ok {
		return fmt.Errorf("service %q which already exists with value: %q", serviceName, v)
	}

	s.hostMappedServiceAddresses[serviceName] = endpoint
	return nil
}

func (s *Session) DockerToDockerServiceAddress(serviceName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr, ok := s.serviceAddresses[serviceName]
	if !ok {
		return "", fmt.Errorf("internal service address not registered for %q", serviceName)
	}

	return addr, nil
}

func (s *Session) HostToDockerServiceAddress(serviceName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr, ok := s.hostMappedServiceAddresses[serviceName]
	if !ok {
		return "", fmt.Errorf("external service address not registered for %q", serviceName)
	}

	return addr, nil
}

func (s *Session) AutoServiceAddress(serviceName string) (string, error) {
	if s.inDocker {
		return s.DockerToDockerServiceAddress(serviceName)
	}
	return s.HostToDockerServiceAddress(serviceName)
}

func (s *Session) WriteToFile(fpath string) error {
	if inDocker() {
		return errors.New("not supported")
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

type sessionDump struct {
	ID                         string
	NetworkID                  string
	ServiceAddresses           map[string]string
	HostMappedServiceAddresses map[string]string
}

func FromFile(fpath string) (*Session, error) {
	if inDocker() {
		return nil, errors.New("not supported inside of docker")
	}

	data, err := ioutil.ReadFile(fpath)
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

func CleanupResources() error {
	re, err := regexp.Compile("^.bake.*")
	if err != nil {
		return err
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err == nil && re.MatchString(info.Name()) {
			cleanupSessionResources(path)
		}
		return nil
	})
	return err
}

func cleanupSessionResources(fname string) error {
	fmt.Println(fname)

	session, err := FromFile(fname)
	if err != nil {
		return err
	}

	var pool *dockertest.Pool
	pool, err = dockertest.NewPool("")
	if err != nil {
		return err
	}

	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.HasPrefix(name, "/"+session.id) {
				fmt.Println(name)
				pool.RemoveContainerByName(name)
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

	err = os.Remove(fname)
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

func FromEnv() (sessionID string, networkID string, err error) {
	sessionID = os.Getenv("BAKE_SESSION_ID")
	if sessionID == "" {
		sessionID = "000"
	}

	networkID = os.Getenv("BAKE_NETWORK_ID")
	if networkID == "" {
		networkID, _ = createNetwork(sessionID)
	}

	return
}

func inDocker() bool {
	_, staterr := os.Stat("/.dockerenv")
	return !os.IsNotExist(staterr)
}
