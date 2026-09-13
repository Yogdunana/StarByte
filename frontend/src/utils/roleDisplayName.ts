import { tx } from '@/i18n/text';

/** Built-in office names follow the UI language; custom role names remain user content. */
export function roleDisplayName(role: { code: string; name: string }): string {
  switch (role.code) {
    case 'super_admin':
      return tx('超级管理员');
    case 'user':
      return tx('用户');
    case 'president':
      return tx('会长');
    case 'vice_president':
      return tx('副会长');
    case 'center_director':
      return tx('中心主任');
    case 'minister':
      return tx('部长');
    case 'vice_minister':
      return tx('副部长');
    case 'officer':
      return tx('干事');
    case 'member':
      return tx('会员');
    case 'probationary':
      return tx('候补成员');
    default:
      return role.name;
  }
}
