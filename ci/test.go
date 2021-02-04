package ci

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/sh"
)

const goCmd = "go"

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
