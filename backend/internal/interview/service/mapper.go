package service

import (
	"github.com/Yogdunana/StarByte/backend/internal/interview/dto"
	"github.com/Yogdunana/StarByte/backend/internal/interview/model"
	"github.com/google/uuid"
)

func mapSession(row *model.SessionWithNames) *dto.SessionResponse {
	out := &dto.SessionResponse{
		ID:             row.ID.String(),
		Title:          row.Title,
		Round:          row.Round,
		DepartmentName: row.DepartmentName,
		StartTime:      row.StartTime,
		EndTime:        row.EndTime,
		Location:       row.Location,
		OnlineLink:     row.OnlineLink,
		Status:         row.Status,
		MaxCandidates:  row.MaxCandidates,
		Description:    row.Description,
		CandidateCount: row.CandidateCount,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
	if row.DepartmentID != nil {
		out.DepartmentID = row.DepartmentID.String()
	}
	return out
}

func mapInterview(row *model.InterviewWithNames, evaluators []dto.Person) *dto.InterviewResponse {
	if evaluators == nil {
		evaluators = []dto.Person{}
	}
	out := &dto.InterviewResponse{
		SessionTitle:    row.SessionTitle,
		Applicant:       dto.Person{ID: row.ApplicantID.String(), Name: row.ApplicantName},
		StudentNo:       row.StudentNo,
		Status:          row.Status,
		ScheduledTime:   row.ScheduledAt,
		ActualStartTime: row.ActualStartTime,
		ActualEndTime:   row.ActualEndTime,
		Result:          row.ResultCode,
		ResultComment:   row.ResultComment,
		Location:        row.Location,
		Duration:        row.Duration,
		Score:           row.Score,
		Evaluators:      evaluators,
		DepartmentName:  row.DepartmentName,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	out.ID = row.ID.String()
	if row.SessionID != nil {
		out.SessionID = row.SessionID.String()
	}
	if row.ApplicationID != nil {
		out.ApplicationID = row.ApplicationID.String()
	}
	return out
}

func groupEvaluators(rows []model.InterviewerNamed) map[uuid.UUID][]dto.Person {
	out := map[uuid.UUID][]dto.Person{}
	for _, r := range rows {
		out[r.InterviewID] = append(out[r.InterviewID], dto.Person{
			ID:   r.InterviewerID.String(),
			Name: r.Name,
		})
	}
	return out
}

func mapDimension(d *model.Dimension) *dto.DimensionResponse {
	return &dto.DimensionResponse{
		ID:        d.ID.String(),
		Name:      d.Name,
		Weight:    d.Weight,
		MaxScore:  d.MaxScore,
		SortOrder: d.SortOrder,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
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

func displayName(u *model.NamedUser) string {
	if u == nil {
		return ""
	}
	if u.RealName != "" {
		return u.RealName
	}
	return u.Username
}
