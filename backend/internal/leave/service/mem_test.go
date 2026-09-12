package service

import (
	"context"
	"sync"
	"time"

	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memRepo struct {
	mu      sync.Mutex
	types   map[uuid.UUID]model.LeaveType
	bals    map[string]model.LeaveBalance
	apps    map[uuid.UUID]model.ApplicationNamed
	names   map[uuid.UUID]string
	depts   map[uuid.UUID]uuid.UUID
	failGet bool
}

func newMemRepo() *memRepo {
	return &memRepo{
		types: map[uuid.UUID]model.LeaveType{},
		bals:  map[string]model.LeaveBalance{},
		apps:  map[uuid.UUID]model.ApplicationNamed{},
		names: map[uuid.UUID]string{},
		depts: map[uuid.UUID]uuid.UUID{},
	}
}

func balKey(userID uuid.UUID, year int, typeID uuid.UUID) string {
	return userID.String() + "/" + itoa(year) + "/" + typeID.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return sign + string(b[i:])
}

func (m *memRepo) addType(t model.LeaveType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.types[t.ID] = t
}

func (m *memRepo) addUser(id uuid.UUID, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.names[id] = name
}

func (m *memRepo) addUserDept(id, dept uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.depts[id] = dept
}

func (m *memRepo) inScope(row model.ApplicationNamed, scope *rbacModel.DataScopeCondition) bool {
	if scope == nil || strings.TrimSpace(scope.Query) == "" {
		return true
	}
	if scope.Query == "1 = 0" {
		return false
	}
	if scope.IsSelf || scope.Query == "leave_applications.applicant_id = ?" {
		if len(scope.Args) == 1 {
			id, ok := scope.Args[0].(uuid.UUID)
			return ok && row.ApplicantID == id
		}
		return false
	}
	if row.ApplicantDepartmentID == nil {
		return false
	}
	dept := *row.ApplicantDepartmentID
	for _, arg := range scope.Args {
		switch v := arg.(type) {
		case uuid.UUID:
			if v == dept {
				return true
			}
		case []uuid.UUID:
			for _, id := range v {
				if id == dept {
					return true
				}
			}
		}
	}
	return false
}

func (m *memRepo) WithTx(_ context.Context, fn func(repo.Repository) error) error {
	m.mu.Lock()
	types := map[uuid.UUID]model.LeaveType{}
	for k, v := range m.types {
		types[k] = v
	}
	bals := map[string]model.LeaveBalance{}
	for k, v := range m.bals {
		bals[k] = v
	}
	apps := map[uuid.UUID]model.ApplicationNamed{}
	for k, v := range m.apps {
		apps[k] = v
	}
	m.mu.Unlock()
	err := fn(m)
	if err != nil {
		m.mu.Lock()
		m.types, m.bals, m.apps = types, bals, apps
		m.mu.Unlock()
	}
	return err
}

func (m *memRepo) DB() *gorm.DB { return nil }

func (m *memRepo) GetAllLeaveTypes(_ context.Context) ([]model.LeaveType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.LeaveType, 0, len(m.types))
	for _, t := range m.types {
		out = append(out, t)
	}
	return out, nil
}

func (m *memRepo) GetLeaveTypeByCode(_ context.Context, code string) (*model.LeaveType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.types {
		if t.Code == code {
			cp := t
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memRepo) CreateLeaveType(_ context.Context, row *model.LeaveType) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.types[row.ID] = *row
	return nil
}

func (m *memRepo) UpdateLeaveType(_ context.Context, row *model.LeaveType) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.types[row.ID] = *row
	return nil
}

func (m *memRepo) GetLeaveTypeByID(_ context.Context, id uuid.UUID) (*model.LeaveType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.types[id]
	if !ok {
		return nil, nil
	}
	cp := t
	return &cp, nil
}

func (m *memRepo) getBalLocked(userID uuid.UUID, year int, leaveTypeID uuid.UUID) *model.LeaveBalance {
	row, ok := m.bals[balKey(userID, year, leaveTypeID)]
	if !ok {
		return nil
	}
	if t, ok := m.types[row.LeaveTypeID]; ok {
		row.LeaveType = t
	}
	cp := row
	return &cp
}

func (m *memRepo) GetLeaveBalance(_ context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getBalLocked(userID, year, leaveTypeID), nil
}

func (m *memRepo) GetLeaveBalanceForUpdate(ctx context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID) (*model.LeaveBalance, error) {
	return m.GetLeaveBalance(ctx, userID, year, leaveTypeID)
}

func (m *memRepo) GetLeaveBalancesByUser(_ context.Context, userID uuid.UUID, year int) ([]model.LeaveBalance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.LeaveBalance
	for _, row := range m.bals {
		if row.UserID == userID && row.Year == year {
			if t, ok := m.types[row.LeaveTypeID]; ok {
				row.LeaveType = t
			}
			out = append(out, row)
		}
	}
	return out, nil
}

func (m *memRepo) CreateLeaveBalance(_ context.Context, balance *model.LeaveBalance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *balance
	m.bals[balKey(balance.UserID, balance.Year, balance.LeaveTypeID)] = cp
	return nil
}

func (m *memRepo) DeductLeaveBalance(_ context.Context, userID uuid.UUID, year int, leaveTypeID uuid.UUID, usedDays float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := balKey(userID, year, leaveTypeID)
	row, ok := m.bals[key]
	if !ok {
		return nil
	}
	row.UsedDays += usedDays
	row.RemainingDays -= usedDays
	m.bals[key] = row
	return nil
}

func (m *memRepo) CreateLeaveApplication(_ context.Context, application *model.LeaveApplication) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	named := model.ApplicationNamed{LeaveApplication: *application, ApplicantName: m.names[application.ApplicantID]}
	if dept, ok := m.depts[application.ApplicantID]; ok {
		d := dept
		named.ApplicantDepartmentID = &d
	}
	if t, ok := m.types[application.LeaveTypeID]; ok {
		named.LeaveType = t
	}
	m.apps[application.ID] = named
	return nil
}

func (m *memRepo) LockApplicant(context.Context, uuid.UUID) error { return nil }

func (m *memRepo) GetLeaveApplicationByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.ApplicationNamed, error) {
	return m.GetLeaveApplicationByID(ctx, id)
}

func (m *memRepo) GetLeaveApplicationByID(_ context.Context, id uuid.UUID) (*model.ApplicationNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failGet {
		return nil, errMem
	}
	row, ok := m.apps[id]
	if !ok {
		return nil, nil
	}
	if t, ok := m.types[row.LeaveTypeID]; ok {
		row.LeaveType = t
	}
	if dept, ok := m.depts[row.ApplicantID]; ok {
		d := dept
		row.ApplicantDepartmentID = &d
	}
	if row.ApproverID != nil {
		row.ApproverName = m.names[*row.ApproverID]
	}
	cp := row
	return &cp, nil
}

func (m *memRepo) GetLeaveApplicationsByUser(_ context.Context, userID uuid.UUID, status string, page, pageSize int, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []model.ApplicationNamed
	for _, row := range m.apps {
		if row.ApplicantID != userID {
			continue
		}
		if status != "" && row.Status != status {
			continue
		}
		if !m.inScope(row, scope) {
			continue
		}
		all = append(all, row)
	}
	return pageSlice(all, page, pageSize), int64(len(all)), nil
}

func (m *memRepo) GetLeaveApplicationsByStatus(_ context.Context, status string, page, pageSize int, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []model.ApplicationNamed
	for _, row := range m.apps {
		if status != "" && row.Status != status {
			continue
		}
		if !m.inScope(row, scope) {
			continue
		}
		all = append(all, row)
	}
	return pageSlice(all, page, pageSize), int64(len(all)), nil
}

func pageSlice(all []model.ApplicationNamed, page, pageSize int) []model.ApplicationNamed {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start >= len(all) {
		return nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end]
}

func (m *memRepo) UpdateApprovalStatus(_ context.Context, id uuid.UUID, approverID uuid.UUID, status, remark string, at time.Time, balanceDeducted bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.apps[id]
	if !ok || row.Status != model.ApprovalStatusPending {
		return repo.ErrNotPending
	}
	row.Status = status
	row.ApproverID = &approverID
	row.ApproveRemark = remark
	row.ApprovedAt = &at
	row.ApproverName = m.names[approverID]
	row.BalanceDeducted = balanceDeducted
	m.apps[id] = row
	return nil
}

func (m *memRepo) GetLeaveApplicationsByUserAndTimeRange(_ context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]model.LeaveApplication, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.LeaveApplication
	for _, row := range m.apps {
		if row.ApplicantID != userID || row.Status == model.ApprovalStatusRejected {
			continue
		}
		if row.StartTime.Before(endTime) && row.EndTime.After(startTime) {
			out = append(out, row.LeaveApplication)
		}
	}
	return out, nil
}

func (m *memRepo) GetUserDepartmentID(_ context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if dept, ok := m.depts[userID]; ok {
		d := dept
		return &d, nil
	}
	return nil, nil
}

func (m *memRepo) SaveApplication(_ context.Context, app *model.LeaveApplication) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.apps[app.ID]
	if !ok {
		row = model.ApplicationNamed{LeaveApplication: *app, ApplicantName: m.names[app.ApplicantID]}
	} else {
		row.LeaveApplication = *app
	}
	if t, ok := m.types[app.LeaveTypeID]; ok {
		row.LeaveType = t
	}
	if app.ApproverID != nil {
		row.ApproverName = m.names[*app.ApproverID]
	}
	m.apps[app.ID] = row
	return nil
}

func (m *memRepo) ListCalendar(_ context.Context, from, to time.Time, userID *uuid.UUID, scope *rbacModel.DataScopeCondition) ([]model.ApplicationNamed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.ApplicationNamed
	for _, row := range m.apps {
		if row.Status == model.ApprovalStatusRejected {
			continue
		}
		if !row.StartTime.Before(to) || !row.EndTime.After(from) {
			continue
		}
		if userID != nil && row.ApplicantID != *userID {
			continue
		}
		if !m.inScope(row, scope) {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func matchBizYear(t time.Time, year int) bool {
	if year <= 0 {
		return true
	}
	return t.In(bizLocation()).Year() == year
}

func (m *memRepo) CountPersonalStats(_ context.Context, userID uuid.UUID, year int) (int64, float64, []repo.TypeCount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	typeAgg := map[uuid.UUID]repo.TypeCount{}
	var n int64
	var days float64
	for _, row := range m.apps {
		if row.ApplicantID != userID || row.Status == model.ApprovalStatusRejected || !matchBizYear(row.StartTime, year) {
			continue
		}
		n++
		days += row.DurationDays
		agg := typeAgg[row.LeaveTypeID]
		agg.LeaveTypeID = row.LeaveTypeID
		agg.Code = row.LeaveType.Code
		agg.Name = row.LeaveType.Name
		agg.Count++
		agg.Days += row.DurationDays
		typeAgg[row.LeaveTypeID] = agg
	}
	var types []repo.TypeCount
	for _, v := range typeAgg {
		types = append(types, v)
	}
	return n, days, types, nil
}

func (m *memRepo) CountDepartmentStats(_ context.Context, year int, scope *rbacModel.DataScopeCondition) ([]repo.DeptCount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	type key struct {
		id   uuid.UUID
		have bool
	}
	agg := map[key]repo.DeptCount{}
	for _, row := range m.apps {
		if row.Status == model.ApprovalStatusRejected || !matchBizYear(row.StartTime, year) || !m.inScope(row, scope) {
			continue
		}
		k := key{}
		if row.ApplicantDepartmentID != nil {
			k.id, k.have = *row.ApplicantDepartmentID, true
		}
		item := agg[k]
		if k.have {
			id := k.id
			item.DepartmentID = &id
		}
		item.Count++
		item.Days += row.DurationDays
		agg[k] = item
	}
	var out []repo.DeptCount
	for _, v := range agg {
		out = append(out, v)
	}
	return out, nil
}

func (m *memRepo) CountMonthlyStats(_ context.Context, year int, scope *rbacModel.DataScopeCondition) ([]repo.MonthCount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	agg := map[string]repo.MonthCount{}
	for _, row := range m.apps {
		if row.Status == model.ApprovalStatusRejected || !m.inScope(row, scope) {
			continue
		}
		local := row.StartTime.In(bizLocation())
		if year > 0 && local.Year() != year {
			continue
		}
		month := local.Format("2006-01")
		item := agg[month]
		item.Month = month
		item.Count++
		item.Days += row.DurationDays
		agg[month] = item
	}
	var out []repo.MonthCount
	for _, v := range agg {
		out = append(out, v)
	}
	return out, nil
}

func (m *memRepo) CountStats(_ context.Context, year int, scope *rbacModel.DataScopeCondition) (int64, map[string]int64, []repo.TypeCount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	byStatus := map[string]int64{
		model.ApprovalStatusPending:  0,
		model.ApprovalStatusApproved: 0,
		model.ApprovalStatusRejected: 0,
	}
	typeAgg := map[uuid.UUID]repo.TypeCount{}
	n := 0
	for _, row := range m.apps {
		if !matchBizYear(row.StartTime, year) || !m.inScope(row, scope) {
			continue
		}
		n++
		byStatus[row.Status]++
		agg := typeAgg[row.LeaveTypeID]
		agg.LeaveTypeID = row.LeaveTypeID
		agg.Code = row.LeaveType.Code
		agg.Name = row.LeaveType.Name
		agg.Count++
		agg.Days += row.DurationDays
		typeAgg[row.LeaveTypeID] = agg
	}
	var types []repo.TypeCount
	for _, v := range typeAgg {
		types = append(types, v)
	}
	return int64(n), byStatus, types, nil
}

var errMem = errString("mem failure")

type errString string

func (e errString) Error() string { return string(e) }
