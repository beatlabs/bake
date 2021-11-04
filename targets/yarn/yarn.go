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
	NpmToken   string
	Cwd        = "web"
	TestSuffix = "--watchAll=false"
)

// Yarn groups yarn related targets.
type Yarn mg.Namespace

func (y Yarn) Install() error {
	args := y.preScript()
	return sh.RunV(cmd, args...)
}

func (y Yarn) Test() error {
	args := y.preScript()

	args = append(args, "test")

	if TestSuffix != "" {
		args = append(args, TestSuffix)
	}

	mg.SerialDeps(y.Install)
	return sh.RunV(cmd, args...)
}

func (y Yarn) setToken() error {
	return sh.RunV("npm", strings.Split(fmt.Sprintf("config set //registry.npmjs.org/:_authToken=%s", NpmToken), " ")...)
}

func (y Yarn) preScript() []string {
	if NpmToken != "" {
		mg.SerialDeps(y.setToken)
	}

	args := []string{}
	if Cwd != "" {
		args = append(args, "--cwd="+Cwd)
	}

	return args
}
