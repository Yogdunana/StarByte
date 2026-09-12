import { tx } from '@/i18n/text';
import axios, { type AxiosError } from 'axios';

/** 主动取消（AbortSignal / 卸载）不算失败，拦截器不应弹 toast。 */
export function isCanceledError(error: unknown): boolean {
  if (axios.isCancel(error)) return true;
  if (typeof error !== 'object' || error === null) return false;
  const e = error as { code?: string; name?: string };
  return e.code === 'ERR_CANCELED' || e.name === 'CanceledError' || e.name === 'AbortError';
}

/**
 * 业务错误码到中文消息的映射
 * 仅包含前端需要特殊处理或展示的常见错误码
 */
const errorCodeMessages: Record<number, string> = {
  get 1001() {
    return tx('参数错误');
  },
  get 1002() {
    return tx('未授权，请先登录');
  },
  get 1003() {
    return tx('没有权限执行此操作');
  },
  get 1004() {
    return tx('资源不存在');
  },
  get 1005() {
    return tx('请求过于频繁，请稍后再试');
  },
  get 2001() {
    return tx('用户不存在');
  },
  get 2002() {
    return tx('用户名或密码错误');
  },
  get 2003() {
    return tx('用户名已存在');
  },
  get 2004() {
    return tx('用户已被禁用');
  },
  get 2005() {
    return tx('原密码错误');
  },
  get 3001() {
    return tx('角色不存在');
  },
  get 3002() {
    return tx('权限不足');
  },
  get 3003() {
    return tx('角色编码已存在');
  },
  get 3004() {
    return tx('部门不存在');
  },
  get 4001() {
    return tx('流程定义不存在');
  },
  get 4002() {
    return tx('流程已结束');
  },
  get 4003() {
    return tx('流程实例不存在');
  },
  get 5001() {
    return tx('审计日志不存在');
  },
  get 5002() {
    return tx('导出格式不支持');
  },
  get 6001() {
    return tx('申请不存在');
  },
  get 6002() {
    return tx('当前状态不允许该操作');
  },
  get 6003() {
    return tx('已有待处理的入会申请');
  },
  get 6004() {
    return tx('档案不存在');
  },
  get 6005() {
    return tx('无权操作该档案');
  },
  get 6006() {
    return tx('学号已存在');
  },
  get 6007() {
    return tx('导出失败');
  },
  get 6008() {
    return tx('请补充必填字段');
  },
  get 7001() {
    return tx('面试场次不存在');
  },
  get 7002() {
    return tx('时间冲突');
  },
  get 8001() {
    return tx('会议不存在');
  },
  get 8002() {
    return tx('投票已结束');
  },
  get 9001() {
    return tx('任务不存在');
  },
  get 9002() {
    return tx('无权操作此任务');
  },
  get 10001() {
    return tx('实习记录不存在');
  },
  get 10002() {
    return tx('无权操作此实习记录');
  },
  get 10003() {
    return tx('实习状态不允许该操作');
  },
  get 10004() {
    return tx('实习已结束，无法修改');
  },
  get 10005() {
    return tx('排行榜暂未开放');
  },
  get 10006() {
    return tx('实习已完成');
  },
  get 12001() {
    return tx('通知不存在');
  },
  get 12003() {
    return tx('通知模板已存在');
  },
  get 22001() {
    return tx('表单不存在');
  },
  get 22002() {
    return tx('表单未发布');
  },
  get 22003() {
    return tx('必填字段缺失');
  },
  get 22004() {
    return tx('字段校验失败');
  },
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
    return tx('请求失败');
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
        return tx('参数错误');
      case 401:
        return tx('未授权，请先登录');
      case 403:
        return tx('没有权限执行此操作');
      case 404:
        return tx('请求的资源不存在');
      case 408:
        return tx('请求超时，请稍后重试');
      case 429:
        return tx('请求过于频繁，请稍后再试');
      case 500:
        return tx('服务器内部错误');
      case 502:
        return tx('网关错误');
      case 503:
        return tx('服务暂时不可用');
      case 504:
        return tx('网关超时');
      default:
        return tx('请求失败（{{value0}}）', { value0: status });
    }
  }

  // 网络错误（无 response）
  if (axiosError.code === 'ECONNABORTED') {
    return tx('请求超时，请稍后重试');
  }
  if (axiosError.code === 'ERR_NETWORK' || !axiosError.response) {
    return tx('网络连接失败，请检查网络');
  }

  return tx('请求失败');
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
