// Package redis exposes a Redis service.
package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/taxibeat/bake/docker"
)

const (
	// ComponentName is the public name of this component.
	ComponentName = "redis"
	// ServiceName is the advertised name of this service.
	ServiceName = "redis"
)

// NewComponent creates a new Redis component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       "redis",
		Repository: "bitnami/redis",
		Tag:        "latest",
		Env: []string{
			"ALLOW_EMPTY_PASSWORD=yes",
		},
		ServicePorts: map[string]string{
			ServiceName: "6379",
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

	opts := &redis.Options{Addr: addr}
	cl := redis.NewClient(opts)

	return docker.Retry(func() error {
		_, err := cl.Ping(context.Background()).Result()
		return err
	})
}
