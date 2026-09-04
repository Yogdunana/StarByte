package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/stretchr/testify/assert"
)

func TestMemberStateMachine(t *testing.T) {
	next, ok := nextStatus(model.ApplicantMember, model.AppPending, actionApprove)
	assert.True(t, ok)
	assert.Equal(t, model.AppApproved, next)

	next, ok = nextStatus(model.ApplicantMember, model.AppPending, actionReject)
	assert.True(t, ok)
	assert.Equal(t, model.AppRejected, next)

	next, ok = nextStatus(model.ApplicantMember, model.AppPending, actionSupplement)
	assert.True(t, ok)
	assert.Equal(t, model.AppSupplement, next)

	_, ok = nextStatus(model.ApplicantMember, model.AppApproved, actionApprove)
	assert.False(t, ok)

	next, ok = nextStatus(model.ApplicantMember, model.AppSupplement, actionResubmit)
	assert.True(t, ok)
	assert.Equal(t, model.AppPending, next)
}

func TestOfficerStateMachine(t *testing.T) {
	next, ok := nextStatus(model.ApplicantOfficer, model.AppPending, actionApprove)
	assert.True(t, ok)
	assert.Equal(t, model.AppReviewing, next)

	next, ok = nextStatus(model.ApplicantOfficer, model.AppReviewing, actionApprove)
	assert.True(t, ok)
	assert.Equal(t, model.AppInterviewing, next)

	next, ok = nextStatus(model.ApplicantOfficer, model.AppInterviewing, actionApprove)
	assert.True(t, ok)
	assert.Equal(t, model.AppApproved, next)

	next, ok = nextStatus(model.ApplicantOfficer, model.AppPending, actionReject)
	assert.True(t, ok)
	assert.Equal(t, model.AppRejected, next)

	next, ok = nextStatus(model.ApplicantOfficer, model.AppReviewing, actionReject)
	assert.True(t, ok)
	assert.Equal(t, model.AppRejected, next)

	_, ok = nextStatus(model.ApplicantOfficer, model.AppInterviewing, actionSupplement)
	assert.False(t, ok)
}

func TestStageLabel(t *testing.T) {
	assert.Equal(t, "待审核", stageLabel(model.AppPending))
	assert.Equal(t, "面试中", stageLabel(model.AppInterviewing))
	assert.Equal(t, "补充材料", stageLabel(model.AppSupplement))
}
