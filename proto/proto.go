// Package proto contains helpers for handling proto schemas.
package proto

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

const (
	skimCMD                  = "skim"
	defaultOwner             = "taxibeat"
	defaultRegistry          = "proto-schemas"
	defaultSchemasLocation   = "proto/schemas"
	defaultGeneratedLocation = "proto/generated"
)

// SchemaValidateAll lints the schemas in the repository against the GitHub schemas.
func SchemaValidateAll(service string) error {
	fmt.Printf("proto schema: validate all schemas for %s\n", service)

	args := []string{
		"-t",
		os.Getenv("GITHUB_TOKEN"),
		"-r",
		defaultRegistry,
		"-o",
		defaultOwner,
		"-n",
		service,
		"validate-all",
		"-s",
		defaultSchemasLocation,
	}

	return sh.RunV(skimCMD, args...)
}

func SchemaGenerate(service, schema, version string) error {
	pathToSchema := fmt.Sprintf("%s/%s/%s.proto", schema, version, schema)
	fmt.Printf("proto schema: generate schema %s\n", pathToSchema)

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir: %s", err)
	}

	args := []string{
		"-r",
		defaultRegistry,
		"-o",
		defaultOwner,
		"-n",
		service,
		"generate",
		"-s",
		defaultSchemasLocation,
		"-out",
		tmpDir,
		"--schema",
		pathToSchema,
	}

	err = sh.RunV(skimCMD, args...)
	if err != nil {
		return err
	}

	generatedFilePath := fmt.Sprintf("%s/go/%s/%s/%s.pb.go", tmpDir, service, schema, schema)
	_, err = os.Open(generatedFilePath)
	if err != nil {
		return fmt.Errorf("failed to open expected generated file: %s", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working dir: %v", wd)
	}

	outDir := fmt.Sprintf("%s/%s/%s", wd, defaultGeneratedLocation, schema)
	err = os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create out dir: %v", err)
	}

	outFile := fmt.Sprintf("%s/%s.pb.go", outDir, schema)
	err = os.Rename(generatedFilePath, outFile)
	if err != nil {
		return fmt.Errorf("failed to move generated file: %v", err)
	}

	return nil
}

func SchemaGenerateAll(service string) error {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir: %s", err)
	}

	args := []string{
		"-r",
		defaultRegistry,
		"-o",
		defaultOwner,
		"-n",
		service,
		"generate-all",
		"-s",
		defaultSchemasLocation,
		"-out",
		tmpDir,
	}

	err = sh.RunV(skimCMD, args...)
	if err != nil {
		return err
	}

	generatedDirPath := fmt.Sprintf("%s/go/%s", tmpDir, service)
	_, err = os.Open(generatedDirPath)
	if err != nil {
		return fmt.Errorf("failed to open expected generated dir: %s", err)
	}

	var schemas []string
	filepath.Walk(generatedDirPath, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.IsDir() {
			return nil
		}
		schema := strings.Split(fInfo.Name(), ".")[0]
		schemas = append(schemas, schema)
		return nil
	})

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working dir: %v", wd)
	}

	for _, schema := range schemas {
		outDir := fmt.Sprintf("%s/%s/%s", wd, defaultGeneratedLocation, schema)
		err = os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create out dir: %v", err)
		}

		outFile := fmt.Sprintf("%s/%s.pb.go", outDir, schema)
		err = os.Rename(fmt.Sprintf("%s/%s/%s.pb.go", generatedDirPath, schema, schema), outFile)
		if err != nil {
			return fmt.Errorf("failed to move generated file: %v", err)
		}
	}

	return nil
}
