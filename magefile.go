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

// CI groups together ci related tasks.
type CI mg.Namespace

// Run CI with Coveralls and default build tags.
func (CI) Run() error {
	goTargets := gocode.Go{}
	mg.SerialDeps(goTargets.FmtCheck, goTargets.CheckVendor, golint.Lint{}.Go, docker.Lint{}.Docker)

	return test.Test{}.Cover()
}
