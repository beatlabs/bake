// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/ci"
	"github.com/taxibeat/bake/doc"
	"github.com/taxibeat/bake/lint"
	"github.com/taxibeat/bake/targets/code"

	// mage:import
	_ "github.com/taxibeat/bake/targets/code"
	// mage:import
	_ "github.com/taxibeat/bake/targets/test"
)

// Doc groups together documentation related tasks.
type Doc mg.Namespace

// Dispatch runs the Dispatch HTTP server.
func (Doc) ConfluenceSync() error {
	docReadme := doc.ConfluenceDoc{
		Path: "",
		File: "README.md",
	}
	return doc.ConfluenceSync(docReadme)
}

// Lint groups together lint related tasks.
type Lint mg.Namespace

// All runs all linters.
func (l Lint) All() {
	mg.Deps(l.Docker, l.Go)
}

// Docker lints the docker file.
func (l Lint) Docker() error {
	return lint.Docker("Dockerfile")
}

// Go runs the go linter.
func (l Lint) Go() error {
	return lint.GoDefault()
}

// CI groups together ci related tasks.
type CI mg.Namespace

// Run CI with CodeCov and default build tags.
func (CI) Run() error {
	goTargets := code.Go{}
	mg.SerialDeps(goTargets.CheckVendor, goTargets.FmtCheck, Lint{}.All)
	return ci.CodeCovDefault()
}
