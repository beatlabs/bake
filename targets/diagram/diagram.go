// Package diagram contains support for compiling your
// [diagram](https://diagrams.mingrammer.com/) written in python to the
// associated png image
package diagram

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	python3CMD = "python3"
)

// InputDiagramPath is a list of files or directories where to search for
// diagrams source (*.py) files to be compiled into diagrams (*.png) files
// e.g.:
// InputDiagramPath = []string{"doc/architecture/arch.py", "doc/topology/servers.py"}
// InputDiagramPath = []string{"doc/architecture"}
var InputDiagramPath = []string{}

// Diagram groups together test related diagram tasks.
type Diagram mg.Namespace

// Generate will search for '*.py' files in the InputDiagramPath list and
// 'compile' them into the associated output diagram image files. The image
// output name and format (png, jpg, svg, and pdf) are defined within the input
// '*.py'. The output will be located in the same folder of the corresponding
// input.
func (Diagram) Generate() error {
	relativePaths := existingInputPythonFiles()

	tmpDir, err := ioutil.TempDir(".", "")
	if err != nil {
		return fmt.Errorf("failed to create tmp wd: %s", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Printf("failed to remove temp wd: %s\n", tmpDir)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %s", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			fmt.Printf("failed to navigate back to working directory: %s\n", wd)
		}
	}()

	absolutePaths := relativeToAbsolute(wd, relativePaths)

	if err := os.Chdir(tmpDir); err != nil {
		return fmt.Errorf("failed to navigate to tmp directory: %s", tmpDir)
	}

	for _, f := range absolutePaths {
		err := sh.RunV(python3CMD, f)
		if err != nil {
			return fmt.Errorf("abort generation of remaining diagrams, failed generating diagram for file [%s]: %v", f, err)
		}
		err = moveAllFilesInDirTo(path.Dir(f))
		if err != nil {
			return fmt.Errorf("abort generation of remaining diagrams, failed to move files: %v", err)
		}
	}

	return nil
}

func moveAllFilesInDirTo(destination string) error {
	matches, err := filepath.Glob("*")
	if err != nil {
		return fmt.Errorf("error while searching matches in current tmp folder: %v", err)
	}
	if len(matches) < 1 {
		return fmt.Errorf("failed to find matches in current tmp folder")
	}

	for _, match := range matches {
		fname := path.Base(match)
		err = os.Rename(match, path.Join(destination, fname))
		if err != nil {
			return fmt.Errorf("failed to move generated file from [%s] to [%s]: %v", match, path.Join(destination, fname), err)
		}
	}

	return nil
}

func relativeToAbsolute(base string, relativePaths []string) []string {
	absolutePaths := make([]string, len(relativePaths))
	for i, rp := range relativePaths {
		absolutePaths[i] = path.Join(base, rp)
	}
	return absolutePaths
}

func existingInputPythonFiles() []string {
	listOfExistingFiles := make([]string, 0)

	for _, inputPath := range InputDiagramPath {
		fi, err := os.Stat(inputPath)
		if os.IsNotExist(err) {
			fmt.Printf("input inputPath [%s] does not exist\n", inputPath)
			continue
		}

		if err != nil {
			fmt.Printf("error retrieving file info: %v\n", err)
			continue
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			// add to listOfExistingFiles all the *.py files in directory
			fmt.Printf("scraping [%s] directory for *.py files\n", inputPath)
			matches, err := filepath.Glob(path.Join(inputPath, "*.py"))
			if err != nil {
				fmt.Printf("error searching for *.py files: %v\n", err)
				continue
			}
			listOfExistingFiles = append(listOfExistingFiles, matches...)
		case mode.IsRegular():
			// add to listOfExistingFiles the file
			listOfExistingFiles = append(listOfExistingFiles, inputPath)
		}
	}

	return listOfExistingFiles
}
