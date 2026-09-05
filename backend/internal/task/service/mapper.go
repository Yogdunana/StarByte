package service

import (
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func mapTask(row *model.TaskWithNames, children []model.Task) *dto.TaskResponse {
	out := &dto.TaskResponse{
		ID:              row.ID.String(),
		Title:           row.Title,
		Description:     row.Description,
		Priority:        row.Priority,
		Status:          row.Status,
		Creator:         dto.Person{ID: row.CreatorID.String(), Name: row.CreatorName},
		Children:        mapChildren(children),
		DueDate:         row.DueDate,
		CompletedAt:     row.CompletedAt,
		Tags:            decodeJSONList(row.Tags),
		CommentCount:    row.CommentCount,
		AttachmentCount: row.AttachmentCount,
		Progress:        row.Progress,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.AssigneeID != nil {
		out.Assignee = &dto.Person{ID: row.AssigneeID.String(), Name: row.AssigneeName, Avatar: row.AssigneeAvatar}
	}
	if row.DepartmentID != nil && row.DepartmentName != "" {
		out.Department = &dto.Person{ID: row.DepartmentID.String(), Name: row.DepartmentName}
	}
	if row.ParentID != nil {
		out.Parent = &dto.ParentRef{ID: row.ParentID.String(), Title: row.ParentTitle}
	}
	return out
}

func mapChildren(rows []model.Task) []dto.TaskBrief {
	out := make([]dto.TaskBrief, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.TaskBrief{ID: r.ID.String(), Title: r.Title, Status: r.Status})
	}
	return out
}

func mapLog(row model.TaskLogNamed) dto.LogResponse {
	return dto.LogResponse{
		ID:         row.ID.String(),
		ActionType: row.ActionType,
		OldValue:   row.OldValue,
		NewValue:   row.NewValue,
		Operator:   dto.Person{ID: row.OperatorID.String(), Name: row.OperatorName},
		Comment:    row.Comment,
		CreatedAt:  row.CreatedAt,
	}
}

func mapComment(row model.TaskCommentNamed) dto.CommentResponse {
	return dto.CommentResponse{
		ID:        row.ID.String(),
		TaskID:    row.TaskID.String(),
		UserID:    row.AuthorID.String(),
		User:      dto.Person{ID: row.AuthorID.String(), Name: row.AuthorName},
		Content:   row.Content,
		Mentions:  decodeJSONList(row.Mentions),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func mapAttachment(row model.TaskAttachmentNamed) dto.AttachmentResponse {
	return dto.AttachmentResponse{
		ID:         row.ID.String(),
		TaskID:     row.TaskID.String(),
		FileID:     row.FileID.String(),
		FileName:   row.FileName,
		FilePath:   row.FilePath,
		FileSize:   row.FileSize,
		FileType:   row.FileType,
		UploadedBy: row.UploaderID,
		CreatedAt:  row.CreatedAt,
	}
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
