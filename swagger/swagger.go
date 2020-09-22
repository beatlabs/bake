// Package swagger contains swagger/openapi related helpers to be used in mage targets.
package swagger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

var (
	defaultSwagCmd   = "swag"
	defaultOutputDir = "internal/infra/http/docs"
	defaultAPIDir    = "api"
)

// Create swagger files from source code annotations.
func Create(main, output, api string) error {
	if err := generate(main, output); err != nil {
		return err
	}

	for _, file := range []string{"swagger.json", "swagger.yaml"} {
		source := filepath.Join(output, file)
		destination := filepath.Join(api, file)
		fmt.Printf("moving %s to directory %s\n", source, api)
		if err := os.Rename(source, destination); err != nil {
			return err
		}
	}
	return nil
}

// CreateDefault creates a swagger files from source code annotations with default arguments.
func CreateDefault(main string) error {
	return Create(main, defaultOutputDir, defaultAPIDir)
}

// Check ensures that the generated files are up to date.
func Check(main, api string) error {
	dir, err := ioutil.TempDir("./", "")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Println(err)
		}
	}()

	if err := generate(main, dir); err != nil {
		return err
	}

	for _, file := range []string{"swagger.json", "swagger.yaml"} {
		generated := filepath.Join(dir, file)
		existing := filepath.Join(api, file)
		fmt.Printf("comparing, generated: %s existing: %s\n", generated, existing)
		if err := compareFiles(generated, existing); err != nil {
			return err
		}
	}
	return nil
}

// CheckDefault ensures that the generated files are up to date with default arguments.
func CheckDefault(main string) error {
	return Check(main, defaultAPIDir)
}

func generate(main, output string) error {
	args := []string{
		"init",
		"--generalInfo",
		main,
		"--output",
		output,
	}

	_, err := exec.LookPath(defaultSwagCmd)
	if err != nil {
		fmt.Printf("Warning: %s command not found, using docker\n", defaultSwagCmd)
		return bake.RunDocker("golang:1.14", "go get -u github.com/swaggo/swag/cmd/swag && "+defaultSwagCmd+" "+
			strings.Join(args, " "))
	}

	return sh.RunV(defaultSwagCmd, args...)
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
