// Package test contains test related mage targets.
package test

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/test"
)

// Test groups together test related tasks.
type Test mg.Namespace

// Unit runs unit tests only.
func (Test) Unit() error {
	return test.Run(nil, nil, "")
}

// All runs all tests.
func (Test) All() error {
	return test.RunDefault()
}

// Cover runs all tests and produces a coverage report.
func (Test) Cover() error {
	return test.CoverDefault()
}
