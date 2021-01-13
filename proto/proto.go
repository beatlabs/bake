// Package proto contains helpers for handling proto schemas.
package proto

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

const skimCMD = "skim"

// SchemaValidateAll lints the schemas in the repository against the GitHub schemas.
func SchemaValidateAll(service string) error {
	fmt.Printf("proto schema: validate all schemas for %s\n", service)

	token := os.Getenv(bake.GitHubTokenEnvVar)
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN env var not set")
	}

	args := []string{
		"-t",
		token,
		"-r",
		"proto-schemas",
		"-o",
		"taxibeat",
		"-n",
		service,
		"validate-all",
		"-s",
		"proto/schemas",
	}

	return sh.RunV(skimCMD, args...)
}
