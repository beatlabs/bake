package testservice

import (
	"fmt"
	"net/http"

	"github.com/taxibeat/bake/docker"
)

const (
	ComponentName = "testservice"
	ServiceName   = "testservice"
)

func NewComponent(redisAddr, mongoAddr, kafkaAddr string) (*docker.SimpleComponent, error) {
	container := docker.SimpleContainerConfig{
		BuildOpts: &docker.BuildOptions{
			Dockerfile: "docker/component/testservice/Dockerfile",
			ContextDir: "../..",
		},
		Name:       "testservice",
		Repository: "testservice",
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
