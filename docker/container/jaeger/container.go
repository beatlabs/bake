// Package jaeger exposes a Jaeger container.
package jaeger

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
)

const (
	// ContainerName name of the container.
	ContainerName = "jaegertracing-all-in-one"

	defaultPort = "16686"
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
		Repository: "jaegertracing/all-in-one",
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
		rsp, err := http.Get("http://" + c.Address(pool) + "/health")
		if err != nil {
			return err
		}

		if rsp.StatusCode != http.StatusOK {
			return errors.New("")
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed waiting for jaeger: %w", err)
	}

	return nil
}
