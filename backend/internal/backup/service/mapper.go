package service

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/google/uuid"
)

func toRecordDTO(r *model.Record) *dto.Record {
	if r == nil {
		return nil
	}
	out := &dto.Record{
		ID:             r.ID.String(),
		TriggerSource:  r.TriggerSource,
		Status:         r.Status,
		Storage:        r.Storage,
		ObjectKey:      r.ObjectKey,
		Filename:       r.Filename,
		ChecksumSHA256: r.ChecksumSHA256,
		SizeBytes:      r.SizeBytes,
		ErrorMessage:   r.ErrorMessage,
		CreatedAt:      r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      r.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if r.StartedAt != nil {
		s := r.StartedAt.UTC().Format(time.RFC3339)
		out.StartedAt = &s
	}
	if r.FinishedAt != nil {
		s := r.FinishedAt.UTC().Format(time.RFC3339)
		out.FinishedAt = &s
	}
	if r.CreatedBy != nil {
		out.CreatedBy = r.CreatedBy.String()
	}
	return out
}

func toRecordDTOs(rows []model.Record) []dto.Record {
	out := make([]dto.Record, 0, len(rows))
	for i := range rows {
		out = append(out, *toRecordDTO(&rows[i]))
	}
	return out
}

func toPolicyDTO(p *model.Policy) *dto.Policy {
	if p == nil {
		return &dto.Policy{
			Enabled:       false,
			RetentionDays: model.DefaultRetentionDays,
			CronExpr:      model.DefaultCronExpr,
			Timezone:      model.DefaultTimezone,
		}
	}
	return &dto.Policy{
		Enabled:       p.Enabled,
		RetentionDays: p.RetentionDays,
		CronExpr:      p.CronExpr,
		Timezone:      p.Timezone,
		UpdatedAt:     p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func ptrUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	v := id
	return &v
}
