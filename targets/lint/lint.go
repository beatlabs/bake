// Package lint contains linting related mage targets.
package lint

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/lint"
)

// Lint groups together lint related tasks.
type Lint mg.Namespace

// All runs all linters.
func (l Lint) All() {
	mg.Deps(l.Docker, l.Go)
}

// Docker lints the docker file.
func (l Lint) Docker() error {
	return lint.Docker()
}

// Go runs the go linter.
func (l Lint) Go() error {
	return lint.GoDefault()
}
