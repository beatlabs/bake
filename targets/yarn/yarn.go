// Package yarn contains yarn related mage targets.
package yarn

import (
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const cmd = "yarn"

// Yarn groups yarn related targets.
type Yarn mg.Namespace

func (y Yarn) Script(script string) error {
	return sh.RunV(cmd, strings.Split(script, " ")...)
}
