import { useEffect, useRef, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import {
  addNotification,
  setWSConnected,
  fetchUnreadCount,
  fetchRecentNotifications,
} from '@/store/slices/notificationSlice';
import { selectIsAuthenticated } from '@/store/slices/authSlice';
import { getToken } from '@/utils/storage';
import type { WSNotificationMessage } from '@/types/api';
import type { AppDispatch } from '@/store';

/** 重连最大次数 */
const MAX_RECONNECT = 5;
/** 重连基础延迟（毫秒） */
const RECONNECT_BASE_DELAY = 2000;
/** 心跳间隔（毫秒） */
const HEARTBEAT_INTERVAL = 30000;

/** 构建 WebSocket URL */
function buildWSUrl(token: string): string {
  const wsBase = import.meta.env.VITE_WS_URL || '/ws';
  const wsPath = `${wsBase}/notifications`;
  let url: string;

  if (wsPath.startsWith('ws://') || wsPath.startsWith('wss://')) {
    url = wsPath;
  } else {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    url = `${protocol}//${host}${wsPath}`;
  }

  url += `?token=${encodeURIComponent(token)}`;
  return url;
}

/**
 * 通知 WebSocket 连接 Hook
 *
 * - 自动在用户登录后建立 WebSocket 连接
 * - 收到新通知时 dispatch 到 Redux store
 * - 支持自动重连（指数退避）
 * - 支持心跳检测
 * - 登出/Token 失效时自动断开
 */
export function useNotificationWebSocket(): void {
  const dispatch = useDispatch<AppDispatch>();
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectCount = useRef(0);
  const heartbeatTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const shouldConnect = useRef(false);

  /** 清理心跳和重连定时器 */
  const clearTimers = useCallback(() => {
    if (heartbeatTimer.current) {
      clearInterval(heartbeatTimer.current);
      heartbeatTimer.current = null;
    }
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
  }, []);

  /** 主动断开连接 */
  const disconnect = useCallback(() => {
    shouldConnect.current = false;
    clearTimers();
    if (wsRef.current) {
      wsRef.current.onopen = null;
      wsRef.current.onmessage = null;
      wsRef.current.onclose = null;
      wsRef.current.onerror = null;
      if (wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.close(1000, 'user disconnect');
      }
      wsRef.current = null;
    }
    dispatch(setWSConnected(false));
  }, [clearTimers, dispatch]);

  /** 处理收到的 WebSocket 消息 */
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const msg: WSNotificationMessage = JSON.parse(event.data);

        switch (msg.type) {
          case 'notification': {
            // 构建完整的 Notification 对象
            const notification = {
              id: '', // WebSocket 通知可能没有 ID，后续 fetch 会补全
              title: msg.data.title || '',
              content: msg.data.content || '',
              category: (msg.data.category as any) || 'other',
              priority: (msg.data.priority as any) || 'normal',
              is_read: false,
              action_url: msg.data.action_url || '',
              sender: msg.data.sender || { id: '', name: '系统' },
              created_at: new Date().toISOString(),
            };
            dispatch(addNotification(notification));
            break;
          }
          case 'auth_result': {
            // 认证结果已在 onopen 后处理
            break;
          }
          case 'pong': {
            // 心跳回应，无需处理
            break;
          }
          default:
            break;
        }
      } catch {
        // 忽略无法解析的消息
      }
    },
    [dispatch],
  );

  /** 建立连接 */
  const connect = useCallback(() => {
    const token = getToken();
    if (!token) return;

    const url = buildWSUrl(token);
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      shouldConnect.current = true;
      reconnectCount.current = 0;
      dispatch(setWSConnected(true));

      // 拉取最新未读数和最近通知
      dispatch(fetchUnreadCount());
      dispatch(fetchRecentNotifications());

      // 启动心跳
      heartbeatTimer.current = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'ping' }));
        }
      }, HEARTBEAT_INTERVAL);
    };

    ws.onmessage = handleMessage;

    ws.onclose = (event: CloseEvent) => {
      dispatch(setWSConnected(false));
      clearTimers();

      // 非正常关闭且仍有 Token → 自动重连
      if (shouldConnect.current && event.code !== 1000 && reconnectCount.current < MAX_RECONNECT) {
        const delay = RECONNECT_BASE_DELAY * Math.pow(2, reconnectCount.current);
        reconnectCount.current += 1;
        reconnectTimer.current = setTimeout(() => {
          connect();
        }, delay);
      }
    };

    ws.onerror = () => {
      dispatch(setWSConnected(false));
    };
  }, [dispatch, handleMessage, clearTimers]);

  useEffect(() => {
    if (isAuthenticated) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated]);

  // 页面卸载时清理
  useEffect(() => {
    return () => {
      disconnect();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
}
