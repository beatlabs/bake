// Package kafka exposes a Kafka-compatible container using Redpanda.
// Redpanda is a C++ Kafka-compatible broker that requires no ZooKeeper
// and uses a fraction of the memory of a JVM-based Kafka deployment.
package kafka

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
	"github.com/beatlabs/bake/docker"
)

const (
	// KafkaServiceName is the advertised name of the Kafka service.
	KafkaServiceName = "kafka"
	// ZookeeperServiceName is kept for backwards compatibility; Redpanda
	// runs without ZooKeeper so this service is never registered.
	ZookeeperServiceName = "zookeeper"
	componentName        = "kafka"
)

// topicSpec holds a parsed topic specification.
type topicSpec struct {
	name              string
	numPartitions     int32
	replicationFactor int16
}

// WithTopics schedules topics to be created after the broker is ready.
// Format: "name:partitions:replication" e.g. "MyTopic:1:1".
func WithTopics(topics ...string) docker.SimpleContainerOptionFunc {
	return func(c *docker.SimpleContainerConfig) {
		c.Env = append(c.Env, "KAFKA_CREATE_TOPICS="+strings.Join(topics, ","))
	}
}

// NewComponent creates a Redpanda-backed Kafka component.
func NewComponent(session *docker.Session, opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	port, _ := docker.GetFreePort()

	container := docker.SimpleContainerConfig{
		Name:       componentName,
		Repository: "redpandadata/redpanda",
		Tag:        "latest",
		ServicePorts: map[string]string{
			KafkaServiceName: "9092",
		},
		StaticServicePorts: map[string]string{
			KafkaServiceName: port,
		},
		RunOpts: &docker.RunOptions{
			Cmd: []string{
				"redpanda", "start",
				"--smp=1",
				"--memory=256M",
				"--reserve-memory=0M",
				"--overprovisioned",
				"--node-id=0",
				"--check=false",
				"--kafka-addr=0.0.0.0:9092",
				fmt.Sprintf("--advertise-kafka-addr=localhost:%s", port),
			},
		},
		MemoryMB: 384,
	}

	for _, opt := range opts {
		opt(&container)
	}

	// Extract topics from the env var set by WithTopics, then remove it —
	// Redpanda doesn't understand KAFKA_CREATE_TOPICS; we create them via
	// the admin API in the ready function instead.
	topics := extractTopics(&container)

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_6_0_0

	container.ReadyFunc = func(s *docker.Session) error {
		addr, err := s.AutoServiceAddress(KafkaServiceName)
		if err != nil {
			return err
		}
		return docker.Retry(func() error {
			admin, err := sarama.NewClusterAdmin([]string{addr}, cfg)
			if err != nil {
				return err
			}
			defer func() { _ = admin.Close() }()

			for _, t := range topics {
				err := admin.CreateTopic(t.name, &sarama.TopicDetail{
					NumPartitions:     t.numPartitions,
					ReplicationFactor: t.replicationFactor,
				}, false)
				// Ignore "already exists" — idempotent startup.
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("create topic %q: %w", t.name, err)
				}
			}
			return nil
		})
	}

	return &docker.SimpleComponent{
		Name:       componentName,
		Containers: []docker.SimpleContainerConfig{container},
	}
}

// extractTopics pulls the KAFKA_CREATE_TOPICS env var out of the container
// config and returns parsed topic specs. The env var is removed because
// Redpanda does not understand it.
func extractTopics(c *docker.SimpleContainerConfig) []topicSpec {
	var specs []topicSpec
	filtered := c.Env[:0]
	for _, e := range c.Env {
		if strings.HasPrefix(e, "KAFKA_CREATE_TOPICS=") {
			val := strings.TrimPrefix(e, "KAFKA_CREATE_TOPICS=")
			for _, raw := range strings.Split(val, ",") {
				if s, ok := parseTopicSpec(raw); ok {
					specs = append(specs, s)
				}
			}
			continue
		}
		filtered = append(filtered, e)
	}
	c.Env = filtered
	return specs
}

// parseTopicSpec parses "name:partitions:replication[:config]".
func parseTopicSpec(raw string) (topicSpec, bool) {
	parts := strings.Split(raw, ":")
	if len(parts) < 3 {
		return topicSpec{}, false
	}
	partitions, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return topicSpec{}, false
	}
	replication, err := strconv.ParseInt(parts[2], 10, 16)
	if err != nil {
		return topicSpec{}, false
	}
	return topicSpec{
		name:              parts[0],
		numPartitions:     int32(partitions),
		replicationFactor: int16(replication),
	}, true
}
