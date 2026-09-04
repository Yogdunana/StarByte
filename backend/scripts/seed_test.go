package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllSeedPermissions_Count(t *testing.T) {
	perms := allSeedPermissions()
	assert.GreaterOrEqual(t, len(perms), 40)

	seen := map[string]struct{}{}
	for _, p := range perms {
		assert.NotEmpty(t, p.Code)
		_, dup := seen[p.Code]
		assert.False(t, dup, "duplicate permission %s", p.Code)
		seen[p.Code] = struct{}{}
	}
}

func TestSeedRoles_IncludesRequired(t *testing.T) {
	codes := map[string]bool{}
	for _, r := range seedRolesData {
		codes[r.Code] = true
	}
	for _, need := range []string{"president", "vice_president", "minister", "vice_minister", "officer", "member"} {
		assert.True(t, codes[need], "missing role %s", need)
	}
}

func TestSeedTemplates_AtLeastFive(t *testing.T) {
	assert.GreaterOrEqual(t, len(seedTemplatesData), 5)
	codes := map[string]bool{}
	for _, tpl := range seedTemplatesData {
		codes[tpl.Code] = true
	}
	for _, need := range []string{"member_approved", "interview_invite", "meeting_notice", "discipline_notice", "task_assigned"} {
		assert.True(t, codes[need], "missing template %s", need)
	}
}

func TestSeedDepartments_Four(t *testing.T) {
	assert.Len(t, seedDepartmentsData, 4)
}
