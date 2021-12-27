// Package env helps output docker env state for localhost debugging
package env

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/taxibeat/bake/docker"
	"github.com/taxibeat/bake/docker/component/mongodb"
)

const (
	inspectEnvFormat = "{{range $index, $value := .Config.Env}}{{$value}}{{println}}{{end}}"
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
		"value": "$value",
		"index": "$index",
	}
)

// GetServiceEnvs inspects docker container envs from given service
// If service does not exist (docker container should exist, it can be stopped), then it fails because there is no service to debug.
func GetServiceEnvs(session *docker.Session, serviceName string) (map[string]string, error) {
	containerName, err := buildContainerName(session, serviceName)
	if err != nil {
		return nil, err
	}

	env := inspectFormatEnv
	args := append(inspectEnvArgs, containerName)
	cmdOut, err := run(env, args)
	if err != nil {
		return nil, err
	}

	envsRaw := strings.Split(cmdOut.String(), "\n")
	envs := make(map[string]string)
	for _, env := range envsRaw {
		envData := strings.SplitN(env, "=", 2)
		if len(envData) == 2 {
			if ok, _ := skipEnvSet[envData[0]]; !ok {
				envs[envData[0]] = envData[1]
			}
		}
	}

	serviceMap, err := buildServiceMap(session)
	if err != nil {
		return nil, fmt.Errorf("could not build service map: %w", err)
	}

	for i := range envs {
		for dockerEndpoint, hostEndpoint := range serviceMap {
			envs[i] = strings.ReplaceAll(envs[i], dockerEndpoint, hostEndpoint)
		}
	}

	return envs, nil
}

// buildContainerName from session id and service name.
// fails if service is not registered in bake session
func buildContainerName(session *docker.Session, serviceName string) (string, error) {
	_, err := session.AutoServiceAddress(serviceName)
	if err != nil {
		return "", fmt.Errorf("service with name %s is not found", serviceName)
	}

	return fmt.Sprintf("%s-%s", session.ID(), serviceName), nil
}

// buildServiceMap where key is docker related endpoint and value is corresponding localhost endpoint
func buildServiceMap(session *docker.Session) (map[string]string, error) {
	serviceMap := map[string]string{}

	for _, svc := range session.ServiceNames() {
		dockerAddress, err := session.DockerToDockerServiceAddress(svc)
		if err != nil {
			return nil, fmt.Errorf("failed to find docker endpoint for service %s", svc)
		}
		localAddress, err := session.AutoServiceAddress(svc)
		if err != nil {
			return nil, fmt.Errorf("failed to find local endpoint for service %s", svc)
		}
		// this hack is required for replica set mongo to be able to connect directly
		// otherwise client is not able to ping mongo container.
		if svc == mongodb.ServiceName {
			localAddress += "/?connect=direct"
		}
		serviceMap[dockerAddress] = localAddress
	}

	return serviceMap, nil
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
