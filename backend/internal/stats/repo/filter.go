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

// truncExpr 按天/周/月分组。
//
// 传进来的列都是 timestamp（不带时区），里面存的是 **UTC 墙钟**：
// pgx 写入时会把 time.Time 转 UTC（discardTimeZone），读取也按 UTC 解释。
// 直接 date_trunc 就变成按 UTC 分天 ——「每天」的分界落在北京时间早上 8 点，
// 统计口径是歪的。所以先把 UTC 墙钟折成北京时间墙钟再截断。
//
// 写成两段 AT TIME ZONE 而不是 + interval '8 hours'，是为了让读代码的人
// 一眼看出基准是什么，也免得以后真换时区时改错常量。
func truncExpr(col, granularity string) string {
	g := strings.ToLower(strings.TrimSpace(granularity))
	unit := "month"
	switch g {
	case "day":
		unit = "day"
	case "week":
		unit = "week"
	}
	return "to_char(date_trunc('" + unit + "', " + shanghaiWallClock(col) + "), 'YYYY-MM-DD')"
}

// shanghaiWallClock 把「UTC 墙钟」的裸 timestamp 折成北京时间墙钟。
//
// 两段 AT TIME ZONE 的顺序不能反：先把裸值按 UTC 解释成一个确定的时刻
// （裸 timestamp 本身不携带时区，不先钉死就无从谈转换），再取它在东八区的墙钟。
// 不用 + interval '8 hours' 是为了让基准显式可读，也免得以后真换时区时漏改常量。
//
// 业务口径是北京时间，与 internal/leave/service/duration.go 的 bizTimezone 一致。
func shanghaiWallClock(col string) string {
	return col + " AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai'"
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

func dateLit(t time.Time) string {
	return t.Format("2006-01-02")
}

func exclusiveDate(t time.Time) time.Time {
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return d.AddDate(0, 0, 1)
}

func internEndExclusiveSQL() string {
	return "(COALESCE(i.end_date, CURRENT_DATE) + 1)"
}

func clipStartSQL(q Query) string {
	if q.Start == nil {
		return "i.start_date"
	}
	return "GREATEST(i.start_date, DATE '" + dateLit(*q.Start) + "')"
}

func clipEndSQL(q Query) string {
	intern := internEndExclusiveSQL()
	if q.End == nil {
		return intern
	}
	return "LEAST(" + intern + ", DATE '" + dateLit(exclusiveDate(*q.End)) + "')"
}

func clippedDaysSQL(q Query) string {
	return "GREATEST(0, (" + clipEndSQL(q) + " - " + clipStartSQL(q) + "))"
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
