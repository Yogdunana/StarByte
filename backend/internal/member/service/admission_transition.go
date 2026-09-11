package service

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

func advanceAdmission(app *model.MemberApplication, now time.Time) {
	switch app.AdmissionStage {
	case model.AdmissionMaterials:
		if app.Type == model.ApplicantMember {
			app.Status = model.AppApproved
			app.AdmissionStage = model.AdmissionApproved
		} else {
			app.Status = model.AppInterviewing
			app.AdmissionStage = model.AdmissionRound1
		}
	case model.AdmissionRound1:
		app.AdmissionStage = model.AdmissionRound2
	case model.AdmissionRound2:
		app.AdmissionStage = model.AdmissionPresident
	case model.AdmissionPresident:
		app.Status = model.AppApproved
		app.AdmissionStage = model.AdmissionProbation
		until := calendarMonthLater(now)
		app.ProbationUntil = &until
	}
}
func admissionStageLabel(stage string) string {
	labels := map[string]string{model.AdmissionMaterials: "资料审核", model.AdmissionRound1: "一面与会签", model.AdmissionRound2: "二面与中心审批", model.AdmissionPresident: "会长最终确认", model.AdmissionProbation: "候补期", model.AdmissionApproved: "正式成员", model.AdmissionRejected: "未通过", model.AdmissionSupplement: "待补充材料", model.AdmissionLegacy: "历史待核验"}
	return labels[stage]
}
