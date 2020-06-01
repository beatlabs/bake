// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/ci"
	"github.com/taxibeat/bake/lint"
	"github.com/taxibeat/bake/targets/code"

	// mage:import
	_ "github.com/taxibeat/bake/targets/code"
	// mage:import
	_ "github.com/taxibeat/bake/targets/test"
)

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Go runs the go linter.
func (l Lint) Go() error {
	return lint.GoDefault()
}

// CI groups together ci related tasks.
type CI mg.Namespace

// Run CI with CodeCov and default build tags.
func (CI) Run() error {
	goTargets := code.Go{}
	mg.SerialDeps(goTargets.CheckVendor, goTargets.FmtCheck, Lint{}.Go())
	return ci.CodeCovDefault()
}
