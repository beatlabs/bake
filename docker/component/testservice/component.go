package testservice

import (
	"fmt"
	"net/http"
	"os"

	"github.com/taxibeat/bake/docker"
	"github.com/taxibeat/bake/docker/component/kafka"
	"github.com/taxibeat/bake/docker/component/mongodb"
	"github.com/taxibeat/bake/docker/component/redis"
)

const (
	ComponentName = "testservice"
	ContainerName = "testservice"
	ServiceName   = "testservice"
)

func NewComponent(session *docker.Session) (*docker.SimpleComponent, error) {
	redisAddr, err := session.DockerToDockerServiceAddress(redis.ServiceName)
	if err != nil {
		return nil, err
	}
	mongoAddr, err := session.DockerToDockerServiceAddress(mongodb.ServiceName)
	if err != nil {
		return nil, err
	}
	kafkaAddr, err := session.DockerToDockerServiceAddress(kafka.KafkaServiceName)
	if err != nil {
		return nil, err
	}

	container := docker.SimpleContainerConfig{
		BuildOpts: &docker.BuildOptions{
			Dockerfile: "docker/component/testservice/Dockerfile",
			ContextDir: "../..",
		},
		Name:       ContainerName,
		Repository: ComponentName,
		Env: []string{
			"REDIS=" + redisAddr,
			"MONGO=" + mongoAddr,
			"KAFKA=" + kafkaAddr,
			"PORT=8080",
		},
		ServicePorts: map[string]string{
			ServiceName: "8080",
		},
		ReadyFunc: readyFunc,
	}

	existing := os.Getenv("EXISTING_TESTSERVICE")
	if existing != "" {
		container.BuildOpts = nil
		container.Tag = existing
	}

	return &docker.SimpleComponent{
		Name:       ComponentName,
		Containers: []docker.SimpleContainerConfig{container},
	}, nil
}

func readyFunc(session *docker.Session) error {
	addr, err := session.AutoServiceAddress(ServiceName)
	if err != nil {
		return err
	}

	return docker.Retry(func() error {
		resp, err := http.Get("http://" + addr)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("got status code: %d", resp.StatusCode)
		}
		return nil
	})
}
