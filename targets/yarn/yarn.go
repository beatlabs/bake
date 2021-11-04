// Package yarn contains yarn related mage targets.
package yarn

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const cmd = "yarn"

var (
	// NpmToken to be added to npm config in order to access private packages. Optional.
	NpmToken string
	// Cwd is the current working directory. Optional.
	Cwd = "web"
	// TestCmd is the test subcommand to be executed.
	TestCmd = "test --watchAll=false"
	// LintCmd is the lint subcommand to be executed.
	LintCmd = "lint"
)

// Yarn groups yarn related targets.
type Yarn mg.Namespace

// Install runs the yarn install subcommand.
func (y Yarn) Install() error {
	return sh.RunV(cmd, y.prepScript()...)
}

// Test runs the yarn test subcommand.
func (y Yarn) Test() error {
	args := append(y.prepScript(), strings.Split(TestCmd, " ")...)
	mg.SerialDeps(y.Install)
	return sh.RunV(cmd, args...)
}

// Lint runs the yarn lint subcommand.
func (y Yarn) Lint() error {
	args := append(y.prepScript(), strings.Split(LintCmd, " ")...)
	mg.SerialDeps(y.Install)
	return sh.RunV(cmd, args...)
}

func (y Yarn) prepScript() []string {
	if NpmToken != "" {
		mg.SerialDeps(y.setToken)
	}

	args := []string{}
	if Cwd != "" {
		args = append(args, "--cwd="+Cwd)
	}

	return args
}

func (y Yarn) setToken() error {
	return sh.RunV("npm", strings.Split(fmt.Sprintf("config set //registry.npmjs.org/:_authToken=%s", NpmToken), " ")...)
}
