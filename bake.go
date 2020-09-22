// Package bake is the main entry point for our mage collection of targets.
package bake

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/sh"
)

const (
	// BuildTagIntegration tag.
	BuildTagIntegration = "integration"
	// BuildTagComponent tag.
	BuildTagComponent = "component"
	// GoCmd defines the std Go command.
	GoCmd = "go"
	// DockerCmd defines the std Docker command.
	DockerCmd = "docker"
)

// RunDocker runs the docker command.
func RunDocker(img string, args ...string) error {
	docker := "docker"

	// todo: set this on init?
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	// todo: extract dockerArgs?
	dockerArgs := []string{
		"run", "--rm",
		"--env", "GOFLAGS=-mod=vendor", "--env", "GO111MODULE=on",
		"--volume", wd + ":/app", "--volume", "/var/run/docker.sock:/var/run/docker.sock",
		"--workdir", "/app", img, "sh", "-c",
		"git config --global url.\"https://golang:$GITHUB_TOKEN@github.com\".insteadOf \"https://github.com\" && " +
			strings.Join(args, " ") +
			" && chown --reference=/app --recursive /app",
	}

	fmt.Printf("Executing cmd: %s %s\n", docker, strings.Join(dockerArgs, " "))
	return sh.RunV(docker, dockerArgs...)
}

// RunGo runs the go command.
func RunGo(args ...string) error {
	forceDocker := false
	if os.Getenv("BAKE_FORCE_DOCKER") != "" {
		forceDocker = true
	}

	_, err := exec.LookPath(GoCmd)
	if err != nil || forceDocker {
		return RunDocker("golang:1.14", append([]string{GoCmd}, args...)...)
	}

	fmt.Printf("Executing cmd: %s %s\n", GoCmd, strings.Join(args, " "))
	return sh.RunV(GoCmd, args...)
}
