package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func TestCreateDepartmentMustMatchAuthorizedScope(t *testing.T) {
	svc, tasks, _, _ := newTestSvc()
	owner, ownDept, otherDept := uuid.New(), uuid.New(), uuid.New()
	tasks.users[owner] = &model.NamedUser{ID: owner, DepartmentID: &ownDept}
	ctx := model.WithViewer(context.Background(), model.Viewer{ID: owner, Scope: &rbac.DataScopeCondition{IsSelf: true}})
	row, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "default department"})
	if err != nil {
		t.Fatal(err)
	}
	stored := tasks.items[uuid.MustParse(row.ID)]
	if stored.DepartmentID == nil || *stored.DepartmentID != ownDept {
		t.Fatalf("missing own department: %+v %v", row, err)
	}
	if _, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "forged department", DepartmentID: otherDept.String()}); err == nil {
		t.Fatal("self scope forged another department")
	}
	if _, err := svc.Create(ctx, uuid.New(), &dto.CreateTaskRequest{Title: "forged operator"}); err == nil {
		t.Fatal("viewer and operator mismatch accepted")
	}
	all := model.WithViewer(context.Background(), model.Viewer{ID: owner, Scope: &rbac.DataScopeCondition{}})
	if _, err := svc.Create(all, owner, &dto.CreateTaskRequest{Title: "authorized cross department", DepartmentID: otherDept.String()}); err != nil {
		t.Fatal(err)
	}
	// A same-department explicit choice needs no broader permission.
	if _, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "own department", DepartmentID: ownDept.String()}); err != nil {
		t.Fatal(err)
	}
	delete(tasks.users, owner)
	if _, err := svc.Create(ctx, owner, &dto.CreateTaskRequest{Title: "unavailable creator"}); err == nil {
		t.Fatal("inactive creator accepted")
	}
}
