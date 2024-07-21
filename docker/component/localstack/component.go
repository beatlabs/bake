// Package localstack exposes a Localstack service.
package localstack

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/beatlabs/bake/docker"
)

const (
	// ServiceName is the advertised name of this service.
	ServiceName   = "localstack"
	componentName = "localstack"
)

// WithServices sets the localstack services in the container config.
func WithServices(services ...string) docker.SimpleContainerOptionFunc {
	return func(c *docker.SimpleContainerConfig) {
		c.Env = append(c.Env, "LOCALSTACK_SERVICES="+strings.Join(services, ","))
	}
}

// NewComponent creates a new Redis component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       componentName,
		Repository: "localstack/localstack",
		Tag:        "latest",
		ServicePorts: map[string]string{
			ServiceName: "4566",
		},
		Env: []string{
			"LOCALSTACK_DEBUG=1",
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
		url := fmt.Sprintf("http://%s/_localstack/health", addr)
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
