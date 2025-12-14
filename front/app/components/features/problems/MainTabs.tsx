'use client';

import React from 'react';

interface MainTabsProps {
  activeTab: 'list' | 'generate';
  onTabChange: (tab: 'list' | 'generate') => void;
}

export default function MainTabs({ activeTab, onTabChange }: MainTabsProps) {
  return (
    <nav className="flex gap-1 border-b-2 border-gray-200 my-6 max-w-5xl mx-auto px-4" aria-label="メインタブ">
      <button
        className={`font-sans text-base font-bold px-5 py-3 bg-transparent border-none transition-colors cursor-pointer relative ${
          activeTab === 'list'
            ? 'text-gray-800'
            : 'text-gray-500 hover:text-gray-800 hover:bg-gray-100'
        }`}
        onClick={() => onTabChange('list')}
        aria-current={activeTab === 'list' ? 'page' : undefined}
      >
        問題一覧
        {activeTab === 'list' && (
          <span className="absolute bottom-0 left-0 right-0 h-[3px] bg-blue-500 translate-y-[2px]"></span>
        )}
      </button>
      <button
        className={`font-sans text-base font-bold px-5 py-3 bg-transparent border-none transition-colors cursor-pointer relative ${
          activeTab === 'generate'
            ? 'text-gray-800'
            : 'text-gray-500 hover:text-gray-800 hover:bg-gray-100'
        }`}
        onClick={() => onTabChange('generate')}
        aria-current={activeTab === 'generate' ? 'page' : undefined}
      >
        問題生成
        {activeTab === 'generate' && (
          <span className="absolute bottom-0 left-0 right-0 h-[3px] bg-blue-500 translate-y-[2px]"></span>
        )}
      </button>
    </nav>
  );
}