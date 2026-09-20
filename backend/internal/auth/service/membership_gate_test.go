package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplicationGates(t *testing.T) {
	member, officer := applicationGates([]string{"user"}, nil)
	require.True(t, member)
	require.False(t, officer)

	member, officer = applicationGates([]string{"member"}, nil)
	require.False(t, member)
	require.False(t, officer)

	member, officer = applicationGates([]string{"member"}, &MemberIdentity{Status: 0, MemberType: 1})
	require.False(t, member)
	require.True(t, officer)

	member, officer = applicationGates([]string{"user"}, &MemberIdentity{Status: 0, MemberType: 1})
	require.False(t, member)
	require.True(t, officer)

	member, officer = applicationGates([]string{"officer"}, &MemberIdentity{Status: 0, MemberType: 2})
	require.False(t, member)
	require.False(t, officer)

	member, officer = applicationGates([]string{"probationary"}, &MemberIdentity{Status: 3, MemberType: 1})
	require.False(t, member)
	require.False(t, officer)

	member, officer = applicationGates([]string{"super_admin"}, &MemberIdentity{Status: 0, MemberType: 4})
	require.False(t, member)
	require.False(t, officer)
}
