package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRefreshRecord_JSONAndLegacy(t *testing.T) {
	uid, jti := parseRefreshRecord(`{"user_id":"u-1","jti":"j-9"}`)
	assert.Equal(t, "u-1", uid)
	assert.Equal(t, "j-9", jti)

	uid, jti = parseRefreshRecord("plain-user-id")
	assert.Equal(t, "plain-user-id", uid)
	assert.Equal(t, "", jti)

	uid, jti = parseRefreshRecord("  {\"user_id\":\"u-2\"}  ")
	assert.Equal(t, "u-2", uid)
	assert.Equal(t, "", jti)

	uid, jti = parseRefreshRecord("{not-json")
	assert.Equal(t, "{not-json", uid)
	assert.Equal(t, "", jti)
}
