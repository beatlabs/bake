package ci

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake/build"
)

const (
	goCmd     = "go"
	coverFile = "coverage.txt"
)

// Coverage runs tests and produces a coverfile with provided build tags.
func Coverage(buildTags ...string) error {
	return runTests(coverFile, buildTags)
}

// CoverageDefault //runs tests and produces a coverfile with default build tags.
func CoveragesDefault() error {
	return runTests(coverFile, build.DefaultTags)
}

func runTests(cf string, tags []string) error {
	fmt.Printf("ci: running tests with tags: %v\n", tags)

	args := []string{
		"test",
		"-mod=vendor",
		"-p=1",
		"-count=1",
		"-cover",
		"-coverprofile=" + cf,
		"-covermode=atomic",
	}

	if len(tags) > 0 {
		args = append(args, getBuildTagFlag(tags))
	}
	args = append(args, "./...")

	return sh.RunV(goCmd, args...)
}

func getBuildTagFlag(tags []string) string {
	return "-tags=" + strings.Join(tags, ",")
}
