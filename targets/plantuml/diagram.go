// Package plantuml contains support for compiling your
// [PlantUML](https://plantuml.com/) text files to the
// associated png image
package plantuml

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	javaCMD = `java`
)

// PlantUML groups together test related diagram tasks.
type PlantUML mg.Namespace

// Generate creates png diagrams from all puml files in current folder and all sub folders.
// The output will be located in the same folder of the corresponding input.
func (PlantUML) Generate() error {
	return sh.RunV(javaCMD, "-jar", "/usr/bin/plantuml.jar", "-tpng", "-verbose", `*/**.puml`)
}
