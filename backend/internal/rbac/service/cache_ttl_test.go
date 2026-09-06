package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJitteredTTL(t *testing.T) {
	assert.Equal(t, time.Minute, jitteredTTL(time.Minute, 0))
	got := jitteredTTL(permCacheTTL, permCacheJitter)
	assert.GreaterOrEqual(t, got, permCacheTTL-permCacheJitter)
	assert.LessOrEqual(t, got, permCacheTTL+permCacheJitter)
}

func TestPermCacheKey(t *testing.T) {
	s := &permissionCacheService{}
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	assert.Equal(t, permCacheKeyPrefix+id.String(), s.permCacheKey(id))
}
