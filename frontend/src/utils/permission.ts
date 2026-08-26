/**
 * 权限检查工具（非 Hook 版本，用于非组件场景）
 */
import { store } from '@/store';
import { selectPermissions } from '@/store/slices/userSlice';

/**
 * 检查单个权限
 */
export function hasPermission(permissionCode: string): boolean {
  const permissions = selectPermissions(store.getState());
  if (permissions.includes('*')) return true;
  return permissions.includes(permissionCode);
}

/**
 * 检查是否有任意一个权限
 */
export function hasAnyPermission(permissionCodes: string[]): boolean {
  const permissions = selectPermissions(store.getState());
  if (permissions.includes('*')) return true;
  return permissionCodes.some((code) => permissions.includes(code));
}

/**
 * 检查是否拥有所有权限
 */
export function hasAllPermissions(permissionCodes: string[]): boolean {
  const permissions = selectPermissions(store.getState());
  if (permissions.includes('*')) return true;
  return permissionCodes.every((code) => permissions.includes(code));
}
