// Package consul exposes a Consul service and client.
package consul

import (
	"github.com/taxibeat/bake/docker"
)

const (
	// ComponentName is the public name of this component.
	ComponentName = "consul"
	// ServiceName is the advertised name of this service.
	ServiceName = "consul"
)

// NewComponent creates a new Consul component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       "consul",
		Repository: "consul",
		Tag:        "1.8.0",
		ServicePorts: map[string]string{
			ServiceName: "8500",
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

	consulClient, err := NewClient(addr)
	if err != nil {
		return err
	}

	return docker.Retry(func() error {
		return consulClient.Live()
	})
}
