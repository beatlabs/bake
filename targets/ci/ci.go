// Package ci contains a ci meta-target.
package ci

import (
	gocode "github.com/beatlabs/bake/targets/code/golang"
	dockerlint "github.com/beatlabs/bake/targets/lint/docker"
	golint "github.com/beatlabs/bake/targets/lint/golang"
	"github.com/beatlabs/bake/targets/test"
	"github.com/magefile/mage/mg"
)

// CI runs the Continuous Integration pipeline.
func CI() error {
	targets := []interface{}{
		gocode.Go{}.FmtCheck,
		dockerlint.Lint{}.Docker,
		gocode.Go{}.CheckVendor,
		golint.Lint{}.Go,
		test.Test{}.CoverAll,
	}

	mg.SerialDeps(targets...)

	return nil
}
