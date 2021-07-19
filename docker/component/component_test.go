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
	"github.com/taxibeat/bake/docker/component/jaeger"
	"github.com/taxibeat/bake/docker/component/kafka"
	"github.com/taxibeat/bake/docker/component/localstack"
	"github.com/taxibeat/bake/docker/component/mockserver"
	"github.com/taxibeat/bake/docker/component/mongodb"
	"github.com/taxibeat/bake/docker/component/redis"
	"github.com/taxibeat/bake/docker/component/testservice"
	"gopkg.in/Shopify/sarama.v1"
)

var session *docker.Session

func TestMain(m *testing.M) {
	var err error
	session, err = docker.LoadSession()
	if err != nil {
		newSession()
	}

	os.Exit(m.Run())
}

func newSession() {
	sessionID, netID, err := docker.GetEnv()
	checkErr(err)

	sessionID += "-bake"

	session, err = docker.NewSession(sessionID, netID)
	checkErr(err)

	err = session.StartComponents(
		kafka.NewComponent(session, kafka.WithTopics("foo:1:1")),
		consul.NewComponent(docker.WithTag("1.8.0")),
		jaeger.NewComponent(),
		localstack.NewComponent(localstack.WithServices("s3")),
		mockserver.NewComponent(),
		redis.NewComponent(),
		mongodb.NewComponent(),
	)
	checkErr(err)

	redisAddr, err := session.DockerToDockerServiceAddress(redis.ServiceName)
	checkErr(err)

	mongoAddr, err := session.DockerToDockerServiceAddress(mongodb.ServiceName)
	checkErr(err)

	kafkaAddr, err := session.DockerToDockerServiceAddress(kafka.KafkaServiceName)
	checkErr(err)

	serviceComponent, err := testservice.NewComponent(redisAddr, mongoAddr, kafkaAddr)
	checkErr(err)

	err = session.StartComponents(serviceComponent)
	checkErr(err)

	// Optional: Store snapshot to filesystem.
	// Should only be used if the tests can be run against dirty resources.
	err = session.Persist()
	checkErr(err)
}

func TestConsul(t *testing.T) {
	consulAddr, err := session.AutoServiceAddress(consul.ServiceName)
	require.NoError(t, err)

	consulClient, err := consul.NewClient(consulAddr)
	assert.NoError(t, err)

	err = consulClient.DeleteTree("services/")
	assert.NoError(t, err)

	err = consulClient.Put("services/foo/bar", "23")
	assert.NoError(t, err)
}

func TestRedis(t *testing.T) {
	redisAddr, err := session.AutoServiceAddress(redis.ServiceName)
	require.NoError(t, err)

	redisClient := redis.NewClient(redisAddr)

	_, err = redisClient.Set(context.Background(), "foo", "bar", time.Second).Result()
	assert.NoError(t, err)
}

func TestMongo(t *testing.T) {
	mongoAddr, err := session.AutoServiceAddress(mongodb.ServiceName)
	require.NoError(t, err)

	mongoClient, err := mongodb.NewClient(context.Background(), mongoAddr)
	assert.NoError(t, err)

	err = mongoClient.Ping(context.Background(), nil)
	assert.NoError(t, err)
}

func TestKafka(t *testing.T) {
	kafkaAddr, err := session.AutoServiceAddress(kafka.KafkaServiceName)
	assert.NoError(t, err)
	kafkaClient, err := sarama.NewClient([]string{kafkaAddr}, nil)
	assert.NoError(t, err)

	topics, err := kafkaClient.Topics()
	assert.NoError(t, err)
	assert.Contains(t, topics, "foo")
}

func TestMockServer(t *testing.T) {
	mockServerAddr, err := session.AutoServiceAddress(mockserver.ServiceName)
	assert.NoError(t, err)
	mockServerClient := mockserver.NewClient(mockServerAddr)
	err = mockServerClient.CreateExpectation(
		mockserver.Expectation{
			Request: mockserver.Request{Method: "GET", Path: "/"},
			Response: mockserver.Response{
				Status: 200,
				Body:   struct{}{},
				Delay:  mockserver.Delay{TimeUnit: mockserver.Milliseconds, Value: 100},
			},
		})
	assert.NoError(t, err)
	err = mockServerClient.Reset()
	assert.NoError(t, err)
}

func TestExampleService(t *testing.T) {
	testServiceAddr, err := session.AutoServiceAddress(testservice.ServiceName)
	assert.NoError(t, err)

	resp, err := http.Get("http://" + testServiceAddr)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func checkErr(err error) {
	if err != nil {
		if session != nil {
			if werr := session.Persist(); werr != nil {
				fmt.Printf("session write failed: %v\n", werr)
			}
		}
		fmt.Printf("test setup failed: %v\n", err)
		os.Exit(1)
	}
}
