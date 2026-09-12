package service

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
)

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func toDocResponse(row *model.DocNamed, attachments []model.AttachmentNamed, includeBody bool) *dto.DocResponse {
	resp := &dto.DocResponse{
		ID:             row.ID.String(),
		Kind:           row.Kind,
		Slug:           row.Slug,
		Path:           model.PublicPath(row.Kind, row.Slug),
		Title:          row.Title,
		Summary:        row.Summary,
		CategoryName:   row.CategoryName,
		Visibility:     row.Visibility,
		PermissionCode: row.PermissionCode,
		Status:         row.Status,
		Version:        row.Version,
		Author:         dto.Person{ID: row.AuthorID.String(), Name: row.AuthorName},
		CreatedAt:      formatTime(row.CreatedAt),
		UpdatedAt:      formatTime(row.UpdatedAt),
	}
	if includeBody {
		resp.Content = row.Content
	}
	if row.CategoryID != nil {
		resp.CategoryID = row.CategoryID.String()
	}
	if row.PublishedAt != nil {
		resp.PublishedAt = formatTime(*row.PublishedAt)
	}
	if len(attachments) > 0 {
		resp.Attachments = make([]dto.AttachmentResponse, 0, len(attachments))
		for _, a := range attachments {
			resp.Attachments = append(resp.Attachments, dto.AttachmentResponse{
				ID:        a.ID.String(),
				FileID:    a.FileID.String(),
				Name:      a.FileName,
				Size:      a.FileSize,
				CreatedAt: formatTime(a.CreatedAt),
			})
		}
	}
	return resp
}

func toVersionResponse(v *model.VersionNamed, current int, includeBody bool) *dto.VersionResponse {
	resp := &dto.VersionResponse{
		Version:   v.Version,
		Title:     v.Title,
		Summary:   v.Summary,
		Editor:    dto.Person{ID: v.EditorID.String(), Name: v.EditorName},
		CreatedAt: formatTime(v.CreatedAt),
		IsCurrent: v.Version == current,
	}
	if includeBody {
		resp.Content = v.Content
	}
	return resp
}

func toCategoryResponse(c *model.Category) *dto.CategoryResponse {
	resp := &dto.CategoryResponse{
		ID:        c.ID.String(),
		Name:      c.Name,
		Slug:      c.Slug,
		SortOrder: c.SortOrder,
	}
	if c.ParentID != nil {
		resp.ParentID = c.ParentID.String()
	}
	return resp
}
