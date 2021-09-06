package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoContainers(t *testing.T) {
	cmp := SimpleComponent{
		Name: "test-cmp",
	}

	sess := Session{
		id:        "simple-component-test",
		networkID: "asd",
		inDocker:  false,
	}
	err := cmp.Start(&sess)
	assert.EqualError(t, err, "component test-cmp has no containers to start")
}

func TestStartOutsideDocker(t *testing.T) {
	cmp := SimpleComponent{
		Name: "test-cmp",
		Containers: []SimpleContainerConfig{
			{
				Name:       "container-1",
				Repository: "bitnami/redis",
				Tag:        "latest",
			},
		},
	}

	sess := Session{
		id:       "simple-component-test",
		inDocker: false,
	}
	err := cmp.Start(&sess)
	assert.NoError(t, err)
}
