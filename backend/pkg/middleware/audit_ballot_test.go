package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBallotChoicesNeverEnterAuditBodies(t *testing.T) {
	path := "/api/v1/votes/680e2013-a925-42de-917b-2ff4b82f71e5/cast"
	actual := sanitizeRequestBody(path, `{"option_key":"SECRET_CHOICE"}`)
	require.NotContains(t, actual, "SECRET_CHOICE")
	require.NotContains(t, actual, "option_key")
	require.Contains(t, actual, "redacted")
}
