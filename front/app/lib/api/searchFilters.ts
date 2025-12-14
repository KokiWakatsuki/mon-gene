// 検索フィルターのAPI関数

import { API_CONFIG } from '../config/api';

export interface SearchFilter {
  id: number;
  name: string;
  keyword?: string;
  subject?: string;
  units?: string[];
  year?: string;
  exam_session?: string;
  is_checked?: boolean | null;
}

export interface CreateSearchFilterRequest {
  name: string;
  keyword?: string;
  subject?: string;
  units?: string[];
  year?: string;
  exam_session?: string;
  is_checked?: boolean | null;
}

export interface SearchFilterResponse {
  success: boolean;
  filters?: SearchFilter[];
  error?: string;
}

// 検索フィルターを作成
export const createSearchFilter = async (
  token: string,
  filter: CreateSearchFilterRequest
): Promise<SearchFilterResponse> => {
  const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/search-filters`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(filter),
  });

  if (!response.ok) {
    throw new Error('検索フィルターの作成に失敗しました');
  }

  return response.json();
};

// ユーザーの検索フィルター一覧を取得
export const getUserSearchFilters = async (token: string): Promise<SearchFilterResponse> => {
  const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/search-filters`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('検索フィルターの取得に失敗しました');
  }

  return response.json();
};

// 検索フィルターを削除
export const deleteSearchFilter = async (token: string, id: number): Promise<void> => {
  const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/search-filters/${id}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('検索フィルターの削除に失敗しました');
  }
};