// Package proto contains proto related mage targets.
package proto

import (
	"errors"

	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/proto"
)

// Proto groups together proto related tasks.
type Proto mg.Namespace

// SchemaGenerate generates a single proto schema with the following arguments: service, schema, version.
func (Proto) SchemaGenerate(service, schema, version string) error {
	if service == "" {
		return errors.New("service is mandatory")
	}
	if schema == "" {
		return errors.New("schema is mandatory")
	}
	if version == "" {
		return errors.New("version is mandatory")
	}
	return proto.SchemaGenerate(service, schema, version)
}

// SchemaGenerateAll generates all the schemas found with the following argument: service.
func (Proto) SchemaGenerateAll(service string) error {
	if service == "" {
		return errors.New("service is mandatory")
	}
	return proto.SchemaGenerateAll(service)
}
