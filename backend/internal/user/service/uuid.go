package service

import (
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// parseRequiredUUID parses a required UUID string. Invalid values return a 400 AppError.
func parseRequiredUUID(value, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "参数错误: 无效的"+field)
	}
	return id, nil
}

// parseOptionalUUID parses an optional UUID string. Empty input yields (nil, nil).
func parseOptionalUUID(value, field string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := parseRequiredUUID(value, field)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
