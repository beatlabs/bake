// Package kafka exposes Kaka and Zookeeper containers.
package kafka

import (
	"errors"
	"fmt"
	"strings"

	"github.com/taxibeat/bake/docker"
	"gopkg.in/Shopify/sarama.v1"
)

const (
	ComponentName          = "kafka"
	KafkaContainerName     = "kafka"
	ZookeeperContainerName = "zookeeper"
	KafkaServiceName       = "kafka"
	ZookeeperServiceName   = "zookeeper"
)

// NewComponent creates a new Redis component.
func NewComponent(session *docker.Session, topics []string, opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	zooContainer := docker.SimpleContainerConfig{
		Name:       ZookeeperContainerName,
		Repository: "wurstmeister/zookeeper",
		Tag:        "latest",
		ServicePorts: map[string]string{
			ZookeeperServiceName: "2181",
		},
		ReadyFunc: zookeeperReadyFunc,
	}

	port, _ := docker.GetFreePort()

	kafkaContainer := docker.SimpleContainerConfig{
		Name:       KafkaContainerName,
		Repository: "wurstmeister/kafka",
		Tag:        "latest",
		ServicePorts: map[string]string{
			KafkaServiceName: "9092",
		},
		FixedHostServicePorts: map[string]string{
			KafkaServiceName: port,
		},
		Env: []string{
			fmt.Sprintf("KAFKA_ZOOKEEPER_CONNECT=%s-%s:2181", session.ID(), ZookeeperContainerName),
			"KAFKA_CREATE_TOPICS=" + strings.Join(topics, ","),
			"KAFKA_LISTENERS=INSIDE://:9092,OUTSIDE://:" + port,
			"KAFKA_ADVERTISED_LISTENERS=INSIDE://:9092,OUTSIDE://localhost:" + port,
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME=INSIDE",
		},
		ReadyFunc: kafkaReadyFunc,
	}

	for _, opt := range opts {
		kafkaContainer = opt(kafkaContainer)
	}

	return &docker.SimpleComponent{
		Name:       ComponentName,
		Containers: []docker.SimpleContainerConfig{zooContainer, kafkaContainer},
	}
}

func zookeeperReadyFunc(session *docker.Session) error {
	return nil
}

func kafkaReadyFunc(session *docker.Session) error {
	addr, err := session.AutoServiceAddress(KafkaServiceName)
	if err != nil {
		return err
	}

	return docker.Retry(func() error {
		cl, err := sarama.NewClient([]string{addr}, nil)
		if err != nil {
			return err
		}
		topics, err := cl.Topics()
		if err != nil {
			return err
		}
		if len(topics) == 0 {
			return errors.New("no topics created")
		}
		return nil
	})
}
