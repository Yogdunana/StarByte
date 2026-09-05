package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *internshipService) DurationStats(ctx context.Context, viewer uuid.UUID, req *dto.DurationStatsRequest, scope *rbacModel.DataScopeCondition) (*dto.DurationStatsResponse, error) {
	rows, err := s.rows.ListForStats(ctx, req.StartDate, req.EndDate, "", rewriteScope(scope, viewer))
	if err != nil {
		return nil, fmt.Errorf("duration stats: %w", err)
	}
	groupBy := req.GroupBy
	if groupBy != "department" && groupBy != "month" {
		groupBy = "user"
	}
	items := aggregateDuration(rows, groupBy)
	return &dto.DurationStatsResponse{GroupBy: groupBy, Items: items}, nil
}

func (s *internshipService) Ranking(ctx context.Context, viewer uuid.UUID, req *dto.RankingRequest, scope *rbacModel.DataScopeCondition) (*dto.RankingResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.RankingVisible && !isAllScope(scope) {
		return nil, response.NewError(response.CodeInternshipRankHidden, "排行榜不可见")
	}
	var listScope *rbacModel.DataScopeCondition
	if !cfg.RankingVisible {
		listScope = rewriteScope(scope, viewer)
	}
	rows, err := s.rows.ListForStats(ctx, "", "", req.DepartmentID, listScope)
	if err != nil {
		return nil, fmt.Errorf("ranking: %w", err)
	}
	items := buildRanking(rows, req.SortBy)
	total := len(items)
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 15
	}
	if limit < len(items) {
		items = items[:limit]
	}
	return &dto.RankingResponse{Rankings: items, Total: total}, nil
}

func (s *internshipService) DepartmentStats(ctx context.Context, viewer uuid.UUID, req *dto.DepartmentStatsRequest, scope *rbacModel.DataScopeCondition) (*dto.DepartmentStatsResponse, error) {
	rows, err := s.rows.ListForStats(ctx, req.StartDate, req.EndDate, "", rewriteScope(scope, viewer))
	if err != nil {
		return nil, fmt.Errorf("department stats: %w", err)
	}
	return &dto.DepartmentStatsResponse{Items: aggregateDepartment(rows)}, nil
}

func aggregateDuration(rows []model.InternshipWithNames, groupBy string) []dto.DurationItem {
	type acc struct {
		name string
		days int
		n    int64
	}
	bucket := map[string]*acc{}
	for i := range rows {
		row := &rows[i]
		end := time.Now()
		if row.EndDate != nil {
			end = *row.EndDate
		}
		days := CalculateDuration(row.StartDate, end)
		switch groupBy {
		case "month":
			for key, n := range CalculateMonthlyDuration(row.StartDate, end) {
				if bucket[key] == nil {
					bucket[key] = &acc{name: key}
				}
				bucket[key].days += n
				bucket[key].n++
			}
		case "department":
			key := "none"
			name := "未分配部门"
			if row.DepartmentID != nil {
				key = row.DepartmentID.String()
				name = row.DepartmentName
			}
			if bucket[key] == nil {
				bucket[key] = &acc{name: name}
			}
			bucket[key].days += days
			bucket[key].n++
		default:
			key := row.UserID.String()
			if bucket[key] == nil {
				bucket[key] = &acc{name: displayName(row.UserName)}
			}
			bucket[key].days += days
			bucket[key].n++
		}
	}
	out := make([]dto.DurationItem, 0, len(bucket))
	for key, v := range bucket {
		out = append(out, dto.DurationItem{Key: key, Name: v.name, DurationDays: v.days, Count: v.n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].DurationDays == out[j].DurationDays {
			return out[i].Name < out[j].Name
		}
		return out[i].DurationDays > out[j].DurationDays
	})
	return out
}

func buildRanking(rows []model.InternshipWithNames, sortBy string) []dto.RankingItem {
	type acc struct {
		user     dto.Person
		dept     *dto.Person
		days     int
		n        int64
		latest   string
		latestAt time.Time
	}
	bucket := map[string]*acc{}
	for i := range rows {
		row := &rows[i]
		key := row.UserID.String()
		if bucket[key] == nil {
			item := &acc{
				user: dto.Person{ID: key, Name: displayName(row.UserName), Avatar: row.UserAvatar},
			}
			if row.DepartmentID != nil {
				item.dept = &dto.Person{ID: row.DepartmentID.String(), Name: row.DepartmentName}
			}
			bucket[key] = item
		}
		item := bucket[key]
		item.days += durationOf(row.StartDate, row.EndDate)
		item.n++
		if row.CreatedAt.After(item.latestAt) {
			item.latestAt = row.CreatedAt
			item.latest = row.Title
		}
	}
	out := make([]dto.RankingItem, 0, len(bucket))
	for _, v := range bucket {
		out = append(out, dto.RankingItem{
			User: v.user, Department: v.dept,
			TotalDurationDays: v.days, InternshipCount: v.n, LatestInternship: v.latest,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if sortBy == "count" {
			if out[i].InternshipCount == out[j].InternshipCount {
				return out[i].TotalDurationDays > out[j].TotalDurationDays
			}
			return out[i].InternshipCount > out[j].InternshipCount
		}
		if out[i].TotalDurationDays == out[j].TotalDurationDays {
			return out[i].InternshipCount > out[j].InternshipCount
		}
		return out[i].TotalDurationDays > out[j].TotalDurationDays
	})
	for i := range out {
		out[i].Rank = i + 1
	}
	return out
}

func aggregateDepartment(rows []model.InternshipWithNames) []dto.DepartmentStatItem {
	type acc struct {
		dept    *dto.Person
		days    int
		n       int64
		ongoing int64
	}
	bucket := map[string]*acc{}
	for i := range rows {
		row := &rows[i]
		key := "none"
		var dept *dto.Person
		if row.DepartmentID != nil {
			key = row.DepartmentID.String()
			dept = &dto.Person{ID: key, Name: row.DepartmentName}
		}
		if bucket[key] == nil {
			bucket[key] = &acc{dept: dept}
		}
		bucket[key].days += durationOf(row.StartDate, row.EndDate)
		bucket[key].n++
		if row.Status == model.StatusOngoing {
			bucket[key].ongoing++
		}
	}
	out := make([]dto.DepartmentStatItem, 0, len(bucket))
	for _, v := range bucket {
		out = append(out, dto.DepartmentStatItem{Department: v.dept, DurationDays: v.days, Count: v.n, Ongoing: v.ongoing})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DurationDays > out[j].DurationDays })
	return out
}
