package redis

import (
	"github.com/go-redis/redis/v8"
)

// NewClient creates a new Client.
func NewClient(address string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: address})
}
