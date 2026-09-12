package service

import (
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// RetentionCutoff is the exclusive lower bound for keepable artifacts.
func RetentionCutoff(now time.Time, days int) time.Time {
	if days < 1 {
		days = model.DefaultRetentionDays
	}
	return now.AddDate(0, 0, -days)
}

// ValidatePolicyFields checks retention + cron + timezone without persisting.
func ValidatePolicyFields(retentionDays int, cronExpr, timezone string) error {
	if retentionDays < 1 || retentionDays > 3650 {
		return errPolicy("保留天数须在 1–3650 之间")
	}
	if strings.TrimSpace(cronExpr) == "" {
		return errPolicy("Cron 表达式不能为空")
	}
	if _, err := cronParser.Parse(cronExpr); err != nil {
		return errPolicy("Cron 表达式无效: " + err.Error())
	}
	if tz := strings.TrimSpace(timezone); tz != "" {
		if _, err := time.LoadLocation(tz); err != nil {
			return errPolicy("时区无效: " + tz)
		}
	}
	return nil
}

func nextRunAt(cronExpr, timezone string, from time.Time) *time.Time {
	sched, err := cronParser.Parse(cronExpr)
	if err != nil {
		return nil
	}
	loc := time.Local
	if timezone == "" {
		timezone = model.DefaultTimezone
	}
	if loaded, err := time.LoadLocation(timezone); err == nil {
		loc = loaded
	}
	n := sched.Next(from.In(loc))
	return &n
}

func applyPolicyPatch(cur *model.Policy, req *dto.UpdatePolicyRequest) error {
	nextDays := cur.RetentionDays
	nextCron := cur.CronExpr
	nextTZ := cur.Timezone
	if req.RetentionDays != nil {
		nextDays = *req.RetentionDays
	}
	if req.CronExpr != nil {
		nextCron = strings.TrimSpace(*req.CronExpr)
	}
	if req.Timezone != nil {
		nextTZ = strings.TrimSpace(*req.Timezone)
	}
	if nextTZ == "" {
		nextTZ = model.DefaultTimezone
	}
	if err := ValidatePolicyFields(nextDays, nextCron, nextTZ); err != nil {
		return err
	}
	if req.Enabled != nil {
		cur.Enabled = *req.Enabled
	}
	cur.RetentionDays = nextDays
	cur.CronExpr = nextCron
	cur.Timezone = nextTZ
	return nil
}

// CanTransition reports whether a job may move from -> to.
func CanTransition(from, to int16) bool {
	switch from {
	case model.StatusPending:
		return to == model.StatusRunning || to == model.StatusFailed
	case model.StatusRunning:
		return to == model.StatusSuccess || to == model.StatusFailed
	case model.StatusSuccess, model.StatusRestored, model.StatusRestoreFailed:
		return to == model.StatusRestoring
	case model.StatusRestoring:
		return to == model.StatusRestored || to == model.StatusRestoreFailed
	default:
		return false
	}
}

func applyTransition(rec *model.Record, to int16, now time.Time, errText string) error {
	if rec == nil {
		return errNotFound()
	}
	if !CanTransition(rec.Status, to) {
		return errInvalidState("当前状态不允许该操作")
	}
	rec.Status = to
	rec.UpdatedAt = now
	switch to {
	case model.StatusRunning, model.StatusRestoring:
		rec.StartedAt = &now
		rec.FinishedAt = nil
		rec.ErrorMessage = ""
	case model.StatusSuccess, model.StatusRestored:
		rec.FinishedAt = &now
		rec.ErrorMessage = ""
	case model.StatusFailed, model.StatusRestoreFailed:
		rec.FinishedAt = &now
		rec.ErrorMessage = errText
	}
	return nil
}

func validateRestoreConfirm(req *dto.RestoreRequest) error {
	if req == nil || !req.Confirm || strings.TrimSpace(req.Confirmation) != model.RestoreConfirmToken {
		return errConfirm()
	}
	return nil
}

func validateDrillConfirm(req *dto.DrillRequest) error {
	if req == nil || !req.Confirm || strings.TrimSpace(req.Confirmation) != model.DrillConfirmToken {
		return errDrillConfirm()
	}
	return nil
}
