package feature

import "github.com/Yogdunana/StarByte/backend/internal/feature/model"

// Enabled is the Go helper for other modules: FeatureEnabled(flag, subject).
func Enabled(flag *model.Flag, sub Subject) bool {
	return Evaluate(flag, sub).Enabled
}
