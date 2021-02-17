// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/targets/code"
	"github.com/taxibeat/bake/targets/lint"
	"github.com/taxibeat/bake/test"

	// mage:import
	_ "github.com/taxibeat/bake/targets/code"
	// mage:import
	_ "github.com/taxibeat/bake/targets/test"
	// mage:import
	_ "github.com/taxibeat/bake/targets/doc"
	// mage:import
	_ "github.com/taxibeat/bake/targets/lint"
)

// CI groups together ci related tasks.
type CI mg.Namespace

// Run CI with Coveralls and default build tags.
func (CI) Run() error {
	goTargets := code.Go{}
	mg.SerialDeps(goTargets.FmtCheck, goTargets.CheckVendor, lint.Lint{}.Go)

	return test.CoverDefault()
}
