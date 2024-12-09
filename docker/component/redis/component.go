// Package redis exposes a Redis service.
package redis

import (
	"context"

	"github.com/beatlabs/bake/docker"
	"github.com/redis/go-redis/v9"
)

const (
	// ServiceName is the advertised name of this service.
	ServiceName   = "redis"
	componentName = "redis"
)

// NewComponent creates a new Redis component.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	container := docker.SimpleContainerConfig{
		Name:       componentName,
		Repository: "bitnami/redis",
		Tag:        "latest",
		Env: []string{
			"ALLOW_EMPTY_PASSWORD=yes",
		},
		ServicePorts: map[string]string{
			ServiceName: "6379",
		},
		ReadyFunc: readyFunc,
		// Disable redis protected mode, in this mode connections are only accepted from the loopback interface
		// https://hub.docker.com/r/bitnami/redis/
		RunOpts: &docker.RunOptions{
			Cmd: []string{"/opt/bitnami/scripts/redis/run.sh", "--protected-mode", "no"},
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

	opts := &redis.Options{Addr: addr}
	cl := redis.NewClient(opts)

	return docker.Retry(func() error {
		_, err := cl.Ping(context.Background()).Result()
		return err
	})
}
