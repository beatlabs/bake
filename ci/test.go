package ci

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

func runTests(cf string, tags []string) error {
	fmt.Printf("ci: running tests with tags: %v\n", tags)

	args := []string{
		"test",
		"-mod=vendor",
		"-count=1",
		"-cover",
		"-coverprofile=" + cf,
		"-covermode=atomic",
	}

	if len(tags) > 0 {
		args = append(args, getBuildTagFlag(tags))
	}
	args = append(args, "./...")

	return sh.RunV(bake.GoCmd, args...)
}

func getBuildTagFlag(tags []string) string {
	return "-tags=" + strings.Join(tags, ",")
}
