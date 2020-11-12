// Package mongodb exposes a Mongo DB container.
package mongodb

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/taxibeat/bake/docker/container"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// ContainerName name of the container.
	ContainerName = "mongodb"

	defaultPort       = "27017"
	defaultReplicaSet = "rs0"
)

// Params contains configuration for Container.
type Params struct {
	Prefix         string
	Version        string
	Env            []string
	ContainerHost  bool
	UseExpiration  bool
	MongoOptions   *options.ClientOptions
	ReplicaSetMode bool
}

// Container for Mongo.
type Container struct {
	params Params
	container.BaseContainer
}

// NewContainer creates a new Mongo container.
func NewContainer(params Params) (*Container, error) {
	var err error

	port := defaultPort
	if params.ReplicaSetMode {
		port, err = getFreePort()
		if err != nil {
			return nil, fmt.Errorf("can not obtain random free port: %w", err)
		}
	}

	return &Container{
		params:        params,
		BaseContainer: *container.NewBaseContainer(params.Prefix, ContainerName, port, params.ContainerHost),
	}, nil
}

// Start container.
func (c *Container) Start(pool *dockertest.Pool, networkID string, expiration uint) error {
	resource, err := pool.RunWithOptions(c.runOptions(networkID))
	if err != nil {
		return err
	}

	if c.params.UseExpiration {
		err = resource.Expire(expiration)
		if err != nil {
			return err
		}
	}

	if c.params.ReplicaSetMode {
		if err := c.initiateReplicaSet(pool, resource); err != nil {
			return fmt.Errorf("mongo replica set initialization failed: %w", err)
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

func (c *Container) runOptions(networkID string) *dockertest.RunOptions {
	runOpts := &dockertest.RunOptions{
		Name:       c.Name(),
		NetworkID:  networkID,
		Repository: "mongo",
		Tag:        c.params.Version,
		Env:        c.params.Env,
	}

	// for the replicaset we are forced to use a custom port,
	// because in a CI environment to avoid potential conflicts we want to have a randomly assigned port
	if c.params.ReplicaSetMode {
		runOpts.Cmd = []string{"--replSet", defaultReplicaSet, "--bind_ip_all", "--port", c.Port()}
		runOpts.ExposedPorts = []string{c.Port()}
		runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
			docker.Port(c.Port()): {{
				HostIP:   "0.0.0.0",
				HostPort: c.Port(),
			}},
		}
	}

	return runOpts
}

func (c *Container) initiateReplicaSet(pool *dockertest.Pool, resource *dockertest.Resource) error {
	command := fmt.Sprintf(`mongo localhost:%s --eval "rs.initiate({
		_id: \"%s\",
		members: [
			{ _id: 0, host : \"%s\" },
		]
    })"`, c.Port(), defaultReplicaSet, c.Address(pool))

	return pool.Retry(func() error {
		ret, err := resource.Exec([]string{"bash", "-c", command},
			dockertest.ExecOptions{
				StdOut: bufio.NewWriter(os.Stdout),
				StdErr: bufio.NewWriter(os.Stdout),
			},
		)
		if err != nil {
			return err
		}
		if ret != 0 {
			return fmt.Errorf("error code: %d", ret)
		}
		return nil
	})
}

func getFreePort() (string, error) {
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
