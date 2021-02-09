// Package mockserver exposes a Mockserver services.
package mockserver

import (
	"fmt"
	"net/http"

	"github.com/taxibeat/bake/docker"
)

const (
	// ServiceName is the advertised name of this service.
	ServiceName   = "mockserver"
	componentName = "mockserver"
)

// NewComponent creates a new Redis component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       componentName,
		Repository: "mockserver/mockserver",
		Tag:        "latest",
		Env: []string{
			"LOG_LEVEL=DEBUG",
		},
		ServicePorts: map[string]string{
			ServiceName: "1080",
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

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/status", addr), nil)
	if err != nil {
		return fmt.Errorf("failed to create status request to mockserver: %w", err)
	}

	return docker.Retry(func() error {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("could not connect to mockserver: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("got status code: %d", resp.StatusCode)
		}
		return nil
	})
}
