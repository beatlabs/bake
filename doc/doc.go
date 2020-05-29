// Package doc contains documentation related helpers to be used in mage targets.
package doc

import (
	"errors"
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
)

type confluenceConfig struct {
	username string
	password string
	baseURL  string
}

// ConfluenceDoc definition of a document to be synced.
type ConfluenceDoc struct {
	Path string
	File string
}

// ConfluenceSync syncs the list of documents provided with Confluence.
func ConfluenceSync(docs ...ConfluenceDoc) error {
	var ok bool
	var cfg confluenceConfig

	cfg.username, ok = os.LookupEnv("CONFLUENCE_USERNAME")
	if !ok {
		return errors.New("env var CONFLUENCE_USERNAME is not set")
	}
	cfg.password, ok = os.LookupEnv("CONFLUENCE_PASSWORD")
	if !ok {
		return errors.New("env var CONFLUENCE_PASSWORD is not set")
	}
	cfg.baseURL, ok = os.LookupEnv("CONFLUENCE_BASEURL")
	if !ok {
		return errors.New("env var CONFLUENCE_BASEURL is not set")
	}

	current, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory")
	}

	for _, doc := range docs {
		err := confluenceSync(current, cfg, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func confluenceSync(currentPath string, cfg confluenceConfig, doc ConfluenceDoc) error {
	fmt.Printf("processing %s%s\n", doc.Path, doc.File)
	defer func() {
		err := os.Chdir(currentPath)
		if err != nil {
			fmt.Printf("failed to revert to previous working directory %s: %v", currentPath, err)
		}
		fmt.Printf("reverted back to working directory: %s\n", currentPath)
	}()
	if doc.Path != "" {
		if err := os.Chdir(doc.Path); err != nil {
			return fmt.Errorf("failed to change directory %s: %w", doc.Path, err)
		}
		fmt.Printf("changed to working directory: %s\n", doc.Path)
	}

	args := []string{"-u", cfg.username, "-p", cfg.password, "-b", cfg.baseURL, "-f", doc.File}

	return sh.RunV("mark", args...)
}
