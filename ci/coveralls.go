package ci

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

// Coveralls runs the actual CI pipeline with Coveralls integration and accepts build tags.
func Coveralls(buildTags ...string) error {
	return coveralls(buildTags)
}

// CoverallsDefault runs the actual CI pipeline with Coveralls integration and default build tags.
func CoverallsDefault() error {
	return coveralls([]string{bake.BuildTagIntegration, bake.BuildTagComponent})
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

	return sh.RunV(
		"goveralls",
		"-coverprofile="+coverFile,
		"-repotoken="+os.Getenv("COVERALLSTOKEN"),
	)
}
