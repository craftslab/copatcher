package report

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	p := New(context.Background(), DefaultConfig())
	assert.NotEqual(t, nil, p)
}
