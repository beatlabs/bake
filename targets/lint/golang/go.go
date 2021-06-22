// Package golang contains linting related mage targets.
package golang

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// GolangciFlags are passed directly to the golanci-lint command in addition to the config file.
var GolangciFlags = []string{}

// GolangciConfigPath sets the path to a config file to be used instead of the default one.
var GolangciConfigPath string

const defaultFile = `
run:
  timeout: 5m
  tests: true
  modules-download-mode: vendor
  build-tags:
    - component
    - integration

linters:
  disable-all: true
  enable:
    - govet
    - revive
    - gofumpt
    - gosec
    - unparam
    - goconst
    - prealloc
    - stylecheck
    - unconvert
    - errcheck
    - deadcode

issues:
  exclude-use-default: false
`

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Go runs the golangci-lint linter.
func (l Lint) Go() error {
	args := "run "

	if mg.Verbose() {
		args += "-v "
	}

	if GolangciConfigPath != "" {
		args += "--config " + GolangciConfigPath
	} else {
		path, err := persistDefaultFile()
		if err != nil {
			return err
		}
		defer func() { _ = os.Remove(path) }()
		args += "--config " + path
	}

	args += " " + strings.Join(GolangciFlags, " ")

	return sh.RunV("golangci-lint", strings.Split(args, " ")...)
}

func persistDefaultFile() (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "bake-*.yml")
	if err != nil {
		return "", err
	}

	if _, err = file.Write([]byte(defaultFile)); err != nil {
		return "", err
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	return file.Name(), nil
}
