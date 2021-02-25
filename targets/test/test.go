// Package test contains test related mage targets.
package test

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake/docker"
)

var (
	// GoBuildTags to use with Golangci-lint.
	GoBuildTags = []string{
		"component",
		"integration",
	}
	// TestArgs used when running tests.
	TestArgs = []string{
		"test",
		"-mod=vendor",
		"-cover",
		"-race",
	}
	CoverArgs = []string{
		"test",
		"-mod=vendor",
		"-coverpkg=./...",
		"-covermode=atomic",
		"-coverprofile=coverage.txt",
		"-race",
	}
	PFlag = 1
	Pkgs  = "./..."
)

const (
	goCmd = "go"
)

// Test groups together test related tasks.
type Test mg.Namespace

// Unit runs unit tests only.
func (Test) Unit() error {
	args := append(TestArgs, Pkgs)
	return run(args)
}

// All runs all tests.
func (Test) All() error {
	args := append(TestArgs, getBuildTagFlag(GoBuildTags))
	args = append(args, getPFlag(PFlag))
	args = append(args, Pkgs)
	return run(args)
}

// Cover runs all tests and produces a coverage report.
func (Test) Cover() error {
	args := append(CoverArgs, Pkgs)
	return run(args)
}

// CoverAll runs all tests and produces a coverage report.
func (Test) CoverAll() error {
	args := append(CoverArgs, getBuildTagFlag(GoBuildTags))
	args = append(args, getPFlag(PFlag))
	args = append(args, Pkgs)
	return run(args)
}

// Cleanup removes any local resources created by `mage test:all`.
func (Test) Cleanup() error {
	return docker.CleanupResources()
}

func run(args []string) error {
	fmt.Printf("test: running tests with args: %v\n", args)
	fmt.Printf("Executing cmd: %s %s\n", goCmd, strings.Join(args, " "))
	return sh.RunV(goCmd, args...)
}

func getBuildTagFlag(buildTags []string) string {
	return "-tags=" + strings.Join(buildTags, ",")
}

// The -p flag controls the
// the number of programs, such as build commands or
// test binaries, that can be run in parallel.
// The default is the number of CPUs available.
func getPFlag(n int) string {
	return fmt.Sprintf("-p=%d", n)
}
