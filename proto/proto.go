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

// SchemaGenerate generates a single schema for a specific version.
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

	generatedFiles, err := getGeneratedFiles(tmpDir)
	if err != nil {
		return err
	}
	return moveGeneratedFiles(generatedFiles)
}

// SchemaGenerateAll generates all the schemas found.
func SchemaGenerateAll(service string) error {
	fmt.Printf("proto schema: generate all schemas for %s\n", service)

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

	generatedFiles, err := getGeneratedFiles(tmpDir)
	if err != nil {
		return err
	}
	return moveGeneratedFiles(generatedFiles)
}

func getGeneratedFiles(tmpDir string) ([]string, error) {
	var generatedFiles []string
	err := filepath.Walk(tmpDir, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		generatedFiles = append(generatedFiles, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list generated files: %v", err)
	}
	return generatedFiles, nil
}

func moveGeneratedFiles(generatedFiles []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working dir: %v", wd)
	}

	for _, generatedFile := range generatedFiles {
		fileName := filepath.Base(generatedFile)
		schemaName := strings.Split(fileName, ".")[0]
		outDir := fmt.Sprintf("%s/%s/%s", wd, defaultGeneratedLocation, schemaName)
		err := os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create out dir: %v", err)
		}

		outFile := fmt.Sprintf("%s/%s", outDir, fileName)
		err = os.Rename(generatedFile, outFile)
		if err != nil {
			return fmt.Errorf("failed to move generated file: %v", err)
		}
		fmt.Printf("schema generated successfully under %s\n", outFile)
	}
	return nil
}
