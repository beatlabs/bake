// Package build contains build related code.
package build

const (
	// TagIntegration tag.
	TagIntegration = "integration"
	// TagComponent tag.
	TagComponent = "component"
)

// DefaultTags defines the default build tags we are using.
var DefaultTags = []string{TagIntegration, TagComponent}
