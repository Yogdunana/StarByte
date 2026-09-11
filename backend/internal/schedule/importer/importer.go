// Package importer 提供可插拔的日程导入解析器（课表 XLSX、ICS，以及后续 Google / 学校 live sync）。
package importer

import (
	"context"
	"time"
)

// DraftEvent 是尚未写入数据库的导入事件。
type DraftEvent struct {
	Title       string
	Description string
	Location    string
	StartAt     time.Time
	EndAt       time.Time
	AllDay      bool
	ExternalUID string
}

// Meta 描述一次导入的来源信息，用于图层 source_key 与日历命名。
type Meta struct {
	Title        string
	StudentName  string
	StudentNo    string
	AcademicYear string
	Term         string
	SourceKey    string
	CalendarName string
	Timezone     string
}

// Options 是解析选项。
//
// SemesterStart 为学期第 1 周星期一的日历日期（ISO date，无时区）。
// 课表格子里的钟点按 Timezone（默认 Asia/Shanghai）展开。
type Options struct {
	SemesterStart time.Time
	Timezone      *time.Location
	Filename      string
}

// Importer 将原始字节解析为草稿事件。Google 等在线源实现 SyncSource。
type Importer interface {
	Kind() string
	Parse(ctx context.Context, raw []byte, opts Options) ([]DraftEvent, Meta, error)
}

// SyncSource 是在线同步源（Google Calendar、学校 live API）的挂钩。
type SyncSource interface {
	Source() string
	Configured() bool
}

func locationOrShanghai(loc *time.Location) *time.Location {
	if loc != nil {
		return loc
	}
	if tz, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return tz
	}
	return time.FixedZone("CST", 8*3600)
}

func mondayOf(day time.Time) time.Time {
	d := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	offset := int(d.Weekday() - time.Monday)
	if offset < 0 {
		offset += 7
	}
	return d.AddDate(0, 0, -offset)
}
