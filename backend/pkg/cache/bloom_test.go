package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloom_AddAndMiss(t *testing.T) {
	b := NewBloom(0, 0)
	assert.False(t, b.MightHave("never"))
	b.Add("user:1")
	assert.True(t, b.MightHave("user:1"))
	assert.False(t, b.MightHave("user:999999-not-added"))
}

func TestBloom_DefaultSize(t *testing.T) {
	b := NewBloom(8, 0)
	assert.GreaterOrEqual(t, int(b.m), 64)
}
