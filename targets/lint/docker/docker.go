// Package docker contains linting related mage targets.
package docker

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// DockerFiles to lint.
var DockerFiles = []string{"./infra/deploy/local/Dockerfile"}

// Args are are the cli args passsed to Hadolint.
var Args = []string{
	"--ignore", "DL3008", // Pin versions in apt get install
	"--ignore", "DL4001", // Either use Wget or Curl but not both
	"--ignore", "DL4006", // Set the SHELL option -o pipefail
}

const cmd = "hadolint"

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Docker lints the docker file.
func (l Lint) Docker() error {
	for _, path := range DockerFiles {
		fmt.Printf("lint: running docker lint for file: %s\n", path)
		args := append(Args, path)
		if err := sh.RunV(cmd, args...); err != nil {
			return err
		}
	}
	return nil
}
