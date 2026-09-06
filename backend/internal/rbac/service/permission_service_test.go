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

func TestPermissionService_Create_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)
	mockRepo.EXPECT().GetByCode(gomock.Any(), "user:read").Return(&model.Permission{Code: "user:read"}, nil)

	_, err := svc.Create(context.Background(), &dto.CreatePermissionRequest{
		Name: "读用户", Code: "user:read", Type: "api",
	})
	require.Error(t, err)
	assert.Equal(t, rbac.ErrCodePermissionCodeExists, err.(*rbac.Error).Code())
}

func TestPermissionService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)
	mockRepo.EXPECT().GetByCode(gomock.Any(), "user:write").Return(nil, nil)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	out, err := svc.Create(context.Background(), &dto.CreatePermissionRequest{
		Name: "写用户", Code: "user:write", Type: "api", Resource: "user", Action: "write",
	})
	require.NoError(t, err)
	assert.Equal(t, "user:write", out.Code)
	assert.Equal(t, "写用户", out.Name)
}

func TestPermissionService_Create_ParentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)
	parentID := uuid.New()
	mockRepo.EXPECT().GetByCode(gomock.Any(), "user:child").Return(nil, nil)
	mockRepo.EXPECT().GetByID(gomock.Any(), parentID).Return(nil, nil)

	_, err := svc.Create(context.Background(), &dto.CreatePermissionRequest{
		Name: "子", Code: "user:child", Type: "button", ParentID: parentID.String(),
	})
	require.Error(t, err)
	assert.Equal(t, rbac.ErrCodePermissionNotFound, err.(*rbac.Error).Code())
}

func TestPermissionService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)
	id := uuid.New()
	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, nil)
	_, err := svc.GetByID(context.Background(), id)
	require.Error(t, err)

	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(&model.Permission{ID: id, Name: "n", Code: "c"}, nil)
	out, err := svc.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "n", out.Name)
}

func TestPermissionService_GetTree(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)
	rootID := uuid.New()
	childID := uuid.New()
	mockRepo.EXPECT().List(gomock.Any()).Return([]model.Permission{
		{ID: rootID, Name: "根", Code: "root", SortOrder: 1},
		{ID: childID, Name: "子", Code: "child", ParentID: &rootID, SortOrder: 1},
	}, nil)
	tree, err := svc.GetTree(context.Background())
	require.NoError(t, err)
	require.Len(t, tree, 1)
	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, "child", tree[0].Children[0].Code)
}

func TestPermissionService_UpdateAndDeleteGuards(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)
	id := uuid.New()
	sys := &model.Permission{ID: id, Name: "sys", Code: "sys", IsSystem: true}
	status := 1
	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(sys, nil)
	_, err := svc.Update(context.Background(), id, &dto.UpdatePermissionRequest{Status: &status})
	require.Error(t, err)

	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(sys, nil)
	err = svc.Delete(context.Background(), id)
	require.Error(t, err)
	assert.Equal(t, rbac.ErrCodeSystemPermissionNoDelete, err.(*rbac.Error).Code())

	plain := &model.Permission{ID: id, Name: "p", Code: "p"}
	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(plain, nil)
	mockRepo.EXPECT().CountChildren(gomock.Any(), id).Return(int64(2), nil)
	err = svc.Delete(context.Background(), id)
	require.Error(t, err)
	assert.Equal(t, rbac.ErrCodePermissionHasChildren, err.(*rbac.Error).Code())
}

func TestPermissionService_Create_InvalidParentAndWithParent(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)

	mockRepo.EXPECT().GetByCode(gomock.Any(), "bad:parent").Return(nil, nil)
	_, err := svc.Create(context.Background(), &dto.CreatePermissionRequest{
		Name: "x", Code: "bad:parent", Type: "api", ParentID: "not-uuid",
	})
	require.Error(t, err)

	parentID := uuid.New()
	mockRepo.EXPECT().GetByCode(gomock.Any(), "ok:child").Return(nil, nil)
	mockRepo.EXPECT().GetByID(gomock.Any(), parentID).Return(&model.Permission{ID: parentID, Code: "p"}, nil)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	out, err := svc.Create(context.Background(), &dto.CreatePermissionRequest{
		Name: "子", Code: "ok:child", Type: "button", ParentID: parentID.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, out.ParentID)
}

func TestPermissionService_UpdateSuccessAndGetTreeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := repo.NewMockPermissionRepo(ctrl)
	svc := NewPermissionService(nil, mockRepo, nil)
	id := uuid.New()
	name := "新名"
	desc := "d"
	sort := 2
	status := 0
	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(&model.Permission{ID: id, Name: "旧", Code: "c"}, nil)
	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	out, err := svc.Update(context.Background(), id, &dto.UpdatePermissionRequest{
		Name: name, Description: &desc, Path: "/p", Icon: "i", SortOrder: &sort, Status: &status,
	})
	require.NoError(t, err)
	assert.Equal(t, "新名", out.Name)

	mockRepo.EXPECT().List(gomock.Any()).Return(nil, assert.AnError)
	_, err = svc.GetTree(context.Background())
	require.Error(t, err)
}
