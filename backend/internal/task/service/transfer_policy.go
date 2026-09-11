package service

import (
	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
)

func transferHasRole(a *model.TransferActor, role string) bool {
	if a == nil {
		return false
	}
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}
	return false
}
func transferSupervisorRole(a *model.TransferActor) string {
	if transferHasRole(a, "president") || transferHasRole(a, "center_director") || transferHasRole(a, "vice_president") {
		return "president"
	}
	if transferHasRole(a, "minister") {
		return "center"
	}
	return "minister"
}
func transferAuthority(a *model.TransferActor, t *model.TaskTransfer, requirement string) (string, bool) {
	if a == nil || a.ID == t.InitiatorID || a.ID == t.FromUserID || a.ID == t.ToUserID {
		return "", false
	}
	department, center, level := t.SourceDepartmentID, t.SourceCenterID, "minister"
	switch requirement {
	case "target_minister":
		department, center = t.TargetDepartmentID, t.TargetCenterID
	case "source_center":
		level = "center"
	case "target_center":
		center = t.TargetCenterID
		level = "center"
	case "supervisor":
		level = t.SupervisorRole
	case "source_minister":
	default:
		return "", false
	}
	same := func(id uuid.UUID) bool { return a.DepartmentID != nil && *a.DepartmentID == id }
	if level == "president" {
		if transferHasRole(a, "president") {
			return "president", false
		}
		return "", false
	}
	if level == "minister" && same(department) && transferHasRole(a, "minister") {
		return "minister", false
	}
	if same(center) && (transferHasRole(a, "center_director") || transferHasRole(a, "vice_president")) {
		return "center", level == "minister"
	}
	return "", false
}
func transferRequirementsComplete(t *model.TaskTransfer, signatures []model.TransferSignature) bool {
	seen := map[string]bool{}
	for _, s := range signatures {
		if s.Decision == "approve" {
			seen[s.Requirement] = true
		}
	}
	for _, r := range engine.TaskTransferRoles(t.Kind) {
		if !seen[r] {
			return false
		}
	}
	return true
}
