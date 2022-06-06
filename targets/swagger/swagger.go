// Package swagger contains swagger/openapi related helpers to be used in mage targets.
package swagger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	swagCmd = "swag"
	// OutputDir is the output directory for the generated files.
	OutputDir = "api"
	// MainGo is the path to the application entrypoint.
	MainGo = ""
	// ExtraArgs passed to the swag command.
	ExtraArgs = []string{"--parseVendor", "--outputTypes", "json,yaml"}
)

// Swagger groups together Swagger related tasks.
type Swagger mg.Namespace

// Create creates a swagger files from source code annotations.
func (Swagger) Create() error {
	return generate(OutputDir)
}

// Check ensures that the generated files are up to date.
func (Swagger) Check() error {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Println(err)
		}
	}()

	if err := generate(dir); err != nil {
		return err
	}

	for _, file := range []string{"swagger.json", "swagger.yaml"} {
		generated := filepath.Join(dir, file)
		existing := filepath.Join(OutputDir, file)
		fmt.Printf("comparing, generated: %s existing: %s\n", generated, existing)
		if err := compareFiles(generated, existing); err != nil {
			return err
		}
	}
	return nil
}

func generate(output string) error {
	args := []string{
		"init",
		"--generalInfo",
		MainGo,
		"--output",
		output,
	}
	args = append(args, ExtraArgs...)
	return sh.RunV(swagCmd, args...)
}

func compareFiles(file1, file2 string) error {
	f1, err := ioutil.ReadFile(filepath.Clean(file1))
	if err != nil {
		return fmt.Errorf("failed to open read %s,: %v", file1, err)
	}
	f2, err := ioutil.ReadFile(filepath.Clean(file2))
	if err != nil {
		return fmt.Errorf("failed to open read %s,: %v", file2, err)
	}

	if bytes.Equal(f1, f2) {
		return nil
	}

	return fmt.Errorf("%s and %s have differences", file1, file2)
}
