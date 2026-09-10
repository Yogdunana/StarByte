package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *taskService) Candidates(ctx context.Context, keyword string) ([]dto.Person, error) {
	keyword = strings.TrimSpace(keyword)
	if utf8.RuneCountInString(keyword) > 100 {
		return nil, response.NewError(response.CodeBadRequest, "搜索词过长")
	}
	rows, err := s.tasks.SearchUsers(ctx, keyword)
	if err != nil {
		return nil, err
	}
	out := make([]dto.Person, 0, len(rows))
	for _, u := range rows {
		out = append(out, dto.Person{ID: u.ID.String(), Name: displayName(&u)})
	}
	return out, nil
}
