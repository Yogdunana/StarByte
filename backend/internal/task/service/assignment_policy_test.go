package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

type assignmentStub struct {
	picked   *model.NamedUser
	excluded []uuid.UUID
	policy   model.AssignmentPolicy
	err      error
}

func (s *assignmentStub) Pick(_ context.Context, p model.AssignmentPolicy, excluded []uuid.UUID) (*model.NamedUser, error) {
	s.policy = p
	s.excluded = excluded
	return s.picked, s.err
}
func (s *assignmentStub) Roles(context.Context, string) ([]model.AssignmentRole, error) {
	return []model.AssignmentRole{{ID: uuid.New(), Name: "干事"}}, s.err
}
func TestAssignmentPolicyBoundaries(t *testing.T) {
	dept, role := uuid.New(), uuid.New()
	cases := []struct {
		config     *dto.AutoAssignmentConfig
		department *uuid.UUID
		bad        bool
	}{
		{nil, nil, false}, {&dto.AutoAssignmentConfig{Mode: "manual"}, nil, false},
		{&dto.AutoAssignmentConfig{Mode: "department"}, nil, true},
		{&dto.AutoAssignmentConfig{Mode: "unknown"}, &dept, true},
		{&dto.AutoAssignmentConfig{Mode: "role"}, &dept, true},
		{&dto.AutoAssignmentConfig{Mode: "role", RoleID: "bad"}, &dept, true},
		{&dto.AutoAssignmentConfig{Mode: "role", RoleID: uuid.Nil.String()}, &dept, true},
		{&dto.AutoAssignmentConfig{Mode: "role", RoleID: role.String()}, &dept, false},
		{&dto.AutoAssignmentConfig{Mode: "round_robin"}, &dept, false},
	}
	for _, tc := range cases {
		_, err := assignmentPolicy(tc.config, tc.department)
		if (err != nil) != tc.bad {
			t.Fatalf("config %+v: %v", tc.config, err)
		}
	}
}
func TestAutomaticAssignmentSnapshotAndSignerExclusion(t *testing.T) {
	svc, _, _, ids := workflowFixture(t)
	dept := uuid.New()
	ctx := context.Background()
	policy := &dto.AutoAssignmentConfig{Mode: "round_robin"}
	selected := uuid.New()
	stub := &assignmentStub{picked: &model.NamedUser{ID: selected}}
	svc.assignments = stub
	row := &model.Task{DepartmentID: &dept, ReviewerID: &ids[2], AcceptorID: &ids[3]}
	if err := svc.autoAssign(ctx, row, policy); err != nil {
		t.Fatal(err)
	}
	if row.AssigneeID == nil || *row.AssigneeID != selected || row.AssignmentPolicy == "" || stub.policy.DepartmentID != dept || len(stub.excluded) != 2 || stub.excluded[0] != ids[2] || stub.excluded[1] != ids[3] {
		t.Fatal("selection context or snapshot lost")
	}
	if err := svc.autoAssign(ctx, row, policy); err == nil {
		t.Fatal("manual and automatic assignee both accepted")
	}
	row.AssigneeID = nil
	stub.picked = nil
	if err := svc.autoAssign(ctx, row, policy); err == nil {
		t.Fatal("empty pool silently accepted")
	}
	stub.err = errors.New("assignment store failed")
	if err := svc.autoAssign(ctx, row, policy); !errors.Is(err, stub.err) {
		t.Fatal("store failure hidden")
	}
	roles, err := svc.AssignmentRoles(ctx, "干事")
	if err == nil || roles != nil {
		t.Fatal("roles store failure hidden")
	}
	stub.err = nil
	roles, err = svc.AssignmentRoles(ctx, "干事")
	if err != nil || len(roles) != 1 || roles[0].Name != "干事" {
		t.Fatal("minimal role candidates unavailable")
	}
}
