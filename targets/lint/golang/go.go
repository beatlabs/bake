// Package golang contains linting related mage targets.
package golang

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const config = `
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
    - ineffassign
    - structcheck
    - tparallel
    - whitespace

issues:
  exclude-use-default: false
`

// Lint groups together lint related tasks.
type Lint mg.Namespace

// GoShowConfig outputs the golangci-lint linter config.
func (l Lint) GoShowConfig() error {
	fmt.Println(strings.TrimSpace(config))
	return nil
}

// Go runs the golangci-lint linter.
func (l Lint) Go() error {
	args := "run "

	if mg.Verbose() {
		args += "-v "
	}

	path, err := persistDefaultFile()
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()
	args += "--config " + path

	return sh.RunV("golangci-lint", strings.Split(args, " ")...)
}

func persistDefaultFile() (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "bake-*.yml")
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
