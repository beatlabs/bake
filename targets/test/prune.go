package test

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func pruneCoverageFile(path string, excludePkgs []string) error {
	if len(excludePkgs) == 0 {
		return nil
	}

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close: %v\n", err)
		}
	}()

	lines, err := pruneCoverageLines(f, excludePkgs)
	if err != nil {
		return err
	}

	f, err = os.Create(path) // nolint:gosec
	if err != nil {
		return err
	}

	for _, line := range lines {
		_, err := fmt.Fprintln(f, line)
		if err != nil {
			return err
		}
	}

	return nil
}

func pruneCoverageLines(input io.Reader, excludePkgs []string) ([]string, error) {
	re, err := regexp.Compile(strings.Join(excludePkgs, "|"))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(input)
	buf := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if !re.MatchString(line) {
			buf = append(buf, line)
		}
	}
	return buf, nil
}
