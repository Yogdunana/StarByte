package service

import (
	"encoding/json"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/model"
	"github.com/google/uuid"
)

func encodeDepends(ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func decodeDepends(raw string) []string {
	if raw == "" || raw == "[]" {
		return []string{}
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return []string{}
	}
	return ids
}

func parseDepends(raw []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func parseTimePtr(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func toTaskDTO(rec *model.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID: rec.ID.String(), Name: rec.Name, Code: rec.Code,
		CronExpr: rec.CronExpr, RunAt: rec.RunAt, Timezone: rec.Timezone,
		HandlerKey: rec.HandlerKey, Payload: rec.Payload,
		DependsOn: decodeDepends(rec.DependsOn), ShardKey: rec.ShardKey,
		Status: rec.Status, MaxRetries: rec.MaxRetries, TimeoutSec: rec.TimeoutSec,
		RetryCount: rec.RetryCount, NextRunAt: rec.NextRunAt, LastRunAt: rec.LastRunAt,
		LastStatus: rec.LastStatus, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
}

func toRunDTO(rec *model.Run) dto.RunResponse {
	return dto.RunResponse{
		ID: rec.ID.String(), TaskID: rec.TaskID.String(),
		ScheduledAt: rec.ScheduledAt, StartedAt: rec.StartedAt, FinishedAt: rec.FinishedAt,
		Status: rec.Status, Attempt: rec.Attempt, WorkerID: rec.WorkerID,
		ErrorText: rec.ErrorText, Output: rec.Output,
	}
}

func toLogDTO(rec *model.RunLog) dto.LogLine {
	return dto.LogLine{ID: rec.ID.String(), Level: rec.Level, Line: rec.Line, CreatedAt: rec.CreatedAt}
}
