package buildkit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildkit(t *testing.T) {
	p := New(context.Background(), DefaultConfig())
	assert.NotEqual(t, nil, p)
}
