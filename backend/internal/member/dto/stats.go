package dto

import (
	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

// StatsQuery 统计查询。
type StatsQuery struct {
	ViewerID  uuid.UUID                     `form:"-" json:"-"`
	Scope     *rbacModel.DataScopeCondition `form:"-" json:"-"`
	StartDate string                        `form:"start_date"`
	EndDate   string                        `form:"end_date"`
	GroupBy   string                        `form:"group_by"`
}

// StatItem 统计桶。
type StatItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// StatsResponse 统计结果。
type StatsResponse struct {
	GroupBy string     `json:"group_by"`
	Items   []StatItem `json:"items"`
}
