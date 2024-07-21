package env

import (
	"fmt"
	"testing"

	"github.com/beatlabs/bake/docker"
	"github.com/beatlabs/bake/docker/component/mockserver"
	"github.com/beatlabs/bake/docker/component/mongodb"
	"github.com/beatlabs/bake/internal/sh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSessionID   = "000"
	testServiceName = "test-service"
)

func TestGetServiceEnvs(t *testing.T) {
	t.Parallel()

	testContainerName := fmt.Sprintf("%s-%s", testSessionID, testServiceName)

	args := []string{
		"run",
		"--name=" + testContainerName,
		"--env=PATRON_HTTP_DEFAULT_PORT=8080",
		"--env=TEST_SERVICE=test_service",
		"--env=TEST_SERVICE_MONGO_URI=mongodb://root:password@000-mongo:27017",
		"--env=TEST_SERVICE_SQS_ENDPOINT=http://000-localstack:4566",
		"--env=TEST_SERVICE_SQS_QUEUE=the_queue",
		"--env=TEST_SERVICE_KAFKA_BROKERS=000-kafka:9092",
		"--env=TEST_SERVICE_API_ENDPOINT=http://000-mockserver:1080",
		"--env=TEST_VALUE=docker-new",
		"alpine",
		"pwd",
	}
	runDockerCmd(t, args)

	extraRules := ReplacementRuleList{
		NewSubstrReplacement("docker", "localhost"),
	}
	session := loadTestSessionFromFile(t, "./testdata/ok.json")
	envs, err := GetServiceEnvs(session, testServiceName, extraRules)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"PATRON_HTTP_DEFAULT_PORT":   "65071",
		"TEST_SERVICE":               "test_service",
		"TEST_SERVICE_MONGO_URI":     "mongodb://root:password@localhost:64952/?connect=direct",
		"TEST_SERVICE_SQS_ENDPOINT":  "http://localhost:64950",
		"TEST_SERVICE_SQS_QUEUE":     "the_queue",
		"TEST_SERVICE_KAFKA_BROKERS": "localhost:64949",
		"TEST_SERVICE_API_ENDPOINT":  "http://localhost:64953",
		"TEST_VALUE":                 "localhost-new",
	}, envs)

	runDockerCmd(t, []string{"rm", testContainerName})
}

func TestGetServiceEnvs_NoService(t *testing.T) {
	t.Parallel()
	session := loadTestSessionFromFile(t, "./testdata/ok.json")
	envs, err := GetServiceEnvs(session, "not-existing-service", ReplacementRuleList{})
	assert.Empty(t, envs)
	assert.EqualError(t, err, "service with name not-existing-service is not found")
}

func TestBuildContainerName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		services         []string
		serviceName      string
		expContainerName string
		expErr           string
	}{
		"not found": {
			services:    []string{mongodb.ServiceName, mockserver.ServiceName, testServiceName},
			serviceName: "invalid-service-name",
			expErr:      "service with name invalid-service-name is not found",
		},
		"ok": {
			services:         []string{mongodb.ServiceName, mockserver.ServiceName, testServiceName},
			serviceName:      testServiceName,
			expContainerName: fmt.Sprintf("%s-%s", testSessionID, testServiceName),
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			session := createTestSession(t, tt.services)
			containerName, err := BuildContainerName(session, tt.serviceName)
			if tt.expErr != "" {
				assert.Empty(t, containerName)
				assert.EqualError(t, err, tt.expErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expContainerName, containerName)
			}
		})
	}
}

func createTestSession(t *testing.T, services []string) *docker.Session {
	session, err := docker.NewSession(testSessionID, "000")
	require.NoError(t, err)
	for _, svc := range services {
		err = session.RegisterHostMappedDockerService(svc, "http://localhost-"+svc)
		require.NoError(t, err)
		err = session.RegisterInternalDockerService(svc, "http://docker-"+svc)
		require.NoError(t, err)
	}
	return session
}

func loadTestSessionFromFile(t *testing.T, filename string) *docker.Session {
	session, err := docker.LoadSessionFromFile(false, filename)
	require.NoError(t, err)
	return session
}

func runDockerCmd(t *testing.T, args []string) {
	err := sh.Run(dockerCmd, args...)
	require.NoError(t, err)
}
