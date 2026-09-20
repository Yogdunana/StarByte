import { expect, it } from 'vitest';
import { applicationGates, blockedApplicationMessage, hasMembershipRole } from './applicationGates';

it('treats association offices as membership for hiding the member form', () => {
  expect(hasMembershipRole(['user'])).toBe(false);
  expect(hasMembershipRole(['member'])).toBe(true);
  expect(hasMembershipRole(['probationary'])).toBe(true);
});

it('hides member apply when the API says so, even if a role leftover exists', () => {
  expect(applicationGates({ roles: ['member'], can_apply_member: false, can_apply_officer: true })).toEqual({
    canApplyMember: false,
    canApplyOfficer: true,
  });
  expect(applicationGates({ roles: ['member'], can_apply_member: false, can_apply_officer: false })).toEqual({
    canApplyMember: false,
    canApplyOfficer: false,
  });
});

it('falls back to roles when the API omits gate flags', () => {
  expect(applicationGates({ roles: ['user'] })).toEqual({ canApplyMember: true, canApplyOfficer: false });
  expect(applicationGates({ roles: ['member'] })).toEqual({ canApplyMember: false, canApplyOfficer: true });
});

it('explains leftover member role without a profile', () => {
  expect(blockedApplicationMessage({ roles: ['member'], can_apply_officer: false })).toBe(
    '当前账号有会员角色但没有有效会员档案，无法申请干事。请管理员核对历史数据。',
  );
  expect(blockedApplicationMessage({ roles: ['probationary'] })).toBe(
    '档案处于预备期，请等待处理后再申请干事。',
  );
});
