import { apiClient } from './client';
import { AuthResponse, LoginRequest, RegisterRequest } from './types';

// 認証API
export const authApi = {
  // ログイン
  login: async (credentials: LoginRequest): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/auth/login', credentials);
  },

  // 新規登録
  register: async (userData: RegisterRequest): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/auth/register', userData);
  },

  // ログアウト
  logout: async (): Promise<void> => {
    return apiClient.post<void>('/auth/logout');
  },

  // トークンの検証
  verifyToken: async (): Promise<{ valid: boolean; user?: any }> => {
    return apiClient.get<{ valid: boolean; user?: any }>('/auth/verify');
  },

  // パスワードリセット要求
  requestPasswordReset: async (email: string): Promise<{ message: string }> => {
    return apiClient.post<{ message: string }>('/auth/password-reset', { email });
  },

  // パスワードリセット実行
  resetPassword: async (token: string, newPassword: string): Promise<{ message: string }> => {
    return apiClient.post<{ message: string }>('/auth/password-reset/confirm', {
      token,
      newPassword,
    });
  },

  // プロフィール取得
  getProfile: async (): Promise<any> => {
    return apiClient.get<any>('/auth/profile');
  },

  // プロフィール更新
  updateProfile: async (profileData: any): Promise<any> => {
    return apiClient.put<any>('/auth/profile', profileData);
  },
};
