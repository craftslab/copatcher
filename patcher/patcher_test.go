package patcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatcher(t *testing.T) {
	p := New(context.Background(), DefaultConfig())
	assert.NotEqual(t, nil, p)
}
