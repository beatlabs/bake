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

// FmtCheck checks if all files are formatted.
func (Go) FmtCheck() error {
	return code.FmtCheck()
}

// CheckVendor checks if vendor is in sync with go.mod.
func (Go) CheckVendor() error {
	return code.CheckVendor()
}
