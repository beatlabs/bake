// Package code contains code related helpers to be used in mage targets.
package code

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
	"golang.org/x/mod/sumdb/dirhash"
)

// ModSync runs go module tidy and vendor.
func ModSync() error {
	fmt.Print("code: running go mod sync\n")

	if err := sh.RunV(bake.GoCmd, "mod", "tidy"); err != nil {
		return err
	}
	return sh.RunV(bake.GoCmd, "mod", "vendor")
}

// Fmt runs go fmt.
func Fmt() error {
	fmt.Print("code: running go fmt\n")

	return sh.RunV(bake.GoCmd, "fmt", "./...")
}

// Fumpt runs gofumpt.
func Fumpt() error {
	fmt.Print("code: running gofumpt\n")

	return sh.RunV("gofumpt", "-s", "-w", "-extra", ".")
}

// FmtCheck checks if all files are formatted.
func FmtCheck() error {
	fmt.Print("code: running go fmt check\n")

	files, err := runCmdOnFiles("gofmt", "-l")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	return fmt.Errorf("go files are not formatted:\n%s", strings.Join(files, "\n"))
}

// FumptCheck checks if all files are formatted with gofumpt.
func FumptCheck() error {
	fmt.Print("code: running gofumpt check\n")

	files, err := runCmdOnFiles("gofumpt", "-l")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	return fmt.Errorf("go files are not gofumpt-ed:\n%s", strings.Join(files, "\n"))
}

func runCmdOnFiles(cmd, args string) ([]string, error) {
	goFiles, err := getAllGoFiles(".")
	if err != nil {
		return nil, err
	}

	if len(goFiles) == 0 {
		return nil, nil
	}

	files := make([]string, 0, len(goFiles))

	for _, f := range goFiles {
		msg, err := sh.Output(cmd, args, f)
		if err != nil {
			return nil, err
		}
		if msg == "" {
			continue
		}
		files = append(files, msg)
	}

	return files, nil
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

	err = sh.RunV(bake.GoCmd, "mod", "vendor")
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
