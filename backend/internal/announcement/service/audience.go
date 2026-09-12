package service

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func parseAudience(typ string, ids []string) (string, model.IDList, error) {
	typ = model.NormalizeAudience(strings.TrimSpace(typ))
	if !model.ValidAudience(typ) {
		return "", nil, invalidAudience()
	}
	clean := make(model.IDList, 0, len(ids))
	seen := map[string]struct{}{}
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, err := uuid.Parse(id); err != nil {
			return "", nil, invalidAudience()
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		clean = append(clean, id)
	}
	if typ != model.AudienceAll && len(clean) == 0 {
		return "", nil, invalidAudience()
	}
	if typ == model.AudienceAll {
		clean = model.IDList{}
	}
	if len(clean) > model.MaxAudienceIDs {
		return "", nil, invalidAudience()
	}
	return typ, clean, nil
}

func parseAttachments(items []dto.Attachment) (model.AttachmentList, error) {
	if len(items) > model.MaxAttachments {
		return nil, response.NewError(response.CodeBadRequest, "附件数量超出限制")
	}
	out := make(model.AttachmentList, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		id := strings.TrimSpace(item.FileID)
		if id == "" {
			continue
		}
		if _, err := uuid.Parse(id); err != nil {
			return nil, response.NewError(response.CodeBadRequest, "附件文件ID不合法")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = id
		}
		if len(name) > 255 {
			name = name[:255]
		}
		size := item.Size
		if size < 0 {
			size = 0
		}
		out = append(out, model.Attachment{FileID: id, Name: name, Size: size})
	}
	return out, nil
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(ids))
	for _, raw := range ids {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, invalidAudience()
		}
		out = append(out, id)
	}
	return out, nil
}

func (s *announcementService) recipients(ctx context.Context, a *model.Announcement) ([]uuid.UUID, error) {
	var (
		ids []uuid.UUID
		err error
	)
	switch model.NormalizeAudience(a.AudienceType) {
	case model.AudienceUsers:
		parsed, perr := parseUUIDs(a.AudienceIDs)
		if perr != nil {
			return nil, perr
		}
		ids, err = s.rows.ListActiveUserIDsAmong(ctx, parsed)
	case model.AudienceDepartment:
		parsed, perr := parseUUIDs(a.AudienceIDs)
		if perr != nil {
			return nil, perr
		}
		ids, err = s.rows.ListActiveUserIDsByDepartments(ctx, parsed)
	case model.AudienceRole:
		parsed, perr := parseUUIDs(a.AudienceIDs)
		if perr != nil {
			return nil, perr
		}
		ids, err = s.rows.ListActiveUserIDsByRoles(ctx, parsed)
	default:
		ids, err = s.rows.ListActiveUserIDs(ctx)
	}
	if err != nil {
		return nil, err
	}
	return ensureAuthor(ids, a.AuthorID), nil
}

func ensureAuthor(ids []uuid.UUID, author uuid.UUID) []uuid.UUID {
	for _, id := range ids {
		if id == author {
			return ids
		}
	}
	return append(ids, author)
}

func (s *announcementService) canViewPublished(ctx context.Context, viewer Viewer, a *model.Announcement) (bool, error) {
	if viewer.CanManage || isAuthor(viewer, a) {
		return true, nil
	}
	return s.rows.UserInAudience(ctx, viewer.UserID, a.AudienceType, a.AudienceIDs)
}
