import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { getCurrentUser } from '@/api/auth';
import type { UserInfo } from '@/types/api';
import type { RootState } from '@/store';

interface UserState {
  currentUser: UserInfo | null;
  permissions: string[];
  loading: boolean;
  error: string | null;
}

const initialState: UserState = {
  currentUser: null,
  permissions: [],
  loading: false,
  error: null,
};

// 获取当前用户信息
export const fetchCurrentUser = createAsyncThunk(
  'user/fetchCurrentUser',
  async () => {
    const response = await getCurrentUser();
    return response;
  }
);

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    clearUser(state) {
      state.currentUser = null;
      state.permissions = [];
      state.error = null;
    },
    setPermissions(state, action: PayloadAction<string[]>) {
      state.permissions = action.payload;
    },
    clearUserError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchCurrentUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchCurrentUser.fulfilled, (state, action: PayloadAction<UserInfo>) => {
        state.loading = false;
        state.error = null;
        state.currentUser = action.payload;
        state.permissions = action.payload.permissions || [];
      })
      .addCase(fetchCurrentUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || '获取用户信息失败';
      });
  },
});

export const { clearUser, setPermissions, clearUserError } = userSlice.actions;

export const selectCurrentUser = (state: RootState) => state.user.currentUser;
export const selectPermissions = (state: RootState) => state.user.permissions;
export const selectRoles = (state: RootState) => state.user.currentUser?.roles ?? [];
export const selectUserLoading = (state: RootState) => state.user.loading;
export const selectUserError = (state: RootState) => state.user.error;

export default userSlice.reducer;
