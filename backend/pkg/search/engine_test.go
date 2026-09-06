package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeAndFloat(t *testing.T) {
	assert.Equal(t, "abc", normalize([]byte("abc")))
	assert.Nil(t, normalize(nil))
	assert.Equal(t, 3.5, asFloat(3.5))
	assert.Equal(t, 2.0, asFloat(int64(2)))
	assert.Equal(t, 4.0, asFloat([]byte("4")))
}

func TestPublicFields(t *testing.T) {
	fs := demoSchema().PublicFields()
	assert.GreaterOrEqual(t, len(fs), 3)
	assert.Contains(t, fs[1]["operators"], OpLike)
}
