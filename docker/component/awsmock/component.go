// Package awsmock exposes an AWS mock service backed by motoserver/moto.
package awsmock

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beatlabs/bake/docker"
)

const (
	// ServiceName is the advertised name of this service.
	ServiceName   = "awsmock"
	componentName = "awsmock"
)

// NewComponent creates a new AWS mock component backed by moto.
// All AWS services are available by default; no explicit service selection is needed.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       componentName,
		Repository: "motoserver/moto",
		Tag:        "latest",
		ServicePorts: map[string]string{
			ServiceName: "5000",
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
		url := fmt.Sprintf("http://%s/moto-api/", addr)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create health request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("got status code: %d from %s", resp.StatusCode, url)
		}
		return nil
	})
}
