package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/stretchr/testify/require"
)

func TestHasMembershipRole(t *testing.T) {
	require.False(t, hasMembershipRole(nil))
	require.False(t, hasMembershipRole([]string{"user"}))
	require.False(t, hasMembershipRole([]string{"super_admin"}))
	require.True(t, hasMembershipRole([]string{"user", "member"}))
	require.True(t, hasMembershipRole([]string{"probationary"}))
	require.True(t, hasMembershipRole([]string{"honorary"}))
}

func TestRejectDuplicateMembership(t *testing.T) {
	t.Run("member role without profile cannot reapply as member", func(t *testing.T) {
		err := rejectDuplicateMembership(int(model.ApplicantMember), nil, []string{"member"})
		requireAppError(t, err, response.CodeMemberAppDuplicate, "已具有该成员身份，无需重复申请")
	})
	t.Run("member role without profile still fails officer via later check", func(t *testing.T) {
		require.NoError(t, rejectDuplicateMembership(int(model.ApplicantOfficer), nil, []string{"member"}))
	})
	t.Run("probation blocks officer with leftover-member message", func(t *testing.T) {
		err := rejectDuplicateMembership(int(model.ApplicantOfficer), &model.MemberProfile{
			Status: model.ProfileProbation, MemberType: model.MemberTypeMember,
		}, nil)
		requireAppError(t, err, response.CodeMemberAppDuplicate, "已有预备期申请，请等待处理")
	})
	t.Run("active member may apply as officer", func(t *testing.T) {
		require.NoError(t, rejectDuplicateMembership(int(model.ApplicantOfficer), &model.MemberProfile{
			Status: model.ProfileActive, MemberType: model.MemberTypeMember,
		}, []string{"member"}))
	})
	t.Run("active officer cannot reapply as officer", func(t *testing.T) {
		err := rejectDuplicateMembership(int(model.ApplicantOfficer), &model.MemberProfile{
			Status: model.ProfileActive, MemberType: model.MemberTypeOfficer,
		}, []string{"officer"})
		requireAppError(t, err, response.CodeMemberAppDuplicate, "已具有该成员身份，无需重复申请")
	})
}
