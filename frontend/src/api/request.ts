import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosError } from 'axios';
import { message } from 'antd';
import { getToken, getRefreshToken, setToken, setRefreshToken, removeToken } from '@/utils/storage';
import { handleApiError } from './error';

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

const request: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 生成请求 ID
const generateRequestId = (): string => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
};

/**
 * 创建请求取消控制器
 *
 * @example
 * ```ts
 * const { controller, signal } = createCancelToken();
 * // 发起请求
 * getUserList(params, signal);
 * // 取消请求
 * controller.abort();
 * ```
 */
export function createCancelToken(): {
  controller: AbortController;
  signal: AbortSignal;
} {
  const controller = new AbortController();
  return { controller, signal: controller.signal };
}

// GET 请求失败自动重试次数
const GET_RETRY_COUNT = 1;
// 重试延迟（ms）
const RETRY_DELAY = 500;

// 判断是否为重试条件：网络错误或 5xx 且是 GET 请求
function shouldRetry(config: InternalAxiosRequestConfig, status?: number): boolean {
  if (config.method?.toUpperCase() !== 'GET') return false;
  const retryCount = (config as { _retryCount?: number })._retryCount ?? 0;
  if (retryCount >= GET_RETRY_COUNT) return false;
  // 网络错误（无 status）或 5xx 服务端错误时重试
  return !status || status >= 500;
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 添加请求 ID
    config.headers['X-Request-ID'] = generateRequestId();

    // 添加 Token
    const token = getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    return config;
  },
  (error) => Promise.reject(error)
);

// 是否正在刷新 Token
let isRefreshing = false;
// 等待刷新的请求队列
let pendingRequests: Array<(token: string) => void> = [];

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    // Blob 响应（文件下载）直接返回原始 response
    if (response.config.responseType === 'blob') {
      return response;
    }

    const { code, message: msg, data } = response.data;

    if (code === 0) {
      return data;
    }

    // 业务错误
    message.error(msg || '请求失败');
    return Promise.reject(new Error(msg || '请求失败'));
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig;

    // 处理 401 Token 过期
    if (error.response?.status === 401) {
      if (!isRefreshing) {
        isRefreshing = true;

        try {
          const refreshToken = getRefreshToken();
          if (!refreshToken) {
            throw new Error('无 refresh token');
          }

          // 调用刷新 Token 接口（用 axios 直接调用，避免走拦截器循环）
          const response = await axios.post(`${baseURL}/auth/refresh`, {
            refresh_token: refreshToken,
          });

          const { access_token, refresh_token } = response.data.data;
          setToken(access_token);
          setRefreshToken(refresh_token);

          // 执行队列中的请求
          pendingRequests.forEach((callback) => callback(access_token));
          pendingRequests = [];

          // 重试当前请求
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return request(originalRequest);
        } catch (refreshError) {
          // 刷新失败，跳转到登录页
          removeToken();
          message.error('登录已过期，请重新登录');
          window.location.href = '/login';
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      } else {
        // 正在刷新，加入队列等待
        return new Promise((resolve) => {
          pendingRequests.push((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            resolve(request(originalRequest));
          });
        });
      }
    }

    // GET 请求失败重试
    if (shouldRetry(originalRequest, error.response?.status)) {
      const retryCount = (originalRequest as { _retryCount?: number })._retryCount ?? 0;
      (originalRequest as { _retryCount?: number })._retryCount = retryCount + 1;
      await delay(RETRY_DELAY);
      return request(originalRequest);
    }

    // 其他错误：统一错误处理
    const errorMsg = handleApiError(error);
    message.error(errorMsg);
    return Promise.reject(error);
  }
);

export default request;
