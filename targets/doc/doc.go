// Package doc contains doc related mage targets.
package doc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type confluenceConfig struct {
	username string
	password string
	baseURL  string
}

type confluenceDoc struct {
	Path string
	File string
}

// Doc groups together doc related tasks.
type Doc mg.Namespace

// ConfluenceSync synchronized annotated docs to confluence.
func (Doc) ConfluenceSync() error {

	fmt.Print("doc: syncing docs with confluence\n")

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

	docs, err := getDocs(".")
	if err != nil {
		return err
	}

	for _, doc := range docs {
		err := confluenceSync(current, cfg, doc)
		if err != nil {
			return err
		}
	}

	return nil
}

func getDocs(root string) ([]confluenceDoc, error) {
	var docs []confluenceDoc
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match("*.md", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			if !strings.HasPrefix(string(content), `<!-- Space: DT -->`) {
				return nil
			}

			docs = append(docs, confluenceDoc{
				Path: filepath.Dir(path),
				File: info.Name(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func confluenceSync(currentPath string, cfg confluenceConfig, doc confluenceDoc) error {
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
