package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorRoundTrip(t *testing.T) {
	enc, err := encodeCursor([]any{float64(3), "id-1"}, "id-1")
	require.NoError(t, err)
	got, err := decodeCursor(enc)
	require.NoError(t, err)
	assert.Equal(t, "id-1", got.ID)
	assert.Len(t, got.Values, 2)
}

func TestDecodeCursorBad(t *testing.T) {
	_, err := decodeCursor("%%%")
	assert.ErrorIs(t, err, ErrInvalidCursor)
	_, err = decodeCursor("")
	assert.Error(t, err)
}

func TestCompileCursorDesc(t *testing.T) {
	schema := demoSchema()
	sorts := []Sort{{Field: "created_at", Desc: true}, {Field: "id", Desc: true}}
	var args []any
	sql, err := compileCursor(schema, sorts, &cursorPayload{
		Values: []any{"t", "i"},
		ID:     "i",
	}, &args)
	require.NoError(t, err)
	assert.Contains(t, sql, " < ?")
	assert.Contains(t, sql, " = ?")
	assert.Len(t, args, 3)
}

func TestAsSliceAndLikeEscape(t *testing.T) {
	got, err := asSlice([]string{"a", "b"})
	require.NoError(t, err)
	assert.Equal(t, []any{"a", "b"}, got)
	assert.Equal(t, `\%`, escapeLike("%"))
	assert.Equal(t, `\_`, escapeLike("_"))
	assert.Equal(t, `\\`, escapeLike(`\`))
}
