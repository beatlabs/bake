package docker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
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

// RunOptions contains docker container run options.
type RunOptions struct {
	Cmd         []string
	InitExecCmd string
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
	RunOpts            *RunOptions
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
		fmt.Printf("Component %q is starting container %q\n", c.Name, container.Name)
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
			Name:           c.Name + ":" + session.id,
			Dockerfile:     conf.BuildOpts.Dockerfile,
			SuppressOutput: true,
			OutputStream:   os.Stdout,
			ErrorStream:    os.Stderr,
			ContextDir:     conf.BuildOpts.ContextDir,
			BuildArgs:      conf.BuildOpts.BuildArgs,
			RmTmpContainer: true,
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

	if conf.RunOpts != nil {
		runOpts.Cmd = conf.RunOpts.Cmd
	}

	const tcpSuffix = "/tcp"

	hostPorts := map[string]string{}
	for serviceName, nativePort := range conf.ServicePorts {
		runOpts.ExposedPorts = append(runOpts.ExposedPorts, nativePort+tcpSuffix)

		if !session.inDocker {
			// staticPort means that we should map this port 1 to 1 on the host,
			// trusting that the component has obtained a random one.
			staticPort, ok := conf.StaticServicePorts[serviceName]
			if ok {
				runOpts.ExposedPorts = append(runOpts.ExposedPorts, staticPort+tcpSuffix)
				runOpts.PortBindings[docker.Port(staticPort+tcpSuffix)] = []docker.PortBinding{
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
			runOpts.PortBindings[docker.Port(nativePort+tcpSuffix)] = []docker.PortBinding{
				{HostIP: "0.0.0.0", HostPort: mappedPort},
			}
			hostPorts[serviceName] = mappedPort
		}
	}

	publishPorts, _ := strconv.ParseBool(os.Getenv("BAKE_PUBLISH_PORTS"))
	hcOpts := func(hc *docker.HostConfig) { hc.PublishAllPorts = publishPorts }
	resource, err := pool.RunWithOptions(runOpts, hcOpts)
	if err != nil {
		return fmt.Errorf("run %s: %w", fullContainerName, err)
	}

	// Update session service registry.
	for serviceName, port := range conf.ServicePorts {
		err := session.RegisterInternalDockerService(serviceName, runOpts.Name+":"+port)
		if err != nil {
			return fmt.Errorf("register service %s: %w", serviceName, err)
		}
		if !session.inDocker {
			hport, ok := hostPorts[serviceName]
			if !ok {
				return fmt.Errorf("host service port not found for service %s", serviceName)
			}
			err := session.RegisterHostMappedDockerService(serviceName, "localhost:"+hport)
			if err != nil {
				return fmt.Errorf("register host service %s: %w", serviceName, err)
			}
		}
	}

	if conf.ReadyFunc != nil {
		err = conf.ReadyFunc(session)
		if err != nil {
			return err
		}
	}

	if conf.RunOpts != nil && conf.RunOpts.InitExecCmd != "" {
		return pool.Retry(func() error {
			_, err := resource.Exec([]string{"bash", "-c", conf.RunOpts.InitExecCmd},
				dockertest.ExecOptions{
					StdOut: bufio.NewWriter(os.Stdout),
					StdErr: bufio.NewWriter(os.Stdout),
				})
			return err
		})
	}

	return nil
}

// RetryMaxTimeout is the timeout for the default retry func.
var RetryMaxTimeout = 5 * time.Minute

// Retry is an exponential backoff retry helper.
// All built-in components use this func to detect whether a container is alive and ready.
// User supplied components may use this helper func or provide their own.
func Retry(op func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 2
	bo.MaxElapsedTime = RetryMaxTimeout
	return backoff.Retry(op, bo)
}

// GetFreePort tries to find a free port on the current machine.
func GetFreePort() (string, error) {
	listernCfg := net.ListenConfig{
		// Configure the listener as needed
	}

	l, err := listernCfg.Listen(context.Background(), "tcp", ":0") // similar to
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
