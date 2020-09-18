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
)

func RunDocker(img, cmd string, args ...string) error {
	docker := "docker"

	// todo: set this on init?
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v\n", err)
	}

	// todo: extract dockerArgs?
	dockerArgs := []string{"run", "--rm", "--volume", wd + ":/volume", "--workdir", "/volume", img, cmd}
	args = append(dockerArgs, args...)

	fmt.Printf("Executing cmd: %s %s\n", docker, strings.Join(args, " "))
	return sh.RunV(docker, args...)
}

func RunGo(args ...string) error {
	cmd := "go"

	_, err := exec.LookPath(cmd)
	if err != nil {
		return RunDocker("golang:1.14", cmd, args...)
	}

	fmt.Printf("Executing cmd: %s %s\n", cmd, strings.Join(args, " "))
	return sh.RunV(cmd, args...)
}
