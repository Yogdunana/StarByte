package feature

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
)

func TestEnabledHelper(t *testing.T) {
	if Enabled(nil, Subject{}) {
		t.Fatal("nil")
	}
	if !Enabled(&model.Flag{FlagType: model.TypeBoolean, Enabled: true}, Subject{}) {
		t.Fatal("boolean")
	}
}
