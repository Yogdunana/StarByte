/**
 * 通用类型定义
 */

/** 通用选项（下拉选择、标签等） */
export interface Option<T = string | number> {
  label: string;
  value: T;
  disabled?: boolean;
  color?: string;
}

/** 面包屑项 */
export interface BreadcrumbItem {
  label: string;
  path?: string;
}

/** 状态映射项 */
export interface StatusMapItem {
  color: string;
  text: string;
}

/** 状态映射表 */
export type StatusMap = Record<number, StatusMapItem>;

/** 搜索字段类型 */
export type SearchFieldType = 'text' | 'select' | 'date' | 'daterange' | 'number';

/** 搜索字段配置 */
export interface SearchField {
  name: string;
  label: string;
  type: SearchFieldType;
  options?: Option[];
  placeholder?: string;
  width?: number;
  initialValue?: unknown;
}

/** 表格分页配置 */
export interface TablePagination {
  page: number;
  pageSize: number;
  total: number;
  onChange: (page: number, pageSize: number) => void;
}

/** 排序信息 */
export interface SortInfo {
  field: string;
  order: 'ascend' | 'descend' | null;
}

/** 操作响应（通用增删改） */
export interface ActionResponse {
  success: boolean;
  message?: string;
}

/** 上传文件信息 */
export interface UploadFileInfo {
  uid: string;
  name: string;
  url?: string;
  status?: 'uploading' | 'done' | 'error' | 'removed';
  size?: number;
  type?: string;
}

/** 性别枚举 */
export const GenderMap: StatusMap = {
  0: { color: 'default', text: '未知' },
  1: { color: 'blue', text: '男' },
  2: { color: 'magenta', text: '女' },
};

/** 用户状态映射 */
export const UserStatusMap: StatusMap = {
  0: { color: 'success', text: '正常' },
  1: { color: 'error', text: '禁用' },
  2: { color: 'warning', text: '锁定' },
};

/** 通用启用/禁用状态 */
export const EnableStatusMap: StatusMap = {
  0: { color: 'success', text: '启用' },
  1: { color: 'error', text: '禁用' },
};
