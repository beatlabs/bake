// Package bake is the main entry point for our mage collection of targets.
package bake

const (
	// BuildTagIntegration tag.
	BuildTagIntegration = "integration"
	// BuildTagComponent tag.
	BuildTagComponent = "component"

	// GoCmd defines the std Go command.
	GoCmd = "go"

	// GitHubTokenEnvVar defines the env var key for the GitHub token.
	GitHubTokenEnvVar = "GITHUB_TOKEN" // nolint:gosec
)
