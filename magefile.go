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
	// mage:import
	_ "github.com/taxibeat/bake/targets/doc"
)

// Lint groups together lint related tasks.
type Lint mg.Namespace

// Go runs the go linter.
func (l Lint) Go() error {
	return lint.GoDefault()
}

// CI groups together ci related tasks.
type CI mg.Namespace

// Run CI with Coveralls and default build tags.
func (CI) Run() error {
	goTargets := code.Go{}
	err := goTargets.FmtCheck()
	if err != nil {
		return err
	}
	err = goTargets.CheckVendor()
	if err != nil {
		return err
	}
	err = Lint{}.Go()
	if err != nil {
		return err
	}
	return ci.CoverageDefault()
}
