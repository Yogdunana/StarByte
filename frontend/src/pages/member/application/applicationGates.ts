import { tx } from '@/i18n/text';
import type { UserInfo } from '@/types/api';

const membershipRoles = new Set([
  'member',
  'officer',
  'probationary',
  'president',
  'vice_president',
  'minister',
  'vice_minister',
  'center_director',
  'vice_center_director',
  'honorary',
]);

export function hasMembershipRole(roles?: string[]): boolean {
  return (roles || []).some((role) => membershipRoles.has(role));
}

export function applicationGates(user?: Partial<UserInfo> | null): {
  canApplyMember: boolean;
  canApplyOfficer: boolean;
} {
  const roles = user?.roles || [];
  return {
    canApplyMember: user?.can_apply_member ?? !hasMembershipRole(roles),
    canApplyOfficer: user?.can_apply_officer ?? roles.includes('member'),
  };
}

export function blockedApplicationMessage(user?: Partial<UserInfo> | null): string {
  const roles = user?.roles || [];
  if (roles.includes('probationary')) {
    return tx('档案处于预备期，请等待处理后再申请干事。');
  }
  if (roles.includes('member') && user?.can_apply_officer === false) {
    return tx('当前账号有会员角色但没有有效会员档案，无法申请干事。请管理员核对历史数据。');
  }
  return tx('已具有该成员身份，无需再提交入会申请。');
}
