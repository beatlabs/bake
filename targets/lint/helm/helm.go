// Package helm contains linting related mage targets.
package helm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Lint groups together lint related tasks.
type Lint mg.Namespace

const cmd = "helm"

var (
	// HelmChartPath is the path to the helm chart to lint.
	HelmChartPath = "infra/deploy/helm/replace_me/Chart.yaml"
	// HelmRepos is a map of repos.
	HelmRepos = map[string]string{
		"beat":      "https://chartmuseum.private.k8s.management.thebeat.co/",
		"stable":    "https://kubernetes-charts.storage.googleapis.com",
		"incubator": "https://kubernetes-charts-incubator.storage.googleapis.com",
		"bitnami":   "https://charts.bitnami.com/bitnami",
		"aws":       "https://aws.github.io/eks-charts",
	}
)

// Helm linting of a specific chart path.
func (l Lint) Helm() error {
	err := helmAddRepos(HelmRepos)
	if err != nil {
		return err
	}

	fmt.Printf("lint: running helm dependency build for chart path: %s\n", HelmChartPath)
	err = sh.RunV(cmd, "dependency", "build", HelmChartPath)
	if err != nil {
		return err
	}

	err = helmCreateTemplateIfNotExists(HelmChartPath)
	if err != nil {
		return err
	}

	fmt.Printf("lint: running helm lint for chart path: %s\n", HelmChartPath)
	return sh.RunV(cmd, "lint", "--strict", HelmChartPath)
}

func helmAddRepos(repos map[string]string) error {
	for key, value := range repos {
		fmt.Printf("lint: running helm add repo %s for registry: %s\n", key, value)
		err := sh.RunV(cmd, "repo", "add", key, value)
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
