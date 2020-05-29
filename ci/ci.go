// Package ci contains all the ci tasks that a project should have.
package ci

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake"
)

const (
	coverFile = "coverage.txt"
)

// CodeCov runs the actual CI pipeline with CodeCov integration and accepts build tags.
func CodeCov(buildTags ...string) error {
	return codeCov(buildTags)
}

// CodeCovDefault runs the actual CI pipeline with CodeCov integration and default build tags.
func CodeCovDefault() error {
	return codeCov([]string{bake.BuildTagIntegration, bake.BuildTagComponent})
}

func codeCov(tags []string) error {

	fmt.Printf("ci: running codecov with tags: %v\n", tags)

	args := []string{
		"test",
		"-mod=vendor",
		"-count=1",
		"-cover",
		"-coverprofile=" + coverFile,
		"-covermode=atomic",
		"-timeout=60s",
	}

	if len(tags) > 0 {
		args = append(args, getBuildTagFlag(tags))
	}

	args = append(args, "./...")

	defer func() {
		err := os.Remove(coverFile)
		if err != nil {
			fmt.Printf("failed to delete coverage file: %v\n", err)
		}
	}()

	err := sh.RunV(bake.GoCmd, args...)
	if err != nil {
		return err
	}

	const codecovFile = "codecov.sh"

	err = downloadFile(codecovFile, "https://codecov.io/bash")
	if err != nil {
		return fmt.Errorf("failed to download codecov bash script: %w", err)
	}
	defer func() {
		err := os.Remove(codecovFile)
		if err != nil {
			fmt.Printf("failed to delete coverage file: %v\n", err)
		}
	}()

	token := os.Getenv("CODECOV_TOKEN")

	args = []string{"./" + codecovFile, "-t", token}

	return sh.RunV("bash", args...)
}

func getBuildTagFlag(tags []string) string {
	return "-tags=" + strings.Join(tags, ",")
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("failed to close codecov response body")
		}
	}()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		err := out.Close()
		if err != nil {
			fmt.Println("failed to close codecov file")
		}
	}()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
