package model

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONStrings_NilWritesEmptyArray(t *testing.T) {
	var j JSONStrings
	v, err := j.Value()
	require.NoError(t, err)
	require.Equal(t, "[]", v)
}

func TestJSONProjects_NilWritesEmptyArray(t *testing.T) {
	var j JSONProjects
	v, err := j.Value()
	require.NoError(t, err)
	require.Equal(t, "[]", v)
}

func TestJSONStrings_GormValueForcesEmptyArray(t *testing.T) {
	var j JSONStrings
	expr := j.GormValue(context.Background(), nil)
	require.Equal(t, "?::jsonb", expr.SQL)
	require.Equal(t, []interface{}{"[]"}, expr.Vars)
}

func TestJSONProjects_GormValueForcesEmptyArray(t *testing.T) {
	var j JSONProjects
	expr := j.GormValue(context.Background(), nil)
	require.Equal(t, "?::jsonb", expr.SQL)
	require.Equal(t, []interface{}{"[]"}, expr.Vars)
}
