package service

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

func applyMeetingPatch(m *model.Meeting, req *dto.UpdateMeetingRequest) {
	if req.Title != nil {
		m.Title = *req.Title
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.StartTime != nil {
		m.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		m.EndTime = *req.EndTime
	}
	if req.Location != nil {
		m.Location = *req.Location
	}
	if req.OnlineLink != nil {
		m.OnlineLink = *req.OnlineLink
	}
	if req.MeetingType != nil {
		m.MeetingType = *req.MeetingType
	}
}

func rewriteMeetingScope(scope *rbacModel.DataScopeCondition, userID uuid.UUID) *rbacModel.DataScopeCondition {
	if scope == nil || scope.IsEmpty() {
		return scope
	}
	if scope.Query == "1 = 0" {
		return &rbacModel.DataScopeCondition{
			Query: "m.organizer_id = ? OR m.id IN (SELECT meeting_id FROM meeting_attendees WHERE user_id = ?)",
			Args:  []interface{}{userID, userID},
		}
	}
	q := strings.ReplaceAll(scope.Query, "department_id", "u.department_id")
	return &rbacModel.DataScopeCondition{Query: q, Args: scope.Args}
}

func newQRToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(uuid.New().String()))
	}
	return hex.EncodeToString(b)
}

func parseUUIDList(raw []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err == nil {
			out = append(out, id)
		}
	}
	return out
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
