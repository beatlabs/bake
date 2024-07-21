// Package jaeger exposes a Jaeger service.
package jaeger

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beatlabs/bake/docker"
)

const (
	// ServiceName is the advertised name of this service.
	ServiceName   = "jaeger"
	componentName = "jaeger"
)

// NewComponent creates a new Redis component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       componentName,
		Repository: "jaegertracing/all-in-one",
		Tag:        "latest",
		ServicePorts: map[string]string{
			ServiceName: "16686",
		},
		ReadyFunc: readyFunc,
	}

	for _, opt := range opts {
		opt(&container)
	}

	return &docker.SimpleComponent{
		Name:       componentName,
		Containers: []docker.SimpleContainerConfig{container},
	}
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
