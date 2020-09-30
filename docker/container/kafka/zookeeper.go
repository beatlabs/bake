package kafka

import (
	"github.com/ory/dockertest/v3"
	"github.com/taxibeat/bake/docker/container"
)

const (
	containerNameZookeeper = "zookeeper"
	defaultPortZookeeper   = "2181"
)

type zookeeperParams struct {
	Prefix        string
	Version       string
	ContainerHost bool
	UseExpiration bool
}

type zookeeperContainer struct {
	params zookeeperParams
	container.BaseContainer
}

func newZookeeperContainer(params zookeeperParams) *zookeeperContainer {
	return &zookeeperContainer{
		params:        params,
		BaseContainer: *container.NewBaseContainer(params.Prefix, containerNameZookeeper, defaultPortZookeeper, params.ContainerHost),
	}
}

// Start a new zookeeper container.
func (c *zookeeperContainer) Start(pool *dockertest.Pool, networkID string, expiration uint) error {
	opts := &dockertest.RunOptions{
		Name:       c.Name(),
		NetworkID:  networkID,
		Repository: "wurstmeister/zookeeper",
		Tag:        c.params.Version,
	}

	resource, err := pool.RunWithOptions(opts)
	if err != nil {
		return err
	}

	if c.params.UseExpiration {
		err = resource.Expire(expiration)
		if err != nil {
			return err
		}
	}

	return nil
}
