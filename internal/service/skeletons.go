// Package service 业务逻辑层实现 - 骨架实现
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件提供 ReminderDispatcher 和 RecurrenceExpander 的骨架实现
//   1. LogReminderDispatcher: 仅记录日志，开发期使用
//   2. NoopRecurrenceExpander: 不展开 RRULE，返回母事件
//   3. SimpleRecurrenceExpander: 简单展开器（基于 FREQ=DAILY/WEEKLY）
// 接入生产时替换为真实实现
package service

import (
	"context"
	"log/slog"
	"time"

	"schedule-service/internal/model"
)

// ============================================================================
// LogReminderDispatcher 提醒发送骨架实现
// ============================================================================

// LogReminderDispatcher 仅记录日志的提醒发送器
// 开发期使用，生产环境替换为接入实际通知通道的实现
type LogReminderDispatcher struct{}

// NewLogReminderDispatcher 构造日志提醒发送器
func NewLogReminderDispatcher() *LogReminderDispatcher {
	return &LogReminderDispatcher{}
}

// Dispatch 发送提醒：仅记录日志
func (d *LogReminderDispatcher) Dispatch(ctx context.Context, reminder *model.ScheduleReminder, event *model.ScheduleEvent) error {
	slog.Info("reminder dispatched",
		"reminder_id", reminder.ID,
		"event_id", reminder.EventID,
		"user_id", reminder.UserID,
		"method", reminder.RemindMethod,
		"event_title", event.Title,
		"start_time", event.StartTime,
	)
	// 生产实现应按 reminder.RemindMethod 分发到不同通道:
	//   - app: 调用推送服务
	//   - email: 调用邮件服务
	//   - sms: 调用短信网关
	//   - webhook: 发起 HTTP 请求
	return nil
}

// ============================================================================
// NoopRecurrenceExpander 空实现展开器
// ============================================================================

// NoopRecurrenceExpander 不展开 RRULE 的空实现
//   Expand 返回空切片，Validate 仅检查非空
//   适用于: 开发期或无需重复事件展开的场景
type NoopRecurrenceExpander struct{}

// NewNoopRecurrenceExpander 构造空实现展开器
func NewNoopRecurrenceExpander() *NoopRecurrenceExpander {
	return &NoopRecurrenceExpander{}
}

// Expand 不展开，返回空切片
func (e *NoopRecurrenceExpander) Expand(ctx context.Context, rule string, baseStartTime time.Time, from, to time.Time) ([]time.Time, error) {
	return []time.Time{}, nil
}

// Validate 仅检查 RRULE 字符串非空
func (e *NoopRecurrenceExpander) Validate(ctx context.Context, rule string) error {
	if rule == "" {
		return nil // 空字符串表示单次事件，合法
	}
	// 空实现不做严格校验，接入生产时替换
	return nil
}

// ============================================================================
// SimpleRecurrenceExpander 简单 RRULE 展开器
// ============================================================================

// SimpleRecurrenceExpander 简单 RRULE 展开器
//   支持 FREQ=DAILY / FREQ=WEEKLY / FREQ=MONTHLY
//   支持 INTERVAL / COUNT / UNTIL
//   不支持 BYDAY/BYMONTH 等复杂规则
//   生产环境建议接入 github.com/teambition/rrule-go 替换
type SimpleRecurrenceExpander struct{}

// NewSimpleRecurrenceExpander 构造简单展开器
func NewSimpleRecurrenceExpander() *SimpleRecurrenceExpander {
	return &SimpleRecurrenceExpander{}
}

// Validate 校验 RRULE 基本格式
func (e *SimpleRecurrenceExpander) Validate(ctx context.Context, rule string) error {
	if rule == "" {
		return nil
	}
	if !startsWith(rule, "FREQ=") {
		return NewBizError(ErrCodeEventRRULEInvalid, "RRULE must start with FREQ=", nil)
	}
	freq := parseRRULEField(rule, "FREQ")
	switch freq {
	case "DAILY", "WEEKLY", "MONTHLY", "YEARLY":
		return nil
	default:
		return NewBizError(ErrCodeEventRRULEInvalid, "unsupported FREQ value: "+freq, nil)
	}
}

// Expand 在 [from, to] 区间内展开重复事件
//   返回实例的开始时间列表
func (e *SimpleRecurrenceExpander) Expand(ctx context.Context, rule string, baseStartTime time.Time, from, to time.Time) ([]time.Time, error) {
	if rule == "" {
		return []time.Time{}, nil
	}
	// 限制展开区间上限，避免恶意大区间耗尽资源
	days := to.Sub(from).Hours() / 24
	if days > 366 {
		return nil, NewBizError(ErrCodeRecurrenceRangeExceed, "expansion range exceeds 1 year", nil)
	}

	freq := parseRRULEField(rule, "FREQ")
	interval := parseRRULEInterval(rule)
	count := parseRRULECount(rule)
	until := parseRRULEUntil(rule)

	var result []time.Time
	current := baseStartTime
	generated := 0

	// 上限保护：单次最多生成 1000 个实例
	for generated < 1000 {
		// 终止条件 1: 到达 UNTIL
		if !until.IsZero() && current.After(until) {
			break
		}
		// 终止条件 2: 达到 COUNT
		if count > 0 && generated >= count {
			break
		}
		// 终止条件 3: 超出查询区间
		if current.After(to) {
			break
		}
		// 当前实例在区间内则加入结果
		if !current.Before(from) {
			result = append(result, current)
		}
		// 按频率推进
		current = advanceByFreq(current, freq, interval)
		generated++
	}

	return result, nil
}

// advanceByFreq 按频率推进时间
func advanceByFreq(t time.Time, freq string, interval int) time.Time {
	if interval <= 0 {
		interval = 1
	}
	switch freq {
	case "DAILY":
		return t.AddDate(0, 0, interval)
	case "WEEKLY":
		return t.AddDate(0, 0, 7*interval)
	case "MONTHLY":
		return t.AddDate(0, interval, 0)
	case "YEARLY":
		return t.AddDate(interval, 0, 0)
	}
	return t
}

// parseRRULEField 从 RRULE 字符串中解析指定字段值
//   例: parseRRULEField("FREQ=WEEKLY;INTERVAL=1", "FREQ") → "WEEKLY"
func parseRRULEField(rule, field string) string {
	pairs := splitRRULE(rule)
	for _, p := range pairs {
		kv := splitKV(p, "=")
		if len(kv) == 2 && kv[0] == field {
			return kv[1]
		}
	}
	return ""
}

// parseRRULEInterval 解析 INTERVAL，默认 1
func parseRRULEInterval(rule string) int {
	v := parseRRULEField(rule, "INTERVAL")
	if v == "" {
		return 1
	}
	n := atoiSafe(v)
	if n <= 0 {
		return 1
	}
	return n
}

// parseRRULECount 解析 COUNT，0 表示无限制
func parseRRULECount(rule string) int {
	v := parseRRULEField(rule, "COUNT")
	if v == "" {
		return 0
	}
	return atoiSafe(v)
}

// parseRRULEUntil 解析 UNTIL 为 time.Time
//   格式: RFC5545 的 UNTIL=20261231T235959Z
func parseRRULEUntil(rule string) time.Time {
	v := parseRRULEField(rule, "UNTIL")
	if v == "" {
		return time.Time{}
	}
	// 尝试解析 RFC3339 / RFC5545 日期时间
	formats := []string{
		"20060102T150405Z",
		"20060102T150405",
		"20060102",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, v); err == nil {
			return t
		}
	}
	return time.Time{}
}

// splitRRULE 按 ; 分割 RRULE 字符串
func splitRRULE(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ';' {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// splitKV 按 sep 分割 key=value
func splitKV(s, sep string) []string {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}

// startsWith 字符串前缀判断
func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

// atoiSafe 安全字符串转 int
func atoiSafe(s string) int {
	if s == "" {
		return 0
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}
