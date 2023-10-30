package pkgmgr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPkgmgr(t *testing.T) {
	p := New(context.Background(), DefaultConfig())
	assert.NotEqual(t, nil, p)
}
