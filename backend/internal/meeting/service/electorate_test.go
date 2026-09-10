package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

func TestElectorateUsesHighestConfiguredRoleOnce(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	candidates := []model.ElectorCandidate{{UserID: first, PositionCode: "officer", RoleCode: "minister"}, {UserID: first, PositionCode: "officer", RoleCode: "president"}, {UserID: second, RoleCode: "unknown"}}
	cfg := DefaultWeightConfig()
	cfg.DefaultWeight = 2
	rows := buildElectorate(uuid.New(), candidates, cfg, model.VoteWeighted)
	require.Len(t, rows, 2)
	for _, row := range rows {
		if row.UserID == first {
			require.Equal(t, 5.0, row.Weight)
		} else {
			require.Equal(t, 2.0, row.Weight)
		}
	}
	for _, row := range buildElectorate(uuid.New(), candidates, cfg, model.VoteEqual) {
		require.Equal(t, 1.0, row.Weight)
	}
}
func TestInvalidWeightsFailClosed(t *testing.T) {
	for _, raw := range []string{`{"weights":{"officer":-1},"default_weight":1}`, `{"weights":{"officer":0.001},"default_weight":1}`, `{"weights":{"vice_minister":2,"deputy":3},"default_weight":1}`, `{"weights":{},"default_weight":0}`} {
		_, err := parseWeightConfig(raw)
		require.Error(t, err)
	}
}
