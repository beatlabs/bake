// Package ci contains a ci meta-target.
package ci

import (
	"os"

	"github.com/magefile/mage/mg"
	gocode "github.com/taxibeat/bake/targets/code/golang"
	dockerlint "github.com/taxibeat/bake/targets/lint/docker"
	golint "github.com/taxibeat/bake/targets/lint/golang"
	"github.com/taxibeat/bake/targets/lint/prometheus"
	"github.com/taxibeat/bake/targets/proto"
	"github.com/taxibeat/bake/targets/swagger"
	"github.com/taxibeat/bake/targets/test"
)

// CI runs the Continuous Integration pipeline.
func CI() error {
	var targets []interface{}

	targets = append(targets,
		gocode.Go{}.FmtCheck,
		dockerlint.Lint{}.Docker,
	)

	if swagger.MainGo != "" {
		targets = append(targets, swagger.Swagger{}.Check)
	}

	if _, err := os.Stat(proto.SchemasLocation); !os.IsNotExist(err) {
		targets = append(targets, proto.Proto{}.SchemaValidateAll)
	}

	if len(prometheus.AlertFiles) > 0 {
		targets = append(targets, prometheus.Lint{}.AlertRules)
	}

	targets = append(targets,
		gocode.Go{}.CheckVendor,
		golint.Lint{}.Go,
		test.Test{}.CoverAll,
	)

	mg.SerialDeps(targets...)

	return nil
}
