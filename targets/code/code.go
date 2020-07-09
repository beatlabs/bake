// Package code contains go code related mage targets.
package code

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/code"
)

// Go groups together go related tasks.
type Go mg.Namespace

// ModSync runs go module tidy and vendor.
func (Go) ModSync() error {
	return code.ModSync()
}

// Fmt runs go fmt.
func (Go) Fmt() error {
	return code.Fmt()
}

// Fumpt runs gofumpt.
func (Go) Fumpt() error {
	return code.Fumpt()
}

// FmtCheck checks if all files are formatted.
func (Go) FmtCheck() error {
	return code.FmtCheck()
}

// FumptCheck checks if all files are formatted with gofumpt.
func (Go) FumptCheck() error {
	return code.FumptCheck()
}

// CheckVendor checks if vendor is in sync with go.mod.
func (Go) CheckVendor() error {
	return code.CheckVendor()
}
