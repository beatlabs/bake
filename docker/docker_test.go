// +build component

package docker

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/taxibeat/bake/docker/component"
	"github.com/taxibeat/bake/docker/container/consul"
	"github.com/taxibeat/bake/docker/container/kafka"
	"github.com/taxibeat/bake/docker/container/localstack"
	"github.com/taxibeat/bake/docker/container/mock"
)

var comp *component.BaseComponent

func TestMain(m *testing.M) {
	runID := os.Getenv("RUN_ID")
	netID := os.Getenv("NETWORK_ID")
	containerHost := runID != ""
	runtime := NewRuntime()

	c, err := newTestComponent(runID, netID, containerHost, true)
	if err != nil {
		fmt.Printf("failed to create docker test environment: %v", err)
		os.Exit(1)
	}
	runtime.WithComponent(c)
	comp = c

	ee := runtime.Start()
	if len(ee) > 0 {
		for _, err := range ee {
			fmt.Printf("failed to start up test environment: %v", err)
		}
		os.Exit(1)
	}

	exitCode := m.Run()

	ee = runtime.Teardown()
	if len(ee) > 0 {
		for _, err := range ee {
			fmt.Printf("failed to teardown test environment: %v", err)
		}
	}

	os.Exit(exitCode)
}

func TestClients(t *testing.T) {
	consulClient, err := consul.NewClient(comp.GetContainer(consul.ContainerName).Address(comp.Pool))
	assert.NoError(t, err)

	err = consulClient.DeleteTree("services/")
	assert.NoError(t, err)

	err = consulClient.Put("services/foo/bar", "23")
	assert.NoError(t, err)

	mockClient := mock.NewClient("http://" + comp.GetContainer("mock-test").Address(comp.Pool))
	err = mockClient.Reset()
	assert.NoError(t, err)
}

func newTestComponent(prefix, existingNetworkID string, containerHost, useExpiration bool) (*component.BaseComponent, error) {
	testComponent, err := component.NewBaseComponent(component.DefaultRuntimeExp, component.DefaultContainerExp, prefix, existingNetworkID)
	if err != nil {
		return nil, err
	}

	kafkaContainer, err := kafka.NewContainer(kafka.Params{
		Prefix:           prefix,
		ContainerHost:    containerHost,
		UseExpiration:    useExpiration,
		Topics:           []string{"test-topic"},
		KafkaVersion:     "2.12-2.4.1",
		ZookeeperVersion: "latest",
	})
	if err != nil {
		return nil, err
	}
	testComponent.WithContainer(kafkaContainer)

	mockServerVersion := "mockserver-5.10.0"

	mockContainer := mock.NewContainer(mock.Params{
		Name:          "mock-test",
		Prefix:        prefix,
		Version:       mockServerVersion,
		ContainerHost: containerHost,
		UseExpiration: useExpiration,
	})
	testComponent.WithContainer(mockContainer)

	consulContainer := consul.NewContainer(consul.Params{
		Prefix:        prefix,
		Version:       "1.8.0",
		ContainerHost: containerHost,
		UseExpiration: useExpiration,
	})
	testComponent.WithContainer(consulContainer)

	localstackContainer, err := localstack.NewContainer(localstack.Params{
		Prefix:        prefix,
		Version:       "0.11.4",
		Services:      []string{localstack.ServiceS3},
		ContainerHost: containerHost,
		UseExpiration: useExpiration,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create Localstack container: %w", err)
	}
	testComponent.WithContainer(localstackContainer)

	return testComponent, nil
}
