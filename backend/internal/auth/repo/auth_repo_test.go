package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRefreshToken(t *testing.T) {
	r := &authRepo{}
	token1 := r.GenerateRefreshToken()
	token2 := r.GenerateRefreshToken()

	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2)
	assert.Len(t, token1, 73) // uuid + "-" + uuid = 36 + 1 + 36 = 73
}

func TestMaxLoginAttempts(t *testing.T) {
	assert.Equal(t, 5, MaxLoginAttempts())
}

func TestLockoutDuration(t *testing.T) {
	assert.Equal(t, 15*time.Minute, LockoutDuration())
}
