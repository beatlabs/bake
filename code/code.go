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
	if err := sh.RunV(bake.GoCmd, "mod", "tidy"); err != nil {
		return err
	}
	return sh.RunV(bake.GoCmd, "mod", "vendor")
}

// Fmt runs go fmt.
func Fmt() error {
	return sh.RunV(bake.GoCmd, "fmt", "./...")
}

// FmtCheck checks if all files are formatted.
func FmtCheck() error {
	goFiles, err := getAllGoFiles(".")
	if err != nil {
		return err
	}

	if len(goFiles) == 0 {
		return nil
	}

	files := make([]string, 0, len(goFiles))

	for _, f := range goFiles {
		msg, err := sh.Output("gofmt", "-l", f)
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
