// Package golang contains go code related mage targets.
package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/beatlabs/bake/internal/sh"
)

// Go groups together go related tasks.
type Go mg.Namespace

const namespace = "code"

// ModSync runs go module tidy and vendor.
func (Go) ModSync() error {
	sh.PrintStartTarget(namespace, "go mod sync")

	if err := sh.RunV(goCmd, "mod", "tidy"); err != nil {
		return err
	}
	return sh.RunV(goCmd, "mod", "vendor")
}

// Fmt runs go fmt.
func (Go) Fmt() error {
	sh.PrintStartTarget(namespace, "go fmt")

	return sh.RunV(goCmd, "fmt", "./...")
}

// FmtCheck checks if all files are formatted.
func (Go) FmtCheck() error {
	sh.PrintStartTarget(namespace, "go fmt check")

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

// CheckVendor checks if vendor is in sync with go.mod.
// The approach is:
// - Delete vendor dir
// - Run go mod vendor
// - Run git add so that any changes resulting from go vendor will be in git staging area
// - Run git diff to find changes
// - If there are change we print them, unstage them and exit with 1.
func (Go) CheckVendor() error {
	sh.PrintStartTarget(namespace, "checkVendor")

	tmpFile, err := os.CreateTemp("", "verify-vendor-*.sh")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	cmd := `
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

tmp="$(mktemp -d)"
go mod vendor -o "$tmp"
if ! _out="$(diff -Naupr vendor $tmp)"; then
  echo "Verify vendor failed" >&2
  echo "If you're seeing this locally, run the below command to fix your go.mod:" >&2
  echo "go mod vendor" >&2
  exit 1
fi
`

	content := []byte(cmd)
	if _, err := tmpFile.Write(content); err != nil {
		return err
	}

	if err := tmpFile.Chmod(0o700); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// 	cmd := `rm -rf vendor && go mod vendor && git add vendor && \
	// git diff --cached --quiet -- vendor || \
	// (git --no-pager diff --cached -- vendor && git reset vendor && exit 1)`

	return sh.RunV("bash", "-c", tmpFile.Name())
}

const goCmd = "go"

func getAllGoFiles(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, _ os.FileInfo, err error) error {
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
