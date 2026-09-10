package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

func TestResolveWeight_EqualAlwaysOne(t *testing.T) {
	cfg := DefaultWeightConfig()
	require.Equal(t, 1.0, ResolveWeight(cfg, "president", model.VoteEqual))
}

func TestResolveWeight_FromConfig(t *testing.T) {
	cfg := DefaultWeightConfig()
	require.Equal(t, 5.0, ResolveWeight(cfg, "president", model.VoteWeighted))
	require.Equal(t, 3.0, ResolveWeight(cfg, "minister", model.VoteWeighted))
	require.Equal(t, 2.0, ResolveWeight(cfg, "vice_minister", model.VoteWeighted))
	require.Equal(t, 2.0, ResolveWeight(cfg, "deputy", model.VoteWeighted))
	require.Equal(t, 1.0, ResolveWeight(cfg, "unknown", model.VoteWeighted))
}

func TestParseWeightConfig_Fallback(t *testing.T) {
	cfg, err := parseWeightConfig(`{"weights":{"president":9},"default_weight":2}`)
	require.NoError(t, err)
	require.Equal(t, 9.0, cfg.Weights["president"])
	require.Equal(t, 2.0, cfg.DefaultWeight)
	_, err = parseWeightConfig("not-json")
	require.Error(t, err)
}
