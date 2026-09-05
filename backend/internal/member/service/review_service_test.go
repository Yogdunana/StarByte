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

func TestApprove_MemberCreatesProfile(t *testing.T) {
	apps := &mockAppRepo{}
	profs := &mockProfRepo{}
	svc := NewMemberService(apps, profs, nil)
	id := uuid.New()
	userID := uuid.New()
	reviewer := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, UserID: userID, Type: model.ApplicantMember, Status: model.AppPending,
		RealName: "李四", StudentNo: "2024002", ContactEmail: "l@b.com",
	}, nil)
	profs.On("GetByUserID", mock.Anything, userID).Return(nil, nil)
	profs.On("GetByStudentNo", mock.Anything, "2024002", (*uuid.UUID)(nil)).Return(nil, nil)
	profs.On("Create", mock.Anything, mock.AnythingOfType("*model.MemberProfile")).Return(nil)
	apps.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberApplication")).Return(nil)
	apps.On("CreateHistory", mock.Anything, mock.AnythingOfType("*model.ApplicationHistory")).Return(nil)
	apps.On("GetByIDWithNames", mock.Anything, id).Return(&model.ApplicationWithNames{
		MemberApplication: model.MemberApplication{ID: id, UserID: userID, Type: 1, Status: model.AppApproved},
	}, nil)

	out, err := svc.Approve(context.Background(), reviewer, id, "同意")
	require.NoError(t, err)
	require.Equal(t, model.AppApproved, out.Status)
	profs.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*model.MemberProfile"))
}

func TestApprove_OfficerStartsWorkflow(t *testing.T) {
	apps := &mockAppRepo{}
	profs := &mockProfRepo{}
	starter := &mockStarter{}
	svc := NewMemberService(apps, profs, starter)
	id := uuid.New()
	reviewer := uuid.New()
	instID := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, UserID: uuid.New(), Type: model.ApplicantOfficer, Status: model.AppReviewing, RealName: "王五",
	}, nil)
	starter.On("StartOfficerInterview", mock.Anything, id, reviewer, mock.Anything).Return(instID, nil)
	apps.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberApplication")).Return(nil).Run(func(args mock.Arguments) {
		app := args.Get(1).(*model.MemberApplication)
		require.Equal(t, model.AppInterviewing, app.Status)
		require.Equal(t, instID, *app.FlowInstanceID)
	})
	apps.On("CreateHistory", mock.Anything, mock.AnythingOfType("*model.ApplicationHistory")).Return(nil)
	apps.On("GetByIDWithNames", mock.Anything, id).Return(&model.ApplicationWithNames{
		MemberApplication: model.MemberApplication{ID: id, Type: 2, Status: model.AppInterviewing, FlowInstanceID: &instID},
	}, nil)

	out, err := svc.Approve(context.Background(), reviewer, id, "进入面试")
	require.NoError(t, err)
	require.Equal(t, model.AppInterviewing, out.Status)
	starter.AssertCalled(t, "StartOfficerInterview", mock.Anything, id, reviewer, mock.Anything)
}

func TestApprove_InvalidStatus(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	id := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, Type: model.ApplicantMember, Status: model.AppApproved,
	}, nil)
	_, err := svc.Approve(context.Background(), uuid.New(), id, "x")
	requireAppError(t, err, response.CodeMemberAppInvalid, "当前状态不允许该操作")
}

func TestSupplement_SetsRequiredFields(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	id := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, Type: model.ApplicantMember, Status: model.AppPending,
	}, nil)
	apps.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberApplication")).Return(nil)
	apps.On("CreateHistory", mock.Anything, mock.AnythingOfType("*model.ApplicationHistory")).Return(nil)
	apps.On("GetByIDWithNames", mock.Anything, id).Return(&model.ApplicationWithNames{
		MemberApplication: model.MemberApplication{ID: id, Status: model.AppSupplement, RequiredFields: model.JSONStrings{"skills"}},
	}, nil)

	out, err := svc.Supplement(context.Background(), uuid.New(), id, &dto.SupplementRequest{
		Comment: "补技能", RequiredFields: []string{"skills"},
	})
	require.NoError(t, err)
	require.Equal(t, model.AppSupplement, out.Status)
}

func TestReject_Pending(t *testing.T) {
	apps := &mockAppRepo{}
	svc := NewMemberService(apps, &mockProfRepo{}, nil)
	id := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, Type: model.ApplicantOfficer, Status: model.AppPending,
	}, nil)
	apps.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberApplication")).Return(nil)
	apps.On("CreateHistory", mock.Anything, mock.AnythingOfType("*model.ApplicationHistory")).Return(nil)
	apps.On("GetByIDWithNames", mock.Anything, id).Return(&model.ApplicationWithNames{
		MemberApplication: model.MemberApplication{ID: id, Status: model.AppRejected},
	}, nil)
	out, err := svc.Reject(context.Background(), uuid.New(), id, "不合适")
	require.NoError(t, err)
	require.Equal(t, model.AppRejected, out.Status)
}

func TestSyncFromInterview_ApproveWhenInterviewing(t *testing.T) {
	apps := &mockAppRepo{}
	profs := &mockProfRepo{}
	svc := NewMemberService(apps, profs, nil)
	id := uuid.New()
	userID := uuid.New()
	apps.On("GetByID", mock.Anything, id).Return(&model.MemberApplication{
		ID: id, UserID: userID, Type: model.ApplicantOfficer, Status: model.AppInterviewing,
		RealName: "王五", StudentNo: "2024003",
	}, nil)
	profs.On("GetByUserID", mock.Anything, userID).Return(nil, nil)
	profs.On("GetByStudentNo", mock.Anything, "2024003", (*uuid.UUID)(nil)).Return(nil, nil)
	profs.On("Create", mock.Anything, mock.AnythingOfType("*model.MemberProfile")).Return(nil)
	apps.On("Update", mock.Anything, mock.AnythingOfType("*model.MemberApplication")).Return(nil)
	apps.On("CreateHistory", mock.Anything, mock.AnythingOfType("*model.ApplicationHistory")).Return(nil)
	apps.On("GetByIDWithNames", mock.Anything, id).Return(&model.ApplicationWithNames{
		MemberApplication: model.MemberApplication{ID: id, Status: model.AppApproved},
	}, nil)
	err := svc.SyncFromInterview(context.Background(), uuid.New(), id, 1, "面试通过")
	require.NoError(t, err)
}
