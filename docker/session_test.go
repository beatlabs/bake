package docker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostMapping(t *testing.T) {
	sess := Session{hostMappedServiceAddresses: map[string]string{}}
	err := sess.RegisterHostMappedDockerService("redis", "8080")
	require.NoError(t, err)

	assert.Equal(t, "8080", sess.hostMappedServiceAddresses["redis"])
}

func TestHostMappingError(t *testing.T) {
	sess := Session{hostMappedServiceAddresses: map[string]string{"redis": "8080"}}
	err := sess.RegisterHostMappedDockerService("redis", "8080")
	assert.EqualError(t, err, `service "redis" which already exists with value: "8080"`)
}

func TestDockerToDocker(t *testing.T) {
	sess := Session{hostMappedServiceAddresses: map[string]string{"redis": "8080"}}
	_, err := sess.HostToDockerServiceAddress("redis")
	require.NoError(t, err)
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
	require.NoError(t, err)

	loadedSession, err := LoadSessionFromFile(false, DefaultSessionFile)
	require.NoError(t, err)
	assert.Equal(t, loadedSession.id, sess.id)

	err = os.Remove(DefaultSessionFile)
	require.NoError(t, err)
}
