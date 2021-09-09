// +build mage

package main

import (
	"github.com/taxibeat/bake/targets/lint/docker"
	"github.com/taxibeat/bake/targets/test"

	// mage:import
	_ "github.com/taxibeat/bake/targets/code/golang"
	// mage:import
	_ "github.com/taxibeat/bake/targets/test"
	// mage:import
	_ "github.com/taxibeat/bake/targets/doc"
	// mage:import
	_ "github.com/taxibeat/bake/targets/lint/docker"
	// mage:import
	_ "github.com/taxibeat/bake/targets/lint/golang"
	// mage:import
	_ "github.com/taxibeat/bake/targets/ci"
)

func init() {
	docker.DockerFiles = []string{"./Dockerfile"}
	test.CoverExcludePatterns = []string{"doc/", "docker/component/testservice/"}
}
