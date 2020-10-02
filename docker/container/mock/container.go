// Package mock exposes a container for Mockserver.
package mock

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
)

const defaultPort = "1080"

// Params of the container.
type Params struct {
	Name          string
	Prefix        string
	Version       string
	ContainerHost bool
	UseExpiration bool
}

// Container docker service definition.
type Container struct {
	params Params
	container.BaseContainer
}

// NewContainer creates a new Mockserver container.
func NewContainer(params Params) *Container {
	return &Container{
		params:        params,
		BaseContainer: *container.NewBaseContainer(params.Prefix, params.Name, defaultPort, params.ContainerHost),
	}
}

// Start a new mockserver container.
func (c *Container) Start(pool *dockertest.Pool, networkID string, expiration uint) error {
	opts := &dockertest.RunOptions{
		Name:       c.Name(),
		NetworkID:  networkID,
		Repository: "mockserver/mockserver",
		Tag:        c.params.Version,
		Env: []string{
			"LOG_LEVEL=DEBUG",
		},
	}

	resource, err := pool.RunWithOptions(opts)
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
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/status", c.Address(pool)), nil)
		if err != nil {
			return fmt.Errorf("failed to create status request to mockserver: %w", err)
		}
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("could not connect to mockserver: %w", err)
		}
		if rsp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("mockserver returned status: %d", rsp.StatusCode)
			return errors.New(msg)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed waiting for mockserver: %w", err)
	}

	return nil
}
