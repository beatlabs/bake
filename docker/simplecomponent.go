package docker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type BuildOptions struct {
	Dockerfile string
	ContextDir string
	BuildArgs  []docker.BuildArg
}

type SimpleContainerConfig struct {
	Name                   string
	Repository             string
	Tag                    string
	Env                    []string
	BuildOpts              *BuildOptions
	ServicePorts           map[string]string
	MappedHostServicePorts map[string]string
	FixedHostServicePorts  map[string]string
	ReadyFunc              func(*Session) error
}

type SimpleContainerOptionFunc func(SimpleContainerConfig) SimpleContainerConfig

func WithTag(tag string) SimpleContainerOptionFunc {
	return func(c SimpleContainerConfig) SimpleContainerConfig {
		c.Tag = tag
		return c
	}
}

// SimpleComponent groups together several containers.
type SimpleComponent struct {
	Name       string
	Containers []SimpleContainerConfig
}

// Start all containers sequentially.
func (c *SimpleComponent) Start(session *Session) error {
	if len(c.Containers) == 0 {
		return fmt.Errorf("component %s has no containers to start", c.Name)
	}

	for _, container := range c.Containers {
		log.Printf("Component %q is starting container %q", c.Name, container.Name)
		err := c.RunContainer(session, container)
		if err != nil {
			return fmt.Errorf("starting component %q: %w", container.Name, err)
		}
	}

	return nil
}

func (c *SimpleComponent) RunContainer(session *Session, conf SimpleContainerConfig) error {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return err
	}

	if conf.BuildOpts != nil {
		err := pool.Client.BuildImage(docker.BuildImageOptions{
			Name:         c.Name + ":" + session.id,
			Dockerfile:   conf.BuildOpts.Dockerfile,
			OutputStream: ioutil.Discard,
			ContextDir:   conf.BuildOpts.ContextDir,
			BuildArgs:    conf.BuildOpts.BuildArgs,
		})
		if err != nil {
			return err
		}
		conf.Repository = c.Name
		conf.Tag = session.id
	}

	fullContainerName := session.id + "-" + conf.Name
	runOpts := &dockertest.RunOptions{
		Name:         fullContainerName,
		NetworkID:    session.networkID,
		Tag:          conf.Tag,
		Repository:   conf.Repository,
		Env:          conf.Env,
		PortBindings: map[docker.Port][]docker.PortBinding{},
		ExposedPorts: []string{},
	}

	for serviceName, nativePort := range conf.ServicePorts {
		runOpts.ExposedPorts = append(runOpts.ExposedPorts, nativePort+"/tcp")
		if !session.inDocker {
			var mappedPort string
			var fixedPort string
			var ok bool
			var err error
			fixedPort, ok = conf.FixedHostServicePorts[serviceName]
			if ok {
				runOpts.ExposedPorts = append(runOpts.ExposedPorts, fixedPort+"/tcp")
				runOpts.PortBindings[docker.Port(fixedPort+"/tcp")] = []docker.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: fixedPort,
					},
				}
			} else {
				mappedPort, err = GetFreePort()
				if err != nil {
					return fmt.Errorf("can not obtain random free port: %w", err)
				}
				if conf.MappedHostServicePorts == nil {
					conf.MappedHostServicePorts = map[string]string{}
				}
				conf.MappedHostServicePorts[serviceName] = mappedPort
			}

			fmt.Println(serviceName, nativePort, mappedPort)
			runOpts.PortBindings[docker.Port(nativePort+"/tcp")] = []docker.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: mappedPort,
				},
			}
		}
	}

	hcOpts := func(hc *docker.HostConfig) { hc.PublishAllPorts = false }
	_, err = pool.RunWithOptions(runOpts, hcOpts)
	if err != nil {
		return err
	}

	// Update session service registry.
	for serviceName, port := range conf.ServicePorts {
		err := session.RegisterInternalDockerSevice(serviceName, runOpts.Name+":"+port)
		if err != nil {
			return nil
		}
		if !session.inDocker {
			pport, ok := conf.FixedHostServicePorts[serviceName]
			if !ok {
				pport, ok = conf.MappedHostServicePorts[serviceName]
				if !ok {
					return fmt.Errorf("mapped service port not found for service %s", serviceName)
				}
			}
			err := session.RegisterHostMappedDockerSevice(serviceName, "localhost:"+pport)
			if err != nil {
				return nil
			}
		}

	}

	if conf.ReadyFunc != nil {
		err = conf.ReadyFunc(session)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *SimpleComponent) GetName() string {
	return c.Name
}

// StreamLogs streams container logs to stdout.
func (c *SimpleComponent) StreamLogs() {
	for i, cont := range c.Containers {
		go streamContainerLogs(cont.Name, colors[i%len(colors)])
	}
}

// Retry is an exponential backoff retry helper.
func Retry(op func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 2
	bo.MaxElapsedTime = time.Second * 15
	return backoff.Retry(op, bo)
}

func GetFreePort() (string, error) {
	/* #nosec */
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", fmt.Errorf("failed to obtain port: %w", err)
	}

	if err := l.Close(); err != nil {
		return "", fmt.Errorf("failed to close listener: %w", err)
	}

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return "", errors.New("failed to cast address to TCPAddr type")
	}

	return strconv.Itoa(tcpAddr.Port), nil
}
