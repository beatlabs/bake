// Package golang contains go code related mage targets.
package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake/internal/shfmt"
)

// Go groups together go related tasks.
type Go mg.Namespace

const namespace = "code"

// ModSync runs go module tidy and vendor.
func (Go) ModSync() error {
	shfmt.PrintStartTarget(namespace, "go mod sync")

	if err := shfmt.RunV(goCmd, "mod", "tidy"); err != nil {
		return err
	}
	return shfmt.RunV(goCmd, "mod", "vendor")
}

// Fmt runs go fmt.
func (Go) Fmt() error {
	shfmt.PrintStartTarget(namespace, "go fmt")

	return shfmt.RunV(goCmd, "fmt", "./...")
}

// FmtCheck checks if all files are formatted.
func (Go) FmtCheck() error {
	shfmt.PrintStartTarget(namespace, "go fmt check")

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
// - If there are change we print them, unstage them and exit with 1
func (Go) CheckVendor() error {
	shfmt.PrintStartTarget(namespace, "checkVendor")

	cmd := `rm -rf vendor && go mod vendor && git add vendor && \
git diff --cached --quiet -- vendor || \
(git --no-pager diff --cached -- vendor && git reset vendor && exit 1)`

	return shfmt.RunV("bash", "-c", cmd)
}

const goCmd = "go"

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
