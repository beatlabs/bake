// Package yarn contains yarn related mage targets.
package yarn

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const cmd = "yarn"

var NpmToken string

// Yarn groups yarn related targets.
type Yarn mg.Namespace

func (y Yarn) Script(script string) error {
	if NpmToken != "" {
		mg.SerialDeps(y.setToken)
	}
	return sh.RunV(cmd, strings.Split(script, " ")...)
}

func (y Yarn) setToken() error {
	return sh.RunV("npm", strings.Split(fmt.Sprintf("config set //registry.npmjs.org/:_authToken=%s", NpmToken), " ")...)
}
