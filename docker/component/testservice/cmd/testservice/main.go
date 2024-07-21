// Package main of the test service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/beatlabs/bake/docker/component/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	redisAddr := os.Getenv("REDIS")
	redisClient := redis.NewClient(redisAddr)
	_, err := redisClient.Set(context.Background(), "testservice", "foo", time.Second).Result()
	if err != nil {
		log.Fatal(err)
	}

	mongoAddr := os.Getenv("MONGO")

	opts := options.Client()
	rs := "rs0"
	opts.ReplicaSet = &rs
	opts.ApplyURI("mongodb://" + mongoAddr)
	mongoClient, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}

	err = mongoClient.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	kafkaAddr := os.Getenv("KAFKA")
	kafkaClient, err := sarama.NewClient([]string{kafkaAddr}, nil)
	if err != nil {
		log.Fatal(err)
	}
	_, err = kafkaClient.Topics()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", Health)
	port := os.Getenv("PORT")
	fmt.Println("Running on port:", port)

	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 3 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Health is a simple health endpoint.
func Health(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "OK")
}
