// Package code contains code related helpers to be used in mage targets.< $(CURDIR)/infra/deploy/local/Dockerfile
package code

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
	"golang.org/x/mod/sumdb/dirhash"
)

// ModSync runs go module tidy and vendor.
func ModSync() error {
	fmt.Print("code: running go mod sync\n")

	if err := bake.RunGo("mod", "tidy"); err != nil {
		return err
	}
	return bake.RunGo("mod", "vendor")
}

// Fmt runs go fmt.
func Fmt() error {
	fmt.Print("code: running go fmt\n")

	return bake.RunGo("fmt", "./...")
}

// FmtCheck checks if all files are formatted.
func FmtCheck() error {
	fmt.Print("code: running go fmt check\n")

	goFiles, err := getAllGoFiles(".")
	if err != nil {
		return err
	}

	if len(goFiles) == 0 {
		return nil
	}

	files := make([]string, 0, len(goFiles))

	for _, f := range goFiles {
		msg, err := runGofmt("-l", f)
		if err != nil {
			return err
		}
		if msg == "" {
			continue
		}
		files = append(files, msg)
	}

	if len(files) == 0 {
		return nil
	}

	return fmt.Errorf("go files are not formatted:\n%s", strings.Join(files, "\n"))
}

func getAllGoFiles(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(path, "vendor/") {
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get go files: %w", err)
	}

	return files, nil
}

// CheckVendor checks if vendor is in sync with go.mod.
func CheckVendor() error {
	fmt.Print("code: running check vendor\n")

	hash1, err := dirhash.HashDir("vendor/", "mod", dirhash.Hash1)
	if err != nil {
		return fmt.Errorf("failed to create vendor hash: %w", err)
	}

	err = bake.RunGo("mod", "vendor")
	if err != nil {
		return err
	}

	hash2, err := dirhash.HashDir("vendor/", "mod", dirhash.Hash1)
	if err != nil {
		return fmt.Errorf("failed to create vendor hash: %w", err)
	}

	if hash1 != hash2 {
		return errors.New("vendor folder is not in sync")
	}

	return nil
}

func runGofmt(args ...string) (string, error) {
	cmd := "gofmt"

	_, err := exec.LookPath(cmd)
	if err != nil {
		cmd = "docker"

		// todo: set this on init?
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %v", err)
		}

		args = append([]string{"run", "--rm", "--volume", wd + ":/volume", "--workdir", "/volume", "golang:1.14", "gofmt"}, args...)

	}

	fmt.Printf("Executing cmd: %s %s\n", cmd, strings.Join(args, " "))
	return sh.Output(cmd, args...)
}
