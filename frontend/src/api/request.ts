import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { message } from 'antd';
import { getToken, getRefreshToken, setToken, setRefreshToken, removeToken } from '@/utils/storage';

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

const request: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
});

// 生成请求ID
const generateRequestId = (): string => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
};

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
    const { code, message: msg, data } = response.data;

    if (code === 0) {
      return data;
    }

    // 业务错误
    message.error(msg || '请求失败');
    return Promise.reject(new Error(msg || '请求失败'));
  },
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401) {
      // Token 过期
      if (!isRefreshing) {
        isRefreshing = true;

        try {
          const refreshToken = getRefreshToken();
          if (!refreshToken) {
            throw new Error('无 refresh token');
          }

          // 调用刷新 Token 接口
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

    // 其他错误
    const errorMsg = error.response?.data?.message || error.message || '网络错误';
    message.error(errorMsg);
    return Promise.reject(error);
  }
);

export default request;
