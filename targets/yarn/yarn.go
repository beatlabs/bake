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
	NpmToken    string
	Cwd         = "web"
	UnitTestCmd = "test --watchAll=false"
	LintCmd     = "lint"
)

// Yarn groups yarn related targets.
type Yarn mg.Namespace

func (y Yarn) Install() error {
	return sh.RunV(cmd, y.prepScript()...)
}

func (y Yarn) Test() error {
	args := append(y.prepScript(), strings.Split(UnitTestCmd, " ")...)
	mg.SerialDeps(y.Install)
	return sh.RunV(cmd, args...)
}

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
