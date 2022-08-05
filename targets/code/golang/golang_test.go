// Package golang contains go code related mage targets.
package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGo_ModUpgradePR(t *testing.T) {
	g := Go{}
	assert.NoError(t, g.ModUpgradePR())
}
