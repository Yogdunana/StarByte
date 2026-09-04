package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubmit_Duplicate(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	userID := uuid.New()
	apps.On("HasOpenApplication", mock.Anything, userID).Return(true, nil)

	_, err := svc.Submit(context.Background(), userID, &dto.SubmitApplicationRequest{
		ApplicantType: 1, RealName: "张三", StudentNo: "2024001",
		Reason: "加入", ContactPhone: "13800000000", ContactEmail: "a@b.com",
	})
	requireAppError(t, err, response.CodeMemberAppDuplicate, "已有待处理的入会申请")
}

func TestSubmit_OK(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	userID := uuid.New()
	apps.On("HasOpenApplication", mock.Anything, userID).Return(false, nil)
	apps.On("Create", mock.Anything, mock.AnythingOfType("*model.MemberApplication")).Return(nil)
	apps.On("CreateHistory", mock.Anything, mock.AnythingOfType("*model.ApplicationHistory")).Return(nil)
	apps.On("GetByIDWithNames", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&model.ApplicationWithNames{
		MemberApplication: model.MemberApplication{
			ID: uuid.New(), UserID: userID, Type: 1, RealName: "张三", Status: model.AppPending,
		},
	}, nil)

	out, err := svc.Submit(context.Background(), userID, &dto.SubmitApplicationRequest{
		ApplicantType: 1, RealName: "张三", StudentNo: "2024001",
		Reason: "加入", Skills: []string{"Go"}, ContactPhone: "13800000000", ContactEmail: "a@b.com",
	})
	require.NoError(t, err)
	require.Equal(t, "张三", out.RealName)
}

func TestResubmit_WrongOwner(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	id := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, UserID: uuid.New(), Status: model.AppSupplement, Type: 1,
	}, nil)
	_, err := svc.Resubmit(context.Background(), uuid.New(), id, &dto.ResubmitApplicationRequest{Experience: "作品"})
	requireAppError(t, err, response.CodeMemberProfileDenied, "无权操作该申请")
}

func TestResubmit_BackToPending(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	userID := uuid.New()
	id := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, UserID: userID, Status: model.AppSupplement, Type: 1, RequiredFields: model.JSONStrings{"experience"},
	}, nil)
	apps.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberApplication")).Return(nil)
	apps.On("CreateHistory", mock.Anything, mock.AnythingOfType("*model.ApplicationHistory")).Return(nil)
	apps.On("GetByIDWithNames", mock.Anything, id).Return(&model.ApplicationWithNames{
		MemberApplication: model.MemberApplication{ID: id, UserID: userID, Status: model.AppPending, Type: 1},
	}, nil)

	out, err := svc.Resubmit(context.Background(), userID, id, &dto.ResubmitApplicationRequest{Experience: "作品集"})
	require.NoError(t, err)
	require.Equal(t, model.AppPending, out.Status)
}

func TestGetApplication_NotFound(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	id := uuid.New()
	apps.On("GetByIDWithNames", mock.Anything, id).Return(nil, nil)
	_, err := svc.GetApplication(context.Background(), uuid.New(), id, nil)
	requireAppError(t, err, response.CodeMemberAppNotFound, "申请不存在")
}
