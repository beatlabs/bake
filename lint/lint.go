// Package lint contains linting related helpers to be used in mage targets.
package lint

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

var defaultLinters = []string{
	"govet",
	"golint",
	"gofumpt",
	"gosec",
	"unparam",
	"goconst",
	"prealloc",
	"stylecheck",
	"unconvert",
}

// Docker lints the docker file.
func Docker(dockerFile string) error {
	fmt.Printf("lint: running docker lint for file: %s\n", dockerFile)
	return sh.RunV("hadolint", dockerFile)
}

// Go lints the Go code and accepts build tags.
func Go(tags []string) error {
	return code(nil, tags)
}

// GoLinters lints the Go code and accepts linters and build tags.
func GoLinters(linters, tags []string) error {
	return code(linters, tags)
}

// GoDefault lints the Go code and uses default linters and build tags.
func GoDefault() error {
	return code(defaultLinters, []string{bake.BuildTagIntegration, bake.BuildTagComponent})
}

func code(linters, tags []string) error {
	fmt.Printf("lint: running go lint. linters: %v tags: %v\n", linters, tags)

	buildTagFlag := ""
	if len(tags) > 0 {
		buildTagFlag = getBuildTagFlag(tags)
	}

	linterFlag := ""
	if len(linters) > 0 {
		linterFlag = getLinterFlag(linters)
	}

	cmd := "golangci-lint"
	args := strings.Split(fmt.Sprintf("run %s %s --exclude-use-default=false --deadline=5m --modules-download-mode=vendor", linterFlag, buildTagFlag), " ")

	fmt.Printf("Executing cmd: %s %s\n", cmd, strings.Join(args, " "))

	return sh.RunV(cmd, args...)
}

func getBuildTagFlag(tags []string) string {
	return "--build-tags=" + strings.Join(tags, ",")
}

func getLinterFlag(linters []string) string {
	return "--enable " + strings.Join(linters, ",")
}
