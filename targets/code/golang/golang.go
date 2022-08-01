// Package golang contains go code related mage targets.
package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	upgradeBranchName = "go-deps-update"
	gitCmd            = "git"
	gitRemoteName     = "origin"
)

// Go groups together go related tasks.
type Go mg.Namespace

// ModSync runs go module tidy and vendor.
func (Go) ModSync() error {
	fmt.Print("code: running go mod sync\n")

	if err := sh.RunV(goCmd, "mod", "tidy"); err != nil {
		return err
	}
	return sh.RunV(goCmd, "mod", "vendor")
}

// Fmt runs go fmt.
func (Go) Fmt() error {
	fmt.Print("code: running go fmt\n")

	return sh.RunV(goCmd, "fmt", "./...")
}

// FmtCheck checks if all files are formatted.
func (Go) FmtCheck() error {
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

// ModUpgrade upgrades all dependencies.
func (Go) ModUpgrade() error {
	fmt.Print("code: running go get -u all\n")

	return sh.RunV(goCmd, "get", "-u", "all")
}

// ModUpgradePR upgrade all dependencies, tidy them up, vendor them and create a PR.
func (g Go) ModUpgradePR() error {
	if err := g.ModUpgrade(); err != nil {
		return err
	}

	// Check if changes exist, if not exit
	err := sh.RunV(gitCmd, "diff", "--exit-code")
	if err == nil || sh.ExitStatus(err) == 0 {
		fmt.Println("no upgrades detected, exiting")
		return nil
	}
	fmt.Println("upgrades detected, continue")

	if err := g.ModSync(); err != nil {
		return err
	}

	// Checkout a new branch
	if err := sh.RunV(gitCmd, "checkout", "-b", upgradeBranchName); err != nil {
		return err
	}

	// Stage all changes
	if err := sh.RunV(gitCmd, "add", "."); err != nil {
		return err
	}

	// Commit to local branch
	if err := sh.RunV(gitCmd, "commit", "-m", "Go dependencies update"); err != nil {
		return err
	}

	// Push local branch to remote
	if err := sh.RunV(gitCmd, "push", "--set-upstream", gitRemoteName, upgradeBranchName); err != nil {
		return err
	}

	// Create a PR in GitHub
	if err := sh.RunV("gh", "pr", "create", "-t", "Go dependencies", "--body", "Go dependencies"); err != nil {
		return err
	}

	return nil
}

// CheckVendor checks if vendor is in sync with go.mod.
// The approach is:
// - Delete vendor dir
// - Run go mod vendor
// - Run git add so that any changes resulting from go vendor will be in git staging area
// - Run git diff to find changes
// - If there are change we print them, unstage them and exit with 1
func (Go) CheckVendor() error {
	cmd := `rm -rf vendor && go mod vendor && git add vendor && \
git diff --cached --quiet -- vendor || \
(git --no-pager diff --cached -- vendor && git reset vendor && exit 1)`

	return sh.RunV("bash", "-c", cmd)
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
