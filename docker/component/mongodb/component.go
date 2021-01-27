// Package mongodb exposes a Mongo DB container.
package mongodb

import (
	"context"
	"fmt"

	"github.com/taxibeat/bake/docker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ComponentName = "mongo"
	ContainerName = "mongo"
	ServiceName   = "mongo"
)

const (
	ReplicaSet = "rs0"
)

// NewComponent creates a new Consul component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       ContainerName,
		Repository: "bitnami/mongodb",
		Tag:        "latest",
		ServicePorts: map[string]string{
			ServiceName: "27017",
		},
		ReadyFunc: readyFunc,
		Env: []string{
			"MONGODB_REPLICA_SET_MODE=primary",
			"MONGODB_ROOT_PASSWORD=password",
			"MONGODB_REPLICA_SET_NAME=" + ReplicaSet,
			"MONGODB_REPLICA_SET_KEY=replicasetkey123",
		},
	}

	for _, opt := range opts {
		container = opt(container)
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
	opts := options.Client()
	opts.ApplyURI("mongodb://" + addr)

	return docker.Retry(func() error {
		cl, err := mongo.Connect(context.Background(), opts)
		if err != nil {
			return fmt.Errorf("failed to create mongo client: %w", err)
		}
		defer func() { _ = cl.Disconnect(context.Background()) }()
		return cl.Ping(context.Background(), nil)
	})
}
