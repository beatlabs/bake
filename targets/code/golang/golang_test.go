// Package golang contains go code related mage targets.
package golang

import "testing"

func TestGo_ModUpgradePR(t *testing.T) {
	g := Go{}
	g.ModFullUpgradePR()
}
