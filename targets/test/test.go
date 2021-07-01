// Package test contains test related mage targets.
package test

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake/docker"
)

const (
	goCmd              = "go"
	componentTestTag   = "component"
	integrationTestTag = "integration"
)

var (
	// GoBuildTags used when running all tests.
	GoBuildTags = []string{componentTestTag, integrationTestTag}

	// TestArgs used in test targets.
	TestArgs = []string{
		"test",
		"-mod=vendor",
		"-cover",
		"-race",
	}
	// CoverArgs used in coverage targets.
	CoverArgs = []string{
		"test",
		"-mod=vendor",
		"-coverpkg=./...",
		"-covermode=atomic",
		"-coverprofile=coverage.txt",
		"-race",
	}
	// Pkgs is the pkg pattern to target.
	Pkgs = "./..."
	// CoverExcludePatterns is a list of pkg patterns to prune from coverage.txt
	CoverExcludePatterns = []string{"proto/generated"}
	// CoverExcludeFile is the coverage file to prune.
	CoverExcludeFile = "coverage.txt"
)

// Test groups together test related tasks.
type Test mg.Namespace

// Unit runs unit tests.
func (Test) Unit() error {
	args := append(TestArgs, Pkgs)
	return run(args)
}

// Integration runs integration tests.
func (Test) Integration() error {
	args := append(appendCacheBustingArg(TestArgs), getBuildTagFlag([]string{integrationTestTag}), Pkgs)
	return run(args)
}

// Component runs unit and component tests.
func (Test) Component() error {
	args := append(appendCacheBustingArg(TestArgs), getBuildTagFlag([]string{componentTestTag}), Pkgs)
	return run(args)
}

// All runs all tests.
func (Test) All() error {
	args := append(appendCacheBustingArg(TestArgs), getBuildTagFlag(GoBuildTags), Pkgs)
	return run(args)
}

// CoverUnit runs unit tests and produces a coverage report.
func (Test) CoverUnit() error {
	args := append(CoverArgs, Pkgs)
	if err := run(args); err != nil {
		return err
	}
	return pruneCoverageFile(CoverExcludeFile, CoverExcludePatterns)
}

// CoverAll runs all tests and produces a coverage report.
func (Test) CoverAll() error {
	args := append(CoverArgs, getBuildTagFlag(GoBuildTags), Pkgs)
	if err := run(args); err != nil {
		return err
	}
	return pruneCoverageFile(CoverExcludeFile, CoverExcludePatterns)
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

func appendCacheBustingArg(args []string) []string {
	return append(args, "-count=1")
}
