// Package kafka exposes Kafka and Zookeeper containers.
package kafka

import (
	"errors"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
	"github.com/beatlabs/bake/docker"
)

const (
	// KafkaServiceName is the advertised name of the Kafka service.
	KafkaServiceName = "kafka"
	// ZookeeperServiceName is the advertised name of the Zookeeper service.
	ZookeeperServiceName = "zookeeper"
	componentName        = "kafka"
)

// WithTopics sets topics in the kafka container config.
// E.g. MyTopic:1:1:compact.
func WithTopics(topics ...string) docker.SimpleContainerOptionFunc {
	return func(c *docker.SimpleContainerConfig) {
		c.Env = append(c.Env, "KAFKA_CREATE_TOPICS="+strings.Join(topics, ","))
	}
}

// NewComponent creates a new Redis component.
func NewComponent(session *docker.Session, opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	zooContainer := docker.SimpleContainerConfig{
		Name:       "zookeeper",
		Repository: "wurstmeister/zookeeper",
		Tag:        "latest",
		ServicePorts: map[string]string{
			ZookeeperServiceName: "2181",
		},
		ReadyFunc: zookeeperReadyFunc,
	}

	port, _ := docker.GetFreePort()

	kafkaContainer := docker.SimpleContainerConfig{
		Name:       "kafka",
		Repository: "wurstmeister/kafka",
		Tag:        "latest",
		ServicePorts: map[string]string{
			KafkaServiceName: "9092",
		},
		StaticServicePorts: map[string]string{
			KafkaServiceName: port,
		},
		Env: []string{
			fmt.Sprintf("KAFKA_ZOOKEEPER_CONNECT=%s-zookeeper:2181", session.ID()),
			"KAFKA_LISTENERS=INSIDE://:9092,OUTSIDE://:" + port,
			"KAFKA_ADVERTISED_LISTENERS=INSIDE://:9092,OUTSIDE://localhost:" + port,
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME=INSIDE",
		},
		ReadyFunc: kafkaReadyFunc,
	}

	for _, opt := range opts {
		opt(&kafkaContainer)
	}

	return &docker.SimpleComponent{
		Name:       componentName,
		Containers: []docker.SimpleContainerConfig{zooContainer, kafkaContainer},
	}
}

func zookeeperReadyFunc(_ *docker.Session) error {
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
