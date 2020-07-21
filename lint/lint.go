// Package lint contains linting related helpers to be used in mage targets.
package lint

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

var defaultLinters = []string{
	"govet",
	"golint",
	"gofumpt",
	"gosec",
	"unparam",
	"goconst",
	"prealloc",
	"stylecheck",
	"unconvert",
}

const (
	helmCmd               = "helm"
	beatHelmRegistry      = "https://chartmuseum.private.k8s.management.thebeat.co/"
	stableHelmRegistry    = "https://kubernetes-charts.storage.googleapis.com"
	incubatorHelmRegistry = "https://kubernetes-charts-incubator.storage.googleapis.com"
	bitnamiHelmRegistry   = "https://charts.bitnami.com/bitnami"
	awsHelmRegistry       = "https://aws.github.io/eks-charts"
)

// Helm linting of a specific chart path.
func Helm(path string) error {
	repos := map[string]string{
		"beat":      beatHelmRegistry,
		"stable":    stableHelmRegistry,
		"incubator": incubatorHelmRegistry,
		"bitnami":   bitnamiHelmRegistry,
		"aws":       awsHelmRegistry,
	}

	err := helmAddRepo(repos)
	if err != nil {
		return err
	}

	fmt.Printf("lint: running helm dependency build for chart path: %s\n", path)
	err = sh.RunV(helmCmd, "dependency", "build", path)
	if err != nil {
		return err
	}

	err = helmCreateTemplateIfNotExists(path)
	if err != nil {
		return err
	}

	fmt.Printf("lint: running helm lint for chart path: %s\n", path)
	return sh.RunV(helmCmd, "lint", "--strict", path)
}

func helmAddRepo(repos map[string]string) error {
	for key, value := range repos {
		fmt.Printf("lint: running helm add repo %s for registry: %s\n", key, value)
		err := sh.RunV(helmCmd, "repo", "add", key, value)
		if err != nil {
			return fmt.Errorf("failed to add helm repo %s %s: %w", key, value, err)
		}
	}
	return nil
}

func helmCreateTemplateIfNotExists(path string) error {
	templatePath := path + "templates"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("lint: creating helm chart template path: %s\n", templatePath)
		err = os.Mkdir(templatePath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create helm chart template path %s: %w", templatePath, err)
		}
	}
	return nil
}

// Docker lints the docker file.
func Docker(dockerFile string) error {
	fmt.Printf("lint: running docker lint for file: %s\n", dockerFile)
	return sh.RunV("hadolint", dockerFile)
}

// Go lints the Go code and accepts build tags.
func Go(tags []string) error {
	return code(nil, tags)
}

// GoLinters lints the Go code and accepts linters and build tags.
func GoLinters(linters, tags []string) error {
	return code(linters, tags)
}

// GoDefault lints the Go code and uses default linters and build tags.
func GoDefault() error {
	return code(defaultLinters, []string{bake.BuildTagIntegration, bake.BuildTagComponent})
}

func code(linters, tags []string) error {
	fmt.Printf("lint: running go lint. linters: %v tags: %v\n", linters, tags)

	buildTagFlag := ""
	if len(tags) > 0 {
		buildTagFlag = getBuildTagFlag(tags)
	}

	linterFlag := ""
	if len(linters) > 0 {
		linterFlag = getLinterFlag(linters)
	}

	cmd := "golangci-lint"
	args := strings.Split(fmt.Sprintf("run %s %s --exclude-use-default=false --deadline=5m --modules-download-mode=vendor", linterFlag, buildTagFlag), " ")

	fmt.Printf("Executing cmd: %s %s\n", cmd, strings.Join(args, " "))

	return sh.RunV(cmd, args...)
}

func getBuildTagFlag(tags []string) string {
	return "--build-tags=" + strings.Join(tags, ",")
}

func getLinterFlag(linters []string) string {
	return "--enable " + strings.Join(linters, ",")
}
