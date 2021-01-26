// Package test contains test related helpers to be used in mage targets.
package test

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake/build"
)

const (
	goCmd      = "go"
	defaultPkg = "./..."
	coverFile  = "coverage.txt"
)

// DefaultTestArgs used when running tests.
var DefaultTestArgs = []string{
	"test",
	"-mod=vendor",
	"-cover",
	"-race",
}

// Run Go tests with cover and race flags enabled. Accepts build tags, extra args and a specific pkg.
func Run(tags, extraArgs []string, pkg string) error {
	args := DefaultTestArgs
	if len(tags) > 0 {
		args = append(args, getBuildTagFlag(tags))
	}

	args = append(args, extraArgs...)

	if pkg == "" {
		pkg = defaultPkg
	}
	args = append(args, pkg)
	return run(args)
}

// RunDefault Go tests with cover, race flags enabled, with default build tags and default pkg.
func RunDefault() error {
	args := DefaultTestArgs
	args = append(args, getBuildTagFlag(build.DefaultTags))
	args = append(args, defaultPkg)
	return run(args)
}

func run(args []string) error {
	fmt.Printf("test: running tests with args: %v\n", args)
	fmt.Printf("Executing cmd: %s %s\n", goCmd, strings.Join(args, " "))
	return sh.RunV(goCmd, args...)
}

// Cover runs Go test and produce full code coverage result and accepts build tags.
func Cover(buildTags ...string) error {
	return cover(buildTags)
}

// CoverDefault runs Go test and produce full code coverage result and uses default build tags.
func CoverDefault() error {
	return cover(build.DefaultTags)
}

func cover(tags []string) error {
	fmt.Printf("test: running cover with tags: %v\n", tags)

	defer func() {
		err := os.Remove(coverFile)
		if err != nil {
			fmt.Printf("failed to delete coverage file: %v\n", err)
		}
	}()

	args := []string{
		"test",
		"-mod=vendor",
		"-coverpkg=./...",
		"-covermode=atomic",
		"-coverprofile=" + coverFile,
		"-race",
	}

	if len(tags) > 0 {
		args = append(args, getBuildTagFlag(tags))
	}

	args = append(args, "./...")

	fmt.Printf("Executing cmd: %s %s\n", goCmd, strings.Join(args, " "))

	err := sh.Run(goCmd, args...)
	if err != nil {
		return err
	}

	return sh.RunV(goCmd, "tool", "cover", "-func="+coverFile)
}

func getBuildTagFlag(buildTags []string) string {
	return "-tags=" + strings.Join(buildTags, ",")
}
