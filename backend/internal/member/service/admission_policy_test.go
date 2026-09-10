package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

func TestAdmissionAuthority(t *testing.T) {
	dept, center, candidate := uuid.New(), uuid.New(), uuid.New()
	app := &model.MemberApplication{UserID: candidate, DepartmentID: &dept}
	tests := []struct {
		name, role         string
		actor              model.AdmissionActor
		allowed, delegated bool
	}{
		{"candidate president", "president", model.AdmissionActor{ID: candidate, Roles: []string{"president"}}, false, false},
		{"technical administrator", "president", model.AdmissionActor{ID: uuid.New(), Roles: []string{"super_admin"}}, false, false},
		{"department minister", "minister", model.AdmissionActor{ID: uuid.New(), DepartmentID: &dept, Roles: []string{"minister"}}, true, false},
		{"unrelated minister", "minister", model.AdmissionActor{ID: uuid.New(), DepartmentID: &center, Roles: []string{"minister"}}, false, true},
		{"center director", "center", model.AdmissionActor{ID: uuid.New(), DepartmentID: &center, Roles: []string{"vice_president"}}, true, false},
		{"president delegation", "minister", model.AdmissionActor{ID: uuid.New(), Roles: []string{"president"}}, true, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allowed, delegated := admissionAuthority(&test.actor, app, &center, test.role)
			require.Equal(t, test.allowed, allowed)
			if allowed {
				require.Equal(t, test.delegated, delegated)
			}
		})
	}
}
func TestCalendarMonthProbation(t *testing.T) {
	for _, test := range [][2]string{{"2026-01-31", "2026-02-28"}, {"2028-01-31", "2028-02-29"}, {"2026-12-31", "2027-01-31"}} {
		start, err := time.Parse("2006-01-02", test[0])
		require.NoError(t, err)
		require.Equal(t, test[1], calendarMonthLater(start).Format("2006-01-02"))
	}
}
func TestMemberApprovalDoesNotAddInterview(t *testing.T) {
	app := &model.MemberApplication{Type: model.ApplicantMember, AdmissionStage: model.AdmissionMaterials}
	advanceAdmission(app, time.Now())
	require.Equal(t, model.AdmissionApproved, app.AdmissionStage)
	require.Nil(t, app.ProbationUntil)
}
func TestOfficerRequiresBothRoundsAndFinalConfirmation(t *testing.T) {
	app := &model.MemberApplication{Type: model.ApplicantOfficer, AdmissionStage: model.AdmissionMaterials}
	now := time.Now()
	for _, next := range []string{model.AdmissionRound1, model.AdmissionRound2, model.AdmissionPresident, model.AdmissionProbation} {
		advanceAdmission(app, now)
		require.Equal(t, next, app.AdmissionStage)
	}
	require.NotNil(t, app.ProbationUntil)
}
