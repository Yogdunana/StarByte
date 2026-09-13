import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import {
  clearNotifications,
  setWSConnected,
  fetchUnreadCount,
  fetchRecentNotifications,
} from '@/store/slices/notificationSlice';
import { selectIsAuthenticated } from '@/store/slices/authSlice';
import { getToken } from '@/utils/storage';
import type { AppDispatch } from '@/store';

function buildWSUrl(token: string): string {
  const base = import.meta.env.VITE_WS_URL || '/ws';
  const path = `${base}/notifications`;
  const origin = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}`;
  return `${/^wss?:\/\//.test(path) ? path : origin + path}?token=${encodeURIComponent(token)}`;
}

/** WebSocket signals refresh persisted messages; polling also delivers durable
 * scheduled notifications and recovers messages missed during disconnection. */
export function useNotificationWebSocket(enabled = true): void {
  const dispatch = useDispatch<AppDispatch>();
  const authenticated = useSelector(selectIsAuthenticated);
  useEffect(() => {
    if (!authenticated || !enabled) {
      dispatch(clearNotifications());
      return;
    }
    let disposed = false;
    let socket: WebSocket | null = null;
    let reconnect: ReturnType<typeof setTimeout> | undefined;
    let refreshTimer: ReturnType<typeof setTimeout> | undefined;
    let attempt = 0;
    const refresh = () => {
      if (disposed) return;
      dispatch(fetchUnreadCount());
      dispatch(fetchRecentNotifications());
    };
    const connect = () => {
      const token = getToken();
      if (disposed || !token) return;
      const current = new WebSocket(buildWSUrl(token));
      socket = current;
      current.onopen = () => {
        if (disposed) {
          current.close();
          return;
        }
        attempt = 0;
        dispatch(setWSConnected(true));
        refresh();
      };
      current.onmessage = (event) => {
        if (disposed) return;
        try {
          const message: unknown = JSON.parse(event.data);
          if (
            typeof message === 'object' &&
            message !== null &&
            'type' in message &&
            message.type === 'notification'
          ) {
            clearTimeout(refreshTimer);
            refreshTimer = setTimeout(refresh, 250);
          }
        } catch {
          /* Ignore malformed transport messages. */
        }
      };
      current.onclose = () => {
        if (disposed) return;
        dispatch(setWSConnected(false));
        reconnect = setTimeout(connect, Math.min(30000, 2000 * 2 ** Math.min(attempt++, 4)));
      };
      current.onerror = () => {
        if (!disposed) dispatch(setWSConnected(false));
      };
    };
    refresh();
    connect();
    const poll = setInterval(() => {
      if (document.visibilityState === 'visible') refresh();
    }, 30000);
    const heartbeat = setInterval(() => {
      if (socket?.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: 'ping' }));
    }, 30000);
    const onVisible = () => {
      if (document.visibilityState === 'visible') refresh();
    };
    document.addEventListener('visibilitychange', onVisible);
    return () => {
      disposed = true;
      clearInterval(poll);
      clearInterval(heartbeat);
      clearTimeout(reconnect);
      clearTimeout(refreshTimer);
      document.removeEventListener('visibilitychange', onVisible);
      if (socket) {
        socket.onopen = socket.onclose = socket.onmessage = socket.onerror = null;
        socket.close(1000, 'page disconnected');
      }
      dispatch(clearNotifications());
    };
  }, [authenticated, dispatch, enabled]);
}
