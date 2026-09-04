import type { AxiosError } from 'axios';

/**
 * 业务错误码到中文消息的映射
 * 仅包含前端需要特殊处理或展示的常见错误码
 */
const errorCodeMessages: Record<number, string> = {
  1001: '参数错误',
  1002: '未授权，请先登录',
  1003: '没有权限执行此操作',
  1004: '资源不存在',
  1005: '请求过于频繁，请稍后再试',
  2001: '用户不存在',
  2002: '用户名或密码错误',
  2003: '用户名已存在',
  2004: '用户已被禁用',
  2005: '原密码错误',
  3001: '角色不存在',
  3002: '权限不足',
  3003: '角色编码已存在',
  3004: '部门不存在',
  4001: '流程定义不存在',
  4002: '流程已结束',
  4003: '流程实例不存在',
  5001: '审计日志不存在',
  5002: '导出格式不支持',
  6001: '申请不存在',
  6002: '当前状态不允许该操作',
  6003: '已有待处理的入会申请',
  6004: '档案不存在',
  6005: '无权操作该档案',
  6006: '学号已存在',
  6007: '导出失败',
  6008: '请补充必填字段',
  7001: '面试场次不存在',
  7002: '时间冲突',
  8001: '会议不存在',
  8002: '投票已结束',
  9001: '任务不存在',
  9002: '无权操作此任务',
  10001: '实习记录不存在',
  10002: '无权操作此实习记录',
  12001: '通知不存在',
  12003: '通知模板已存在',
};

/**
 * 从 API 错误中提取用户友好的错误消息
 *
 * 优先级：
 * 1. 后端返回的业务错误消息（response.data.message）
 * 2. 错误码映射的中文消息
 * 3. HTTP 状态码对应的通用消息
 * 4. 网络错误提示
 * 5. 默认「请求失败」
 */
export function handleApiError(error: unknown): string {
  // 非 Axios 错误
  if (!isAxiosError(error)) {
    if (error instanceof Error) return error.message;
    return '请求失败';
  }

  const axiosError = error as AxiosError<{ code?: number; message?: string }>;

  // 后端返回了业务错误消息
  if (axiosError.response?.data?.message) {
    return axiosError.response.data.message;
  }

  // 错误码映射
  const code = axiosError.response?.data?.code;
  if (code && errorCodeMessages[code]) {
    return errorCodeMessages[code];
  }

  // HTTP 状态码通用消息
  const status = axiosError.response?.status;
  if (status) {
    switch (status) {
      case 400:
        return '参数错误';
      case 401:
        return '未授权，请先登录';
      case 403:
        return '没有权限执行此操作';
      case 404:
        return '请求的资源不存在';
      case 408:
        return '请求超时，请稍后重试';
      case 429:
        return '请求过于频繁，请稍后再试';
      case 500:
        return '服务器内部错误';
      case 502:
        return '网关错误';
      case 503:
        return '服务暂时不可用';
      case 504:
        return '网关超时';
      default:
        return `请求失败（${status}）`;
    }
  }

  // 网络错误（无 response）
  if (axiosError.code === 'ECONNABORTED') {
    return '请求超时，请稍后重试';
  }
  if (axiosError.code === 'ERR_NETWORK' || !axiosError.response) {
    return '网络连接失败，请检查网络';
  }

  return '请求失败';
}

/**
 * 判断是否为 Axios 错误
 */
function isAxiosError(error: unknown): error is AxiosError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'isAxiosError' in error &&
    (error as AxiosError).isAxiosError === true
  );
}
