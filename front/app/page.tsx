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
        <img
          src="/images/モンジェネロゴタイプ.svg"
          alt="Mongene"
          className="h-10 w-auto object-contain mx-auto mb-4"
        />
        <div className="text-mongene-muted">リダイレクトしています...</div>
      </div>
    </div>
  );
}
