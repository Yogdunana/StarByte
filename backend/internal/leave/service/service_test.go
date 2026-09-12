package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func monday() time.Time {
	// 2026-09-14 is a Monday. 业务日历按 Asia/Shanghai，不用 UTC。
	return time.Date(2026, 9, 14, 9, 0, 0, 0, bizLocation())
}

func fixture() (*leaveService, *memRepo, uuid.UUID, uuid.UUID, uuid.UUID) {
	mem := newMemRepo()
	annual := model.LeaveType{
		ID: uuid.New(), Name: "年假", Code: model.TypeAnnual,
		Deductible: true, DefaultDays: 5, Enabled: true, SortOrder: 10,
	}
	personal := model.LeaveType{
		ID: uuid.New(), Name: "事假", Code: model.TypePersonal,
		Deductible: false, DefaultDays: 0, Enabled: true, SortOrder: 30,
	}
	mem.addType(annual)
	mem.addType(personal)
	applicant := uuid.New()
	approver := uuid.New()
	mem.addUser(applicant, "申请人")
	mem.addUser(approver, "审批人")
	svc := &leaveService{rows: mem, now: monday}
	return svc, mem, applicant, approver, annual.ID
}

func applicantViewer(id uuid.UUID) Viewer {
	return Viewer{UserID: id}
}

func approverViewer(id uuid.UUID) Viewer {
	return Viewer{UserID: id, CanRead: true, CanApprove: true, Scope: &rbacModel.DataScopeCondition{}}
}

func deptViewer(id, dept uuid.UUID, approve bool) Viewer {
	return Viewer{
		UserID:     id,
		CanRead:    true,
		CanApprove: approve,
		Scope:      &rbacModel.DataScopeCondition{Query: "department_id = ?", Args: []interface{}{dept}},
	}
}

func submitReq(typeID uuid.UUID, start, end time.Time) *dto.SubmitLeaveRequest {
	return &dto.SubmitLeaveRequest{
		LeaveTypeID: typeID.String(),
		StartTime:   start,
		EndTime:     end,
		Reason:      "参加会议",
	}
}

func TestCalculateDurationDays_SameDay(t *testing.T) {
	loc := bizLocation()
	start := time.Date(2026, 9, 14, 9, 0, 0, 0, loc)
	end := time.Date(2026, 9, 14, 18, 0, 0, 0, loc)
	assert.Equal(t, 1.0, CalculateDurationDays(start, end))
}

func TestCalculateDurationDays_WeekdaysAndWeekend(t *testing.T) {
	loc := bizLocation()
	start := time.Date(2026, 9, 14, 9, 0, 0, 0, loc) // Mon
	end := time.Date(2026, 9, 18, 18, 0, 0, 0, loc)  // Fri
	assert.Equal(t, 5.0, CalculateDurationDays(start, end))

	endSat := time.Date(2026, 9, 19, 18, 0, 0, 0, loc)
	assert.Equal(t, 5.0, CalculateDurationDays(start, endSat))

	// RangePicker 默认结束到周六 00:00 时，不应把周五扣成半天。
	endSatMidnight := time.Date(2026, 9, 19, 0, 0, 0, 0, loc)
	assert.Equal(t, 5.0, CalculateDurationDays(start, endSatMidnight))
	assert.Equal(t, 0.0, CalculateDurationDays(
		time.Date(2026, 9, 19, 9, 0, 0, 0, loc),
		time.Date(2026, 9, 19, 18, 0, 0, 0, loc),
	))
	assert.Equal(t, 0.0, CalculateDurationDays(
		time.Date(2026, 9, 19, 9, 0, 0, 0, loc),
		time.Date(2026, 9, 20, 18, 0, 0, 0, loc),
	))
}

func TestCalculateDurationDays_HalfDay(t *testing.T) {
	loc := bizLocation()
	start := time.Date(2026, 9, 14, 9, 0, 0, 0, loc)
	end := time.Date(2026, 9, 15, 10, 0, 0, 0, loc) // 上海 10:00，未满 12 小时
	assert.Equal(t, 1.5, CalculateDurationDays(start, end))
}

func TestCalendarDateUsesShanghai(t *testing.T) {
	// UTC 9 月 13 日 16:00 = 上海 9 月 14 日 00:00
	instant := time.Date(2026, 9, 13, 16, 0, 0, 0, time.UTC)
	got := calendarDate(instant)
	assert.True(t, got.Equal(time.Date(2026, 9, 14, 0, 0, 0, 0, bizLocation())))
	assert.Equal(t, 2026, bizYear(instant))
}

func TestSubmitShanghaiMidnightTodayAllowed(t *testing.T) {
	svc, _, applicant, _, annualID := fixture()
	loc := bizLocation()
	start := time.Date(2026, 9, 14, 0, 0, 0, 0, loc)
	end := time.Date(2026, 9, 14, 8, 0, 0, 0, loc)
	got, err := svc.Submit(context.Background(), applicantViewer(applicant), submitReq(annualID, start, end))
	require.NoError(t, err)
	assert.Equal(t, 1.0, got.DurationDays)
}

func TestSubmitDeductsDeductibleBalance(t *testing.T) {
	svc, mem, applicant, _, annualID := fixture()
	ctx := context.Background()
	start := monday()
	end := start.Add(8 * time.Hour)
	got, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, end))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, model.ApprovalStatusPending, got.Status)
	assert.Equal(t, 1.0, got.DurationDays)

	bals, err := svc.Balances(ctx, applicantViewer(applicant), "", 2026)
	require.NoError(t, err)
	var annual *dto.LeaveBalanceResponse
	for _, b := range bals {
		if b.LeaveType.ID == annualID.String() {
			annual = b
		}
	}
	require.NotNil(t, annual)
	assert.Equal(t, 1.0, annual.UsedDays)
	assert.Equal(t, 4.0, annual.RemainingDays)

	rows, err := mem.GetLeaveBalancesByUser(ctx, applicant, 2026)
	require.NoError(t, err)
	require.NotEmpty(t, rows)
}

func TestSubmitPersonalDoesNotDeduct(t *testing.T) {
	svc, mem, applicant, _, _ := fixture()
	var personalID uuid.UUID
	for _, tpe := range mustTypes(mem) {
		if tpe.Code == model.TypePersonal {
			personalID = tpe.ID
		}
	}
	start := monday()
	got, err := svc.Submit(context.Background(), applicantViewer(applicant), submitReq(personalID, start, start.Add(8*time.Hour)))
	require.NoError(t, err)
	assert.Equal(t, 1.0, got.DurationDays)
	bals, err := mem.GetLeaveBalancesByUser(context.Background(), applicant, 2026)
	require.NoError(t, err)
	for _, b := range bals {
		if b.LeaveTypeID == personalID {
			assert.Equal(t, 0.0, b.UsedDays)
		}
	}
}

func TestSubmitInsufficientBalance(t *testing.T) {
	svc, _, applicant, _, annualID := fixture()
	start := monday()
	end := start.Add(6 * 24 * time.Hour) // Mon-Sun = 5 weekdays, still within 5; use 10 weekdays
	end = time.Date(2026, 9, 25, 18, 0, 0, 0, time.UTC)
	_, err := svc.Submit(context.Background(), applicantViewer(applicant), submitReq(annualID, start, end))
	require.Error(t, err)
	appErr, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeLeaveInsufficient, appErr.Code)
}

func TestSubmitOverlapAndInvalidTime(t *testing.T) {
	svc, _, applicant, _, annualID := fixture()
	ctx := context.Background()
	start := monday()
	end := start.Add(8 * time.Hour)
	_, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, end))
	require.NoError(t, err)

	_, err = svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start.Add(time.Hour), end.Add(time.Hour)))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveOverlap, err.(*response.AppError).Code)

	_, err = svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, end, start))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveInvalidTime, err.(*response.AppError).Code)

	_, err = svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start.Add(24*time.Hour), start.Add(24*time.Hour)))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveInvalidTime, err.(*response.AppError).Code)

	past := start.Add(-48 * time.Hour)
	_, err = svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, past, past.Add(time.Hour)))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveInvalidTime, err.(*response.AppError).Code)
}

func TestApproveDoesNotChangeBalanceAndRejectRestores(t *testing.T) {
	svc, _, applicant, approver, annualID := fixture()
	ctx := context.Background()
	start := monday()
	app, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, start.Add(8*time.Hour)))
	require.NoError(t, err)

	require.NoError(t, svc.Approve(ctx, approverViewer(approver), uuid.MustParse(app.ID), "ok"))
	bals, err := svc.Balances(ctx, applicantViewer(applicant), "", 2026)
	require.NoError(t, err)
	assert.Equal(t, 4.0, findAnnual(bals, annualID).RemainingDays)

	err = svc.Approve(ctx, approverViewer(approver), uuid.MustParse(app.ID), "again")
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveInvalidState, err.(*response.AppError).Code)

	svc2, _, applicant2, approver2, annual2 := fixture()
	app2, err := svc2.Submit(ctx, applicantViewer(applicant2), submitReq(annual2, start, start.Add(8*time.Hour)))
	require.NoError(t, err)
	require.NoError(t, svc2.Reject(ctx, approverViewer(approver2), uuid.MustParse(app2.ID), "no"))
	bals2, err := svc2.Balances(ctx, applicantViewer(applicant2), "", 2026)
	require.NoError(t, err)
	assert.Equal(t, 5.0, findAnnual(bals2, annual2).RemainingDays)
	assert.Equal(t, 0.0, findAnnual(bals2, annual2).UsedDays)

	err = svc2.Reject(ctx, approverViewer(approver2), uuid.MustParse(app2.ID), "again")
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveInvalidState, err.(*response.AppError).Code)
	bals2, err = svc2.Balances(ctx, applicantViewer(applicant2), "", 2026)
	require.NoError(t, err)
	assert.Equal(t, 5.0, findAnnual(bals2, annual2).RemainingDays)
	assert.Equal(t, 0.0, findAnnual(bals2, annual2).UsedDays)
}

func TestSelfApproveAndIDOR(t *testing.T) {
	svc, _, applicant, approver, annualID := fixture()
	ctx := context.Background()
	start := monday()
	app, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, start.Add(8*time.Hour)))
	require.NoError(t, err)
	id := uuid.MustParse(app.ID)

	err = svc.Approve(ctx, Viewer{UserID: applicant, CanApprove: true}, id, "self")
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	err = svc.Approve(ctx, Viewer{UserID: approver}, id, "no perm")
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	stranger := uuid.New()
	_, err = svc.Get(ctx, applicantViewer(stranger), id)
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	got, err := svc.Get(ctx, applicantViewer(applicant), id)
	require.NoError(t, err)
	assert.Equal(t, app.ID, got.ID)

	got, err = svc.Get(ctx, approverViewer(approver), id)
	require.NoError(t, err)
	assert.Equal(t, app.ID, got.ID)

	_, _, err = svc.ListAll(ctx, applicantViewer(applicant), &dto.ListLeaveRequest{})
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	list, total, err := svc.ListMine(ctx, applicantViewer(applicant), &dto.ListLeaveRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	all, total, err := svc.ListAll(ctx, approverViewer(approver), &dto.ListLeaveRequest{Status: model.ApprovalStatusPending})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, all, 1)
}

func TestDepartmentScopeBlocksCrossDept(t *testing.T) {
	svc, mem, applicant, approver, annualID := fixture()
	ctx := context.Background()
	deptA, deptB := uuid.New(), uuid.New()
	mem.addUserDept(applicant, deptA)
	outsider := uuid.New()
	mem.addUser(outsider, "外部门申请人")
	mem.addUserDept(outsider, deptB)

	start := monday()
	mine, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, start.Add(8*time.Hour)))
	require.NoError(t, err)
	theirs, err := svc.Submit(ctx, applicantViewer(outsider), submitReq(annualID, start.Add(24*time.Hour), start.Add(32*time.Hour)))
	require.NoError(t, err)

	viewer := deptViewer(approver, deptA, true)
	got, err := svc.Get(ctx, viewer, uuid.MustParse(mine.ID))
	require.NoError(t, err)
	assert.Equal(t, mine.ID, got.ID)

	_, err = svc.Get(ctx, viewer, uuid.MustParse(theirs.ID))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	list, total, err := svc.ListAll(ctx, viewer, &dto.ListLeaveRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	assert.Equal(t, mine.ID, list[0].ID)

	_, _, err = svc.ListAll(ctx, viewer, &dto.ListLeaveRequest{UserID: outsider.String()})
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	_, err = svc.Balances(ctx, viewer, outsider.String(), 2026)
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	err = svc.Approve(ctx, viewer, uuid.MustParse(theirs.ID), "no")
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	require.NoError(t, svc.Approve(ctx, viewer, uuid.MustParse(mine.ID), "ok"))
}

func TestBalanceIDORAndStats(t *testing.T) {
	svc, _, applicant, approver, annualID := fixture()
	ctx := context.Background()
	start := monday()
	_, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, start, start.Add(8*time.Hour)))
	require.NoError(t, err)

	_, err = svc.Balances(ctx, applicantViewer(applicant), approver.String(), 2026)
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	bals, err := svc.Balances(ctx, approverViewer(approver), applicant.String(), 2026)
	require.NoError(t, err)
	require.NotEmpty(t, bals)

	_, err = svc.Stats(ctx, applicantViewer(applicant), 2026)
	require.Error(t, err)
	stats, err := svc.Stats(ctx, approverViewer(approver), 2026)
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Total)
	assert.Equal(t, int64(1), stats.ByStatus[model.ApprovalStatusPending])
	assert.Equal(t, int64(0), stats.Personal.Total)
	assert.NotEmpty(t, stats.ByMonth)

	empty, err := svc.Stats(ctx, approverViewer(approver), 2025)
	require.NoError(t, err)
	assert.Equal(t, int64(0), empty.Total)
	assert.Empty(t, empty.ByMonth)
	assert.Empty(t, empty.Departments)
}

func TestStatsExcludeRejectedAndHonorYear(t *testing.T) {
	svc, mem, applicant, approver, annualID := fixture()
	ctx := context.Background()
	app, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, monday(), monday().Add(8*time.Hour)))
	require.NoError(t, err)
	require.NoError(t, svc.Reject(ctx, approverViewer(approver), uuid.MustParse(app.ID), "no"))

	oldID := uuid.New()
	oldStart := time.Date(2025, 6, 2, 9, 0, 0, 0, bizLocation())
	mem.apps[oldID] = model.ApplicationNamed{
		LeaveApplication: model.LeaveApplication{
			ID: oldID, ApplicantID: applicant, LeaveTypeID: annualID,
			StartTime: oldStart, EndTime: oldStart.Add(8 * time.Hour),
			DurationDays: 1, Status: model.ApprovalStatusApproved,
			LeaveType: mem.types[annualID],
		},
	}

	self := Viewer{UserID: applicant, CanRead: true, CanApprove: true, Scope: &rbacModel.DataScopeCondition{}}
	year2026, err := svc.Stats(ctx, self, 2026)
	require.NoError(t, err)
	assert.Equal(t, int64(1), year2026.Total)
	assert.Equal(t, int64(1), year2026.ByStatus[model.ApprovalStatusRejected])
	assert.Equal(t, int64(0), year2026.Personal.Total)
	assert.Equal(t, 0.0, year2026.Personal.Days)
	assert.Empty(t, year2026.Personal.ByType)
	assert.Empty(t, year2026.ByMonth)

	year2025, err := svc.Stats(ctx, self, 2025)
	require.NoError(t, err)
	assert.Equal(t, int64(1), year2025.Total)
	assert.Equal(t, int64(1), year2025.Personal.Total)
	assert.Equal(t, 1.0, year2025.Personal.Days)
	require.Len(t, year2025.Personal.ByType, 1)
	assert.Equal(t, int64(1), year2025.Personal.ByType[0].Count)
	assert.Equal(t, 1.0, year2025.Personal.ByType[0].Days)

	inbox, err := svc.Stats(ctx, self, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), inbox.Total)
	assert.Equal(t, int64(1), inbox.ByStatus[model.ApprovalStatusRejected])
	assert.Equal(t, int64(1), inbox.ByStatus[model.ApprovalStatusApproved])
	assert.Equal(t, int64(1), inbox.Personal.Total)
	assert.Equal(t, 1.0, inbox.Personal.Days)
}

func TestRejectRestoresUsingDeductionSnapshot(t *testing.T) {
	svc, mem, applicant, approver, annualID := fixture()
	ctx := context.Background()
	app, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, monday(), monday().Add(8*time.Hour)))
	require.NoError(t, err)
	bals, err := svc.Balances(ctx, applicantViewer(applicant), "", 2026)
	require.NoError(t, err)
	assert.Equal(t, 4.0, findAnnual(bals, annualID).RemainingDays)

	row := mem.types[annualID]
	row.Deductible = false
	mem.types[annualID] = row

	require.NoError(t, svc.Reject(ctx, approverViewer(approver), uuid.MustParse(app.ID), "changed type"))
	bals, err = svc.Balances(ctx, applicantViewer(applicant), "", 2026)
	require.NoError(t, err)
	assert.Equal(t, 5.0, findAnnual(bals, annualID).RemainingDays)
	stored := mem.apps[uuid.MustParse(app.ID)]
	assert.False(t, stored.BalanceDeducted)
}

func TestGetMissingAndUnknownType(t *testing.T) {
	svc, _, applicant, _, _ := fixture()
	_, err := svc.Get(context.Background(), applicantViewer(applicant), uuid.New())
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNotFound, err.(*response.AppError).Code)

	_, err = svc.Submit(context.Background(), applicantViewer(applicant), submitReq(uuid.New(), monday(), monday().Add(time.Hour)))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveTypeNotFound, err.(*response.AppError).Code)
}

func TestDefaultPage(t *testing.T) {
	p, s := defaultPage(0, 0)
	assert.Equal(t, 1, p)
	assert.Equal(t, 10, s)
	p, s = defaultPage(2, 200)
	assert.Equal(t, 2, p)
	assert.Equal(t, 100, s)
}

func mustTypes(mem *memRepo) []model.LeaveType {
	rows, _ := mem.GetAllLeaveTypes(context.Background())
	return rows
}

func TestCreateAndUpdateType(t *testing.T) {
	svc, _, applicant, approver, _ := fixture()
	ctx := context.Background()
	_, err := svc.CreateType(ctx, applicantViewer(applicant), &dto.UpsertLeaveTypeRequest{Name: "婚假", Code: "marriage", DefaultDays: 3})
	require.Error(t, err)
	enabled := true
	got, err := svc.CreateType(ctx, approverViewer(approver), &dto.UpsertLeaveTypeRequest{
		Name: "婚假", Code: "Marriage", Deductible: false, DefaultDays: 3, Enabled: &enabled, SortOrder: 50,
	})
	require.NoError(t, err)
	assert.Equal(t, "marriage", got.Code)
	assert.True(t, got.Enabled)

	disabled := false
	got, err = svc.UpdateType(ctx, approverViewer(approver), uuid.MustParse(got.ID), &dto.UpsertLeaveTypeRequest{
		Name: "婚假", Code: "marriage", Enabled: &disabled,
	})
	require.NoError(t, err)
	assert.False(t, got.Enabled)
	_, err = svc.Submit(ctx, applicantViewer(applicant), submitReq(uuid.MustParse(got.ID), monday(), monday().Add(time.Hour)))
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveTypeDisabled, err.(*response.AppError).Code)
}

func TestDepartmentApproverCannotMutateTypes(t *testing.T) {
	svc, _, _, approver, annualID := fixture()
	ctx := context.Background()
	dept := uuid.New()
	minister := deptViewer(approver, dept, true)
	_, err := svc.CreateType(ctx, minister, &dto.UpsertLeaveTypeRequest{Name: "调休加码", Code: "bonus", Deductible: true, DefaultDays: 366})
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	enabled := false
	_, err = svc.UpdateType(ctx, minister, annualID, &dto.UpsertLeaveTypeRequest{Name: "年假", Code: "annual", Enabled: &enabled})
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)

	_, err = svc.CreateType(ctx, Viewer{UserID: approver, CanApprove: true}, &dto.UpsertLeaveTypeRequest{Name: "无范围", Code: "noscope"})
	require.Error(t, err)
	assert.Equal(t, response.CodeLeaveNoAccess, err.(*response.AppError).Code)
}

func TestCalendarAndTodosAndAttachments(t *testing.T) {
	svc, _, applicant, approver, annualID := fixture()
	ctx := context.Background()
	start := monday()
	req := submitReq(annualID, start, start.Add(8*time.Hour))
	req.Attachments = []dto.Attachment{{FileID: uuid.New().String(), Name: "note.pdf", Size: 12}}
	app, err := svc.Submit(ctx, applicantViewer(applicant), req)
	require.NoError(t, err)
	require.Len(t, app.Attachments, 1)

	cal, err := svc.Calendar(ctx, applicantViewer(applicant), &dto.ListLeaveRequest{From: "2026-09-01", To: "2026-09-30"})
	require.NoError(t, err)
	require.Len(t, cal, 1)

	_, _, err = svc.ListTodos(ctx, applicantViewer(applicant), &dto.ListLeaveRequest{})
	require.Error(t, err)
	todos, total, err := svc.ListTodos(ctx, approverViewer(approver), &dto.ListLeaveRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, todos, 1)
}

func TestListTodosOnlyAssignedWorkflowStage(t *testing.T) {
	svc, mem, flow, applicant, minister, annualID := workflowFixture(t)
	president := uuid.New()
	mem.addUser(president, "社长")
	ctx := context.Background()
	app, err := svc.Submit(ctx, applicantViewer(applicant), submitReq(annualID, monday(), monday().Add(8*time.Hour)))
	require.NoError(t, err)
	id := uuid.MustParse(app.ID)
	mem.assignTodo(id, minister)

	mine, total, err := svc.ListTodos(ctx, approverViewer(minister), &dto.ListLeaveRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, mine, 1)

	theirs, total, err := svc.ListTodos(ctx, approverViewer(president), &dto.ListLeaveRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, theirs)

	require.NoError(t, svc.Approve(ctx, approverViewer(minister), id, "dept ok"))
	assert.Equal(t, 1, flow.idx)
	mem.assignees[id] = []uuid.UUID{president}
	next, total, err := svc.ListTodos(ctx, approverViewer(president), &dto.ListLeaveRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, next, 1)
}

func findAnnual(bals []*dto.LeaveBalanceResponse, id uuid.UUID) *dto.LeaveBalanceResponse {
	for _, b := range bals {
		if b.LeaveType.ID == id.String() {
			return b
		}
	}
	return &dto.LeaveBalanceResponse{}
}
