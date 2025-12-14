// API型定義

// 認証関連の型
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  refreshToken?: string;
}

// 問題関連の型
export interface Problem {
  id: string;
  title: string;
  description: string;
  difficulty: 'easy' | 'medium' | 'hard';
  subject: string;
  topic: string;
  requirements?: string;
  pdfUrl?: string;
  imageUrl?: string;
  createdAt: string;
  updatedAt: string;
  userId: string;
}

export interface ProblemGenerateRequest {
  subject: string;
  difficulty: 'easy' | 'medium' | 'hard';
  topic: string;
  requirements?: string;
}

export interface ProblemGenerateResponse {
  problem: Problem;
  pdfUrl?: string;
}

export interface ProblemsListResponse {
  problems: Problem[];
  total: number;
  page: number;
  limit: number;
}

// API共通レスポンス型
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

export interface ApiError {
  message: string;
  code?: string;
  details?: any;
}

// ページネーション関連
export interface PaginationParams {
  page?: number;
  limit?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface PaginationResponse {
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

// フィルター関連
export interface ProblemFilters {
  subject?: string;
  difficulty?: 'easy' | 'medium' | 'hard';
  topic?: string;
  dateFrom?: string;
  dateTo?: string;
}
