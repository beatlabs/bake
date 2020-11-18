// Package component contains pre-configured components that can be used to run sets of containers connected together.
package component

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	// DefaultRuntimeExp expiration.
	DefaultRuntimeExp = 20 * time.Second
	// DefaultContainerExp expiration.
	DefaultContainerExp = 300
)

type colorFunc func(string, ...interface{}) string

var colors = []colorFunc{
	color.BlueString,
	color.CyanString,
	color.GreenString,
	color.MagentaString,
	color.RedString,
	color.YellowString,
}

// Container is the shared interface for Container implementations.
type Container interface {
	Name() string
	Address(pool *dockertest.Pool) string
	InternalAddress() string
	ExternalAddress(pool *dockertest.Pool) string
	Start(pool *dockertest.Pool, networkID string, expiration uint) error
	Stop(pool *dockertest.Pool) error
}

// BaseComponent groups together several containers and can run in a runtime.
type BaseComponent struct {
	Pool                *dockertest.Pool
	NetworkID           string
	Network             *dockertest.Network
	Containers          []Container
	Prefix              string
	ContainerExpiration uint
}

// NewBaseComponent creates a new base component that can be used to run in a runtime.
func NewBaseComponent(runtimeExp time.Duration, containerExp uint, prefix, existingNetworkID string) (*BaseComponent, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}
	pool.MaxWait = runtimeExp

	// Create a new network only if we didn't ask to use an existing one.
	var network *dockertest.Network
	if existingNetworkID == "" {
		network, err = pool.CreateNetwork(uuid.New().String())
		if err != nil {
			return nil, fmt.Errorf("failed to create network: %w", err)
		}
		existingNetworkID = network.Network.ID
	}

	return &BaseComponent{
		Pool:                pool,
		Prefix:              prefix,
		NetworkID:           existingNetworkID,
		Network:             network,
		ContainerExpiration: containerExp,
	}, nil
}

// WithContainer adds a container to the runtime.
func (c *BaseComponent) WithContainer(container Container) *BaseComponent {
	c.Containers = append(c.Containers, container)
	return c
}

// Start all containers.
func (c *BaseComponent) Start() error {
	for _, container := range c.Containers {
		existingResource, ok := c.Pool.ContainerByName(container.Name())
		if ok {
			fmt.Println("purging", container.Name())
			if err := c.Pool.Purge(existingResource); err != nil {
				return fmt.Errorf("failed to purge existing container %s: %w", container.Name(), err)
			}
		}
	}

	for _, container := range c.Containers {
		fmt.Println("starting", container.Name())
		err := container.Start(c.Pool, c.NetworkID, c.ContainerExpiration)
		if err != nil {
			return fmt.Errorf("failed to start %s container: %w", container.Name(), err)
		}
	}

	return nil
}

// Teardown removes all containers and networks from the component.
func (c *BaseComponent) Teardown() []error {
	errors := make([]error, 0)

	for _, container := range c.Containers {
		err := container.Stop(c.Pool)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to stop container %s: %w", container.Name(), err))
		}
	}

	if c.Network != nil {
		err := c.Pool.RemoveNetwork(c.Network)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to close network: %w", err))
		}
	}

	return errors
}

// GetContainer tries to find a container by name, by automatically prefixing the name.
func (c *BaseComponent) GetContainer(name string) Container {
	prefixedName := c.Prefix + name
	for _, container := range c.Containers {
		if container.Name() == prefixedName {
			return container
		}
	}
	return nil
}

// StreamLogs streams container logs to stdout.
func (c *BaseComponent) StreamLogs() {
	for i, cont := range c.Containers {
		go c.streamContainerLogs(cont.Name(), colors[i%len(colors)])
	}
}

func (c *BaseComponent) streamContainerLogs(name string, cf colorFunc) {
	w := &prefixedWriter{target: os.Stdout, prefix: cf(name + " | ")}
	err := c.Pool.Client.Logs(
		docker.LogsOptions{
			Container:    name,
			OutputStream: w,
			ErrorStream:  w,
			Follow:       true,
			Stdout:       true,
			Stderr:       true,
		},
	)
	if err != nil {
		fmt.Printf("Could not attach logs to %s: %v\n", name, err)
	}
}

type prefixedWriter struct {
	target io.Writer
	prefix string
}

func (w *prefixedWriter) Write(b []byte) (int, error) {
	fmt.Fprintf(w.target, w.prefix)
	return w.target.Write(b)
}
