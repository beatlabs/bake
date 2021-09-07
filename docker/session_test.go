package docker

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestHostMapping(t *testing.T) {
	sess := Session{hostMappedServiceAddresses: map[string]string{}}
	err := sess.RegisterHostMappedDockerService("redis", "8080")
	assert.NoError(t, err)

	assert.Equal(t, sess.hostMappedServiceAddresses["redis"], "8080")
}

func TestHostMappingError(t *testing.T) {
	sess := Session{hostMappedServiceAddresses: map[string]string{"redis": "8080"}}
	err := sess.RegisterHostMappedDockerService("redis", "8080")
	assert.EqualError(t, err, `service "redis" which already exists with value: "8080"`)
}

func TestDockerToDocker(t *testing.T) {
	sess := Session{hostMappedServiceAddresses: map[string]string{"redis": "8080"}}
	_, err := sess.HostToDockerServiceAddress("redis")
	assert.NoError(t, err)
}

func TestDockerToDockerError(t *testing.T) {
	sess := Session{hostMappedServiceAddresses: map[string]string{}}
	_, err := sess.HostToDockerServiceAddress("redis")
	assert.EqualError(t, err, `external service address not registered for "redis"`)
}

func TestPersistAndLoadingSession(t *testing.T) {
	sess := Session{
		id:        "test-id-1",
		networkID: "test-network",
		inDocker:  false,
	}
	err := sess.PersistToFile(DefaultSessionFile)
	assert.NoError(t, err)

	loadedSession, err := LoadSessionFromFile(false, DefaultSessionFile)
	require.NoError(t, err)
	assert.Equal(t, loadedSession.id, sess.id)

	_ = CleanupResources()
}
