package env

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Dumper dumps envs
type Dumper interface {
	Dump(envs map[string]string) error
}

// StdoutDumper simple stdout dumper
type StdoutDumper struct {
	writer io.Writer
}

// NewStdoutDumper creates new stdout dumper
func NewStdoutDumper(writer io.Writer) *StdoutDumper {
	return &StdoutDumper{writer: writer}
}

// Dump envs to stdout
func (d StdoutDumper) Dump(envs map[string]string) error {
	for key, val := range envs {
		_, err := fmt.Fprintf(d.writer, "%s=%s\n", key, val)
		if err != nil {
			return err
		}
	}
	return nil
}

// FileDumper simple file dumper
type FileDumper struct {
	filename string
}

// NewFileDumper creates a file dumper
func NewFileDumper(filename string) (*FileDumper, error) {
	if filename == "" {
		return nil, errors.New("filename must be provided")
	}
	return &FileDumper{filename: filename}, nil
}

// Dump envs to configured file
func (d FileDumper) Dump(envs map[string]string) error {
	var content string
	for key, val := range envs {
		content += fmt.Sprintf("%s=%s\n", key, val)
	}

	f, err := os.Create(d.filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", d.filename, err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			fmt.Printf("failed to close file %s: %v", d.filename, err)
		}
	}()

	_, err = f.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", d.filename, err)
	}

	return nil
}
