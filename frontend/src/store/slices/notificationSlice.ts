import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import {
  getUnreadCount,
  getNotificationList,
  markAsRead as markAsReadApi,
  markAllAsRead as markAllAsReadApi,
} from '@/api/notification';
import type { Notification, ListNotificationParams } from '@/types/api';
import type { RootState } from '@/store';

interface NotificationState {
  unreadCount: number;
  recentNotifications: Notification[];
  wsConnected: boolean;
  loading: boolean;
  error: string | null;
}

const initialState: NotificationState = {
  unreadCount: 0,
  recentNotifications: [],
  wsConnected: false,
  loading: false,
  error: null,
};

/** 拉取未读通知数量 */
export const fetchUnreadCount = createAsyncThunk(
  'notification/fetchUnreadCount',
  async () => {
    const res = await getUnreadCount();
    return res.count;
  },
);

/** 拉取最近通知（前 10 条，用于铃铛下拉） */
export const fetchRecentNotifications = createAsyncThunk(
  'notification/fetchRecentNotifications',
  async (_, { rejectWithValue }) => {
    try {
      const params: ListNotificationParams = { page: 1, page_size: 10 };
      const res = await getNotificationList(params);
      return res.list;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取通知失败');
    }
  },
);

/** 标记单条已读 */
export const markNotificationAsRead = createAsyncThunk(
  'notification/markAsRead',
  async (id: string, { rejectWithValue }) => {
    try {
      await markAsReadApi(id);
      return id;
    } catch (error: any) {
      return rejectWithValue(error.message || '标记已读失败');
    }
  },
);

/** 全部标记已读 */
export const markAllNotificationsAsRead = createAsyncThunk(
  'notification/markAllAsRead',
  async (category?: string) => {
    await markAllAsReadApi(category);
  },
);

const notificationSlice = createSlice({
  name: 'notification',
  initialState,
  reducers: {
    /** WebSocket 收到新通知时添加到列表头部 */
    addNotification(state, action: PayloadAction<Notification>) {
      state.recentNotifications.unshift(action.payload);
      // 保持最多 20 条
      if (state.recentNotifications.length > 20) {
        state.recentNotifications = state.recentNotifications.slice(0, 20);
      }
      state.unreadCount += 1;
    },
    /** 设置 WebSocket 连接状态 */
    setWSConnected(state, action: PayloadAction<boolean>) {
      state.wsConnected = action.payload;
    },
    /** 清空通知（登出时调用） */
    clearNotifications(state) {
      state.unreadCount = 0;
      state.recentNotifications = [];
      state.wsConnected = false;
      state.error = null;
    },
    /** 通知列表中某条标记为已读（UI 即时更新） */
    markReadInList(state, action: PayloadAction<string>) {
      const n = state.recentNotifications.find((item) => item.id === action.payload);
      if (n && !n.is_read) {
        n.is_read = true;
        state.unreadCount = Math.max(0, state.unreadCount - 1);
      }
    },
    /** 全部已读（UI 即时更新） */
    markAllReadInList(state) {
      state.recentNotifications.forEach((n) => {
        n.is_read = true;
      });
      state.unreadCount = 0;
    },
  },
  extraReducers: (builder) => {
    builder
      // 未读计数
      .addCase(fetchUnreadCount.fulfilled, (state, action: PayloadAction<number>) => {
        state.unreadCount = action.payload;
      })
      // 最近通知
      .addCase(fetchRecentNotifications.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchRecentNotifications.fulfilled, (state, action: PayloadAction<Notification[]>) => {
        state.loading = false;
        state.recentNotifications = action.payload;
      })
      .addCase(fetchRecentNotifications.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // 标记已读
      .addCase(markNotificationAsRead.fulfilled, (state, action: PayloadAction<string>) => {
        const n = state.recentNotifications.find((item) => item.id === action.payload);
        if (n && !n.is_read) {
          n.is_read = true;
          state.unreadCount = Math.max(0, state.unreadCount - 1);
        }
      })
      // 全部已读
      .addCase(markAllNotificationsAsRead.fulfilled, (state) => {
        state.recentNotifications.forEach((n) => {
          n.is_read = true;
        });
        state.unreadCount = 0;
      });
  },
});

export const {
  addNotification,
  setWSConnected,
  clearNotifications,
  markReadInList,
  markAllReadInList,
} = notificationSlice.actions;

export const selectUnreadCount = (state: RootState) => state.notification.unreadCount;
export const selectRecentNotifications = (state: RootState) => state.notification.recentNotifications;
export const selectWSConnected = (state: RootState) => state.notification.wsConnected;
export const selectNotificationLoading = (state: RootState) => state.notification.loading;
export const selectNotificationError = (state: RootState) => state.notification.error;

export default notificationSlice.reducer;
