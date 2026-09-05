package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCalculateDuration(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	require.Equal(t, 15, CalculateDuration(start, end))
	require.Equal(t, 0, CalculateDuration(end, start))
	require.Greater(t, CalculateDuration(start, time.Time{}), 0)
}

func TestCalculateMonthlyDuration(t *testing.T) {
	start := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	got := CalculateMonthlyDuration(start, end)
	require.Equal(t, 12, got["2026-08"])
	require.Equal(t, 4, got["2026-09"])
	require.Equal(t, 16, CalculateDuration(start, end))
	require.Empty(t, CalculateMonthlyDuration(end, start))
}
