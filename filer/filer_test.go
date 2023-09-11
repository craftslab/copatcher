package filer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiler(t *testing.T) {
	r := New(context.Background(), DefaultConfig())
	assert.NotEqual(t, nil, r)
}
