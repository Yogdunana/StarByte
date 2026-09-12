package service

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/google/uuid"
)

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func mapType(t model.LeaveType) dto.LeaveTypeResponse {
	return dto.LeaveTypeResponse{
		ID: t.ID.String(), Name: t.Name, Code: t.Code,
		Deductible: t.Deductible, DefaultDays: t.DefaultDays, Description: t.Description,
		Enabled: t.Enabled, SortOrder: t.SortOrder,
	}
}

func mapAttachments(items model.AttachmentList) []dto.Attachment {
	if len(items) == 0 {
		return nil
	}
	out := make([]dto.Attachment, 0, len(items))
	for _, item := range items {
		out = append(out, dto.Attachment{FileID: item.FileID, Name: item.Name, Size: item.Size})
	}
	return out
}

func mapBalance(b model.LeaveBalance) *dto.LeaveBalanceResponse {
	return &dto.LeaveBalanceResponse{
		ID: b.ID.String(), UserID: b.UserID.String(), Year: b.Year,
		TotalDays: b.TotalDays, UsedDays: b.UsedDays, RemainingDays: b.RemainingDays,
		LeaveType: mapType(b.LeaveType),
	}
}

func mapApplication(row *model.ApplicationNamed) *dto.LeaveApplicationResponse {
	out := &dto.LeaveApplicationResponse{
		ID:            row.ID.String(),
		Applicant:     dto.Person{ID: row.ApplicantID.String(), Name: row.ApplicantName},
		LeaveType:     mapType(row.LeaveType),
		StartTime:     formatTime(row.StartTime),
		EndTime:       formatTime(row.EndTime),
		DurationDays:  row.DurationDays,
		Reason:        row.Reason,
		Status:        row.Status,
		WorkflowStage: row.WorkflowStage,
		Attachments:   mapAttachments(row.Attachments),
		ApproveRemark: row.ApproveRemark,
		ApprovedAt:    formatTimePtr(row.ApprovedAt),
		CreatedAt:     formatTime(row.CreatedAt),
		UpdatedAt:     formatTime(row.UpdatedAt),
	}
	if row.WorkflowInstanceID != nil && *row.WorkflowInstanceID != uuid.Nil {
		out.WorkflowInstanceID = row.WorkflowInstanceID.String()
	}
	if row.ApproverID != nil && *row.ApproverID != uuid.Nil {
		out.Approver = &dto.Person{ID: row.ApproverID.String(), Name: row.ApproverName}
	}
	return out
}
