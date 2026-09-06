package service

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func compactArchiveJSON() []byte {
	return []byte(`[{"path":"/api/v1/users/keep-a","action":"CREATE"},{"path":"/api/v1/users/` + strings.Repeat("b", 40) + `","action":"UPDATE"}]`)
}

func TestDecodeArchiveLogs_OversizedTruncates(t *testing.T) {
	old := archiveDecodeLimit
	archiveDecodeLimit = 60
	t.Cleanup(func() { archiveDecodeLimit = old })

	raw := compactArchiveJSON()
	require.Greater(t, len(raw), archiveDecodeLimit)

	got, truncated, err := decodeArchiveLogs(raw)
	require.NoError(t, err)
	assert.True(t, truncated)
	require.NotEmpty(t, got)
	assert.Equal(t, "/api/v1/users/keep-a", got[0].Path)
}

func TestDecodeArchiveLogs_GzipOversizedTruncates(t *testing.T) {
	old := archiveDecodeLimit
	archiveDecodeLimit = 60
	t.Cleanup(func() { archiveDecodeLimit = old })

	payload := compactArchiveJSON()
	require.Greater(t, len(payload), archiveDecodeLimit)

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	got, truncated, err := decodeArchiveLogs(buf.Bytes())
	require.NoError(t, err)
	assert.True(t, truncated)
	require.NotEmpty(t, got)
	assert.Equal(t, "/api/v1/users/keep-a", got[0].Path)
}

func TestDecodeArchiveLogs_ValidSmall(t *testing.T) {
	got, truncated, err := decodeArchiveLogs([]byte(`[{"path":"/ok","action":"LOGIN"}]`))
	require.NoError(t, err)
	assert.False(t, truncated)
	require.Len(t, got, 1)
	assert.Equal(t, "/ok", got[0].Path)
}
