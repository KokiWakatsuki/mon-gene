// 出典リストのAPI関数

import { API_CONFIG } from '../config/api';

export interface SourceListItem {
  id: number;
  year: string;
  exam_session: string;
  created_at?: string;
}

export interface SourceListResponse {
  success: boolean;
  items?: SourceListItem[];
  error?: string;
}

// 出典リストに項目を追加
export const addSourceListItem = async (
  token: string,
  year: string,
  examSession: string
): Promise<SourceListResponse> => {
  const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/source-list`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      year,
      exam_session: examSession,
    }),
  });

  if (!response.ok) {
    throw new Error('出典リストへの追加に失敗しました');
  }

  return response.json();
};

// ユーザーの出典リストを取得
export const getUserSourceList = async (token: string): Promise<SourceListResponse> => {
  const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/source-list`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('出典リストの取得に失敗しました');
  }

  return response.json();
};

// 出典リストから項目を削除
export const deleteSourceListItem = async (token: string, id: number): Promise<void> => {
  const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/source-list/${id}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('出典リストからの削除に失敗しました');
  }
};