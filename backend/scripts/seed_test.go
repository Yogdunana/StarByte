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
	for _, need := range []string{"member_approved", "interview_invite", "meeting_notice", "discipline_notice", "task_assigned", "activity_registered", "activity_waitlist", "schedule_reminder", "announcement_published"} {
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

	assert.NotEmpty(t, departmentRefTables)
	seenTbl := map[string]bool{}
	for _, tbl := range departmentRefTables {
		assert.NotEmpty(t, tbl)
		assert.False(t, seenTbl[tbl], "duplicate ref table %s", tbl)
		seenTbl[tbl] = true
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
	for _, need := range []string{
		"system:config", "config:create", "config:update", "config:delete",
		"cache:delete", "cache:manage",
		"scheduler:create", "scheduler:update", "scheduler:delete", "scheduler:run", "scheduler:manage",
	} {
		assert.True(t, excluded[need], "vice_president must not inherit %s", need)
	}
	assert.False(t, excluded["config:read"])
	assert.False(t, excluded["cache:read"])
	assert.False(t, excluded["scheduler:read"])
	assert.False(t, excluded["search:read"])
}

func TestAllSeedPermissions_IncludesSearch(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	assert.True(t, seen["search:read"], "missing permission search:read")
}

func TestAllSeedPermissions_IncludesCache(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	for _, need := range []string{"cache:read", "cache:delete", "cache:manage"} {
		assert.True(t, seen[need], "missing permission %s", need)
	}
	assert.False(t, seen["cache:create"])
	assert.False(t, seen["cache:update"])
}

func TestAllSeedPermissions_IncludesScheduler(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	for _, need := range []string{
		"scheduler:read", "scheduler:create", "scheduler:update",
		"scheduler:delete", "scheduler:run", "scheduler:manage",
	} {
		assert.True(t, seen[need], "missing permission %s", need)
	}
}

func TestSeedDicts_SystemTypes(t *testing.T) {
	codes := map[string]bool{}
	for _, typ := range seedDictTypesData {
		codes[typ.Code] = true
		assert.NotEmpty(t, typ.Items)
	}
	for _, need := range []string{"applicant_type", "interview_result", "task_status", "task_priority", "internship_status", "internship_type", "meeting_status"} {
		assert.True(t, codes[need], "missing dict type %s", need)
	}
}

func TestAllSeedPermissions_IncludesDict(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	for _, need := range []string{"dict:read", "dict:create", "dict:update", "dict:delete"} {
		assert.True(t, seen[need], "missing permission %s", need)
	}
}

func TestAllSeedPermissions_IncludesSession(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	for _, need := range []string{"session:read", "session:delete"} {
		assert.True(t, seen[need], "missing permission %s", need)
	}
	assert.False(t, seen["session:create"])
	assert.False(t, seen["session:update"])
}

func TestAllSeedPermissions_IncludesExport(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	for _, need := range []string{
		"export:excel", "export:csv", "export:pdf", "export:json",
		"export:template", "export:read", "export:download",
	} {
		assert.True(t, seen[need], "missing permission %s", need)
	}
	assert.False(t, seen["export:create"])
	assert.False(t, seen["export:update"])
	assert.False(t, seen["export:delete"])
}

func TestAllSeedPermissions_IncludesSchedule(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	for _, need := range []string{"schedule:read", "schedule:create", "schedule:update", "schedule:delete"} {
		assert.True(t, seen[need], "missing permission %s", need)
	}
}

func TestOfficerAndMemberPerms_IncludeSchedule(t *testing.T) {
	assert.Contains(t, officerPermCodes(), "schedule:read")
	assert.Contains(t, officerPermCodes(), "schedule:create")
	assert.Contains(t, memberPermCodes(), "schedule:read")
	assert.Contains(t, memberPermCodes(), "schedule:update")
}

func TestAllSeedPermissions_IncludesOpsModules(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range allSeedPermissions() {
		seen[p.Code] = true
	}
	for _, need := range []string{
		"finance:read", "finance:create", "finance:manage",
		"discipline:read", "discipline:create", "discipline:approve", "discipline:revoke",
		"contract:read", "contract:create", "contract:manage",
		"activity:read", "activity:create", "activity:update", "activity:delete", "activity:manage",
		"announcement:read", "announcement:create", "announcement:update", "announcement:delete",
		"announcement:publish", "announcement:manage",
		"leave:read", "leave:approve",
	} {
		assert.True(t, seen[need], "missing permission %s", need)
	}
}

func TestOfficerAndMemberPerms_NonEmpty(t *testing.T) {
	assert.GreaterOrEqual(t, len(officerPermCodes()), 8)
	assert.GreaterOrEqual(t, len(memberPermCodes()), 5)
	assert.Contains(t, officerPermCodes(), "task:create")
	assert.Contains(t, memberPermCodes(), "file:read")
	assert.Contains(t, memberPermCodes(), "discipline:read")
	assert.Contains(t, officerPermCodes(), "discipline:read")
	assert.Contains(t, memberPermCodes(), "activity:read")
	assert.Contains(t, officerPermCodes(), "activity:read")
	assert.Contains(t, memberPermCodes(), "announcement:read")
	assert.Contains(t, officerPermCodes(), "announcement:read")
	assert.Contains(t, officerPermCodes(), "leave:read")
	assert.NotContains(t, memberPermCodes(), "leave:read")
}

func TestSeedMemberProfiles_StudentNos(t *testing.T) {
	assert.Len(t, seedProfileData, 2)
	seenUser := map[string]bool{}
	seenNo := map[string]bool{}
	for _, row := range seedProfileData {
		assert.NotEmpty(t, row.Username)
		assert.NotEmpty(t, row.StudentNo)
		assert.NotEmpty(t, row.RealName)
		assert.False(t, seenUser[row.Username], "duplicate username %s", row.Username)
		assert.False(t, seenNo[row.StudentNo], "duplicate student_no %s", row.StudentNo)
		seenUser[row.Username] = true
		seenNo[row.StudentNo] = true
	}
	assert.Equal(t, "20210001", seedProfileData[0].StudentNo)
	assert.Equal(t, "20210002", seedProfileData[1].StudentNo)
}
