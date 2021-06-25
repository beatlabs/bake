package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewClient creates a new Client.
func NewClient(ctx context.Context, address string) (*mongo.Client, error) {
	opts := options.Client()
	opts.SetDirect(true)
	opts.ApplyURI("mongodb://" + address)

	return mongo.Connect(ctx, opts)
}
