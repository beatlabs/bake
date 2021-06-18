// Package golang contains linting related mage targets.
package golang

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// GoLinters to be used with Golangci-lint.
var GoLinters = []string{
	"govet",
	"revive",
	"gofumpt",
	"gosec",
	"unparam",
	"goconst",
	"prealloc",
	"stylecheck",
	"unconvert",
}

// GoBuildTags to use with Golangci-lint.
var GoBuildTags = []string{
	"component",
	"integration",
}

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Go runs the go linter.
func (l Lint) Go() error {
	fmt.Printf("lint: running go lint. linters: %v tags: %v\n", GoLinters, GoBuildTags)

	buildTagFlag := ""
	if len(GoBuildTags) > 0 {
		buildTagFlag = getBuildTagFlag(GoBuildTags)
	}

	linterFlag := ""
	if len(GoLinters) > 0 {
		linterFlag = getLinterFlag(GoLinters)
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
