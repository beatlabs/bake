package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

// Client is a Consul client.
type Client struct {
	client *api.Client
}

// NewClient creates a new Client.
func NewClient(address string) (Client, error) {
	config := api.DefaultConfig()
	config.Address = address
	config.Datacenter = "dc1"
	client, err := api.NewClient(config)
	if err != nil {
		return Client{}, fmt.Errorf("failed to create consul client: %w", err)
	}

	return Client{
		client: client,
	}, nil
}

// Put stores a value.
func (c Client) Put(key, value string) error {
	_, err := c.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: []byte(value),
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to put key %s value %s: %w", key, value, err)
	}
	return nil
}

// Delete removes a value.
func (c Client) Delete(key string) error {
	_, err := c.client.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// DeleteTree removes values under a prefix.
func (c Client) DeleteTree(prefix string) error {
	_, err := c.client.KV().DeleteTree(prefix, nil)
	if err != nil {
		return fmt.Errorf("failed to delete tree %s: %w", prefix, err)
	}
	return nil
}

// Live is a liveness check.
func (c Client) Live() error {
	_, _, err := c.client.KV().List("p", nil)
	return err
}
