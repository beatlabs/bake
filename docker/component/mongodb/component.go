// Package mongodb exposes a Mongo DB service.
package mongodb

import (
	"context"
	"fmt"

	"github.com/taxibeat/bake/docker"
)

const (
	// ServiceName is the advertised name of this service.
	ServiceName = "mongo"
	// ReplicaSet is the replica set name.
	ReplicaSet    = "rs0"
	componentName = "mongo"
)

// NewComponent creates a new Consul component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       componentName,
		Repository: "mongo",
		Tag:        "latest",
		ServicePorts: map[string]string{
			ServiceName: "27017",
		},

		ReadyFunc: readyFunc,
		Env:       []string{},
		RunOpts: &docker.RunOptions{
			Cmd:         []string{"--replSet", ReplicaSet},
			InitExecCmd: `mongo --eval "rs.initiate()"`,
		},
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
		cl, err := NewClient(context.Background(), addr)
		if err != nil {
			return fmt.Errorf("failed to create mongo client: %w", err)
		}
		defer func() { _ = cl.Disconnect(context.Background()) }()
		return cl.Ping(context.Background(), nil)
	})
}
