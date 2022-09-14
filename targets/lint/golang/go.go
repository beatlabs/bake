// Package golang contains linting related mage targets.
package golang

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

//go:embed golangci.config.yml
var config string

// Lint groups together lint related tasks.
type Lint mg.Namespace

// GoShowConfig outputs the golangci-lint linter config.
func (l Lint) GoShowConfig() error {
	_, _ = fmt.Print(config)
	return nil
}

// Go runs the golangci-lint linter.
func (l Lint) Go() error {
	args := "run "

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		args += "--out-format=github-actions "
	}

	if mg.Verbose() {
		args += "-v "
	}

	path, err := persistDefaultFile()
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()
	args += "--config " + path

	return sh.Run("golangci-lint", strings.Split(args, " ")...)
}

func persistDefaultFile() (string, error) {
	file, err := os.CreateTemp(os.TempDir(), "bake-*.yml")
	if err != nil {
		return "", err
	}

	if _, err = file.Write([]byte(config)); err != nil {
		return "", err
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	return file.Name(), nil
}
