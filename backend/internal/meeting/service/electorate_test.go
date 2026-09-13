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
			require.Equal(t, 2.0, row.Weight)
		} else {
			require.Equal(t, 2.0, row.Weight)
		}
	}
	for _, row := range buildElectorate(uuid.New(), candidates, cfg, model.VoteEqual) {
		require.Equal(t, 1.0, row.Weight)
	}
}
func TestInvalidWeightsFailClosed(t *testing.T) {
	for _, raw := range []string{`{"weights":{"officer":-1},"default_weight":1}`, `{"weights":{"officer":0.001},"default_weight":1}`, `{"weights":{"vice_minister":2,"deputy":3},"default_weight":1}`, `{"weights":{},"default_weight":-1}`} {
		_, err := parseWeightConfig(raw)
		require.Error(t, err)
	}
}

func TestCharterRoleOnlyWeightAndNoTechnicalVotes(t *testing.T) {
	user, admin, member := uuid.New(), uuid.New(), uuid.New()
	rows := buildElectorate(uuid.New(), []model.ElectorCandidate{{UserID: user, RoleCode: "president"}, {UserID: user, RoleCode: "minister"}, {UserID: user, RoleCode: "center_director"}, {UserID: admin, RoleCode: "super_admin"}, {UserID: member, RoleCode: "member"}}, DefaultWeightConfig(), model.VoteWeighted)
	require.Len(t, rows, 1)
	require.Equal(t, user, rows[0].UserID)
	require.Equal(t, 2.0, rows[0].Weight)
	cfg, err := parseWeightConfig(`{"weights":{"officer":0.25},"default_weight":0}`)
	require.NoError(t, err)
	require.Zero(t, cfg.DefaultWeight)
}
