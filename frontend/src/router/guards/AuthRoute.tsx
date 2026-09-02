import React, { useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { Spin } from 'antd';
import { useDispatch, useSelector } from 'react-redux';

import { getToken } from '@/utils/storage';
import { logout as logoutAction } from '@/store/slices/authSlice';
import { fetchCurrentUser, selectCurrentUser, selectUserLoading, selectUserError } from '@/store/slices/userSlice';
import type { AppDispatch } from '@/store';

/**
 * 全屏加载占位
 */
const LoadingScreen: React.FC = () => (
  <div style={{
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100vh',
    flexDirection: 'column',
    gap: 16,
  }}>
    <Spin size="large" />
    <span style={{ color: '#999', fontSize: 14 }}>正在加载用户信息...</span>
  </div>
);

export interface AuthRouteProps {
  children: React.ReactNode;
}

/**
 * 认证路由守卫
 *
 * 职责：
 * 1. 检查 localStorage 中是否存在 token，不存在则重定向到 /login
 * 2. token 存在但用户信息未加载时，自动触发 fetchCurrentUser
 * 3. 加载过程中显示全屏 Loading
 * 4. 加载失败（token 过期等）时清除认证状态并重定向到 /login
 */
const AuthRoute: React.FC<AuthRouteProps> = ({ children }) => {
  const location = useLocation();
  const dispatch = useDispatch<AppDispatch>();

  const token = getToken();
  const currentUser = useSelector(selectCurrentUser);
  const userLoading = useSelector(selectUserLoading);
  const userError = useSelector(selectUserError);

  useEffect(() => {
    if (token && !currentUser && !userError) {
      dispatch(fetchCurrentUser());
    }
  }, [dispatch, token, currentUser, userError]);

  // 无 token → 登录页
  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  // token 存在但用户信息获取失败 → 清除状态并重定向
  if (userError && !currentUser) {
    dispatch(logoutAction());
    return <Navigate to="/login" replace />;
  }

  // token 存在且正在加载用户信息 → 全屏 Loading
  if (!currentUser && (userLoading || !userError)) {
    return <LoadingScreen />;
  }

  // 认证通过 → 渲染子节点
  return <>{children}</>;
};

export default AuthRoute;
