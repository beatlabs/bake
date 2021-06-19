// Package golang contains linting related mage targets.
package golang

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// GoLinters set the --enable flag in the golangci-lint command.
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

// GoBuildTags set the --built-tags flag in the golangci-lint command.
var GoBuildTags = []string{
	"component",
	"integration",
}

// GoRawFlags are passed directly to the golanci-lint command.
var GoRawFlags = []string{
	"--no-config",
	"--disable-all",
	"--exclude-use-default=false",
	"--deadline=5m",
	"--modules-download-mode=vendor",
}

// ConfigFilePath sets the --config flag in the golangci-lint command.
var ConfigFilePath string

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Go runs the golangci-lint linter.
// If ConfigFilePath is set a config file is used, otherwise
// command flags are generated from GoRawFlags, GoLinters and GoBuildTags.
func (l Lint) Go() error {
	var args string
	if ConfigFilePath != "" {
		args = "--config " + ConfigFilePath
	} else {
		if len(GoRawFlags) > 0 {
			args = strings.Join(GoRawFlags, " ")
		}

		if len(GoBuildTags) > 0 {
			args += " --build-tags=" + strings.Join(GoBuildTags, ",")
		}

		if len(GoLinters) > 0 {
			args += " --enable " + strings.Join(GoLinters, ",")
		}
	}

	args = "run -v " + args
	cmd := "golangci-lint"
	fmt.Printf("Executing cmd: %s %s\n", cmd, args)

	return sh.RunV(cmd, strings.Split(args, " ")...)
}
