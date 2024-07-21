// Package testservice exposes a simple test service.
package testservice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beatlabs/bake/docker"
)

const (
	// ServiceName is the advertised name of this service.
	ServiceName   = "testservice"
	componentName = "testservice"
)

// NewComponent constructs a component.
func NewComponent(redisAddr, mongoAddr, kafkaAddr string) (*docker.SimpleComponent, error) {
	container := docker.SimpleContainerConfig{
		BuildOpts: &docker.BuildOptions{
			Dockerfile: "docker/component/testservice/Dockerfile",
			ContextDir: "../..",
		},
		Name:       componentName,
		Repository: componentName,
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
		Name:       componentName,
		Containers: []docker.SimpleContainerConfig{container},
	}, nil
}

func readyFunc(session *docker.Session) error {
	addr, err := session.AutoServiceAddress(ServiceName)
	if err != nil {
		return err
	}

	return docker.Retry(func() error {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("http://%s/health", addr), nil)
		if err != nil {
			return fmt.Errorf("failed to create health request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("got status code: %d", resp.StatusCode)
		}
		return nil
	})
}
