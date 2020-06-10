// Package ci contains ci related mage targets.
package ci

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/ci"
	"github.com/taxibeat/bake/targets/code"
	"github.com/taxibeat/bake/targets/lint"
)

// CI groups together ci related tasks.
type CI mg.Namespace

// Run CI with CodeCov and default build tags.
func (CI) Run() error {
	goTargets := code.Go{}
	mg.SerialDeps(goTargets.CheckVendor, goTargets.FmtCheck, lint.Lint{}.All)
	return ci.CodeCovDefault()
}
