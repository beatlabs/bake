// +build component

package component

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/bake/docker"
	"github.com/taxibeat/bake/docker/component/consul"
	"github.com/taxibeat/bake/docker/component/kafka"
	"github.com/taxibeat/bake/docker/component/mongodb"
	"github.com/taxibeat/bake/docker/component/redis"
	"github.com/taxibeat/bake/docker/component/testservice"
	"gopkg.in/Shopify/sarama.v1"
)

var session *docker.Session

func TestMain(m *testing.M) {
	var err error
	session, err = docker.FromFile(docker.SessionFile)
	if err != nil {
		sessionID, netID, err := docker.FromEnv()
		checkErr(err)

		session, err = docker.NewSession(sessionID, netID)
		checkErr(err)

		err = session.StartComponents(
			consul.NewComponent(),
			redis.NewComponent(),
			mongodb.NewComponent(),
			kafka.NewComponent(session, []string{"foo:1:1"}),
		)
		checkErr(err)

		serviceComponent, err := testservice.NewComponent(session)
		checkErr(err)

		err = session.StartComponents(serviceComponent)
		checkErr(err)
	}

	session.WriteToFile(docker.SessionFile)

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestClients(t *testing.T) {
	consulAddr, err := session.AutoServiceAddress(consul.ServiceName)
	require.NoError(t, err)

	consulClient, err := consul.NewClient(consulAddr)
	assert.NoError(t, err)

	err = consulClient.DeleteTree("services/")
	assert.NoError(t, err)

	err = consulClient.Put("services/foo/bar", "23")
	assert.NoError(t, err)

	redisAddr, err := session.AutoServiceAddress(redis.ServiceName)
	require.NoError(t, err)

	redisClient := redis.NewClient(redisAddr)

	_, err = redisClient.Set(context.Background(), "foo", "bar", time.Second).Result()
	assert.NoError(t, err)

	mongoAddr, err := session.AutoServiceAddress(mongodb.ServiceName)
	require.NoError(t, err)

	mongoClient, err := mongodb.NewClient(context.Background(), mongoAddr)
	assert.NoError(t, err)

	err = mongoClient.Ping(context.Background(), nil)
	assert.NoError(t, err)

	kafkaAddr, err := session.AutoServiceAddress(kafka.KafkaServiceName)
	assert.NoError(t, err)
	kafkaClient, err := sarama.NewClient([]string{kafkaAddr}, nil)
	assert.NoError(t, err)

	topics, err := kafkaClient.Topics()
	assert.NoError(t, err)
	assert.Contains(t, topics, "foo")

	testServiceAddr, err := session.AutoServiceAddress(testservice.ServiceName)
	assert.NoError(t, err)

	resp, err := http.Get("http://" + testServiceAddr)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func checkErr(err error) {
	if err != nil {
		_ = session.WriteToFile(docker.SessionFile)
		fmt.Printf("test setup failed: %v\n", err)
		os.Exit(1)
	}
}
