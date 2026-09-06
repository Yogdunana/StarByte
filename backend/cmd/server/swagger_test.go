package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSwaggerEnabled(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	require.True(t, swaggerEnabled())
	t.Setenv("APP_ENV", "test")
	require.True(t, swaggerEnabled())
	t.Setenv("APP_ENV", "")
	require.True(t, swaggerEnabled())
	t.Setenv("APP_ENV", "prod")
	require.False(t, swaggerEnabled())
}
