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
		// CoverUnit instead of CoverAll: component tests start Kafka, MongoDB,
		// Redis, Consul, Jaeger and more simultaneously, exceeding the 7 GB
		// memory limit of GitHub-hosted runners. Run CoverAll locally or on a
		// self-hosted runner with sufficient memory.
		test.Test{}.CoverUnit,
	}

	mg.SerialDeps(targets...)

	return nil
}
