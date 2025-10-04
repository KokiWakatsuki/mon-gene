'use client';

import { useEffect } from 'react';

export default function RootPage() {
  useEffect(() => {
    // 認証状態をチェックしてリダイレクト
    const token = localStorage.getItem('authToken');
    if (token) {
      // 認証済みの場合は問題生成ページへ
      window.location.href = '/problems';
    } else {
      // 未認証の場合はログインページへ
      window.location.href = '/login';
    }
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center bg-mongene-bg">
      <div className="text-center">
        <div className="w-8 h-8 bg-mongene-blue rounded-lg mx-auto mb-4"></div>
        <div className="font-extrabold text-mongene-blue mb-2">Mongene</div>
        <div className="text-mongene-muted">リダイレクトしています...</div>
      </div>
    </div>
  );
}
