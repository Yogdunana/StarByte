/**
 * 后端 associationOffice 的前端镜像。
 *
 * 源码在 backend/internal/rbac/handler/role_membership.go（同名函数）。
 * 两边必须保持一致，否则会出现「后端其实允许，但前端把入口藏了」的鬼打墙 ——
 * 指导老师、荣誉会员、队长、队员、副中心主任之前都踩过：它们 IsSystem=true，
 * 前端用 is_system 判断可编辑性，于是登记成员的入口直接不渲染。
 *
 * - unscoped：协会职务但不绑任何部门/中心（会长、指导老师、荣誉会员、队长、队员）
 * - center：需要选一个中心（副会长、中心主任、副中心主任）
 * - department：需要选一个职能部门（部长）
 * - ''：普通角色，不属于章程里的协会职务
 */
export type AssociationOffice = 'unscoped' | 'center' | 'department' | '';

export function associationOffice(code?: string): AssociationOffice {
  switch (code) {
    case 'president':
    case 'advisor':
    case 'honorary':
    case 'captain':
    case 'teammate':
      return 'unscoped';
    case 'vice_president':
    case 'center_director':
    case 'vice_center_director':
      return 'center';
    case 'minister':
      return 'department';
    default:
      return '';
  }
}

/** 是否需要选任职范围（中心或部门）。会长是 unscoped，不需要。 */
export const isScopedOffice = (code?: string): boolean => {
  const office = associationOffice(code);
  return office === 'center' || office === 'department';
};
