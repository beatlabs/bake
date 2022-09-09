// Package session contains targets which helps to work with bake session
package session

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/taxibeat/bake/internal/sh"

	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/docker"
	"github.com/taxibeat/bake/docker/env"
)

var (
	// BakeSessionLocation where current bake session file is located
	BakeSessionLocation = "test/.bakesession"
	// ServiceName name of service under test
	// use name as it appears in the .bakesession file
	ServiceName = ""
	// ExtraRules allows to add extra replacement rules
	ExtraRules = env.ReplacementRuleList{}
	// OutputFileLocation where to dump output envs
	OutputFileLocation = ".env.localhost"
)

// Session groups together interactions with bake session services
type Session mg.Namespace

const namespace = "session"

// DumpEnv outputs envs variables from the service.
// It Replaces docker hosts with corresponding localhost endpoints
// substituting docker-to-docker addresses to host-to-docker ones.
// Extra replacement rules can be added.
func (Session) DumpEnv() error {
	sh.PrintStartTarget(namespace, "dump env")

	if ServiceName == "" {
		return errors.New("please set session.ServiceName in your magefile")
	}

	// load current bake session if any, otherwise fail.
	session, err := docker.LoadSessionFromFile(docker.InDocker(), BakeSessionLocation)
	if err != nil {
		return fmt.Errorf("bake session is not found in %s", BakeSessionLocation)
	}
	containerName, err := env.BuildContainerName(session, ServiceName)
	if err != nil {
		return err
	}
	// load env variables from existing service.
	envs, err := env.GetServiceEnvs(session, ServiceName, ExtraRules)
	if err != nil {
		return fmt.Errorf("failed to fetch env from service %s: %w", ServiceName, err)
	}

	err = dumpToFile(envs, OutputFileLocation)
	if err != nil {
		return fmt.Errorf("failed to write envs to file %s: %w", OutputFileLocation, err)
	}
	fmt.Printf("environment variables from %s are dumped to file '%s'\n", ServiceName, OutputFileLocation)
	fmt.Printf("please stop service %s (use command below):\n", ServiceName)
	fmt.Printf("docker stop %s\n", containerName)
	fmt.Println("now you can run service on localhost")
	return nil
}

// dumpToFile envs to a file
func dumpToFile(envs map[string]string, filename string) error {
	lines := make([]string, 0, len(envs))
	for key, val := range envs {
		lines = append(lines, fmt.Sprintf("%s=%s", key, val))
	}
	sort.Strings(lines)
	content := strings.Join(lines, "\n")

	f, err := os.Create(filename) // nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			fmt.Printf("failed to close file %s: %v", filename, err)
		}
	}()

	_, err = f.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	return nil
}
