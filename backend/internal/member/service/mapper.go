package service

import (
	"encoding/json"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/google/uuid"
)

func mapApplication(row *model.ApplicationWithNames) *dto.ApplicationResponse {
	if row == nil {
		return nil
	}
	out := &dto.ApplicationResponse{
		ID:             row.ID.String(),
		UserID:         row.UserID.String(),
		Username:       row.Username,
		ApplicantType:  row.Type,
		RealName:       row.RealName,
		StudentNo:      row.StudentNo,
		DepartmentName: row.DepartmentName,
		Reason:         row.Reason,
		Skills:         []string(row.Skills),
		Experience:     row.Experience,
		ContactPhone:   row.ContactPhone,
		ContactEmail:   row.ContactEmail,
		Status:         row.Status,
		CurrentStage:   row.CurrentStage,
		ReviewComment:  row.ReviewComment,
		RequiredFields: []string(row.RequiredFields),
		ReviewedAt:     row.ReviewedAt,
		SubmittedAt:    row.SubmittedAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
	if out.Skills == nil {
		out.Skills = []string{}
	}
	if row.DepartmentID != nil {
		out.DepartmentID = row.DepartmentID.String()
	}
	if row.FlowInstanceID != nil {
		out.FlowInstanceID = row.FlowInstanceID.String()
	}
	if row.ReviewerID != nil {
		out.Reviewer = &dto.ReviewerInfo{ID: row.ReviewerID.String(), Name: row.ReviewerName}
	}
	return out
}

func mapApplications(rows []model.ApplicationWithNames) []*dto.ApplicationResponse {
	out := make([]*dto.ApplicationResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapApplication(&rows[i]))
	}
	return out
}

func mapProfile(row *model.ProfileWithNames) *dto.ProfileResponse {
	if row == nil {
		return nil
	}
	projects := make([]dto.ProjectItem, 0, len(row.Projects))
	for _, p := range row.Projects {
		projects = append(projects, dto.ProjectItem{Name: p.Name, Role: p.Role, Period: p.Period})
	}
	out := &dto.ProfileResponse{
		ID:           row.ID.String(),
		UserID:       row.UserID.String(),
		Username:     row.Username,
		RealName:     row.RealName,
		StudentNo:    row.StudentNo,
		Gender:       row.Gender,
		Grade:        row.Grade,
		Major:        row.Major,
		MemberType:   row.MemberType,
		Status:       row.Status,
		JoinDate:     row.JoinDate,
		LeaveDate:    row.LeaveDate,
		Skills:       []string(row.Skills),
		Projects:     projects,
		Bio:          row.Bio,
		ContactPhone: row.ContactPhone,
		ContactEmail: row.ContactEmail,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if out.Skills == nil {
		out.Skills = []string{}
	}
	if row.DepartmentID != nil {
		out.Department = &dto.NamedRef{ID: row.DepartmentID.String(), Name: row.DepartmentName}
	}
	if row.PositionID != nil {
		out.Position = &dto.NamedRef{ID: row.PositionID.String(), Name: row.PositionName}
	}
	return out
}

func mapProfiles(rows []model.ProfileWithNames) []*dto.ProfileResponse {
	out := make([]*dto.ProfileResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapProfile(&rows[i]))
	}
	return out
}

func parseOptionalUUID(raw string) *uuid.UUID {
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func jsonExtra(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
