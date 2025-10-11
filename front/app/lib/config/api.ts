// API設定ファイル
// 環境変数を使用してAPIキーを安全に管理

export const API_CONFIG = {
  // テスト版と実際のAPI版の切り替え
  // true: 実際のClaude APIを使用, false: テスト版（ダミーデータ）
  // バックエンドサーバー経由でClaude APIを呼び出します
  USE_REAL_API: process.env.NEXT_PUBLIC_USE_REAL_API === 'true',
  
  // Claude APIキー（環境変数から取得 - 直接呼び出し用、現在はCORSエラーのため使用不可）
  CLAUDE_API_KEY: process.env.NEXT_PUBLIC_CLAUDE_API_KEY || '',
  
  // Claude API設定（直接呼び出し用 - CORSエラーのため現在は使用不可）
  CLAUDE_API_URL: 'https://api.anthropic.com/v1/messages',
  CLAUDE_MODEL: 'claude-3-5-sonnet-20241022',
  CLAUDE_MAX_TOKENS: 1000,
  CLAUDE_VERSION: '2023-06-01',
  
  // バックエンドサーバー経由でのAPI呼び出し設定
  API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  BACKEND_API_URL: process.env.NEXT_PUBLIC_BACKEND_API_URL || 'http://localhost:8080/api/generate-problem',
  USER_INFO_API_URL: process.env.NEXT_PUBLIC_USER_INFO_API_URL || 'http://localhost:8080/api/user-info',
  
  // 認証関連のAPI設定
  LOGIN_API_URL: (process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080') + '/api/login',
  FORGOT_PASSWORD_API_URL: (process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080') + '/api/forgot-password',
  
  // 他のAPI設定もここに追加できます
};
