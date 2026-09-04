package model

import (
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
