// Package lint contains linting related helpers to be used in mage targets.
package lint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	err = runHelm("dependency", "build", path)
	if err != nil {
		return err
	}

	err = helmCreateTemplateIfNotExists(path)
	if err != nil {
		return err
	}

	fmt.Printf("lint: running helm lint for chart path: %s\n", path)
	return runHelm("lint", "--strict", path)
}

func helmAddRepo(repos map[string]string) error {
	for key, value := range repos {
		fmt.Printf("lint: running helm add repo %s for registry: %s\n", key, value)
		err := runHelm("repo", "add", key, value)
		if err != nil {
			return fmt.Errorf("failed to add helm repo %s %s: %w", key, value, err)
		}
	}
	return nil
}

func helmCreateTemplateIfNotExists(path string) error {
	templatePath := filepath.Join(path, "templates")
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
	wd, _ := os.Getwd()
	return sh.RunV("docker", "run", "--rm", "-v", wd+":/app", "-w", "/app", "hadolint/hadolint", "hadolint", dockerFile)
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

	cmd := "docker"
	wd, _ := os.Getwd()
	args := strings.Split(fmt.Sprintf("run --env=GOFLAGS=-mod=vendor --rm --volume %s:/app -w /app golangci/golangci-lint:v1.28.1 golangci-lint run %s %s --exclude-use-default=false --deadline=5m --modules-download-mode=vendor", wd, linterFlag, buildTagFlag), " ")

	fmt.Printf("Executing cmd: %s %s\n", cmd, strings.Join(args, " "))
	return sh.RunV(cmd, args...)
}

func getBuildTagFlag(tags []string) string {
	return "--build-tags=" + strings.Join(tags, ",")
}

func getLinterFlag(linters []string) string {
	return "--enable " + strings.Join(linters, ",")
}

func runHelm(args ...string) error {
	cmd := "helm"

	_, err := exec.LookPath(cmd)
	if err != nil {
		cmd = "docker"

		// todo: set this on init?
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %v\n", err)
		}

		// todo: extract dockerArgs?
		dockerArgs := []string{"run", "--rm", "--volume", wd + ":/volume", "--workdir", "/volume", "alpine/helm", "helm"}
		args = append(dockerArgs, args...)

	}

	fmt.Printf("Executing cmd: %s %s\n", cmd, strings.Join(args, " "))
	return sh.RunV(cmd, args...)
}
