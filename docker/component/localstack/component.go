// Package localstack exposes a Localstack service.
package localstack

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/taxibeat/bake/docker"
)

const (
	// ComponentName is the public name of this component.
	ComponentName = "localstack"
	// ServiceName is the advertised name of this service.
	ServiceName = "localstack"
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
		Name:       "localstack",
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
