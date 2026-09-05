package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/stretchr/testify/require"
)

func TestCalculateVoteResult_EqualAndWeighted(t *testing.T) {
	recs := []model.VoteRecord{
		{OptionKey: "web", Weight: 1},
		{OptionKey: "web", Weight: 5},
		{OptionKey: "ai", Weight: 3},
	}
	agg, total, tw := CalculateVoteResult(recs)
	require.Equal(t, 3, total)
	require.Equal(t, 9.0, tw)
	require.Equal(t, 2, agg["web"].Count)
	require.Equal(t, 6.0, agg["web"].Weight)
	require.Equal(t, 1, agg["ai"].Count)
	require.Equal(t, 3.0, agg["ai"].Weight)
}
