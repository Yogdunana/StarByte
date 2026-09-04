package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateProfile_WritesFieldDiff(t *testing.T) {
	profs := &mockProfRepo{}
	svc := NewMemberService(&mockAppRepo{}, profs, nil)
	id := uuid.New()
	viewer := uuid.New()
	row := &model.ProfileWithNames{MemberProfile: model.MemberProfile{
		ID: id, UserID: viewer, RealName: "旧名", Bio: "旧简介",
	}}
	profs.On("GetByIDWithNames", mock.Anything, id).Return(row, nil).Once()
	profs.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberProfile")).Return(nil)
	profs.On("CreateHistories", mock.Anything, mock.AnythingOfType("[]model.ProfileHistory")).Return(nil).Run(func(args mock.Arguments) {
		rows := args.Get(1).([]model.ProfileHistory)
		require.NotEmpty(t, rows)
		found := false
		for _, h := range rows {
			if h.FieldName == "real_name" {
				found = true
				require.Equal(t, "旧名", h.OldValue)
				require.Equal(t, "新名", h.NewValue)
			}
		}
		require.True(t, found)
	})
	updated := *row
	updated.RealName = "新名"
	profs.On("GetByIDWithNames", mock.Anything, id).Return(&updated, nil)

	out, err := svc.UpdateProfile(context.Background(), viewer, id, &dto.UpdateProfileRequest{RealName: "新名"}, nil)
	require.NoError(t, err)
	require.Equal(t, "新名", out.RealName)
}

func TestGetProfile_Denied(t *testing.T) {
	profs := &mockProfRepo{}
	svc := NewMemberService(&mockAppRepo{}, profs, nil)
	id := uuid.New()
	profs.On("GetByIDWithNames", mock.Anything, id).Return(&model.ProfileWithNames{
		MemberProfile: model.MemberProfile{ID: id, UserID: uuid.New()},
	}, nil)
	_, err := svc.GetProfile(context.Background(), uuid.New(), id, &rbacModel.DataScopeCondition{Query: "1 = 0"})
	requireAppError(t, err, response.CodeMemberProfileDenied, "无权操作该档案")
}

func TestUpdateProfileStatus_Leave(t *testing.T) {
	profs := &mockProfRepo{}
	svc := NewMemberService(&mockAppRepo{}, profs, nil)
	id := uuid.New()
	p := &model.MemberProfile{ID: id, UserID: uuid.New(), Status: model.ProfileActive}
	profs.On("GetByID", mock.Anything, id).Return(p, nil)
	profs.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberProfile")).Return(nil)
	profs.On("CreateHistories", mock.Anything, mock.AnythingOfType("[]model.ProfileHistory")).Return(nil)
	profs.On("GetByIDWithNames", mock.Anything, id).Return(&model.ProfileWithNames{
		MemberProfile: model.MemberProfile{ID: id, Status: model.ProfileLeft},
	}, nil)

	out, err := svc.UpdateProfileStatus(context.Background(), uuid.New(), id, &dto.UpdateProfileStatusRequest{
		Status: model.ProfileLeft, Reason: "毕业离会",
	})
	require.NoError(t, err)
	require.Equal(t, model.ProfileLeft, out.Status)
}

func TestExportProfiles_Empty(t *testing.T) {
	profs := &mockProfRepo{}
	svc := NewMemberService(&mockAppRepo{}, profs, nil)
	profs.On("List", mock.Anything, mock.Anything, mock.Anything).Return([]model.ProfileWithNames{}, int64(0), nil)
	_, err := svc.ExportProfiles(context.Background(), uuid.New(), &dto.ListProfileRequest{}, nil)
	requireAppError(t, err, response.CodeMemberExportFail, "没有可导出的档案")
}

func TestExportProfiles_PDFHeader(t *testing.T) {
	profs := &mockProfRepo{}
	svc := NewMemberService(&mockAppRepo{}, profs, nil)
	id := uuid.New()
	profs.On("List", mock.Anything, mock.Anything, mock.Anything).Return([]model.ProfileWithNames{{
		MemberProfile: model.MemberProfile{ID: id, RealName: "张三", StudentNo: "1", Bio: "简介"},
	}}, int64(1), nil)
	pdf, err := svc.ExportProfiles(context.Background(), uuid.New(), &dto.ListProfileRequest{}, nil)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(pdf), "%PDF-1.4"))
	require.Contains(t, string(pdf), "%%EOF")
}

func TestApplicationStats(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	apps.On("Stats", mock.Anything, "", "", "type").Return([]model.StatBucket{{Key: "1", Label: "会员", Count: 3}}, nil)
	out, err := svc.ApplicationStats(context.Background(), &dto.StatsQuery{GroupBy: "type"})
	require.NoError(t, err)
	require.Equal(t, int64(3), out.Items[0].Count)
}
