// Package docker contains linting related mage targets.
package docker

import (
	"github.com/taxibeat/bake/internal/shfmt"

	"github.com/magefile/mage/mg"
)

// DockerFiles to lint.
var DockerFiles = []string{"./infra/deploy/local/Dockerfile"}

// Args are are the cli args passsed to Hadolint.
var Args = []string{
	"--ignore", "DL3008", // Pin versions in apt get install
	"--ignore", "DL4001", // Either use Wget or Curl but not both
	"--ignore", "DL4006", // Set the SHELL option -o pipefail
	"--ignore", "DL3059", // Multiple consecutive `RUN` instructions. Consider consolidation.
}

const (
	cmd       = "hadolint"
	namespace = "lint"
)

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Docker lints the docker file.
func (l Lint) Docker() error {
	shfmt.PrintStartTarget(namespace, "docker")
	for _, path := range DockerFiles {
		args := append(Args, path) // nolint:gocritic
		if err := shfmt.RunV(cmd, args...); err != nil {
			return err
		}
	}
	return nil
}
