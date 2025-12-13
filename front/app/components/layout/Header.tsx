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
    <header className="relative flex items-center justify-between bg-white shadow-[0_1px_3px_rgba(0,0,0,0.05)]" style={{
      padding: '16px max(16px, calc((100vw - 1024px) / 2 + 16px))',
      zIndex: 'var(--z-header)'
    }}>
      <a href="#" className="flex items-center no-underline">
        <img
          src="/images/モンジェネロゴタイプ.svg"
          alt="Mongene"
          className="h-10 w-auto object-contain"
        />
      </a>
      
      <div className="relative" ref={menuRef}>
        <button
          className="w-10 h-10 rounded-full bg-transparent border-none cursor-pointer flex items-center justify-center transition-colors hover:bg-gray-100"
          onClick={() => setShowLogoutMenu(!showLogoutMenu)}
          aria-label="ユーザー設定"
          aria-expanded={showLogoutMenu}
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-6 h-6 text-gray-500">
            <path strokeLinecap="round" strokeLinejoin="round" d="M17.982 18.725A7.488 7.488 0 0 0 12 15.75a7.488 7.488 0 0 0-5.982 2.975m11.963 0a9 9 0 1 0-11.963 0m11.963 0A8.966 8.966 0 0 1 12 21a8.966 8.966 0 0 1-5.982-2.275M15 9.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
          </svg>
        </button>
        
        {showLogoutMenu && (
          <div
            className="absolute right-0 top-12 bg-white border border-gray-200 rounded-lg shadow-[0_4px_12px_rgba(0,0,0,0.1)] py-2 min-w-[200px] flex flex-col"
            style={{ zIndex: 'var(--z-modal)' }}
          >
            {user?.role === 'admin' && (
              <button
                className="w-full px-4 py-2.5 text-left text-sm font-medium text-gray-800 hover:bg-gray-50 transition-colors border-none bg-transparent cursor-pointer"
                onClick={handleSettings}
              >
                アカウント設定
              </button>
            )}
            <button
              className="w-full px-4 py-2.5 text-left text-sm font-medium text-red-600 hover:bg-gray-50 transition-colors border-none bg-transparent cursor-pointer border-t border-gray-200 mt-1 pt-3"
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
