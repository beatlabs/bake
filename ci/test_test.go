package ci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getBuildTagFlag(t *testing.T) {
	assert.Equal(t, "-tags=integration,component", getBuildTagFlag([]string{"integration", "component"}))
}
