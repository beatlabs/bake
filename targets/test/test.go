// Package test contains test related mage targets.
package test

import (
	"strings"

	"github.com/beatlabs/bake/docker"
	"github.com/beatlabs/bake/internal/sh"
	"github.com/magefile/mage/mg"
)

const (
	goCmd              = "go"
	componentTestTag   = "component"
	integrationTestTag = "integration"
	namespace          = "test"
)

var (
	// GoBuildTags used when running all tests.
	GoBuildTags = []string{componentTestTag, integrationTestTag}

	// TestArgs used in test targets.
	TestArgs = []string{
		"test",
		"-mod=vendor",
		"-cover",
		"-race",
		"-shuffle=on",
	}
	// CoverArgs used in coverage targets.
	CoverArgs = []string{
		"test",
		"-mod=vendor",
		"-coverpkg=./...",
		"-covermode=atomic",
		"-coverprofile=coverage.txt",
		"-race",
		"-shuffle=on",
	}
	// Pkgs is the pkg pattern to target.
	Pkgs = "./..."
	// CoverExcludePatterns is a list of pkg patterns to prune from coverage.txt.
	CoverExcludePatterns = []string{}
	// CoverExcludeFile is the coverage file to prune.
	CoverExcludeFile = "coverage.txt"
)

// Test groups together test related tasks.
type Test mg.Namespace

// Unit runs unit tests.
func (Test) Unit() error {
	sh.PrintStartTarget(namespace, "unit")

	args := append(TestArgs, Pkgs) // nolint:gocritic
	return run(args)
}

// Integration runs unit and integration tests.
func (Test) Integration() error {
	sh.PrintStartTarget(namespace, "integration")

	args := append(appendCacheBustingArg(TestArgs), getBuildTagFlag([]string{integrationTestTag}), Pkgs)
	return run(args)
}

// Component runs unit and component tests.
func (Test) Component() error {
	sh.PrintStartTarget(namespace, "component")

	args := append(appendCacheBustingArg(TestArgs), getBuildTagFlag([]string{componentTestTag}), Pkgs)
	return run(args)
}

// All runs all tests.
func (Test) All() error {
	sh.PrintStartTarget(namespace, "all")

	args := append(appendCacheBustingArg(TestArgs), getBuildTagFlag(GoBuildTags), Pkgs)
	return run(args)
}

// CoverUnit runs unit tests and produces a coverage report.
func (Test) CoverUnit() error {
	sh.PrintStartTarget(namespace, "coverUnit")

	args := append(CoverArgs, Pkgs) // nolint:gocritic
	if err := run(args); err != nil {
		return err
	}
	return pruneCoverageFile(CoverExcludeFile, CoverExcludePatterns)
}

// CoverAll runs all tests and produces a coverage report.
func (Test) CoverAll() error {
	sh.PrintStartTarget(namespace, "coverAll")

	coveragePkgs, err := packagesExcluding("/docker/component")
	if err != nil {
		return err
	}

	coverageArgs := withoutArgs(CoverArgs, "-race")
	coverageArgs = append(coverageArgs, "-p=1", getBuildTagFlag([]string{integrationTestTag}))
	coverageArgs = append(coverageArgs, coveragePkgs...)
	if err := runWithEnv(coverAllEnv(), coverageArgs); err != nil {
		return err
	}

	return pruneCoverageFile(CoverExcludeFile, CoverExcludePatterns)
}

// Cleanup removes any local resources created by `mage test:all`.
func (Test) Cleanup() error {
	sh.PrintStartTarget(namespace, "cleanup")

	return docker.CleanupResources()
}

func run(args []string) error {
	return sh.RunV(goCmd, args...)
}

func runWithEnv(env map[string]string, args []string) error {
	return sh.RunWithV(env, goCmd, args...)
}

func coverAllEnv() map[string]string {
	return map[string]string{
		"GOGC":       "25",
		"GOMEMLIMIT": "1GiB",
	}
}

func packagesExcluding(excludePatterns ...string) ([]string, error) {
	output, err := sh.Output(goCmd, "list", "-mod=vendor", Pkgs)
	if err != nil {
		return nil, err
	}

	pkgs := make([]string, 0)
	for _, pkg := range strings.Fields(output) {
		if isExcludedPackage(pkg, excludePatterns) {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func isExcludedPackage(pkg string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if strings.Contains(pkg, pattern) {
			return true
		}
	}
	return false
}

func withoutArgs(args []string, excluded ...string) []string {
	exclusions := make(map[string]struct{}, len(excluded))
	for _, arg := range excluded {
		exclusions[arg] = struct{}{}
	}

	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if _, ok := exclusions[arg]; ok {
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered
}

func getBuildTagFlag(buildTags []string) string {
	return "-tags=" + strings.Join(buildTags, ",")
}

func appendCacheBustingArg(args []string) []string {
	return append(args, "-count=1")
}
