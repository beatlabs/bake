// Package mongodb exposes a Mongo DB container.
package mongodb

import (
	"context"
	"fmt"

	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// ContainerName name of the container.
	ContainerName = "mongodb"

	defaultPort = "27017"
)

// Params contains configuration for Container.
type Params struct {
	Prefix        string
	Version       string
	Env           []string
	ContainerHost bool
	UseExpiration bool
	MongoOptions  *options.ClientOptions
}

// Container for Mongo.
type Container struct {
	params Params
	container.BaseContainer
}

// NewContainer creates a new Mongo container.
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
		Repository: "mongo",
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
		opts := c.params.MongoOptions
		if opts == nil {
			opts = options.Client()
		}
		opts.ApplyURI("mongodb://" + c.Address(pool))
		cl, err := mongo.Connect(context.Background(), opts)
		if err != nil {
			return fmt.Errorf("failed to create mongo client: %w", err)
		}

		if err := cl.Ping(context.Background(), nil); err != nil {
			return fmt.Errorf("could not ping mongo: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed waiting for mongo: %w", err)
	}

	return nil
}
