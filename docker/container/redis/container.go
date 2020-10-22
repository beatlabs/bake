// Package redis exposes a Redis container.
package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
)

const (
	// ContainerName name of the container.
	ContainerName = "redis"

	defaultPort = "6379"
)

// Params contains configuration for Container.
type Params struct {
	Prefix        string
	Version       string
	Env           []string
	ContainerHost bool
	UseExpiration bool
	RedisOptions  *redis.Options
}

// Container for Localstack.
type Container struct {
	params Params
	container.BaseContainer
}

// NewContainer creates a new Redis container.
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
		Repository: "bitnami/redis",
		Tag:        c.params.Version,
		Env:        c.params.Env,
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
		opts := c.params.RedisOptions
		if opts == nil {
			opts = &redis.Options{}
		}
		if opts.Addr == "" {
			opts.Addr = c.Address(pool)
		}
		cl := redis.NewClient(opts)
		_, err := cl.Ping(context.Background()).Result()
		if err != nil {
			return fmt.Errorf("could not ping redis: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed waiting for redis: %w", err)
	}

	return nil
}
