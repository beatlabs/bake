// Package swagger contains swagger/openapi related mage targets.
package swagger

import (
	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/swagger"
)

// Swagger groups together swagger/openapi related tasks.
type Swagger mg.Namespace

// Create creates a swagger files from source code annotations.
func (s Swagger) Create() error {
	return swagger.CreateDefault()
}

// Check ensures that the generated files are up to date.
func (Swagger) Check() error {
	return swagger.CheckDefault()
}
