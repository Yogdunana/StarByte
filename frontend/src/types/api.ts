// API 响应通用类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  request_id: string;
  timestamp: number;
}

// 分页响应
export interface PageResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

// 分页请求参数
export interface PageParams {
  page?: number;
  page_size?: number;
  keyword?: string;
}

// 用户信息
export interface UserInfo {
  id: string;
  username: string;
  real_name: string;
  avatar_url: string;
  email: string;
  phone: string;
  gender: number;
  status: number;
  department_id: string;
  position_id: string;
  roles: RoleInfo[];
  permissions: string[];
  created_at: string;
}

// 角色信息
export interface RoleInfo {
  id: string;
  name: string;
  code: string;
}

// 登录响应
export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  access_token_expires: number;
  refresh_token_expires: number;
  user: UserInfo;
}

// 登录请求
export interface LoginRequest {
  username: string;
  password: string;
}

// 注册请求
export interface RegisterRequest {
  username: string;
  password: string;
  real_name?: string;
  email?: string;
  phone?: string;
}
