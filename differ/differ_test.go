package differ

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffer(t *testing.T) {
	buf := New()
	assert.NotEqual(t, nil, buf)
}
