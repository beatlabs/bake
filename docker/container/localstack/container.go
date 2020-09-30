// Package localstack exposes a Localstack container.
package localstack

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
)

const (
	// ContainerName name of the container.
	ContainerName = "localstack"
	// ServiceS3 exposed in Localstack.
	ServiceS3 = "s3"

	defaultPort = "4566"
)

// Params contains configuration for Container.
type Params struct {
	Prefix        string
	Version       string
	ContainerHost bool
	UseExpiration bool
	Services      []string
}

// Container for Localstack.
type Container struct {
	params Params
	container.BaseContainer
}

// NewContainer creates a new Localstack container.
func NewContainer(params Params) (*Container, error) {
	if len(params.Services) == 0 {
		return nil, errors.New("please specify at least one service")
	}

	return &Container{
		params:        params,
		BaseContainer: *container.NewBaseContainer(params.Prefix, ContainerName, defaultPort, params.ContainerHost),
	}, nil
}

// Start container.
func (c *Container) Start(pool *dockertest.Pool, networkID string, expiration uint) error {
	runOpts := &dockertest.RunOptions{
		Name:       c.Name(),
		NetworkID:  networkID,
		Repository: "localstack/localstack",
		Tag:        c.params.Version,
		Env: []string{
			"LOCALSTACK_SERVICES=" + strings.Join(c.params.Services, ","),
			"LOCALSTACK_DEBUG=1",
		},
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
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", c.Address(pool)), nil)
		if err != nil {
			return fmt.Errorf("create status request: %w", err)
		}
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("make status request: %w", err)
		}

		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return fmt.Errorf("read body: %w", err)
		}
		if string(body) != `{"status": "running"}` {
			return fmt.Errorf("status request returned: %s", string(body))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed waiting for Localstack: %w", err)
	}

	return nil
}
