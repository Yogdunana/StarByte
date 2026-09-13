package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

func hasAdmissionRole(actor *model.AdmissionActor, role string) bool {
	if actor == nil {
		return false
	}
	for _, code := range actor.Roles {
		if code == role {
			return true
		}
	}
	return false
}
func sameDepartment(left, right *uuid.UUID) bool {
	return left != nil && right != nil && *left == *right
}
func admissionAuthority(actor *model.AdmissionActor, app *model.MemberApplication, parent *uuid.UUID, role string) (allowed, delegated bool) {
	return admissionRoleAuthority(actor, app, parent, role, false)
}

func admissionRoleAuthority(actor *model.AdmissionActor, app *model.MemberApplication, parent *uuid.UUID, role string, departmentScope bool) (allowed, delegated bool) {
	if actor == nil || actor.ID == app.UserID {
		return false, false
	}
	president := hasAdmissionRole(actor, "president")
	department := app.DepartmentID
	if app.CharterPolicy {
		department = app.ReviewDepartmentID
	}
	minister := hasAdmissionRole(actor, "minister") && roleDepartment(actor, "minister", department)
	officer := hasAdmissionRole(actor, "officer") && roleDepartment(actor, "officer", department)
	center := hasAdmissionRole(actor, "vice_president") && roleDepartment(actor, "vice_president", parent) || hasAdmissionRole(actor, "center_director") && roleDepartment(actor, "center_director", parent)
	switch role {
	case "standing_committee":
		return president || hasAdmissionRole(actor, "minister") || hasAdmissionRole(actor, "vice_president") || hasAdmissionRole(actor, "center_director"), false
	case "center_director":
		return center || president, !center
	case "materials":
		return president || minister || center, false
	case "officer":
		return officer || minister || center || president, !officer
	case "minister":
		return minister || center || president, !minister
	case "center":
		return center || president, !center
	case "president":
		return president, false
	default:
		holds := hasAdmissionRole(actor, role)
		if holds && (!departmentScope || roleDepartment(actor, role, department)) {
			return true, false
		}
		if president {
			return true, true
		}
		if departmentScope && center {
			return true, true
		}
		return false, false
	}
}
func requiredAdmissionRoles(stage string) []string {
	switch stage {
	case model.AdmissionMaterials:
		return []string{"materials"}
	case model.AdmissionRound1:
		return []string{"minister", "center"}
	case model.AdmissionRound2:
		return []string{"center"}
	case model.AdmissionPresident:
		return []string{"president"}
	}
	return []string{}
}
func signedAdmissionRole(signatures []model.AdmissionSignature, app *model.MemberApplication, role string) bool {
	for _, item := range signatures {
		if item.Revision == app.AdmissionRevision && item.Stage == app.AdmissionStage && item.SignerRole == role {
			return true
		}
	}
	return false
}
func admissionRound(stage string) int16 {
	if stage == model.AdmissionRound1 {
		return 1
	}
	if stage == model.AdmissionRound2 {
		return 2
	}
	return 0
}

// calendarMonthLater clamps month-end dates instead of overflowing into the following month.
func calendarMonthLater(now time.Time) time.Time {
	first := time.Date(now.Year(), now.Month()+1, 1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
	last := first.AddDate(0, 1, -1).Day()
	day := now.Day()
	if day > last {
		day = last
	}
	return first.AddDate(0, 0, day-1)
}

func roleDepartment(actor *model.AdmissionActor, role string, department *uuid.UUID) bool {
	if department == nil {
		return false
	}
	if scopes, ok := actor.RoleDepartments[role]; ok {
		for _, id := range scopes {
			if id == *department {
				return true
			}
		}
		return false
	}
	return sameDepartment(actor.DepartmentID, department)
}

func approvalDepartment(app *model.MemberApplication) *uuid.UUID {
	if app.CharterPolicy {
		return app.ReviewDepartmentID
	}
	return app.DepartmentID
}
