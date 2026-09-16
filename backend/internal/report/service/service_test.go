package service

import (
	"context"
	"errors"
	"testing"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/report/dto"
	"github.com/Yogdunana/StarByte/backend/internal/report/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReportRepo struct {
	rows         []model.Report
	total        int64
	err          error
	lastRequest  *dto.ListReportRequest
	lastViewerID uuid.UUID
	lastScope    *rbacModel.DataScopeCondition
}

func (f *fakeReportRepo) Create(context.Context, *model.Report) error { return nil }
func (f *fakeReportRepo) Update(context.Context, *model.Report) error { return nil }
func (f *fakeReportRepo) GetByID(context.Context, uuid.UUID) (*model.Report, error) {
	return nil, nil
}
func (f *fakeReportRepo) GetByUserAndPeriod(context.Context, uuid.UUID, model.ReportType, time.Time, time.Time) (*model.Report, error) {
	return nil, nil
}
func (f *fakeReportRepo) List(_ context.Context, req *dto.ListReportRequest, viewerID uuid.UUID, scope *rbacModel.DataScopeCondition) ([]model.Report, int64, error) {
	requestCopy := *req
	f.lastRequest = &requestCopy
	f.lastViewerID = viewerID
	f.lastScope = scope
	return f.rows, f.total, f.err
}

func TestListSuccessMapsResponseAndRequest(t *testing.T) {
	viewerID := uuid.New()
	departmentID := uuid.New()
	reportID := uuid.New()
	periodStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	submittedAt := periodEnd.Add(time.Hour)
	createdAt := submittedAt.Add(time.Minute)
	updatedAt := createdAt.Add(time.Minute)
	reports := &fakeReportRepo{
		total: 1,
		rows: []model.Report{{
			ID: reportID, UserID: viewerID, DepartmentID: &departmentID,
			ReportType: model.ReportTypeWeekly, PeriodStart: periodStart, PeriodEnd: periodEnd,
			WorkContent: "work", NextPlan: "plan", ProblemsSuggestions: "none",
			ReviewStatus: model.ReviewStatusPending, SubmittedAt: submittedAt,
			CreatedAt: createdAt, UpdatedAt: updatedAt,
		}},
	}
	scope := &rbacModel.DataScopeCondition{}
	req := &dto.ListReportRequest{
		Page: 2, PageSize: 25, ReportType: "weekly", ReviewStatus: "pending",
		UserID: viewerID.String(), DepartmentID: departmentID.String(),
		PeriodStart: &periodStart, PeriodEnd: &periodEnd,
	}

	list, total, page, pageSize, err := New(reports).List(context.Background(), viewerID, req, scope)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 2, page)
	assert.Equal(t, 25, pageSize)
	require.Len(t, list, 1)
	assert.Equal(t, reportID.String(), list[0].ID)
	assert.Equal(t, viewerID.String(), list[0].UserID)
	require.NotNil(t, list[0].DepartmentID)
	assert.Equal(t, departmentID.String(), *list[0].DepartmentID)
	assert.Equal(t, "weekly", list[0].ReportType)
	assert.Equal(t, periodStart, list[0].PeriodStart)
	assert.Equal(t, periodEnd, list[0].PeriodEnd)
	assert.Equal(t, "work", list[0].WorkContent)
	assert.Equal(t, "plan", list[0].NextPlan)
	assert.Equal(t, "none", list[0].ProblemsSuggestions)
	assert.Equal(t, "pending", list[0].ReviewStatus)
	assert.Equal(t, submittedAt, list[0].SubmittedAt)
	assert.Equal(t, createdAt, list[0].CreatedAt)
	assert.Equal(t, updatedAt, list[0].UpdatedAt)
	assert.Equal(t, req, reports.lastRequest)
	assert.Equal(t, viewerID, reports.lastViewerID)
	assert.Same(t, scope, reports.lastScope)
}

func TestListHandlesNilDepartmentAndEmptyRows(t *testing.T) {
	t.Run("nil department", func(t *testing.T) {
		reports := &fakeReportRepo{rows: []model.Report{{ID: uuid.New(), UserID: uuid.New()}}}
		list, _, _, _, err := New(reports).List(context.Background(), uuid.New(), nil, nil)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Nil(t, list[0].DepartmentID)
	})

	t.Run("empty list is non nil", func(t *testing.T) {
		reports := &fakeReportRepo{}
		list, total, _, _, err := New(reports).List(context.Background(), uuid.New(), nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Empty(t, list)
		assert.Zero(t, total)
	})
}

func TestListWrapsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	reports := &fakeReportRepo{err: wantErr}

	list, total, page, pageSize, err := New(reports).List(context.Background(), uuid.New(), nil, nil)

	assert.Nil(t, list)
	assert.Zero(t, total)
	assert.Zero(t, page)
	assert.Zero(t, pageSize)
	assert.ErrorIs(t, err, wantErr)
	assert.Contains(t, err.Error(), "list reports")
}

func TestListNormalizesPagination(t *testing.T) {
	tests := []struct {
		name                   string
		request                *dto.ListReportRequest
		wantPage, wantPageSize int
	}{
		{name: "nil request", wantPage: 1, wantPageSize: 20},
		{name: "defaults", request: &dto.ListReportRequest{}, wantPage: 1, wantPageSize: 20},
		{name: "caps page size", request: &dto.ListReportRequest{Page: 3, PageSize: 101}, wantPage: 3, wantPageSize: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reports := &fakeReportRepo{}
			_, _, page, pageSize, err := New(reports).List(context.Background(), uuid.New(), tt.request, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantPageSize, pageSize)
			require.NotNil(t, reports.lastRequest)
			assert.Equal(t, tt.wantPage, reports.lastRequest.Page)
			assert.Equal(t, tt.wantPageSize, reports.lastRequest.PageSize)
		})
	}
}

func TestListPassesDataScopeThroughUnchanged(t *testing.T) {
	viewerID := uuid.New()
	departmentID := uuid.New()
	tests := []struct {
		name  string
		scope *rbacModel.DataScopeCondition
	}{
		{name: "nil", scope: nil},
		{name: "all", scope: &rbacModel.DataScopeCondition{}},
		{name: "self", scope: &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}},
		{name: "department", scope: &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{departmentID}}},
		{name: "deny", scope: &rbacModel.DataScopeCondition{Query: "1 = 0"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reports := &fakeReportRepo{}
			_, _, _, _, err := New(reports).List(context.Background(), viewerID, &dto.ListReportRequest{}, tt.scope)
			require.NoError(t, err)
			assert.Same(t, tt.scope, reports.lastScope)
			assert.Equal(t, viewerID, reports.lastViewerID)
		})
	}
}
