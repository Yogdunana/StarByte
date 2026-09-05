package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCRUDAndMine(t *testing.T) {
	svc, mem := newTestSvc()
	ctx := context.Background()
	owner := uuid.New()
	dept := uuid.New()
	mem.users[owner] = &model.NamedUser{ID: owner, RealName: "张三", DepartmentID: &dept}

	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	created, err := svc.Create(ctx, owner, &dto.CreateInternshipRequest{
		Title: "StarByte 后端开发实习", Organization: "计算机协会技术部",
		Description: "参与后端开发", StartDate: start, Type: 0,
		Skills: []string{"Go", "PostgreSQL"},
	}, nil)
	require.NoError(t, err)
	require.Equal(t, "张三", created.User.Name)
	require.Equal(t, []string{"Go", "PostgreSQL"}, created.Skills)
	require.GreaterOrEqual(t, created.DurationDays, 0)
	id := uuid.MustParse(created.ID)

	list, total, err := svc.List(ctx, owner, &dto.ListInternshipRequest{}, nil)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)

	got, err := svc.Get(ctx, owner, id, nil)
	require.NoError(t, err)
	require.Equal(t, created.Title, got.Title)

	title := "StarByte 全栈实习"
	updated, err := svc.Update(ctx, owner, id, &dto.UpdateInternshipRequest{Title: &title, Skills: []string{"Go", "React"}}, nil)
	require.NoError(t, err)
	require.Equal(t, title, updated.Title)

	mine, err := svc.ListMine(ctx, owner, nil)
	require.NoError(t, err)
	require.Len(t, mine, 1)

	require.NoError(t, svc.Delete(ctx, owner, id, nil))
	_, err = svc.Get(ctx, owner, id, nil)
	require.Error(t, err)
}

func TestCreateRejectsInvalidRangeAndUnknownUser(t *testing.T) {
	svc, mem := newTestSvc()
	ctx := context.Background()
	owner := uuid.New()
	mem.users[owner] = &model.NamedUser{ID: owner, RealName: "李四"}
	start := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(ctx, owner, &dto.CreateInternshipRequest{
		Title: "A", Organization: "B", StartDate: start, EndDate: &end,
	}, nil)
	require.Error(t, err)

	_, err = svc.Create(ctx, owner, &dto.CreateInternshipRequest{
		Title: "A", Organization: "B", StartDate: start, UserID: uuid.New().String(),
	}, nil)
	require.Error(t, err)
}

func TestGetDeniedByScope(t *testing.T) {
	svc, mem := newTestSvc()
	ctx := context.Background()
	owner := uuid.New()
	other := uuid.New()
	dept := uuid.New()
	mem.users[owner] = &model.NamedUser{ID: owner, RealName: "本人", DepartmentID: &dept}
	created, err := svc.Create(ctx, owner, &dto.CreateInternshipRequest{
		Title: "A", Organization: "B", StartDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}, nil)
	require.NoError(t, err)
	_, err = svc.Get(ctx, other, uuid.MustParse(created.ID), &rbacModel.DataScopeCondition{Query: "1 = 0"})
	require.Error(t, err)
	ae, ok := err.(*response.AppError)
	require.True(t, ok)
	require.Equal(t, response.CodeInternshipNoAccess, ae.Code)
}

func TestUpdateClosedDenied(t *testing.T) {
	svc, mem := newTestSvc()
	ctx := context.Background()
	owner := uuid.New()
	mem.users[owner] = &model.NamedUser{ID: owner, RealName: "王五"}
	created, err := svc.Create(ctx, owner, &dto.CreateInternshipRequest{
		Title: "A", Organization: "B", StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}, nil)
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	_, err = svc.Complete(ctx, owner, id, &dto.CompleteRequest{Report: "done"}, nil)
	require.NoError(t, err)
	title := "改不了"
	_, err = svc.Update(ctx, owner, id, &dto.UpdateInternshipRequest{Title: &title}, nil)
	require.Error(t, err)
	ae, ok := err.(*response.AppError)
	require.True(t, ok)
	require.Equal(t, response.CodeInternshipClosed, ae.Code)
}
