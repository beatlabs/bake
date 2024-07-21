//go:build mage

package main

import (
	"github.com/beatlabs/bake/targets/lint/docker"
	"github.com/beatlabs/bake/targets/prometheus"
	"github.com/beatlabs/bake/targets/test"

	// mage:import
	_ "github.com/beatlabs/bake/targets/code/golang"
	// mage:import
	_ "github.com/beatlabs/bake/targets/test"
	// mage:import
	_ "github.com/beatlabs/bake/targets/doc"
	// mage:import
	_ "github.com/beatlabs/bake/targets/diagram"
	// mage:import
	_ "github.com/beatlabs/bake/targets/lint/docker"
	// mage:import
	_ "github.com/beatlabs/bake/targets/lint/golang"
	// mage:import
	_ "github.com/beatlabs/bake/targets/prometheus"
	// mage:import
	_ "github.com/beatlabs/bake/targets/ci"
)

func init() {
	docker.DockerFiles = []string{"./Dockerfile"}
	prometheus.AlertsDir = "./targets/prometheus/examples"
	test.CoverExcludePatterns = []string{"doc/", "docker/component/testservice/"}
}
