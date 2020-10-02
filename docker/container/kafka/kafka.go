// Package kafka exposes a Kafka container. It runs a Zookeeper container behind the scenes for coordination.
package kafka

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
)

const (
	// ContainerName name of the container.
	ContainerName = "kafka"

	defaultPort         = "9094"
	defaultPortInternal = "9092"
)

// Params contains configuration for Container.
type Params struct {
	Prefix           string
	Topics           []string
	KafkaVersion     string
	ZookeeperVersion string
	ContainerHost    bool
	UseExpiration    bool
}

// Container for Kafka.
type Container struct {
	params Params
	container.BaseContainer
	topicsCreationString string
	zookeeperContainer   *zookeeperContainer
}

// NewContainer creates a new Kafka container.
func NewContainer(params Params) (*Container, error) {
	if len(params.Topics) == 0 {
		return nil, errors.New("please specify at least one topic to create")
	}
	topicsCreationStrings := make([]string, 0, len(params.Topics))
	for _, topic := range params.Topics {
		topicsCreationStrings = append(topicsCreationStrings, fmt.Sprintf("%s:1:1", topic))
	}

	zookeeperContainer := newZookeeperContainer(zookeeperParams{
		Prefix:        params.Prefix,
		Version:       params.ZookeeperVersion,
		ContainerHost: params.ContainerHost,
	})

	port := defaultPort
	if params.ContainerHost {
		port = defaultPortInternal
	}
	return &Container{
		params:               params,
		BaseContainer:        *container.NewBaseContainer(params.Prefix, ContainerName, port, params.ContainerHost),
		topicsCreationString: strings.Join(topicsCreationStrings, ","),
		zookeeperContainer:   zookeeperContainer,
	}, nil
}

// Start container.
func (c *Container) Start(pool *dockertest.Pool, networkID string, expiration uint) error {
	err := c.zookeeperContainer.Start(pool, networkID, expiration)
	if err != nil {
		return fmt.Errorf("failed to start zookeeper container: %w", err)
	}

	runOpts := &dockertest.RunOptions{
		Name:       c.Name(),
		NetworkID:  networkID,
		Repository: "wurstmeister/kafka",
		Tag:        c.params.KafkaVersion,
		Env: []string{
			"KAFKA_ZOOKEEPER_CONNECT=" + c.zookeeperContainer.InternalAddress(),
			"KAFKA_ADVERTISED_HOST_NAME=" + c.Name(),
			"KAFKA_ADVERTISED_PORT=" + c.Port(),
			"KAFKA_CREATE_TOPICS=" + c.topicsCreationString,
		},
		ExposedPorts: []string{c.Port()},
	}

	if !c.params.ContainerHost {
		runOpts.Env = []string{
			"KAFKA_ZOOKEEPER_CONNECT=" + c.zookeeperContainer.InternalAddress(),
			"KAFKA_CREATE_TOPICS=" + c.topicsCreationString,
			"PORT_COMMAND=docker port $(hostname) 9094/tcp | cut -d: -f2",
			//# For more details see See https://rmoff.net/2018/08/02/kafka-listeners-explained/
			"KAFKA_LISTENERS=INSIDE://:9092,OUTSIDE://:9094",
			"KAFKA_ADVERTISED_LISTENERS=INSIDE://:9092,OUTSIDE://localhost:_{PORT_COMMAND}",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME=INSIDE",
		}
		runOpts.Mounts = []string{"/var/run/docker.sock:/var/run/docker.sock"}
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
		ret, err := resource.Exec([]string{
			"/bin/sh", "-c",
			"kafka-topics.sh --bootstrap-server localhost:" + defaultPortInternal + " --describe --topic " + c.params.Topics[0] + " | grep " + c.params.Topics[0],
		}, dockertest.ExecOptions{
			StdOut: bufio.NewWriter(os.Stdout),
			StdErr: bufio.NewWriter(os.Stdout),
		})
		if err != nil {
			return err
		}

		if ret != 0 {
			return fmt.Errorf("kafka check command returned error code: %d", ret)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed waiting for kafka: %w", err)
	}

	return nil
}

// Stop kafka and zookeeper containers.
func (c *Container) Stop(pool *dockertest.Pool) error {
	err := c.BaseContainer.Stop(pool)
	if err != nil {
		return err
	}

	return c.zookeeperContainer.Stop(pool)
}

// InternalAddress of the container.
func (c *Container) InternalAddress() string {
	if !c.params.ContainerHost {
		return c.Name() + ":" + defaultPortInternal
	}
	return c.BaseContainer.InternalAddress()
}
