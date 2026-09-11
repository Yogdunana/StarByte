package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApprovalThresholds(t *testing.T) {
	tests := []struct {
		name, kind       string
		ratio            float64
		yes, no, pending int
		pass, fail       bool
	}{
		{"all waits", "all", 0, 1, 0, 1, false, false},
		{"all passes", "all", 0, 2, 0, 0, true, false},
		{"all veto", "all", 0, 0, 1, 1, false, true},
		{"any waits after refusal", "any", 0, 0, 1, 1, false, false},
		{"any passes", "any", 0, 1, 0, 2, true, false},
		{"any impossible", "any", 0, 0, 2, 0, false, true},
		{"ratio rounds up", "ratio", 67, 2, 0, 1, false, false},
		{"ratio sufficient", "ratio", 66, 2, 0, 1, true, false},
		{"ratio impossible", "ratio", 66, 0, 2, 1, false, true},
		{"empty never passes", "all", 0, 0, 0, 0, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, fail := approvalDecision(map[string]interface{}{"approvalType": tt.kind, "passRatio": tt.ratio}, tt.yes, tt.no, tt.pending)
			require.Equal(t, tt.pass, pass)
			require.Equal(t, tt.fail, fail)
		})
	}
	require.Error(t, ValidateApprovalConfig(map[string]interface{}{"approvalType": "ratio", "passRatio": float64(0)}))
}
