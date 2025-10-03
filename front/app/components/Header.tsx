'use client';

import React, { useState, useEffect, useRef } from 'react';

export default function Header() {
  const [showLogoutMenu, setShowLogoutMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  const handleLogout = () => {
    // ローカルストレージからトークンを削除
    localStorage.removeItem('authToken');
    
    // ログインページにリダイレクト
    window.location.href = '/login';
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
          <div className="absolute right-0 top-12 bg-white border border-mongene-border rounded-xl shadow-lg py-2 min-w-32 z-10">
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
