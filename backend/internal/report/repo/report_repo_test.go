package repo

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/report/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportRepoGetByIDNotFound(t *testing.T) {
	db, _ := openTestDB(t, queryResult{columns: []string{"id"}})

	got, err := NewReportRepo(db).GetByID(context.Background(), uuid.New())

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestReportRepoReturnsDatabaseError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	db, _ := openTestDB(t, queryResult{err: wantErr})

	got, err := NewReportRepo(db).GetByID(context.Background(), uuid.New())

	assert.Nil(t, got)
	assert.ErrorIs(t, err, wantErr)
}

func TestReportRepoGetByUserAndPeriod(t *testing.T) {
	id, userID := uuid.New(), uuid.New()
	start := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.September, 7, 0, 0, 0, 0, time.UTC)
	db, _ := openTestDB(t, queryResult{
		columns: []string{"id", "user_id", "report_type", "period_start", "period_end"},
		rows:    [][]driver.Value{{id.String(), userID.String(), "weekly", start, end}},
	})

	got, err := NewReportRepo(db).GetByUserAndPeriod(
		context.Background(), userID, "weekly", start, end,
	)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, start, got.PeriodStart)
	assert.Equal(t, end, got.PeriodEnd)
}

func TestReportRepoListPaginationAndSorting(t *testing.T) {
	firstID, secondID := uuid.New(), uuid.New()
	start := time.Date(2026, time.September, 8, 0, 0, 0, 0, time.UTC)
	db, backend := openTestDB(t,
		countResult(3),
		queryResult{
			columns: []string{"id", "period_start"},
			rows: [][]driver.Value{
				{firstID.String(), start},
				{secondID.String(), start.AddDate(0, 0, -7)},
			},
		},
	)

	reports, total, err := NewReportRepo(db).List(
		context.Background(),
		&dto.ListReportRequest{Page: 2, PageSize: 2},
		uuid.New(),
		&rbacModel.DataScopeCondition{},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, reports, 2)
	assert.Equal(t, firstID, reports[0].ID)
	assert.Equal(t, secondID, reports[1].ID)
	queries := backend.captured(t)
	require.Len(t, queries, 2)
	assert.Contains(t, queries[1].sql, "ORDER BY period_start DESC, created_at DESC")
	assert.Contains(t, queries[1].sql, "LIMIT")
	assert.Contains(t, queries[1].sql, "OFFSET")
	assert.Equal(t, []string{"2", "2"}, argumentStrings(queries[1].args))
}

func TestReportRepoListTimeIntersectionFilter(t *testing.T) {
	from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	db, backend := openTestDB(t, countResult(0), emptyReportResult())

	_, _, err := NewReportRepo(db).List(
		context.Background(),
		&dto.ListReportRequest{PeriodStart: &from, PeriodEnd: &to},
		uuid.New(),
		&rbacModel.DataScopeCondition{},
	)

	require.NoError(t, err)
	queries := backend.captured(t)
	require.Len(t, queries, 2)
	assert.Contains(t, queries[0].sql, "period_end >=")
	assert.Contains(t, queries[0].sql, "period_start <=")
	assert.Equal(t, []string{from.String(), to.String()}, argumentStrings(queries[0].args))
}

func TestReportRepoListNilScopeFailClosed(t *testing.T) {
	from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	db, backend := openTestDB(t, countResult(0), emptyReportResult())

	reports, total, err := NewReportRepo(db).List(
		context.Background(),
		&dto.ListReportRequest{ReportType: "weekly", PeriodStart: &from},
		uuid.New(),
		nil,
	)

	require.NoError(t, err)
	assert.Empty(t, reports)
	assert.Equal(t, int64(0), total)
	queries := backend.captured(t)
	require.Len(t, queries, 2)
	assert.Contains(t, queries[0].sql, "WHERE")
	assert.Contains(t, queries[0].sql, "1 = 0")
	assert.Contains(t, queries[0].sql, "report_type =")
	assert.Contains(t, queries[1].sql, "WHERE")
	assert.Contains(t, queries[1].sql, "1 = 0")
}

func TestReportRepoListDataScopes(t *testing.T) {
	viewerID := uuid.New()
	departmentID := uuid.New()
	tests := []struct {
		name        string
		scope       *rbacModel.DataScopeCondition
		wantSQL     string
		wantArgs    []string
		wantNoWhere bool
	}{
		{name: "all", scope: &rbacModel.DataScopeCondition{}, wantArgs: []string{}, wantNoWhere: true},
		{
			name:     "self",
			scope:    &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true},
			wantSQL:  "user_id =",
			wantArgs: []string{viewerID.String()},
		},
		{
			name:     "department",
			scope:    &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{departmentID}},
			wantSQL:  "department_id =",
			wantArgs: []string{departmentID.String()},
		},
		{name: "deny", scope: &rbacModel.DataScopeCondition{Query: "1 = 0"}, wantSQL: "1 = 0", wantArgs: []string{}},
		{name: "nil fail-closed", scope: nil, wantSQL: "1 = 0", wantArgs: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, backend := openTestDB(t, countResult(0), emptyReportResult())
			_, _, err := NewReportRepo(db).List(context.Background(), &dto.ListReportRequest{}, viewerID, tt.scope)
			require.NoError(t, err)
			queries := backend.captured(t)
			require.Len(t, queries, 2)
			if tt.wantNoWhere {
				assert.NotContains(t, queries[0].sql, "WHERE")
			} else {
				assert.Contains(t, queries[0].sql, tt.wantSQL)
			}
			assert.Equal(t, tt.wantArgs, argumentStrings(queries[0].args))
		})
	}
}

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{name: "defaults", wantPage: 1, wantPageSize: 20},
		{name: "keeps valid values", page: 2, pageSize: 50, wantPage: 2, wantPageSize: 50},
		{name: "caps page size", page: 1, pageSize: 101, wantPage: 1, wantPageSize: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, pageSize := normalizePage(tt.page, tt.pageSize)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantPageSize, pageSize)
		})
	}
}

func countResult(total int64) queryResult {
	return queryResult{columns: []string{"count"}, rows: [][]driver.Value{{total}}}
}

func emptyReportResult() queryResult {
	return queryResult{columns: []string{"id"}}
}

func argumentStrings(args []driver.NamedValue) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, fmt.Sprint(arg.Value))
	}
	return out
}
