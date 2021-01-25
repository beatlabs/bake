// Package ci contains all the ci tasks that a project should have.
package ci

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake/build"
)

const coverFile = "coverage.txt"

// Coveralls runs the actual CI pipeline with Coveralls integration and accepts build tags.
func Coveralls(buildTags ...string) error {
	return coveralls(buildTags)
}

// CoverallsDefault runs the actual CI pipeline with Coveralls integration and default build tags.
func CoverallsDefault() error {
	return coveralls(build.DefaultTags)
}

func coveralls(tags []string) error {
	err := runTests(coverFile, tags)
	if err != nil {
		return err
	}

	defer func() {
		err := os.Remove(coverFile)
		if err != nil {
			fmt.Printf("failed to delete coverage file: %v\n", err)
		}
	}()

	if os.Getenv("COVERALLS_TOKEN") == "" {
		fmt.Printf("coveralls token is not set, skipping code coverage upload step.\n")
		return nil
	}

	return sh.RunV(
		"goveralls",
		"-coverprofile="+coverFile,
	)
}
