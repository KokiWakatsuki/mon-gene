'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  id: number;
  school_code: string;
  email: string;
  role: string;
  profile_image?: string;
}

export default function AccountPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  
  // パスワード変更
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  
  // メールアドレス変更
  const [newEmail, setNewEmail] = useState('');
  const [emailError, setEmailError] = useState('');
  
  // プロフィール画像
  const [profileImage, setProfileImage] = useState<string | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    fetchUserData();
  }, []);

  const fetchUserData = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        router.push('/login');
        return;
      }

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user/profile`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
        setNewEmail(userData.email);
        if (userData.profile_image) {
          setProfileImage(userData.profile_image);
          setImagePreview(userData.profile_image);
        }
      } else {
        router.push('/login');
      }
    } catch (error) {
      console.error('ユーザーデータの取得に失敗しました:', error);
      router.push('/login');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');

    // バリデーション
    if (!currentPassword || !newPassword || !confirmPassword) {
      setPasswordError('すべてのフィールドを入力してください');
      return;
    }

    if (newPassword.length < 8) {
      setPasswordError('新しいパスワードは8文字以上である必要があります');
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('新しいパスワードが一致しません');
      return;
    }

    setSaving(true);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user/password`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword
        })
      });

      if (response.ok) {
        alert('パスワードを変更しました');
        setCurrentPassword('');
        setNewPassword('');
        setConfirmPassword('');
      } else {
        const error = await response.json();
        setPasswordError(error.error || 'パスワードの変更に失敗しました');
      }
    } catch (error) {
      console.error('パスワード変更に失敗しました:', error);
      setPasswordError('パスワードの変更に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  const handleEmailChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setEmailError('');

    // バリデーション
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(newEmail)) {
      setEmailError('有効なメールアドレスを入力してください');
      return;
    }

    if (newEmail === user?.email) {
      setEmailError('現在のメールアドレスと同じです');
      return;
    }

    setSaving(true);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user/email`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          email: newEmail
        })
      });

      if (response.ok) {
        alert('メールアドレスを変更しました');
        fetchUserData();
      } else {
        const error = await response.json();
        setEmailError(error.error || 'メールアドレスの変更に失敗しました');
      }
    } catch (error) {
      console.error('メールアドレス変更に失敗しました:', error);
      setEmailError('メールアドレスの変更に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  const handleImageSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // 画像形式チェック（JPG、PNG、GIFのみ）
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];
    if (!allowedTypes.includes(file.type)) {
      alert('JPG、PNG、GIF形式の画像のみアップロード可能です');
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
      return;
    }

    // ファイルサイズ制限 (5MB)
    if (file.size > 5 * 1024 * 1024) {
      alert('ファイルサイズは5MB以下にしてください');
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      const result = e.target?.result as string;
      
      // Base64エンコード後のサイズをチェック（約5MB）
      // Base64は元のサイズの約1.37倍になるため、3.65MB以下に制限
      const base64Size = result.length * 0.75; // Base64から元のバイト数を推定
      if (base64Size > 5 * 1024 * 1024) {
        alert('エンコード後のファイルサイズが5MBを超えています。より小さい画像を選択してください');
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
        return;
      }
      
      setImagePreview(result);
    };
    reader.readAsDataURL(file);
  };

  const handleImageUpload = async () => {
    if (!imagePreview) return;

    setSaving(true);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user/profile-image`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          image: imagePreview
        })
      });

      if (response.ok) {
        alert('プロフィール画像を更新しました');
        setProfileImage(imagePreview);
        fetchUserData();
      } else {
        const errorData = await response.json();
        console.error('エラーレスポンス:', errorData);
        alert(`プロフィール画像の更新に失敗しました: ${errorData.error || '不明なエラー'}`);
      }
    } catch (error) {
      console.error('プロフィール画像の更新に失敗しました:', error);
      alert('プロフィール画像の更新に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  const handleImageDelete = async () => {
    if (!confirm('プロフィール画像を削除しますか？')) return;

    setSaving(true);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user/profile-image`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        alert('プロフィール画像を削除しました');
        setProfileImage(null);
        setImagePreview(null);
        fetchUserData();
      } else {
        alert('プロフィール画像の削除に失敗しました');
      }
    } catch (error) {
      console.error('プロフィール画像の削除に失敗しました:', error);
      alert('プロフィール画像の削除に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="text-lg">読み込み中...</div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-lg shadow-md p-6">
            <div className="flex justify-between items-center mb-6">
              <h1 className="text-2xl font-bold text-gray-800">アカウント設定</h1>
              <button
                onClick={() => router.push('/problems')}
                className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600 transition-colors"
              >
                戻る
              </button>
            </div>

            <div className="space-y-8">
              {/* プロフィール画像セクション */}
              <div className="border-b pb-6">
                <h2 className="text-lg font-semibold mb-4 flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  プロフィール画像
                </h2>
                <div className="flex items-center space-x-6">
                  <div className="relative">
                    <div className="w-32 h-32 rounded-full overflow-hidden bg-gray-200 flex items-center justify-center">
                      {imagePreview ? (
                        <img src={imagePreview} alt="Profile" className="w-full h-full object-cover" />
                      ) : (
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                        </svg>
                      )}
                    </div>
                  </div>
                  <div className="flex-1">
                    <input
                      type="file"
                      ref={fileInputRef}
                      onChange={handleImageSelect}
                      accept="image/*"
                      className="hidden"
                    />
                    <div className="space-y-2">
                      <button
                        onClick={() => fileInputRef.current?.click()}
                        className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 transition-colors"
                        disabled={saving}
                      >
                        画像を選択
                      </button>
                      {imagePreview && imagePreview !== profileImage && (
                        <button
                          onClick={handleImageUpload}
                          className="ml-2 bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 transition-colors"
                          disabled={saving}
                        >
                          {saving ? '保存中...' : '保存'}
                        </button>
                      )}
                      {profileImage && (
                        <button
                          onClick={handleImageDelete}
                          className="ml-2 bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 transition-colors"
                          disabled={saving}
                        >
                          削除
                        </button>
                      )}
                    </div>
                    <p className="text-sm text-gray-500 mt-2">
                      JPG、PNG、GIF形式（最大5MB）
                    </p>
                  </div>
                </div>
              </div>

              {/* 基本情報セクション */}
              <div className="border-b pb-6">
                <h2 className="text-lg font-semibold mb-4 flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  基本情報
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <p className="text-sm text-gray-600 mb-1">学校コード</p>
                    <p className="font-medium">{user.school_code}</p>
                  </div>
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <p className="text-sm text-gray-600 mb-1">権限</p>
                    <p className="font-medium">{user.role}</p>
                  </div>
                </div>
              </div>

              {/* メールアドレス変更セクション */}
              <div className="border-b pb-6">
                <h2 className="text-lg font-semibold mb-4 flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  メールアドレス
                </h2>
                <form onSubmit={handleEmailChange} className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      現在のメールアドレス
                    </label>
                    <input
                      type="email"
                      value={user.email}
                      disabled
                      className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100 text-gray-600"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      新しいメールアドレス
                    </label>
                    <input
                      type="email"
                      value={newEmail}
                      onChange={(e) => setNewEmail(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="new@example.com"
                    />
                  </div>
                  {emailError && (
                    <div className="text-red-600 text-sm">{emailError}</div>
                  )}
                  <button
                    type="submit"
                    disabled={saving || newEmail === user.email}
                    className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
                  >
                    {saving ? '変更中...' : 'メールアドレスを変更'}
                  </button>
                </form>
              </div>

              {/* パスワード変更セクション */}
              <div>
                <h2 className="text-lg font-semibold mb-4 flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                  パスワード変更
                </h2>
                <form onSubmit={handlePasswordChange} className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      現在のパスワード
                    </label>
                    <input
                      type="password"
                      value={currentPassword}
                      onChange={(e) => setCurrentPassword(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="現在のパスワードを入力"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      新しいパスワード
                    </label>
                    <input
                      type="password"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="8文字以上"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      新しいパスワード（確認）
                    </label>
                    <input
                      type="password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="もう一度入力"
                    />
                  </div>
                  {passwordError && (
                    <div className="text-red-600 text-sm">{passwordError}</div>
                  )}
                  <button
                    type="submit"
                    disabled={saving}
                    className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
                  >
                    {saving ? '変更中...' : 'パスワードを変更'}
                  </button>
                </form>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}