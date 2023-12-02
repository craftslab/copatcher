package connhelpers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildx(t *testing.T) {
	u := url.URL{}
	u.Path = "invalid"

	_, err := Buildx(&u)
	assert.NotEqual(t, nil, err)
}

func TestBuildxContextDialer(t *testing.T) {
	// TODO: FIXME
	assert.Equal(t, nil, nil)
}

func TestContainerContextDialer(t *testing.T) {
	// TODO: FIXME
	assert.Equal(t, nil, nil)
}
