package redis

import (
	"github.com/redis/go-redis/v9"
)

// NewClient creates a new Client.
func NewClient(address string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: address})
}
