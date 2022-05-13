// Package prometheus contains api to prometheus util called promtool
// @see https://prometheus.io/docs/prometheus/latest/configuration/unit_testing_rules/
package prometheus

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Prometheus groups together lint related tasks.
type Prometheus mg.Namespace

const (
	cmd = "promtool"
)

var (
	// AlertsDir directory where to find alert files to be checked by promtool
	AlertsDir = "./infra/observe/alerting"
	// TestsDir directory where alerts tests are located if any, leave empty if no tests.
	TestsDir = ""
)

// Lint checks if the prometheus alert rules are valid or not.
func (p Prometheus) Lint() error {
	alertFiles, err := loadFiles(AlertsDir)
	if err != nil {
		return err
	}
	args := []string{"check", "rules"}
	args = append(args, alertFiles...)

	return sh.RunV(cmd, args...)
}

// Test run tests on prometheus alerts.
func (p Prometheus) Test() error {
	if TestsDir == "" {
		return errors.New("please provide prometheus.TestsDir variable")
	}

	testFiles, err := loadFiles(TestsDir)
	if err != nil {
		return err
	}
	args := []string{"test", "rules"}
	args = append(args, testFiles...)

	return sh.RunV(cmd, args...)
}

func loadFiles(dir string) ([]string, error) {
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, item := range items {
		if !item.IsDir() {
			files = append(files, filepath.Join(dir, item.Name()))
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("directory %s does not have files", dir)
	}
	return files, nil
}
