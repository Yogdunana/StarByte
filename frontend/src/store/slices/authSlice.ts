import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { login as loginApi, refreshToken as refreshTokenApi } from '@/api/auth';
import type { LoginRequest, LoginResponse } from '@/types/api';
import { setToken as saveToken, getRefreshToken, removeToken, setRefreshToken } from '@/utils/storage';
import type { RootState } from '@/store';

interface AuthState {
  token: string;
  refreshToken: string;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  token: '',
  refreshToken: '',
  isAuthenticated: false,
  loading: false,
  error: null,
};

// 登录
export const login = createAsyncThunk(
  'auth/login',
  async (params: LoginRequest, { rejectWithValue }) => {
    try {
      const response = await loginApi(params);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '登录失败');
    }
  }
);

// 刷新 Token
export const refreshToken = createAsyncThunk(
  'auth/refreshToken',
  async (_, { rejectWithValue }) => {
    try {
      const token = getRefreshToken();
      if (!token) {
        throw new Error('无 refresh token');
      }
      const response = await refreshTokenApi(token);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '刷新 Token 失败');
    }
  }
);

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    logout(state) {
      state.token = '';
      state.refreshToken = '';
      state.isAuthenticated = false;
      removeToken();
    },
    setToken(state, action: PayloadAction<{ accessToken: string; refreshToken: string }>) {
      state.token = action.payload.accessToken;
      state.refreshToken = action.payload.refreshToken;
      state.isAuthenticated = true;
      saveToken(action.payload.accessToken);
      setRefreshToken(action.payload.refreshToken);
    },
    clearError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(login.fulfilled, (state, action: PayloadAction<LoginResponse>) => {
        state.loading = false;
        state.token = action.payload.access_token;
        state.refreshToken = action.payload.refresh_token;
        state.isAuthenticated = true;
        saveToken(action.payload.access_token);
        setRefreshToken(action.payload.refresh_token);
      })
      .addCase(login.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(refreshToken.fulfilled, (state, action: PayloadAction<LoginResponse>) => {
        state.token = action.payload.access_token;
        state.refreshToken = action.payload.refresh_token;
        saveToken(action.payload.access_token);
        setRefreshToken(action.payload.refresh_token);
      });
  },
});

export const { logout, setToken, clearError } = authSlice.actions;

export const selectIsAuthenticated = (state: RootState) => state.auth.isAuthenticated;
export const selectAuthLoading = (state: RootState) => state.auth.loading;
export const selectAuthError = (state: RootState) => state.auth.error;
export const selectToken = (state: RootState) => state.auth.token;

export default authSlice.reducer;
