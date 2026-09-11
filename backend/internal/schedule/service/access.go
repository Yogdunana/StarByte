package service

import (
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/schedule/model"
	"github.com/google/uuid"
)

func isAllScope(scope *rbacModel.DataScopeCondition) bool {
	return scope == nil || scope.IsEmpty()
}

func deptInScope(scope *rbacModel.DataScopeCondition, deptID *uuid.UUID) bool {
	if deptID == nil || scope == nil {
		return false
	}
	for _, arg := range scope.Args {
		switch v := arg.(type) {
		case uuid.UUID:
			if v == *deptID {
				return true
			}
		case []uuid.UUID:
			for _, id := range v {
				if id == *deptID {
					return true
				}
			}
		}
	}
	return false
}

func canViewCalendar(scope *rbacModel.DataScopeCondition, cal *model.Calendar, viewer uuid.UUID, memberRole int16) bool {
	if isAllScope(scope) {
		return true
	}
	if cal.OwnerID == viewer || memberRole > 0 {
		return true
	}
	if scope != nil && scope.IsSelf {
		return false
	}
	if scope != nil && scope.Query == "1 = 0" {
		return false
	}
	return cal.CalendarType == model.CalendarDepartment && deptInScope(scope, cal.DepartmentID)
}

func canEditCalendar(scope *rbacModel.DataScopeCondition, cal *model.Calendar, viewer uuid.UUID, memberRole int16) bool {
	if isAllScope(scope) {
		return true
	}
	if cal.OwnerID == viewer {
		return true
	}
	if memberRole == model.MemberEditor {
		return true
	}
	if scope != nil && scope.IsSelf {
		return false
	}
	return cal.CalendarType == model.CalendarDepartment && deptInScope(scope, cal.DepartmentID)
}

func canViewEvent(scope *rbacModel.DataScopeCondition, ev *model.EventNamed, viewer uuid.UUID, memberRole int16, isAttendee bool) bool {
	cal := &model.Calendar{
		OwnerID: ev.OwnerID, CalendarType: ev.CalendarType, DepartmentID: ev.DepartmentID,
	}
	return canViewCalendar(scope, cal, viewer, memberRole) || isAttendee
}

func canEditEvent(scope *rbacModel.DataScopeCondition, ev *model.EventNamed, viewer uuid.UUID, memberRole int16) bool {
	cal := &model.Calendar{
		OwnerID: ev.OwnerID, CalendarType: ev.CalendarType, DepartmentID: ev.DepartmentID,
	}
	if canEditCalendar(scope, cal, viewer, memberRole) {
		return true
	}
	return ev.CreatedBy == viewer && canViewCalendar(scope, cal, viewer, memberRole)
}
