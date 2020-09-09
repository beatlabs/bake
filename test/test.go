// Package test contains test related helpers to be used in mage targets.
package test

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

const (
	// DefaultPkg to run test against.
	DefaultPkg = "./..."
	coverFile  = "coverage.txt"
)

// Run Go tests with cover and race flags enabled. Accepts extra args and build tags.
func Run(extraArgs, tags []string, pkg string) error {
	return run(extraArgs, tags, pkg)
}

// RunDefault Go tests with cover and race flags enabled and with default build tags.
func RunDefault() error {
	return run(nil, []string{bake.BuildTagIntegration, bake.BuildTagComponent}, DefaultPkg)
}

func run(extraArgs, tags []string, pkg string) error {
	if pkg == "" {
		pkg = DefaultPkg
	}

	fmt.Printf("test: running tests with extra args: %v, tags: %v on pkg: %s\n", extraArgs, tags, pkg)

	args := []string{
		"test",
		"-mod=vendor",
		"-cover",
		"-race",
	}
	args = append(args, extraArgs...)

	if len(tags) > 0 {
		args = append(args, getBuildTagFlag(tags))
	}

	args = append(args, pkg)

	fmt.Printf("Executing cmd: %s %s\n", bake.GoCmd, strings.Join(args, " "))

	return sh.RunV(bake.GoCmd, args...)
}

// Cover runs Go test and produce full code coverage result and accepts build tags.
func Cover(buildTags ...string) error {
	return cover(buildTags)
}

// CoverDefault runs Go test and produce full code coverage result and uses default build tags.
func CoverDefault() error {
	return cover([]string{bake.BuildTagIntegration, bake.BuildTagComponent})
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

	fmt.Printf("Executing cmd: %s %s\n", bake.GoCmd, strings.Join(args, " "))

	err := sh.Run(bake.GoCmd, args...)
	if err != nil {
		return err
	}

	return sh.RunV(bake.GoCmd, "tool", "cover", "-func="+coverFile)
}

func getBuildTagFlag(buildTags []string) string {
	return "-tags=" + strings.Join(buildTags, ",")
}
