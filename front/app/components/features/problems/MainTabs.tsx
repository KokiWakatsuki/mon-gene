'use client';

import React from 'react';

interface MainTabsProps {
  activeTab: 'list' | 'generate';
  onTabChange: (tab: 'list' | 'generate') => void;
}

export default function MainTabs({ activeTab, onTabChange }: MainTabsProps) {
  return (
    <nav className="flex gap-1 border-b-2 border-gray-200 my-6 max-w-6xl mx-auto px-4" aria-label="メインタブ">
      <button
        className={`font-sans text-base font-bold px-5 py-3 bg-transparent border-none border-b-[3px] transition-colors translate-y-0.5 cursor-pointer ${
          activeTab === 'list'
            ? 'text-gray-800 border-b-blue-500'
            : 'text-gray-500 border-b-transparent hover:text-gray-800 hover:bg-gray-100'
        }`}
        onClick={() => onTabChange('list')}
        aria-current={activeTab === 'list' ? 'page' : undefined}
      >
        問題一覧
      </button>
      <button
        className={`font-sans text-base font-bold px-5 py-3 bg-transparent border-none border-b-[3px] transition-colors translate-y-0.5 cursor-pointer ${
          activeTab === 'generate'
            ? 'text-gray-800 border-b-blue-500'
            : 'text-gray-500 border-b-transparent hover:text-gray-800 hover:bg-gray-100'
        }`}
        onClick={() => onTabChange('generate')}
        aria-current={activeTab === 'generate' ? 'page' : undefined}
      >
        問題生成
      </button>
    </nav>
  );
}