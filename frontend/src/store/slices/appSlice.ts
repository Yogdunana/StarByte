import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { RootState } from '@/store';

interface AppState {
  collapsed: boolean;
  theme: 'light' | 'dark';
  notificationCount: number;
}

const initialState: AppState = {
  collapsed: false,
  theme: 'light',
  notificationCount: 0,
};

const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    toggleCollapsed(state) {
      state.collapsed = !state.collapsed;
    },
    setCollapsed(state, action: PayloadAction<boolean>) {
      state.collapsed = action.payload;
    },
    setTheme(state, action: PayloadAction<'light' | 'dark'>) {
      state.theme = action.payload;
    },
    setNotificationCount(state, action: PayloadAction<number>) {
      state.notificationCount = action.payload;
    },
  },
});

export const { toggleCollapsed, setCollapsed, setTheme, setNotificationCount } = appSlice.actions;

export const selectCollapsed = (state: RootState) => state.app.collapsed;
export const selectTheme = (state: RootState) => state.app.theme;
export const selectNotificationCount = (state: RootState) => state.app.notificationCount;

export default appSlice.reducer;
