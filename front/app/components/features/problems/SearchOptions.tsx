'use client';

import React, { useState } from 'react';

interface SearchOptionsProps {
  opinionProfile: any;
  onOpinionProfileChange: (profile: any) => void;
}

export default function SearchOptions({ opinionProfile, onOpinionProfileChange }: SearchOptionsProps) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <section className="mb-6">
      <button
        className="w-full flex justify-between items-center bg-white border border-gray-200 rounded-xl px-6 py-5 font-sans text-lg font-semibold text-gray-800 cursor-pointer transition-all hover:bg-gray-50"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-controls="searchOptionsPanel"
      >
        <span className="flex items-center gap-2">
          <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 3c2.755 0 5.455.232 8.083.678.533.09.917.556.917 1.096v1.044a2.25 2.25 0 01-.659 1.591l-5.432 5.432a2.25 2.25 0 00-.659 1.591v2.927a2.25 2.25 0 01-1.244 2.013L9.75 21v-6.568a2.25 2.25 0 00-.659-1.591L3.659 7.409A2.25 2.25 0 013 5.818V4.774c0-.54.384-1.006.917-1.096A48.32 48.32 0 0112 3z" />
          </svg>
          検索オプション（Coming Soon）
        </span>
        <svg
          width="14"
          height="9"
          viewBox="0 0 14 9"
          className={`transition-transform ${isOpen ? 'rotate-180' : ''}`}
        >
          <path d="M1 1l6 6 6-6" fill="none" stroke="#333" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </button>

      {isOpen && (
        <div
          id="searchOptionsPanel"
          className="bg-white border border-gray-200 border-t-0 rounded-b-xl px-6 py-6 -mt-3 shadow-sm"
        >
          <input
            type="text"
            placeholder="フィルター内を検索 (例: 相似, 円周角...)"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg mb-6 text-base focus:outline-none focus:ring-2 focus:ring-blue-500"
          />

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
            {/* 主要な解法 */}
            <div>
              <h4 className="text-base font-semibold text-gray-800 mb-4 pb-3 border-b border-gray-200">
                主要な解法
              </h4>
              <div className="flex flex-col gap-3">
                {['三平方の定理', '相似', '円周角の定理', '解の公式'].map((item) => (
                  <label key={item} className="flex items-center gap-2 text-base cursor-pointer">
                    <input
                      type="checkbox"
                      className="w-4 h-4 accent-blue-500"
                    />
                    <span>{item}</span>
                  </label>
                ))}
                <button type="button" className="text-sm text-blue-500 bg-transparent border-none cursor-pointer p-0 mt-3 text-left hover:underline">
                  + もっと見る
                </button>
              </div>
            </div>

            {/* 出典 (年度) */}
            <div>
              <h4 className="text-base font-semibold text-gray-800 mb-4 pb-3 border-b border-gray-200">
                出典 (年度)
              </h4>
              <div className="flex flex-col gap-3">
                {['新潟県 2024', '東京都 2024', 'AI生成'].map((item) => (
                  <label key={item} className="flex items-center gap-2 text-base cursor-pointer">
                    <input
                      type="checkbox"
                      className="w-4 h-4 accent-blue-500"
                    />
                    <span>{item}</span>
                  </label>
                ))}
                <button type="button" className="text-sm text-blue-500 bg-transparent border-none cursor-pointer p-0 mt-3 text-left hover:underline">
                  + もっと見る
                </button>
              </div>
            </div>

            {/* 問題構造 */}
            <div>
              <h4 className="text-base font-semibold text-gray-800 mb-4 pb-3 border-b border-gray-200">
                問題構造
              </h4>
              <div className="flex flex-col gap-3">
                {['動点(点Pなど)を含む', '補助線が必要', '記述・証明問題'].map((item) => (
                  <label key={item} className="flex items-center gap-2 text-base cursor-pointer">
                    <input
                      type="checkbox"
                      className="w-4 h-4 accent-blue-500"
                    />
                    <span>{item}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* 難易度 (総合) */}
            <div>
              <h4 className="text-base font-semibold text-gray-800 mb-4 pb-3 border-b border-gray-200">
                難易度 (総合)
              </h4>
              <div className="flex items-center gap-2.5 text-sm font-medium">
                <span>Lv1</span>
                <input
                  type="range"
                  min="1"
                  max="5"
                  defaultValue="3"
                  className="flex-1 accent-blue-500"
                />
                <span>Lv5</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}