// Package ci contains a ci meta-target.
package ci

import (
	"os"

	gocode "github.com/beatlabs/bake/targets/code/golang"
	dockerlint "github.com/beatlabs/bake/targets/lint/docker"
	golint "github.com/beatlabs/bake/targets/lint/golang"
	"github.com/beatlabs/bake/targets/prometheus"
	"github.com/beatlabs/bake/targets/proto"
	"github.com/beatlabs/bake/targets/swagger"
	"github.com/beatlabs/bake/targets/test"
	"github.com/magefile/mage/mg"
)

// CI runs the Continuous Integration pipeline.
func CI() error {
	var targets []interface{}

	targets = append(targets,
		gocode.Go{}.FmtCheck,
		dockerlint.Lint{}.Docker,
		prometheus.Prometheus{}.Lint,
	)

	if swagger.MainGo != "" {
		targets = append(targets, swagger.Swagger{}.Check)
	}

	if _, err := os.Stat(proto.SchemasLocation); !os.IsNotExist(err) {
		targets = append(targets, proto.Proto{}.SchemaValidateAll)
	}

	if prometheus.TestsDir != "" {
		targets = append(targets, prometheus.Prometheus{}.Test)
	}

	targets = append(targets,
		gocode.Go{}.CheckVendor,
		golint.Lint{}.Go,
		test.Test{}.CoverAll,
	)

	mg.SerialDeps(targets...)

	return nil
}
