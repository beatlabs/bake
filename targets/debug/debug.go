// Package debug contains handy commands for debugging tests while bake session is running
// the most useful case is debugging component or integration tests
package debug

import (
	"errors"
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/taxibeat/bake/docker"
	"github.com/taxibeat/bake/targets/debug/env"
)

type output string

const (
	stdout output = "stdout"
	file   output = "file"
	// envFilename where envs will be dumped from service
	envFilename = ".env.localhost"
)

var (
	// BakeSessionLocation where current bake session file is located
	BakeSessionLocation = "test/.bakesession"
	// ServiceName name of service under test
	// use name as it appears in the .bakesession file
	ServiceName = ""
	// ExtraRules allows to add extra replacement rules
	ExtraRules = env.ReplacementRuleList{}
)

// Debug groups together debugging tests.
type Debug mg.Namespace

// Env outputs envs variables from given service, replaces docker hosts with corresponding localhost endpoints
// serviceName is the name of service to output env variables from.
// output is where to dump loaded envs, there are two options: stdout, file
func (Debug) Env(output string) error {
	// load current bake session if any, otherwise fail.
	session, err := docker.LoadSessionFromFile(docker.InDocker(), BakeSessionLocation)
	if err != nil {
		return errors.New("bake session is not running")
	}
	// load env variables from existing service.
	envs, err := env.GetServiceEnvs(session, ServiceName, ExtraRules)
	if err != nil {
		return fmt.Errorf("failed to fetch env from service %s: %w", ServiceName, err)
	}

	var dumper env.Dumper
	switch output {
	case string(stdout):
		dumper = env.NewStdoutDumper(os.Stdout)
		break
	case string(file):
		dumper, _ = env.NewFileDumper(envFilename)
	}

	if dumper == nil {
		return fmt.Errorf("unknown output: %s", output)
	}

	err = dumper.Dump(envs)
	if err != nil {
		return fmt.Errorf("failed to write envs to output %s: %w", output, err)
	}
	fmt.Println("envs dumped successfully")
	return nil
}
