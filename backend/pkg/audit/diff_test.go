package audit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffJSON_TopLevelAndNested(t *testing.T) {
	before := `{"name":"旧","status":0,"dept":{"id":"a"}}`
	after := `{"name":"新","status":0,"dept":{"id":"b"},"phone":"13812345678"}`
	raw := DiffJSON(before, after)
	var changes []FieldChange
	require.NoError(t, json.Unmarshal([]byte(raw), &changes))
	got := map[string]FieldChange{}
	for _, c := range changes {
		got[c.Path] = c
	}
	assert.Equal(t, "旧", got["name"].Before)
	assert.Equal(t, "新", got["name"].After)
	assert.Equal(t, "a", got["dept.id"].Before)
	assert.Equal(t, "b", got["dept.id"].After)
	assert.Equal(t, "13812345678", got["phone"].After)
	_, hasStatus := got["status"]
	assert.False(t, hasStatus)
}

func TestDiffJSON_Empty(t *testing.T) {
	assert.Equal(t, "", DiffJSON("", ""))
	assert.Equal(t, "[]", DiffJSON(`{"a":1}`, `{"a":1}`))
}

func TestCompactJSON_Desensitizes(t *testing.T) {
	s := CompactJSON(map[string]any{"password": "secret123", "name": "部长"}, 4096)
	assert.Contains(t, s, "部长")
	assert.NotContains(t, s, "secret123")
}

func TestParseFieldChanges(t *testing.T) {
	assert.Nil(t, ParseFieldChanges(""))
	assert.Nil(t, ParseFieldChanges("not-json"))
	raw := `[{"path":"name","before":"a","after":"b"}]`
	got := ParseFieldChanges(raw)
	require.Len(t, got, 1)
	assert.Equal(t, "name", got[0].Path)
}
