// Package container contains the base elements to create a Docker container.
package container

import (
	"fmt"

	"github.com/ory/dockertest/v3"
)

// BaseContainer contains similar logic shared across all containers.
type BaseContainer struct {
	prefixedName  string
	host          string
	port          string
	containerHost bool
}

// NewBaseContainer creates a new BaseContainer.
func NewBaseContainer(prefix, name, port string, containerHost bool) *BaseContainer {
	prefixedName := prefix + name
	return &BaseContainer{
		prefixedName:  prefixedName,
		host:          fmt.Sprintf("%s:%s", prefixedName, port),
		port:          port,
		containerHost: containerHost,
	}
}

// Name of the container.
func (c *BaseContainer) Name() string {
	return c.prefixedName
}

// InternalAddress of the container.
func (c *BaseContainer) InternalAddress() string {
	return c.host
}

// Port of the container.
func (c *BaseContainer) Port() string {
	return c.port
}

// ExternalAddress of the container.
func (c *BaseContainer) ExternalAddress(pool *dockertest.Pool) string {
	res, ok := pool.ContainerByName(c.Name())
	if !ok {
		return ""
	}
	return res.GetHostPort(c.port + "/tcp")
}

// Address of the container.
func (c *BaseContainer) Address(pool *dockertest.Pool) string {
	if c.containerHost {
		return c.InternalAddress()
	}
	return c.ExternalAddress(pool)
}

// Stop container.
func (c *BaseContainer) Stop(pool *dockertest.Pool) error {
	res, ok := pool.ContainerByName(c.Name())
	if !ok {
		return fmt.Errorf("could not find container %s", c.Name())
	}

	err := res.Close()
	if err != nil {
		return fmt.Errorf("failed to purge container %s: %w", c.Name(), err)
	}

	return nil
}
