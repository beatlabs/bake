// Package localstack is deprecated. Use github.com/beatlabs/bake/docker/component/awsmock instead.
//
// Deprecated: Use github.com/beatlabs/bake/docker/component/awsmock instead.
package localstack

import (
	"github.com/beatlabs/bake/docker"
	"github.com/beatlabs/bake/docker/component/awsmock"
)

const (
	// ServiceName is the advertised name of this service.
	//
	// Deprecated: Use awsmock.ServiceName instead.
	ServiceName = awsmock.ServiceName
)

// WithServices is a no-op retained for backward compatibility.
// Moto exposes all AWS services by default.
//
// Deprecated: awsmock.NewComponent does not require service selection.
func WithServices(_ ...string) docker.SimpleContainerOptionFunc {
	return func(_ *docker.SimpleContainerConfig) {}
}

// NewComponent creates a new AWS mock component.
//
// Deprecated: Use awsmock.NewComponent instead.
func NewComponent(opts ...docker.SimpleContainerOptionFunc) *docker.SimpleComponent {
	return awsmock.NewComponent(opts...)
}
