// Package jaeger exposes a Jaeger service.
package jaeger

import (
	"fmt"
	"net/http"

	"github.com/taxibeat/bake/docker"
)

const (
	// ComponentName is the public name of this component.
	ComponentName = "jaeger"
	// ServiceName is the advertised name of this service.
	ServiceName = "jaeger"
)

// NewComponent creates a new Redis component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       "jaeger",
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
		Name:       ComponentName,
		Containers: []docker.SimpleContainerConfig{container},
	}
}

func readyFunc(session *docker.Session) error {
	addr, err := session.AutoServiceAddress(ServiceName)
	if err != nil {
		return err
	}

	return docker.Retry(func() error {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", addr))
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("got status code: %d", resp.StatusCode)
		}
		return nil
	})
}
