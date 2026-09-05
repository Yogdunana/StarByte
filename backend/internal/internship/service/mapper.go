package service

import (
	"encoding/json"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	"github.com/google/uuid"
)

func mapInternship(row *model.InternshipWithNames) *dto.InternshipResponse {
	out := &dto.InternshipResponse{
		ID:            row.ID.String(),
		User:          dto.Person{ID: row.UserID.String(), Name: displayName(row.UserName), Avatar: row.UserAvatar},
		Title:         row.Title,
		Organization:  row.Organization,
		Description:   row.Description,
		StartDate:     row.StartDate,
		EndDate:       row.EndDate,
		Status:        row.Status,
		Type:          row.Type,
		Skills:        decodeSkills(row.Skills),
		Achievements:  row.Achievements,
		Report:        row.Report,
		MentorComment: row.MentorComment,
		DurationDays:  durationOf(row.StartDate, row.EndDate),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.MentorID != nil {
		out.Mentor = &dto.Person{ID: row.MentorID.String(), Name: displayName(row.MentorName)}
	}
	if row.SupervisorID != nil {
		out.Supervisor = &dto.Person{ID: row.SupervisorID.String(), Name: displayName(row.SupervisorName)}
	}
	if row.DepartmentID != nil {
		out.Department = &dto.Person{ID: row.DepartmentID.String(), Name: row.DepartmentName}
	}
	return out
}

func displayName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "-"
	}
	return name
}

func encodeSkills(skills []string) string {
	clean := make([]string, 0, len(skills))
	for _, s := range skills {
		s = strings.TrimSpace(s)
		if s != "" {
			clean = append(clean, s)
		}
	}
	raw, err := json.Marshal(clean)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func decodeSkills(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out
	}
	return []string{}
}

func parseUUIDPtr(raw string) *uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func parseOptionalUUIDField(raw *string) (*uuid.UUID, bool) {
	if raw == nil {
		return nil, false
	}
	if strings.TrimSpace(*raw) == "" {
		return nil, true
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, false
	}
	return &id, true
}
