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

// BuildOptions contains simple docker build options.
type BuildOptions struct {
	Dockerfile string
	ContextDir string
	BuildArgs  []docker.BuildArg
}

// SimpleContainerConfig defines a Docker container with associated service ports.
type SimpleContainerConfig struct {
	Name               string
	Repository         string
	Tag                string
	Env                []string
	BuildOpts          *BuildOptions
	ServicePorts       map[string]string
	StaticServicePorts map[string]string
	ReadyFunc          func(*Session) error
}

// SimpleContainerOptionFunc allows for customization of SimpleContainerConfigs.
type SimpleContainerOptionFunc func(*SimpleContainerConfig)

// WithTag sets a Docker tag in a SimpleContainerConfig.
func WithTag(tag string) SimpleContainerOptionFunc {
	return func(c *SimpleContainerConfig) {
		c.Tag = tag
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
		err := c.runContainer(session, container)
		if err != nil {
			return fmt.Errorf("starting component %q: %w", container.Name, err)
		}
	}

	return nil
}

func (c *SimpleComponent) runContainer(session *Session, conf SimpleContainerConfig) error {
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
			return fmt.Errorf("build image %s: %w", c.Name+":"+session.id, err)
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

	hostPorts := map[string]string{}
	for serviceName, nativePort := range conf.ServicePorts {
		runOpts.ExposedPorts = append(runOpts.ExposedPorts, nativePort+"/tcp")

		if !session.inDocker {
			// staticPort means that we should map this port 1 to 1 on the host,
			// trusting that the component has obtained a random one.
			staticPort, ok := conf.StaticServicePorts[serviceName]
			if ok {
				runOpts.ExposedPorts = append(runOpts.ExposedPorts, staticPort+"/tcp")
				runOpts.PortBindings[docker.Port(staticPort+"/tcp")] = []docker.PortBinding{
					{HostIP: "0.0.0.0", HostPort: staticPort},
				}
				hostPorts[serviceName] = staticPort
				continue
			}

			// by default we obtain a random port to publish for the native port for this service.
			mappedPort, err := GetFreePort()
			if err != nil {
				return fmt.Errorf("can not obtain random free port: %w", err)
			}
			runOpts.PortBindings[docker.Port(nativePort+"/tcp")] = []docker.PortBinding{
				{HostIP: "0.0.0.0", HostPort: mappedPort},
			}
			hostPorts[serviceName] = mappedPort
		}
	}

	hcOpts := func(hc *docker.HostConfig) { hc.PublishAllPorts = false }
	_, err = pool.RunWithOptions(runOpts, hcOpts)
	if err != nil {
		return fmt.Errorf("run %s: %w", fullContainerName, err)
	}

	// Update session service registry.
	for serviceName, port := range conf.ServicePorts {
		err := session.RegisterInternalDockerSevice(serviceName, runOpts.Name+":"+port)
		if err != nil {
			return nil
		}
		if !session.inDocker {
			hport, ok := hostPorts[serviceName]
			if !ok {
				return fmt.Errorf("host service port not found for service %s", serviceName)
			}
			err := session.RegisterHostMappedDockerSevice(serviceName, "localhost:"+hport)
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
	bo.MaxElapsedTime = time.Minute
	return backoff.Retry(op, bo)
}

// GetFreePort tries to find a free port on the current machine.
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
