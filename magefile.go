// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	gocode "github.com/taxibeat/bake/targets/code/golang"
	"github.com/taxibeat/bake/targets/lint/docker"
	golint "github.com/taxibeat/bake/targets/lint/golang"
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
)

func init() {
	docker.DockerFiles = []string{"./Dockerfile"}
}

// CI executes all CI targets.
func CI() error {
	goTargets := gocode.Go{}
	mg.SerialDeps(goTargets.FmtCheck, goTargets.CheckVendor, golint.Lint{}.Go, docker.Lint{}.Docker)

	return test.Test{}.CoverAll()
}
