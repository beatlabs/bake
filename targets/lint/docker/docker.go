// Package docker contains linting related mage targets.
package docker

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// DockerFiles to lint.
var DockerFiles = []string{"./infra/deploy/local/Dockerfile"}

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Docker lints the docker file.
func (l Lint) Docker() error {
	for _, path := range DockerFiles {
		fmt.Printf("lint: running docker lint for file: %s\n", path)
		if err := sh.RunV("hadolint", path); err != nil {
			return err
		}
	}
	return nil
}
