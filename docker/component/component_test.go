//go:build component

package component

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/beatlabs/bake/docker"
	"github.com/beatlabs/bake/docker/component/consul"
	"github.com/beatlabs/bake/docker/component/jaeger"
	"github.com/beatlabs/bake/docker/component/kafka"
	"github.com/beatlabs/bake/docker/component/localstack"
	"github.com/beatlabs/bake/docker/component/mockserver"
	"github.com/beatlabs/bake/docker/component/mongodb"
	"github.com/beatlabs/bake/docker/component/redis"
	"github.com/beatlabs/bake/docker/component/testservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	err = consulClient.DeleteTree("services/")
	require.NoError(t, err)

	err = consulClient.Put("services/foo/bar", "23")
	require.NoError(t, err)

	err = consulClient.Delete("services/foo/bar")
	require.NoError(t, err)
}

func TestRedis(t *testing.T) {
	redisAddr, err := session.AutoServiceAddress(redis.ServiceName)
	require.NoError(t, err)

	redisClient := redis.NewClient(redisAddr)

	_, err = redisClient.Set(context.Background(), "foo", "bar", time.Second).Result()
	require.NoError(t, err)
}

func TestMongo(t *testing.T) {
	mongoAddr, err := session.AutoServiceAddress(mongodb.ServiceName)
	require.NoError(t, err)

	mongoClient, err := mongodb.NewClient(context.Background(), mongoAddr)
	require.NoError(t, err)

	err = mongoClient.Ping(context.Background(), nil)
	require.NoError(t, err)
}

func TestKafka(t *testing.T) {
	kafkaAddr, err := session.AutoServiceAddress(kafka.KafkaServiceName)
	require.NoError(t, err)
	kafkaClient, err := sarama.NewClient([]string{kafkaAddr}, nil)
	require.NoError(t, err)

	topics, err := kafkaClient.Topics()
	require.NoError(t, err)
	assert.Contains(t, topics, "foo")
}

func TestMockServer(t *testing.T) {
	mockServerAddr, err := session.AutoServiceAddress(mockserver.ServiceName)
	require.NoError(t, err)
	mockServerClient := mockserver.NewClient(mockServerAddr)
	err = mockServerClient.CreateExpectation(
		mockserver.Expectation{
			Request: mockserver.Request{Method: "GET", Path: "/"},
			Response: mockserver.Response{
				Status:  200,
				Body:    struct{}{},
				Delay:   &mockserver.Delay{TimeUnit: mockserver.Milliseconds, Value: 50},
				Headers: map[string][]string{"X-TEST": {"test"}},
			},
			Times: mockserver.CallTimes{Unlimited: true},
		})
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+mockServerAddr, nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "test", res.Header.Get("X-Test"))

	err = mockServerClient.Reset()
	require.NoError(t, err)
}

func TestExampleService(t *testing.T) {
	testServiceAddr, err := session.AutoServiceAddress(testservice.ServiceName)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+testServiceAddr, nil)
	require.NoError(t, err)
	rsp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, rsp.Body.Close())
	assert.Equal(t, 200, rsp.StatusCode)
}

func checkErr(err error) {
	if err == nil {
		return
	}

	if session != nil {
		if cerr := docker.CleanupSessionResources(session); cerr != nil {
			fmt.Printf("failed to cleanup resources: %v\n", err)
		}
	}

	fmt.Printf("test setup failed: %v\n", err)
	os.Exit(1)
}
