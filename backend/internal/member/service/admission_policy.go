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
	if actor == nil || actor.ID == app.UserID {
		return false, false
	}
	president := hasAdmissionRole(actor, "president")
	minister := hasAdmissionRole(actor, "minister") && sameDepartment(actor.DepartmentID, app.DepartmentID)
	center := (hasAdmissionRole(actor, "vice_president") || hasAdmissionRole(actor, "center_director")) && sameDepartment(actor.DepartmentID, parent)
	switch role {
	case "materials":
		return president || minister || center, false
	case "minister":
		return minister || center || president, !minister
	case "center":
		return center || president, !center
	case "president":
		return president, false
	}
	return false, false
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
