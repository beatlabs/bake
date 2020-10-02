// Package consul exposes a Consul container and client.
package consul

import (
	"fmt"

	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
)

const (
	// ContainerName name of the container.
	ContainerName = "consul"

	defaultPort = "8500"
)

// Params contains configuration for Container.
type Params struct {
	Prefix        string
	Version       string
	ContainerHost bool
	UseExpiration bool
}

// Container for Consul.
type Container struct {
	params Params
	container.BaseContainer
}

// NewContainer creates a new Consul container.
func NewContainer(params Params) *Container {
	return &Container{
		params:        params,
		BaseContainer: *container.NewBaseContainer(params.Prefix, ContainerName, defaultPort, params.ContainerHost),
	}
}

// Start container.
func (c *Container) Start(pool *dockertest.Pool, networkID string, expiration uint) error {
	runOpts := &dockertest.RunOptions{
		Name:       c.Name(),
		NetworkID:  networkID,
		Repository: "consul",
		Tag:        c.params.Version,
	}

	resource, err := pool.RunWithOptions(runOpts)
	if err != nil {
		return err
	}

	if c.params.UseExpiration {
		err = resource.Expire(expiration)
		if err != nil {
			return err
		}
	}

	err = pool.Retry(func() error {
		consulClient, err := NewClient(c.Address(pool))
		if err != nil {
			return err
		}
		err = consulClient.Live()
		return err
	})
	if err != nil {
		return fmt.Errorf("failed waiting for consul: %w", err)
	}

	return nil
}
