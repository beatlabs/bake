// Package sh implements shell output formatting utilities
// that decorate the execution of targets and commands with useful messages.
package sh

import (
	"fmt"
	"io"
	"strings"

	"github.com/magefile/mage/sh"
)

// PrintStartTarget prints a message to indicate the start of execution of the specified namespace and target.
func PrintStartTarget(namespace, target string) {
	fmt.Println("###################################")
	fmt.Printf(" Executing target %s:%s\n", namespace, target)
	fmt.Println("###################################")
	fmt.Println()
}

// RunV decorates mage's `sh.RunV` with a printed message of the command to be executed.
func RunV(cmd string, args ...string) error {
	quotedArgs := quote(args)
	fmt.Printf(">> Running Command: `%s %s`\n\n", cmd, strings.Join(quotedArgs, " "))

	err := sh.RunV(cmd, args...)
	if err == nil {
		fmt.Printf("Command finished successfully.\n\n")
	}
	return err
}

// RunWithV decorates mage's `sh.RunWithV` with a printed message of the command to be executed.
func RunWithV(env map[string]string, cmd string, args ...string) error {
	quotedArgs := quote(args)
	fmt.Printf(">> Running Command: `%s %s`\n\n", cmd, strings.Join(quotedArgs, " "))

	err := sh.RunWithV(env, cmd, args...)
	if err == nil {
		fmt.Printf("Command finished successfully.\n\n")
	}
	return err
}

// Output is a wrapper of mage's `sh.Output` and exists in this package for grouping purposes.
func Output(cmd string, args ...string) (string, error) {
	return sh.Output(cmd, args...)
}

// Exec is a wrapper of mage's `sh.Exec` and exists in this package for grouping purposes.
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (bool, error) {
	return sh.Exec(env, stdout, stderr, cmd, args...)
}

// Run is a wrapper of mage's `sh.Run` and exists in this package for grouping purposes.
func Run(cmd string, args ...string) error {
	return sh.Run(cmd, args...)
}

func quote(args []string) []string {
	quoted := []string{}
	for i := range args {
		quoted = append(quoted, fmt.Sprintf("%q", args[i]))
	}
	return quoted
}
