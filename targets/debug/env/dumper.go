package env

import (
	"bytes"
	"fmt"
	"os"
)

// Dumper dumps envs
type Dumper interface {
	Dump(envs map[string]string) error
}

// StdoutDumper simple stdout dumper
type StdoutDumper struct{}

// NewStdoutDumper creates stdout dumper
func NewStdoutDumper() *StdoutDumper {
	return &StdoutDumper{}
}

// Dump envs to stdout
func (d StdoutDumper) Dump(envs map[string]string) error {
	for key, val := range envs {
		fmt.Printf("%s=%s\n", key, val)
	}
	return nil
}

// FileDumper simple file dumper
type FileDumper struct {
	filename string
}

// NewFileDumper creates a file dumper
func NewFileDumper(filename string) *FileDumper {
	return &FileDumper{filename: filename}
}

// Dump envs to configured file
func (d FileDumper) Dump(envs map[string]string) error {
	var buf bytes.Buffer
	for key, val := range envs {
		buf.WriteString(fmt.Sprintf("%s=%s\n", key, val))
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

	_, err = f.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", d.filename, err)
	}

	return nil
}
