// Package swagger contains swagger/openapi related helpers to be used in mage targets.
package swagger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/magefile/mage/sh"
)

var (
	defaultSwagCmd  = "swag"
	defaultSwagArgs = []string{
		"init", "--generalInfo", "../../../cmd/dispatch/main.go", "--dir", "./internal/infra/http", "-output",
	}
	defaultDocsDir = "internal/infra/http/docs"
	defaultAPIDir  = "api"
)

// Create creates a swagger files from source code annotations.
func Create(cmd string, args []string, docsDir, apiDir string) error {
	args = append(args, docsDir)
	if err := sh.RunV(cmd, args...); err != nil {
		return err
	}

	for _, ext := range []string{"json", "yaml"} {
		fname := "swagger." + ext
		if err := sh.RunV("mv", filepath.Join(docsDir, fname), apiDir); err != nil {
			return err
		}
	}
	return nil

}

// CreateDefault creates a swagger files from source code annotations with default arguments.
func CreateDefault() error {
	return Create(defaultSwagCmd, defaultSwagArgs, defaultDocsDir, defaultAPIDir)
}

// Check ensures that the generated files are up to date.
func Check(cmd string, args []string, apiDir string) error {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Println(err)
		}
	}()

	args = append(args, dir)
	if err := sh.RunV(cmd, args...); err != nil {
		return err
	}

	for _, ext := range []string{"json", "yaml"} {
		fname := "swagger." + ext
		if err := sh.RunV("cmp", filepath.Join(dir, fname), filepath.Join(apiDir, fname)); err != nil {
			return err
		}
	}
	return nil
}

// CheckDefault ensures that the generated files are up to date with default arguments.
func CheckDefault() error {
	return Check(defaultSwagCmd, defaultSwagArgs, defaultAPIDir)
}
