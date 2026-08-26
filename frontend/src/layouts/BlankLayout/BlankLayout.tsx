import React from 'react';

export interface BlankLayoutProps {
  children: React.ReactNode;
}

/**
 * 空白布局 — 用于登录页、404 等无需侧边栏的页面
 */
const BlankLayout: React.FC<BlankLayoutProps> = ({ children }) => {
  return (
    <div style={{ minHeight: '100vh', width: '100%' }}>
      {children}
    </div>
  );
};

export default BlankLayout;
