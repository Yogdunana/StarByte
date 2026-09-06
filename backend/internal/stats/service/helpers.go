package service

import (
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/stats/dto"
	"github.com/Yogdunana/StarByte/backend/internal/stats/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func toQuery(p *dto.StatsQuery) repo.Query {
	if p == nil {
		return repo.Query{}
	}
	g := p.Granularity
	if g == "" {
		g = "month"
	}
	return repo.Query{
		Start: p.StartDate, End: p.EndDate, DepartmentID: p.DepartmentID,
		DeptIDs: p.ScopeDeptIDs, Granularity: g, Denied: p.Denied, AllScope: p.AllScope,
	}
}

func toSeries(name, typ string, buckets []repo.Bucket) dto.DataSeries {
	data := make([]dto.DataPoint, 0, len(buckets))
	x := make([]string, 0, len(buckets))
	for _, b := range buckets {
		label := b.Label
		if label == "" {
			label = b.Key
		}
		data = append(data, dto.DataPoint{Label: label, Value: b.Value})
		x = append(x, label)
	}
	return dto.DataSeries{Name: name, Type: typ, Data: data, XAxis: x}
}

func round4(v float64) float64 {
	return float64(int(v*10000+0.5)) / 10000
}

func validateQuery(q *dto.StatsQuery) error {
	if q == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(q.Granularity)) {
	case "", "day", "week", "month":
	default:
		return response.NewError(response.CodeStatsInvalidParam, "查询参数无效")
	}
	switch strings.ToLower(strings.TrimSpace(q.GroupBy)) {
	case "", "date", "department", "type", "status", "grade":
	default:
		return response.NewError(response.CodeStatsInvalidParam, "查询参数无效")
	}
	if q.StartDate != nil && q.EndDate != nil && q.EndDate.Before(*q.StartDate) {
		return response.NewError(response.CodeStatsInvalidParam, "查询参数无效")
	}
	return nil
}
