package component

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

func Test_Component_GetContainer(t *testing.T) {
	names := []string{"foo", "bar"}
	ports := []string{"1", "2"}
	hosts := []string{"pref_foo:1", "pref_foo:2"}

	c := &BaseComponent{}
	c1 := &stubContainer{
		name:     names[0],
		host:     hosts[0],
		port:     ports[0],
		startErr: nil,
	}
	c2 := &stubContainer{
		name:     names[0],
		host:     hosts[0],
		port:     ports[0],
		startErr: nil,
	}
	c.WithContainer(c1)
	c.WithContainer(c2)

	testCases := map[string]struct {
		name     string
		expected Container
	}{
		"existing container": {
			name:     names[0],
			expected: c1,
		},
		"container not found 1": {
			name:     "",
			expected: nil,
		},
		"container not found 2": {
			name:     hosts[0],
			expected: nil,
		},
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			found := c.GetContainer(tt.name)
			assert.Equal(t, tt.expected, found)
		})
	}
}

type stubContainer struct {
	name     string
	host     string
	port     string
	startErr error
	stopErr  error
}

func (c *stubContainer) Name() string {
	return c.name
}

func (c *stubContainer) Start(pool *dockertest.Pool, networkID string, expiration uint) error {
	return c.startErr
}

func (c *stubContainer) Stop(pool *dockertest.Pool) error {
	return c.stopErr
}

func (c *stubContainer) Address(pool *dockertest.Pool) string {
	return ""
}

func (c *stubContainer) ExternalAddress(pool *dockertest.Pool) string {
	return ""
}

func (c *stubContainer) InternalAddress() string {
	return ""
}
