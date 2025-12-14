'use client';

import React, { useState, useEffect } from 'react';

interface SearchOptionsProps {
  opinionProfile: any;
  onOpinionProfileChange: (profile: any) => void;
  searchFilters?: {
    units?: string[];
    year?: string;
    examSession?: string;
    isChecked?: boolean;
  };
  onSearchFiltersChange?: (filters: any) => void;
  onReset?: () => void;
}

export default function SearchOptions({
  opinionProfile,
  onOpinionProfileChange,
  searchFilters = {},
  onSearchFiltersChange,
  onReset
}: SearchOptionsProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [showMoreUnits, setShowMoreUnits] = useState(false);
  const [selectedUnits, setSelectedUnits] = useState<string[]>(searchFilters.units || []);
  const [selectedYear, setSelectedYear] = useState(searchFilters.year || '');
  const [selectedExam, setSelectedExam] = useState(searchFilters.examSession || '');
  const [checkedStatus, setCheckedStatus] = useState<'all' | 'checked' | 'unchecked'>('all');
  
  // searchFiltersが外部から変更された時に内部状態を更新
  useEffect(() => {
    setSelectedUnits(searchFilters.units || []);
    setSelectedYear(searchFilters.year || '');
    setSelectedExam(searchFilters.examSession || '');
    if (searchFilters.isChecked === true) {
      setCheckedStatus('checked');
    } else if (searchFilters.isChecked === false) {
      setCheckedStatus('unchecked');
    } else {
      setCheckedStatus('all');
    }
  }, [searchFilters]);
  
  // リセット関数
  const handleReset = () => {
    setSelectedUnits([]);
    setSelectedYear('');
    setSelectedExam('');
    setCheckedStatus('all');
    onSearchFiltersChange?.({});
    onReset?.();
  };

  return (
    <div className={`mb-4 ${isOpen ? 'is-open' : ''}`}>
      <button
        type="button"
        className={`w-full flex justify-between items-center border rounded-[10px] px-4 py-3 cursor-pointer transition-all shadow-[0_1px_2px_rgba(0,0,0,0.05)] ${
          isOpen
            ? 'bg-gray-800 border-gray-800'
            : 'bg-white border-gray-200 hover:bg-gray-50 hover:border-gray-500'
        }`}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
      >
        <span className={`flex items-center gap-2.5 ${isOpen ? 'text-white' : 'text-gray-800'}`}>
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M12 2H2v10l9.29 9.29c.94.94 2.48.94 3.42 0l6.58-6.58c.94-.94.94-2.48 0-3.42L12 2Z" />
            <path d="M7 7h.01" />
          </svg>
          <span className="text-[15px] font-bold tracking-[0.02em]">タグを選択</span>
        </span>
        <svg
          className={`transition-transform duration-300 ${isOpen ? 'rotate-180 text-white' : 'text-gray-500'}`}
          xmlns="http://www.w3.org/2000/svg"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M6 9l6 6 6-6"/>
        </svg>
      </button>

      {isOpen && (
        <div className="border border-gray-200 border-t border-gray-200 rounded-xl bg-white mt-2 shadow-[0_4px_15px_rgba(0,0,0,0.05)] overflow-y-auto">
          {/* アクティブタグエリアと検索 */}
          <div className="p-6 pb-0">
            <div className="flex flex-wrap gap-2 mb-3 min-h-0">
              {/* アクティブタグがここに表示される */}
            </div>
            <div className="relative w-full mb-0">
              <input
                type="text"
                placeholder="タグを検索 (例: 相似, 二次関数...)"
                className="w-full font-sans text-[15px] border border-gray-200 rounded-lg px-3.5 py-3 mb-0"
              />
              <div className="suggestions-list"></div>
            </div>
          </div>

          {/* フィルターグリッド */}
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {/* 単元 */}
            <div>
              <h4 className="text-base font-semibold text-gray-800 m-0 mb-4 pb-3 border-b border-gray-200">📚 単元</h4>
              <div className="flex flex-col gap-3">
                {['多項式（展開・因数分解）', '平方根', '二次方程式'].map((item) => (
                  <label key={item} className="flex items-center gap-2 text-[15px] cursor-pointer">
                    <input
                      type="checkbox"
                      className="w-4 h-4"
                      style={{accentColor: 'var(--primary)'}}
                      checked={selectedUnits.includes(item)}
                      onChange={(e) => {
                        const newUnits = e.target.checked
                          ? [...selectedUnits, item]
                          : selectedUnits.filter(u => u !== item);
                        setSelectedUnits(newUnits);
                        onSearchFiltersChange?.({ ...searchFilters, units: newUnits });
                      }}
                    />
                    <span>{item}</span>
                  </label>
                ))}
                {showMoreUnits && (
                  <div className="flex flex-col gap-3">
                    {['関数 y=ax²', '図形の相似', '円の性質（円周角）', '三平方の定理', '標本調査'].map((item) => (
                      <label key={item} className="flex items-center gap-2 text-[15px] cursor-pointer">
                        <input
                          type="checkbox"
                          className="w-4 h-4"
                          style={{accentColor: 'var(--primary)'}}
                          checked={selectedUnits.includes(item)}
                          onChange={(e) => {
                            const newUnits = e.target.checked
                              ? [...selectedUnits, item]
                              : selectedUnits.filter(u => u !== item);
                            setSelectedUnits(newUnits);
                            onSearchFiltersChange?.({ ...searchFilters, units: newUnits });
                          }}
                        />
                        <span>{item}</span>
                      </label>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  onClick={() => setShowMoreUnits(!showMoreUnits)}
                  className={`flex items-center justify-center gap-1.5 w-full mt-3 px-0 py-2 bg-transparent border border-dashed border-gray-200 rounded-md text-gray-500 text-[13px] font-bold cursor-pointer transition-all hover:bg-gray-100 hover:text-blue-500 hover:border-blue-500 ${showMoreUnits ? 'is-open' : ''}`}
                >
                  <span>{showMoreUnits ? '閉じる' : 'もっと見る'}</span>
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className={`transition-transform duration-300 ${showMoreUnits ? 'rotate-180' : ''}`}
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
              </div>
            </div>

            {/* 出典 (年度・回数) */}
            <div>
              <h4 className="text-base font-semibold text-gray-800 m-0 mb-4 pb-3 border-b border-gray-200">🏫 出典 (年度・回数)</h4>
              <div className="flex flex-col gap-3">
                <div className="flex gap-3">
                  <div className="relative flex-1">
                    <select
                      value={selectedYear}
                      onChange={(e) => {
                        setSelectedYear(e.target.value);
                        onSearchFiltersChange?.({ ...searchFilters, year: e.target.value });
                      }}
                      className="w-full px-3.5 py-2.5 pr-9 font-sans text-sm text-gray-800 bg-white border border-gray-200 rounded-lg cursor-pointer appearance-none transition-all hover:bg-gray-50 focus:outline-none focus:border-blue-500 focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)]"
                      style={{backgroundImage: "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%236b7280' stroke-width='2'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' d='M19 9l-7 7-7-7'/%3E%3C/svg%3E\")", backgroundRepeat: 'no-repeat', backgroundPosition: 'right 12px center', backgroundSize: '16px'}}
                    >
                      <option value="">年度</option>
                      <option value="2025">2025年</option>
                      <option value="2024">2024年</option>
                      <option value="2023">2023年</option>
                      <option value="2022">2022年</option>
                      <option value="2021">2021年</option>
                      <option value="2020">2020年</option>
                    </select>
                  </div>
                  <div className="relative flex-1">
                    <select
                      value={selectedExam}
                      onChange={(e) => {
                        setSelectedExam(e.target.value);
                        onSearchFiltersChange?.({ ...searchFilters, examSession: e.target.value });
                      }}
                      className="w-full px-3.5 py-2.5 pr-9 font-sans text-sm text-gray-800 bg-white border border-gray-200 rounded-lg cursor-pointer appearance-none transition-all hover:bg-gray-50 focus:outline-none focus:border-blue-500 focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)]"
                      style={{backgroundImage: "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%236b7280' stroke-width='2'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' d='M19 9l-7 7-7-7'/%3E%3C/svg%3E\")", backgroundRepeat: 'no-repeat', backgroundPosition: 'right 12px center', backgroundSize: '16px'}}
                    >
                      <option value="">回数</option>
                      <option value="第1回">第1回</option>
                      <option value="第2回">第2回</option>
                      <option value="第3回">第3回</option>
                      <option value="プレ">プレ</option>
                      <option value="追試">追試</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>

            {/* ステータス */}
            <div>
              <h4 className="text-base font-semibold text-gray-800 m-0 mb-4 pb-3 border-b border-gray-200">✅ ステータス</h4>
              <div className="flex flex-col gap-3">
                <label className="flex items-center gap-2 text-[15px] cursor-pointer">
                  <input
                    type="radio"
                    name="status"
                    className="w-4 h-4"
                    style={{accentColor: 'var(--primary)'}}
                    checked={checkedStatus === 'checked'}
                    onChange={() => {
                      setCheckedStatus('checked');
                      onSearchFiltersChange?.({ ...searchFilters, isChecked: true });
                    }}
                  />
                  <span>チェック済み</span>
                </label>
                <label className="flex items-center gap-2 text-[15px] cursor-pointer">
                  <input
                    type="radio"
                    name="status"
                    className="w-4 h-4"
                    style={{accentColor: 'var(--primary)'}}
                    checked={checkedStatus === 'unchecked'}
                    onChange={() => {
                      setCheckedStatus('unchecked');
                      onSearchFiltersChange?.({ ...searchFilters, isChecked: false });
                    }}
                  />
                  <span>未チェック</span>
                </label>
                <label className="flex items-center gap-2 text-[15px] cursor-pointer">
                  <input
                    type="radio"
                    name="status"
                    className="w-4 h-4"
                    style={{accentColor: 'var(--primary)'}}
                    checked={checkedStatus === 'all'}
                    onChange={() => {
                      setCheckedStatus('all');
                      const { isChecked, ...rest } = searchFilters;
                      onSearchFiltersChange?.(rest);
                    }}
                  />
                  <span>すべて</span>
                </label>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}