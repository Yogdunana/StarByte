package dto

import (
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

func TestToFlagAndAudit(t *testing.T) {
	actor := uuid.New()
	id := uuid.New()
	flag := &model.Flag{
		ID: id, FlagKey: "cms.public", Name: "CMS", FlagType: model.TypeBoolean,
		CreatedBy: &actor, UpdatedBy: &actor, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	out := ToFlag(flag)
	if out.ID != id.String() || out.CreatedBy != actor.String() || out.UpdatedBy != actor.String() {
		t.Fatalf("%+v", out)
	}
	audit := &model.Audit{ID: uuid.New(), FlagID: &id, FlagKey: "cms.public", Action: "toggle", ActorID: &actor, CreatedAt: time.Now()}
	a := ToAudit(audit)
	if a.FlagID != id.String() || a.ActorID != actor.String() {
		t.Fatalf("%+v", a)
	}
	empty := ToAudit(&model.Audit{ID: uuid.New(), FlagKey: "x", Action: "create"})
	if empty.FlagID != "" || empty.ActorID != "" {
		t.Fatalf("%+v", empty)
	}
}
