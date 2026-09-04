import { useMemo } from 'react';
import { useSelector } from 'react-redux';
import { selectPermissions, selectRoles } from '@/store/slices/userSlice';

/**
 * 检查单个权限
 */
export function usePermission(permissionCode: string): boolean {
  const permissions = useSelector(selectPermissions);
  return useMemo(() => {
    // 超级管理员拥有所有权限（这里简化判断，实际应该从后端返回）
    if (permissions.includes('*')) {
      return true;
    }
    return permissions.includes(permissionCode);
  }, [permissions, permissionCode]);
}

/**
 * 批量检查权限
 */
export function usePermissions(permissionCodes: string[]): boolean[] {
  const permissions = useSelector(selectPermissions);
  return useMemo(() => {
    const isSuperAdmin = permissions.includes('*');
    return permissionCodes.map((code) => isSuperAdmin || permissions.includes(code));
  }, [permissions, permissionCodes]);
}

/**
 * 检查是否有任意一个权限
 */
export function useAnyPermission(permissionCodes: string[]): boolean {
  const permissions = useSelector(selectPermissions);
  return useMemo(() => {
    if (permissions.includes('*')) {
      return true;
    }
    return permissionCodes.some((code) => permissions.includes(code));
  }, [permissions, permissionCodes]);
}

/**
 * 检查是否拥有所有权限
 */
export function useAllPermissions(permissionCodes: string[]): boolean {
  const permissions = useSelector(selectPermissions);
  return useMemo(() => {
    if (permissions.includes('*')) {
      return true;
    }
    return permissionCodes.every((code) => permissions.includes(code));
  }, [permissions, permissionCodes]);
}

/**
 * 检查当前用户是否拥有指定角色。
 * super_admin 视为拥有全部角色。
 */
export function useHasRole(roleCode: string): boolean {
  const roles = useSelector(selectRoles);
  return useMemo(() => {
    if (roles.includes('super_admin')) {
      return true;
    }
    return roles.includes(roleCode);
  }, [roles, roleCode]);
}

/** Issue #20：usePermission 模块提供 hasRole */
export const hasRole = useHasRole;
