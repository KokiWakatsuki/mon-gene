'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  school_code: string;
  email: string;
  role: string;
  preferred_api: string;
  preferred_model: string;
}

export default function Header() {
  const router = useRouter();
  const [showLogoutMenu, setShowLogoutMenu] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchUserInfo();
  }, []);

  const fetchUserInfo = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user-info`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
      }
    } catch (error) {
      console.error('ユーザー情報の取得に失敗しました:', error);
    }
  };

  const handleLogout = () => {
    // ローカルストレージからトークンを削除
    localStorage.removeItem('token');
    localStorage.removeItem('authToken');
    
    // ログインページにリダイレクト
    router.push('/login');
  };

  const handleSettings = () => {
    setShowLogoutMenu(false);
    router.push('/settings');
  };

  // メニューの外側をクリックした時にメニューを閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setShowLogoutMenu(false);
      }
    };

    if (showLogoutMenu) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showLogoutMenu]);

  return (
    <header className="flex items-center justify-between mb-5">
      <div className="flex items-center gap-2.5">
        <div 
          className="w-8 h-8 bg-mongene-blue rounded-lg" 
          aria-hidden="true"
        />
        <div className="font-extrabold text-mongene-blue text-lg">
          Mongene
        </div>
      </div>
      
      <div className="relative" ref={menuRef}>
        <button
          className="w-10 h-10 bg-mongene-border rounded-full hover:bg-gray-300 transition-colors cursor-pointer border-none"
          onClick={() => setShowLogoutMenu(!showLogoutMenu)}
          aria-label="ユーザーメニュー"
        />
        
        {showLogoutMenu && (
          <div className="absolute right-0 top-12 bg-white border border-mongene-border rounded-xl shadow-lg py-2 min-w-48 z-10">
            {/* デバッグ情報 */}
            <div className="px-4 py-2 text-xs text-gray-500 border-b">
              {user ? `Role: ${user.role}` : 'User: null'}
            </div>
            
            {user?.role === 'admin' && (
              <button
                className="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors border-none bg-transparent cursor-pointer"
                onClick={handleSettings}
              >
                設定
              </button>
            )}
            <button
              className="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors border-none bg-transparent cursor-pointer"
              onClick={handleLogout}
            >
              ログアウト
            </button>
          </div>
        )}
      </div>
    </header>
  );
}
