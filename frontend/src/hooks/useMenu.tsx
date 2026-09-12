import React, { useEffect, useMemo, useState } from 'react';
import type { MenuProps } from 'antd';
import { useLocation } from 'react-router-dom';
import { useSelector } from 'react-redux';
import {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  ScheduleOutlined,
  CalendarOutlined,
  CarryOutOutlined,
  CheckCircleOutlined,
  ApartmentOutlined,
  ReadOutlined,
  BarChartOutlined,
  SettingOutlined,
  BellOutlined,
  FolderOutlined,
  BookOutlined,
  FormOutlined,
  DollarOutlined,
  AlertOutlined,
  FileProtectOutlined,
  NotificationOutlined,
  FieldTimeOutlined,
  MonitorOutlined,
  CloudServerOutlined,
} from '@ant-design/icons';

import routes, { AppRouteObject, RouteMeta } from '@/router/routes';
import { selectCollapsed } from '@/store/slices/appSlice';
import { selectPermissions } from '@/store/slices/userSlice';
import { useTranslation } from 'react-i18next';

export type MenuItem = NonNullable<MenuProps['items']>[number];

export interface UseMenuResult {
  menuItems: MenuItem[];
  selectedKeys: string[];
  openKeys: string[];
  setOpenKeys: (keys: string[]) => void;
  collapsed: boolean;
  searchKeyword: string;
  setSearchKeyword: (keyword: string) => void;
}

interface MenuNode {
  key: string;
  icon?: React.ReactNode;
  label: string;
  children?: MenuNode[];
}

const iconMap: Record<string, React.FC> = {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  ScheduleOutlined,
  CalendarOutlined,
  CarryOutOutlined,
  CheckCircleOutlined,
  ApartmentOutlined,
  ReadOutlined,
  BarChartOutlined,
  SettingOutlined,
  BellOutlined,
  FolderOutlined,
  BookOutlined,
  FormOutlined,
  DollarOutlined,
  AlertOutlined,
  FileProtectOutlined,
  NotificationOutlined,
  FieldTimeOutlined,
  MonitorOutlined,
  CloudServerOutlined,
};

function hasMenuPermission(permissions: string[], meta?: RouteMeta): boolean {
  if (permissions.includes('*')) return true;
  if (!meta?.permission) return true;
  return permissions.includes(meta.permission);
}

function hasVisibleChildren(route: AppRouteObject, permissions: string[]): boolean {
  if (!route.children || route.children.length === 0) return false;
  return route.children.some((child) => {
    if (child.meta?.hidden) return false;
    if (child.children && child.children.length > 0) {
      return hasVisibleChildren(child, permissions);
    }
    return hasMenuPermission(permissions, child.meta);
  });
}

function buildMenuNodes(
  routeList: AppRouteObject[],
  permissions: string[],
  parentPath: string,
  labelOf: (path: string, fallback: string) => string,
): MenuNode[] {
  return routeList
    .filter((route) => !route.meta?.hidden && route.path && route.path !== '*')
    .filter((route) => {
      if (route.children && route.children.length > 0) {
        return hasVisibleChildren(route, permissions);
      }
      return hasMenuPermission(permissions, route.meta);
    })
    .map((route) => {
      const fullPath = route.path!.startsWith('/')
        ? route.path!
        : `${parentPath}/${route.path}`;
      const IconComponent = route.meta?.icon ? iconMap[route.meta.icon] : null;
      const node: MenuNode = {
        key: fullPath,
        icon: IconComponent ? <IconComponent /> : undefined,
        label: labelOf(fullPath, route.meta?.title || route.path || ''),
      };
      if (route.children && route.children.length > 0) {
        const childItems = buildMenuNodes(
          route.children.filter((c) => c.path !== 'index'),
          permissions,
          fullPath,
          labelOf,
        );
        if (childItems.length > 0) {
          node.children = childItems;
        }
      }
      return node;
    });
}

function filterMenuNodes(nodes: MenuNode[], keyword: string): MenuNode[] {
  const kw = keyword.trim().toLowerCase();
  if (!kw) return nodes;
  const result: MenuNode[] = [];
  for (const node of nodes) {
    const selfMatch = node.label.toLowerCase().includes(kw);
    const children = node.children ? filterMenuNodes(node.children, keyword) : [];
    if (selfMatch) {
      result.push(node);
    } else if (children.length > 0) {
      result.push({ ...node, children });
    }
  }
  return result;
}

function collectOpenKeys(nodes: MenuNode[]): string[] {
  const keys: string[] = [];
  for (const node of nodes) {
    if (node.children && node.children.length > 0) {
      keys.push(node.key, ...collectOpenKeys(node.children));
    }
  }
  return keys;
}

function toAntdItems(nodes: MenuNode[]): MenuItem[] {
  return nodes.map((node) => ({
    key: node.key,
    icon: node.icon,
    label: node.label,
    children: node.children ? toAntdItems(node.children) : undefined,
  })) as MenuItem[];
}

/**
 * 侧栏菜单：按权限过滤、关键词搜索、折叠与 selected/open keys。
 */
export function useMenu(): UseMenuResult {
  const location = useLocation();
  const collapsed = useSelector(selectCollapsed);
  const permissions = useSelector(selectPermissions);
  const { t } = useTranslation();
  const [searchKeyword, setSearchKeyword] = useState('');
  const [openKeys, setOpenKeys] = useState<string[]>([]);

  const allNodes = useMemo(() => {
    const layoutRoute = (routes as AppRouteObject[]).find((r) =>
      r.children?.some((child) => child.path === 'dashboard'),
    );
    const labelOf = (path: string, fallback: string) => t(`menu.${path}`, { defaultValue: fallback });
    return layoutRoute?.children
      ? buildMenuNodes(layoutRoute.children, permissions, '', labelOf)
      : [];
  }, [permissions, t]);

  const visibleNodes = useMemo(
    () => filterMenuNodes(allNodes, searchKeyword),
    [allNodes, searchKeyword],
  );

  const menuItems = useMemo(() => {
    const sections = [
      { title: '我的工作', keys: ['/dashboard', '/notification'] },
      { title: '成员与招新', keys: ['/member', '/interview', '/discipline'] },
      { title: '协作与活动', keys: ['/task', '/meeting', '/internship', '/announcement', '/knowledge', '/leave'] },
      { title: '资源与财务', keys: ['/files', '/finance', '/contract', '/stats'] },
      { title: '组织与系统', keys: ['/user', '/workflow', '/forms', '/system', '/monitor', '/backup'] },
    ];
    const known = new Set(sections.flatMap(section => section.keys));
    const otherNodes = visibleNodes.filter(node => !known.has(node.key));
    return [...sections.flatMap(section => {
      const children = visibleNodes.filter(node => section.keys.includes(node.key));
      return children.length ? [{ type: 'group' as const, key: section.title, label: section.title, children: toAntdItems(children) }] : [];
    }), ...toAntdItems(otherNodes)];
  }, [visibleNodes]);
  const selectedKeys = useMemo(() => [location.pathname], [location.pathname]);

  useEffect(() => {
    if (searchKeyword.trim()) {
      setOpenKeys(collectOpenKeys(visibleNodes));
      return;
    }
    const parts = location.pathname.split('/').filter(Boolean);
    setOpenKeys(parts.length < 2 ? [] : [`/${parts[0]}`]);
  }, [location.pathname, searchKeyword, visibleNodes]);

  return {
    menuItems,
    selectedKeys,
    openKeys,
    setOpenKeys,
    collapsed,
    searchKeyword,
    setSearchKeyword,
  };
}
