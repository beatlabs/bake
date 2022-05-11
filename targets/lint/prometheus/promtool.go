// Package prometheus contains api to prometheus util called promtool
package prometheus

import (
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

// AlertRules check if the rule files are valid or not using promtool
// @see https://prometheus.io/docs/prometheus/latest/configuration/unit_testing_rules/
func (l Lint) AlertRules() error {
	args := []string{
		"check",
		"rules",
	}
	for _, alertFile := range AlertFiles {
		args = append(args, filepath.Join(alertingDir, alertFile))
	}

	return sh.RunV(cmd, args...)
}
