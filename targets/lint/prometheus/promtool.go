// Package prometheus contains api to prometheus util called promtool
// @see https://prometheus.io/docs/prometheus/latest/configuration/unit_testing_rules/
package prometheus

import (
	"errors"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Lint groups together lint related tasks.
type Lint mg.Namespace

const (
	cmd         = "promtool"
	alertingDir = "infra/observe/alerting"
)

// AlertFiles list of alert files to be checked by promtool
var AlertFiles []string

// AlertRules check if the prometheus alert rules are valid or not.
func (l Lint) AlertRules() error {
	if len(AlertFiles) == 0 {
		return errors.New("prometheus.AlertFiles variable must be filled in mage file")
	}

	args := []string{
		"check",
		"rules",
	}
	for _, alertFile := range AlertFiles {
		args = append(args, filepath.Join(alertingDir, alertFile))
	}

	return sh.RunV(cmd, args...)
}
