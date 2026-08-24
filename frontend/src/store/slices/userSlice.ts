import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { getCurrentUser } from '@/api/auth';
import type { UserInfo } from '@/types/api';
import type { RootState } from '@/store';

interface UserState {
  currentUser: UserInfo | null;
  permissions: string[];
  loading: boolean;
}

const initialState: UserState = {
  currentUser: null,
  permissions: [],
  loading: false,
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
    },
    setPermissions(state, action: PayloadAction<string[]>) {
      state.permissions = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchCurrentUser.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchCurrentUser.fulfilled, (state, action: PayloadAction<UserInfo>) => {
        state.loading = false;
        state.currentUser = action.payload;
        state.permissions = action.payload.permissions || [];
      })
      .addCase(fetchCurrentUser.rejected, (state) => {
        state.loading = false;
      });
  },
});

export const { clearUser, setPermissions } = userSlice.actions;

export const selectCurrentUser = (state: RootState) => state.user.currentUser;
export const selectPermissions = (state: RootState) => state.user.permissions;
export const selectUserLoading = (state: RootState) => state.user.loading;

export default userSlice.reducer;
