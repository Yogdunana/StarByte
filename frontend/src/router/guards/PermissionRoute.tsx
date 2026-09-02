import React from 'react';
import { Navigate } from 'react-router-dom';

import { usePermission } from '@/hooks/usePermission';

export interface PermissionRouteProps {
  /** 权限码，不传或为空字符串时视为无权限限制 */
  permission?: string;
  children: React.ReactNode;
}

/**
 * 权限路由守卫
 *
 * 职责：
 * - 当 route meta 中声明了 permission 字段时，检查当前用户是否持有该权限
 * - 无权限时重定向到 /403 页面
 * - 有权限或未声明 permission 时直接渲染 children
 */
const PermissionRoute: React.FC<PermissionRouteProps> = ({ permission, children }) => {
  const hasPermission = usePermission(permission || '');

  // 无权限码 → 不做限制
  if (!permission) {
    return <>{children}</>;
  }

  // 无权限 → 403
  if (!hasPermission) {
    return <Navigate to="/403" replace />;
  }

  return <>{children}</>;
};

export default PermissionRoute;
