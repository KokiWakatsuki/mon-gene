import { apiClient } from './client';
import { 
  Problem, 
  ProblemGenerateRequest, 
  ProblemGenerateResponse, 
  ProblemsListResponse,
  PaginationParams,
  ProblemFilters 
} from './types';

// 問題生成API
export const problemsApi = {
  // 問題一覧取得
  getProblems: async (
    params?: PaginationParams & ProblemFilters
  ): Promise<ProblemsListResponse> => {
    const queryParams = new URLSearchParams();
    
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          queryParams.append(key, value.toString());
        }
      });
    }
    
    const queryString = queryParams.toString();
    const endpoint = queryString ? `/problems?${queryString}` : '/problems';
    
    return apiClient.get<ProblemsListResponse>(endpoint);
  },

  // 問題詳細取得
  getProblem: async (id: string): Promise<Problem> => {
    return apiClient.get<Problem>(`/problems/${id}`);
  },

  // 問題生成
  generateProblem: async (request: ProblemGenerateRequest): Promise<ProblemGenerateResponse> => {
    return apiClient.post<ProblemGenerateResponse>('/problems/generate', request);
  },

  // 問題削除
  deleteProblem: async (id: string): Promise<void> => {
    return apiClient.delete<void>(`/problems/${id}`);
  },

  // 問題更新
  updateProblem: async (id: string, updates: Partial<Problem>): Promise<Problem> => {
    return apiClient.put<Problem>(`/problems/${id}`, updates);
  },

  // PDF生成
  generatePDF: async (id: string): Promise<{ pdfUrl: string }> => {
    return apiClient.post<{ pdfUrl: string }>(`/problems/${id}/pdf`);
  },

  // PDF取得
  getPDF: async (id: string): Promise<Blob> => {
    const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/problems/${id}/pdf`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
      },
    });
    
    if (!response.ok) {
      throw new Error('PDF取得に失敗しました');
    }
    
    return response.blob();
  },

  // 問題の複製
  duplicateProblem: async (id: string): Promise<Problem> => {
    return apiClient.post<Problem>(`/problems/${id}/duplicate`);
  },

  // 問題のお気に入り登録/解除
  toggleFavorite: async (id: string): Promise<{ isFavorite: boolean }> => {
    return apiClient.post<{ isFavorite: boolean }>(`/problems/${id}/favorite`);
  },

  // お気に入り問題一覧取得
  getFavoriteProblems: async (
    params?: PaginationParams
  ): Promise<ProblemsListResponse> => {
    const queryParams = new URLSearchParams();
    
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          queryParams.append(key, value.toString());
        }
      });
    }
    
    const queryString = queryParams.toString();
    const endpoint = queryString ? `/problems/favorites?${queryString}` : '/problems/favorites';
    
    return apiClient.get<ProblemsListResponse>(endpoint);
  },
};
