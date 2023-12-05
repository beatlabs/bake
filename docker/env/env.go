// Package env contains helpers to export a service's env vars with some substitution in order to replace docker-to-docker to host-to-docker addresses.
package env

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/beatlabs/bake/docker"
	"github.com/beatlabs/bake/internal/sh"
)

const (
	inspectEnvFormat = "{{range $index, $new := .Config.Env}}{{$new}}{{println}}{{end}}"
	dockerCmd        = "docker"
)

var (
	skipEnvSet = map[string]bool{
		"PATH":     true,
		"HOME":     true,
		"HOSTNAME": true,
	}

	inspectEnvArgs = []string{
		"inspect",
		"-f",
		inspectEnvFormat,
	}

	// inspectFormatEnv in order to make format pass correctly without replacing $ variables
	inspectFormatEnv = map[string]string{
		"new":   "$new",
		"index": "$index",
	}
)

// GetServiceEnvs inspects docker container envs from given service
// If service does not exist (docker container should exist, it can be stopped), then it fails because there is no service to debug.
func GetServiceEnvs(session *docker.Session, serviceName string, extraRules ReplacementRuleList) (map[string]string, error) {
	containerName, err := BuildContainerName(session, serviceName)
	if err != nil {
		return nil, err
	}

	env := inspectFormatEnv
	args := append(inspectEnvArgs, containerName) // nolint:gocritic
	cmdOut, err := run(env, args)
	if err != nil {
		return nil, err
	}

	envsRaw := strings.Split(cmdOut.String(), "\n")
	envs := make(map[string]string)
	for _, envRaw := range envsRaw {
		envData := strings.SplitN(envRaw, "=", 2)
		if len(envData) == 2 {
			if skip := skipEnvSet[envData[0]]; !skip {
				envs[envData[0]] = envData[1]
			}
		}
	}

	replacementRulesList, err := newReplacementRulesList(session, serviceName)
	if err != nil {
		return nil, fmt.Errorf("could not create replacement rules: %w", err)
	}

	envs = replacementRulesList.
		Merge(extraRules).
		Replace(envs)

	return envs, nil
}

// BuildContainerName from session id and service name.
// fails if service is not registered in bake session
func BuildContainerName(session *docker.Session, serviceName string) (string, error) {
	_, err := session.AutoServiceAddress(serviceName)
	if err != nil {
		return "", fmt.Errorf("service with name %s is not found", serviceName)
	}

	return fmt.Sprintf("%s-%s", session.ID(), serviceName), nil
}

func run(env map[string]string, args []string) (bytes.Buffer, error) {
	var cmdOut, cmdErr bytes.Buffer

	_, err := sh.Exec(env, &cmdOut, &cmdErr, dockerCmd, args...)
	if err != nil {
		return cmdOut, err
	}
	if cmdErr.String() != "" {
		return cmdOut, errors.New(cmdErr.String())
	}
	return cmdOut, nil
}
