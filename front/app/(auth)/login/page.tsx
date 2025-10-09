'use client';

import React, { useState } from 'react';
import BackgroundShapes from '../../components/layout/BackgroundShapes';

export default function LoginPage() {
  const [formData, setFormData] = useState({
    schoolCode: '',
    password: '',
    remember: false,
  });
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showForgotPassword, setShowForgotPassword] = useState(false);
  const [forgotPasswordEmail, setForgotPasswordEmail] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    // バリデーション
    if (!formData.schoolCode.trim()) {
      setError('塾コードを入力してください。');
      setIsLoading(false);
      return;
    }
    if (!formData.password.trim()) {
      setError('パスワードを入力してください。');
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          schoolCode: formData.schoolCode,
          password: formData.password,
          remember: formData.remember,
        }),
      });
      
      const data = await response.json();
      
      if (!data.success) {
        throw new Error(data.error || 'ログインに失敗しました');
      }
      
      // トークンをローカルストレージに保存
      localStorage.setItem('authToken', data.token);
      
      // トップページへリダイレクト
        window.location.href = '/problems';
    } catch (error) {
      setError(error instanceof Error ? error.message : '塾コードまたはパスワードが正しくありません。');
    } finally {
      setIsLoading(false);
    }
  };

  const handleForgotPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    if (!forgotPasswordEmail.trim()) {
      setError('塾コードを入力してください。');
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/forgot-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          schoolCode: forgotPasswordEmail,
        }),
      });
      
      const data = await response.json();
      
      if (!data.success) {
        throw new Error(data.error || 'パスワードリセットに失敗しました');
      }
      
      alert(data.message);
      setShowForgotPassword(false);
      setForgotPasswordEmail('');
    } catch (error) {
      setError(error instanceof Error ? error.message : 'パスワードリセットに失敗しました。');
    } finally {
      setIsLoading(false);
    }
  };

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  return (
    <div className="relative min-h-screen overflow-hidden bg-mongene-bg">
      <BackgroundShapes />
      
      <div className="relative z-10 max-w-7xl mx-auto p-6">
        {/* Header */}
        <header className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 bg-mongene-blue rounded-lg" aria-hidden="true"></div>
            <div className="font-extrabold text-mongene-blue">Mongene</div>
          </div>
        </header>

        {/* Main Content */}
        <main className="grid place-items-center p-6" aria-labelledby="loginTitle">
          <section className="w-full max-w-lg bg-white border border-mongene-border rounded-2xl shadow-xl p-5" role="region" aria-label="ログインフォーム">
            <h1 id="loginTitle" className="text-2xl font-semibold mb-1.5">ログイン</h1>
            <p className="text-mongene-muted text-sm mb-4">塾コードとパスワードを入力してください。</p>

            {error && (
              <div className="border border-red-200 bg-red-50 text-red-800 rounded-lg p-2.5 text-sm mb-3" role="alert">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} noValidate className="grid gap-3">
              <div>
                <label htmlFor="schoolCode" className="block font-bold text-sm mb-1">塾コード</label>
                <input
                  className="w-full border border-mongene-border rounded-xl px-3.5 py-3 text-base focus:outline-none focus:ring-3 focus:ring-mongene-blue/25 focus:ring-offset-2"
                  id="schoolCode"
                  name="schoolCode"
                  type="text"
                  placeholder="例: ABC123"
                  value={formData.schoolCode}
                  onChange={handleInputChange}
                  required
                />
              </div>

              <div>
                <div className="flex items-center justify-between gap-2 mb-1.5">
                  <label htmlFor="password" className="font-bold text-sm">パスワード</label>
                  <button 
                    type="button"
                    className="text-sm text-blue-600 no-underline hover:underline bg-transparent border-none cursor-pointer"
                    onClick={() => setShowForgotPassword(true)}
                  >
                    パスワードをお忘れですか？
                  </button>
                </div>
                <div className="relative">
                  <input
                    className="w-full border border-mongene-border rounded-xl px-3.5 py-3 text-base focus:outline-none focus:ring-3 focus:ring-mongene-blue/25 focus:ring-offset-2"
                    id="password"
                    name="password"
                    type={showPassword ? 'text' : 'password'}
                    autoComplete="current-password"
                    placeholder="••••••••"
                    minLength={8}
                    value={formData.password}
                    onChange={handleInputChange}
                    required
                  />
                  <button
                    type="button"
                    className="absolute right-2 top-1/2 -translate-y-1/2 border border-mongene-border bg-white rounded-lg px-2 py-1.5 font-bold text-xs cursor-pointer"
                    aria-controls="password"
                    aria-label="パスワードの表示切替"
                    onClick={togglePasswordVisibility}
                  >
                    {showPassword ? '非表示' : '表示'}
                  </button>
                </div>
              </div>

              <div className="flex items-center justify-between gap-2">
                <label className="flex items-center gap-2 font-semibold">
                  <input
                    type="checkbox"
                    id="remember"
                    name="remember"
                    checked={formData.remember}
                    onChange={handleInputChange}
                  />
                  次回から自動的にログイン
                </label>
                <button
                  className="appearance-none border-0 rounded-xl px-4 py-3 font-extrabold cursor-pointer bg-mongene-green text-white shadow-lg hover:brightness-98 hover:-translate-y-0.5 transition-all focus:outline-none focus:ring-3 focus:ring-mongene-green/25 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
                  type="submit"
                  aria-label="サインイン"
                  disabled={isLoading}
                >
                  {isLoading ? 'ログイン中...' : 'サインイン'}
                </button>
              </div>
            </form>
          </section>
        </main>
      </div>

      {/* パスワードリセットモーダル */}
      {showForgotPassword && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-2xl p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-semibold mb-4">パスワードリセット</h2>
            <p className="text-mongene-muted text-sm mb-4">
              塾コードを入力してください。登録されたメールアドレスにパスワードを送信します。
            </p>
            
            <form onSubmit={handleForgotPassword} className="space-y-4">
              <div>
                <label htmlFor="forgotPasswordEmail" className="block font-bold text-sm mb-1">
                  塾コード
                </label>
                <input
                  className="w-full border border-mongene-border rounded-xl px-3.5 py-3 text-base focus:outline-none focus:ring-3 focus:ring-mongene-blue/25 focus:ring-offset-2"
                  id="forgotPasswordEmail"
                  type="text"
                  placeholder="例: ABC123"
                  value={forgotPasswordEmail}
                  onChange={(e) => setForgotPasswordEmail(e.target.value)}
                  required
                />
              </div>
              
              <div className="flex gap-3 justify-end">
                <button
                  type="button"
                  className="px-4 py-2 border border-mongene-border rounded-xl font-semibold hover:bg-gray-50 transition-colors"
                  onClick={() => {
                    setShowForgotPassword(false);
                    setForgotPasswordEmail('');
                    setError('');
                  }}
                  disabled={isLoading}
                >
                  キャンセル
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-mongene-green text-white rounded-xl font-semibold hover:brightness-98 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={isLoading}
                >
                  {isLoading ? '送信中...' : '送信'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
