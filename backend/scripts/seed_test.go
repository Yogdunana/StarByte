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
	for _, need := range []string{"config:read", "config:create", "config:update", "config:delete"} {
		_, ok := seen[need]
		assert.True(t, ok, "missing permission %s", need)
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

func TestSeedDepartments_CharterLayout(t *testing.T) {
	assert.Len(t, seedCentersData, 3)
	assert.Len(t, seedDepartmentsData, 7)
	assert.Len(t, seedLegacyDeptRemaps, 4)

	centers := map[string]bool{}
	for _, c := range seedCentersData {
		assert.NotEmpty(t, c.Code)
		assert.Empty(t, c.ParentCode)
		assert.False(t, centers[c.Code], "duplicate center %s", c.Code)
		centers[c.Code] = true
	}

	depts := map[string]bool{}
	for _, d := range seedDepartmentsData {
		assert.NotEmpty(t, d.Code)
		assert.True(t, centers[d.ParentCode], "department %s parent %s missing", d.Code, d.ParentCode)
		assert.False(t, depts[d.Code], "duplicate department %s", d.Code)
		assert.False(t, centers[d.Code], "department code clashes with center %s", d.Code)
		depts[d.Code] = true
	}

	seenOld := map[string]bool{}
	for _, m := range seedLegacyDeptRemaps {
		assert.True(t, depts[m.New], "legacy remap target %s missing", m.New)
		assert.False(t, seenOld[m.Old], "duplicate legacy code %s", m.Old)
		seenOld[m.Old] = true
	}
}

func TestSeedRoles_OnlyTopRolesAreSystem(t *testing.T) {
	for _, r := range seedRolesData {
		switch r.Code {
		case "super_admin", "president":
			assert.True(t, r.IsSystem, "%s should be system", r.Code)
		default:
			assert.False(t, r.IsSystem, "%s should not be system", r.Code)
		}
	}
}

func TestSeedRuntimeConfigs_Keys(t *testing.T) {
	assert.GreaterOrEqual(t, len(seedConfigsData), 3)
	seen := map[string]bool{}
	for _, c := range seedConfigsData {
		assert.NotEmpty(t, c.Key)
		assert.False(t, seen[c.Key], "duplicate config %s", c.Key)
		seen[c.Key] = true
	}
}

func TestVicePresident_ExcludesConfigWrites(t *testing.T) {
	excluded := map[string]bool{}
	for _, c := range vicePresidentExcludedPerms() {
		excluded[c] = true
	}
	for _, need := range []string{"system:config", "config:create", "config:update", "config:delete"} {
		assert.True(t, excluded[need], "vice_president must not inherit %s", need)
	}
	assert.False(t, excluded["config:read"])
}

func TestOfficerAndMemberPerms_NonEmpty(t *testing.T) {
	assert.GreaterOrEqual(t, len(officerPermCodes()), 8)
	assert.GreaterOrEqual(t, len(memberPermCodes()), 5)
	assert.Contains(t, officerPermCodes(), "task:create")
	assert.Contains(t, memberPermCodes(), "file:read")
}
