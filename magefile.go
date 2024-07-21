//go:build mage

package main

import (
	"github.com/beatlabs/bake/targets/lint/docker"
	"github.com/beatlabs/bake/targets/test"

	// mage:import
	_ "github.com/beatlabs/bake/targets/code/golang"
	// mage:import
	_ "github.com/beatlabs/bake/targets/test"
	// mage:import
	_ "github.com/beatlabs/bake/targets/lint/docker"
	// mage:import
	_ "github.com/beatlabs/bake/targets/lint/golang"
	// mage:import
	_ "github.com/beatlabs/bake/targets/ci"
)

func init() {
	docker.DockerFiles = []string{"./Dockerfile"}
	test.CoverExcludePatterns = []string{"doc/", "docker/component/testservice/"}
}
