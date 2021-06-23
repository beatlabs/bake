// Package ci contains a ci meta-target.
package ci

import (
	"github.com/magefile/mage/mg"
	gocode "github.com/taxibeat/bake/targets/code/golang"
	dockerlint "github.com/taxibeat/bake/targets/lint/docker"
	golint "github.com/taxibeat/bake/targets/lint/golang"
	"github.com/taxibeat/bake/targets/swagger"
	"github.com/taxibeat/bake/targets/test"
)

// CI runs the Continuous Integration pipeline.
func CI() error {
	targets := []interface{}{
		gocode.Go{}.CheckVendor,
		gocode.Go{}.FmtCheck,
		golint.Lint{}.Go,
		dockerlint.Lint{}.Docker,
		test.Test{}.CoverAll,
	}

	if swagger.MainGo != "" {
		targets = append(targets, swagger.Swagger{}.Check)
	}

	mg.SerialDeps(targets...)

	return nil
}
