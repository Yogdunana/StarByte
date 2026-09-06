package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/rbac"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPositionService_Create_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPositionRepo(ctrl)
	svc := NewPositionService(nil, mockRepo)
	mockRepo.EXPECT().GetByCode(gomock.Any(), "intern").Return(&model.Position{Code: "intern"}, nil)

	_, err := svc.Create(context.Background(), &dto.CreatePositionRequest{Name: "实习生", Code: "intern"})
	require.Error(t, err)
	assert.Equal(t, rbac.ErrCodePositionCodeExists, err.(*rbac.Error).Code())
}

func TestPositionService_CreateAndGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPositionRepo(ctrl)
	svc := NewPositionService(nil, mockRepo)
	level, weight, sort := 1, 1.5, 3
	mockRepo.EXPECT().GetByCode(gomock.Any(), "lead").Return(nil, nil)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	out, err := svc.Create(context.Background(), &dto.CreatePositionRequest{
		Name: "组长", Code: "lead", Level: &level, VoteWeight: &weight, SortOrder: &sort,
	})
	require.NoError(t, err)
	assert.Equal(t, "lead", out.Code)
	assert.Equal(t, 1, out.Level)

	id := uuid.New()
	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, nil)
	_, err = svc.GetByID(context.Background(), id)
	require.Error(t, err)

	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(&model.Position{ID: id, Name: "组长", Code: "lead"}, nil)
	got, err := svc.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "组长", got.Name)
}

func TestPositionService_ListUpdateDeleteGuards(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPositionRepo(ctrl)
	svc := NewPositionService(nil, mockRepo)
	id := uuid.New()

	mockRepo.EXPECT().List(gomock.Any(), 1, 10, "").Return([]model.Position{{ID: id, Name: "a", Code: "a"}}, int64(1), nil)
	list, total, err := svc.List(context.Background(), &dto.ListPositionRequest{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)

	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, nil)
	_, err = svc.Update(context.Background(), id, &dto.UpdatePositionRequest{Name: "b"})
	require.Error(t, err)

	status := 1
	desc := "d"
	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(&model.Position{ID: id, Name: "a", Code: "a"}, nil)
	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	out, err := svc.Update(context.Background(), id, &dto.UpdatePositionRequest{Name: "b", Description: &desc, Status: &status})
	require.NoError(t, err)
	assert.Equal(t, "b", out.Name)

	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, nil)
	err = svc.Delete(context.Background(), id)
	require.Error(t, err)
	assert.Equal(t, rbac.ErrCodePositionNotFound, err.(*rbac.Error).Code())
}
