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

	args := []string{
		"-t",
		os.Getenv(bake.GitHubTokenEnvVar),
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
