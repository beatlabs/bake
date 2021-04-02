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
	OutputDir = "internal/infra/http/docs"
	// APIDir is the final directory.
	APIDir = "api"
	// MainGo is the path to the application entrypoint.
	MainGo = "cmd/replaceme/main.go"
)

// Swagger groups together Swagger related tasks.
type Swagger mg.Namespace

// Create creates a swagger files from source code annotations.
func (Swagger) Create() error {
	if err := generate(MainGo, OutputDir); err != nil {
		return err
	}

	for _, file := range []string{"swagger.json", "swagger.yaml"} {
		source := filepath.Join(OutputDir, file)
		destination := filepath.Join(APIDir, file)
		fmt.Printf("moving %s to directory %s\n", source, APIDir)
		if err := os.Rename(source, destination); err != nil {
			return err
		}
	}
	return nil
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

	if err := generate(MainGo, dir); err != nil {
		return err
	}

	for _, file := range []string{"swagger.json", "swagger.yaml"} {
		generated := filepath.Join(dir, file)
		existing := filepath.Join(APIDir, file)
		fmt.Printf("comparing, generated: %s existing: %s\n", generated, existing)
		if err := compareFiles(generated, existing); err != nil {
			return err
		}
	}
	return nil
}

func generate(main, output string) error {
	args := []string{
		"init",
		"--generalInfo",
		main,
		"--output",
		output,
	}
	if err := sh.RunV(swagCmd, args...); err != nil {
		return err
	}

	return nil
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
