package repo

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Query 仓储层筛选。
type Query struct {
	Start        *time.Time
	End          *time.Time
	DepartmentID *uuid.UUID
	DeptIDs      []uuid.UUID
	Granularity  string
	Denied       bool
	AllScope     bool
	HideRanking  bool
}

// Bucket 分组桶。
type Bucket struct {
	Key   string
	Label string
	Value float64
}

func truncExpr(col, granularity string) string {
	g := strings.ToLower(strings.TrimSpace(granularity))
	unit := "month"
	switch g {
	case "day":
		unit = "day"
	case "week":
		unit = "week"
	}
	return "to_char(date_trunc('" + unit + "', " + col + "), 'YYYY-MM-DD')"
}

func applyRange(db *gorm.DB, col string, q Query) *gorm.DB {
	if q.Start != nil {
		db = db.Where(col+" >= ?", *q.Start)
	}
	if q.End != nil {
		db = db.Where(col+" <= ?", *q.End)
	}
	return db
}

func applyDept(db *gorm.DB, col string, q Query) *gorm.DB {
	if q.Denied {
		return db.Where("1 = 0")
	}
	if q.DepartmentID != nil {
		db = db.Where(col+" = ?", *q.DepartmentID)
	}
	if len(q.DeptIDs) > 0 {
		db = db.Where(col+" IN ?", q.DeptIDs)
	}
	return db
}

func applyOverlap(db *gorm.DB, startCol, endCol string, q Query) *gorm.DB {
	if q.Start != nil {
		db = db.Where("("+endCol+" IS NULL OR "+endCol+" >= ?)", *q.Start)
	}
	if q.End != nil {
		db = db.Where(startCol+" <= ?", *q.End)
	}
	return db
}

func scanBuckets(db *gorm.DB) ([]Bucket, error) {
	var rows []struct {
		K string  `gorm:"column:k"`
		L string  `gorm:"column:l"`
		V float64 `gorm:"column:v"`
	}
	if err := db.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]Bucket, 0, len(rows))
	for _, r := range rows {
		label := r.L
		if label == "" {
			label = r.K
		}
		out = append(out, Bucket{Key: r.K, Label: label, Value: r.V})
	}
	return out, nil
}
