'use client';

import React, { useState, useEffect } from 'react';
import { createSearchFilter, getUserSearchFilters, deleteSearchFilter, SearchFilter } from '@/app/lib/api/searchFilters';
import { addSourceListItem, getUserSourceList, deleteSourceListItem, SourceListItem } from '@/app/lib/api/sourceList';

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
  keyword?: string;
  subject?: string;
  onKeywordChange?: (keyword: string) => void;
}

export default function SearchOptions({
  opinionProfile,
  onOpinionProfileChange,
  searchFilters = {},
  onSearchFiltersChange,
  onReset,
  keyword = '',
  subject = '',
  onKeywordChange
}: SearchOptionsProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [showMoreUnits, setShowMoreUnits] = useState(false);
  const [showMoreSources, setShowMoreSources] = useState(false);
  const [selectedUnits, setSelectedUnits] = useState<string[]>(searchFilters.units || []);
  const [selectedYear, setSelectedYear] = useState(searchFilters.year || '');
  const [selectedExam, setSelectedExam] = useState(searchFilters.examSession || '');
  const [checkedStatus, setCheckedStatus] = useState<'all' | 'checked' | 'unchecked'>('all');
  const [sourceList, setSourceList] = useState<SourceListItem[]>([]);
  const [savedFilters, setSavedFilters] = useState<SearchFilter[]>([]);
  const [showSaveModal, setShowSaveModal] = useState(false);
  const [filterName, setFilterName] = useState('');
  const [showLoadModal, setShowLoadModal] = useState(false);
  
  // タグ検索機能の状態
  const [tagSearchInput, setTagSearchInput] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);
  
  // 利用可能なすべてのタグ
  const allAvailableTags = [
    '多項式（展開・因数分解）',
    '平方根',
    '二次方程式',
    '関数 y=ax²',
    '図形の相似',
    '円の性質（円周角）',
    '三平方の定理',
    '標本調査'
  ];
  
  // searchFiltersが外部から変更された時に内部状態を更新（年度・回数は除外）
  useEffect(() => {
    setSelectedUnits(searchFilters.units || []);
    // 年度・回数は独立しているため、searchFiltersからは更新しない
    // setSelectedYear(searchFilters.year || '');
    // setSelectedExam(searchFilters.examSession || '');
    if (searchFilters.isChecked === true) {
      setCheckedStatus('checked');
    } else if (searchFilters.isChecked === false) {
      setCheckedStatus('unchecked');
    } else {
      setCheckedStatus('all');
    }
  }, [searchFilters.units, searchFilters.isChecked]); // 依存配列を明示的に指定
  
  // リセット関数
  const handleReset = () => {
    setSelectedUnits([]);
    setSelectedYear('');
    setSelectedExam('');
    setCheckedStatus('all');
    onSearchFiltersChange?.({});
    onKeywordChange?.('');
    onReset?.();
  };
  
  // 出典リストを読み込む
  const loadSourceList = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await getUserSourceList(token);
      if (response.success && response.items) {
        setSourceList(response.items);
      }
    } catch (error) {
      console.error('出典リストの読み込みに失敗しました:', error);
    }
  };

  // 出典をリストに追加
  const handleAddSource = async () => {
    if (selectedYear && selectedExam) {
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          alert('ログインが必要です');
          return;
        }

        await addSourceListItem(token, selectedYear, selectedExam);
        await loadSourceList();
        
        // 選択をリセット
        setSelectedYear('');
        setSelectedExam('');
      } catch (error) {
        console.error('出典リストへの追加に失敗しました:', error);
        alert('出典リストへの追加に失敗しました');
      }
    }
  };
  
  // 出典をリストから削除
  const handleRemoveSource = async (id: number) => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      await deleteSourceListItem(token, id);
      await loadSourceList();
    } catch (error) {
      console.error('出典リストからの削除に失敗しました:', error);
      alert('出典リストからの削除に失敗しました');
    }
  };

  // 保存された検索条件を読み込む（新しい順に並べ替え）
  const loadSavedFilters = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await getUserSearchFilters(token);
      if (response.success && response.filters) {
        // IDの降順（新しいものが上）にソート
        const sortedFilters = [...response.filters].sort((a, b) => b.id - a.id);
        setSavedFilters(sortedFilters);
      }
    } catch (error) {
      console.error('検索条件の読み込みに失敗しました:', error);
    }
  };

  // 検索条件を保存
  const handleSaveFilter = async () => {
    if (!filterName.trim()) {
      alert('検索条件の名前を入力してください');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        alert('ログインが必要です');
        return;
      }

      const isCheckedValue = checkedStatus === 'all' ? null : checkedStatus === 'checked';

      await createSearchFilter(token, {
        name: filterName,
        keyword: keyword || undefined,
        subject: subject || undefined,
        units: selectedUnits.length > 0 ? selectedUnits : undefined,
        year: selectedYear || undefined,
        exam_session: selectedExam || undefined,
        is_checked: isCheckedValue,
      });

      alert('検索条件を保存しました');
      setFilterName('');
      setShowSaveModal(false);
      loadSavedFilters();
    } catch (error) {
      console.error('検索条件の保存に失敗しました:', error);
      alert('検索条件の保存に失敗しました');
    }
  };

  // 検索条件を読み込む
  const handleLoadFilter = (filter: SearchFilter) => {
    setSelectedUnits(filter.units || []);
    // 年度・回数も保存された条件から復元
    setSelectedYear(filter.year || '');
    setSelectedExam(filter.exam_session || '');
    
    if (filter.is_checked === true) {
      setCheckedStatus('checked');
    } else if (filter.is_checked === false) {
      setCheckedStatus('unchecked');
    } else {
      setCheckedStatus('all');
    }

    // キーワードを更新
    onKeywordChange?.(filter.keyword || '');

    onSearchFiltersChange?.({
      units: filter.units,
      year: filter.year,
      examSession: filter.exam_session,
      isChecked: filter.is_checked,
    });

    setShowLoadModal(false);
  };

  // 検索条件を削除
  const handleDeleteFilter = async (id: number) => {
    if (!confirm('この検索条件を削除しますか？')) return;

    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      await deleteSearchFilter(token, id);
      loadSavedFilters();
    } catch (error) {
      console.error('検索条件の削除に失敗しました:', error);
      alert('検索条件の削除に失敗しました');
    }
  };

  // 初回読み込み時に保存された検索条件と出典リストを取得
  useEffect(() => {
    loadSavedFilters();
    loadSourceList();
  }, []);
  
  // タグ検索のサジェスト機能
  const getFilteredSuggestions = () => {
    if (!tagSearchInput.trim()) return [];
    
    return allAvailableTags.filter(tag =>
      tag.includes(tagSearchInput) && !selectedUnits.includes(tag)
    );
  };
  
  // タグを選択
  const handleSelectTag = (tag: string) => {
    const newUnits = [...selectedUnits, tag];
    setSelectedUnits(newUnits);
    setTagSearchInput('');
    setShowSuggestions(false);
    
    onSearchFiltersChange?.({
      ...searchFilters,
      units: newUnits,
      year: selectedYear || searchFilters.year,
      examSession: selectedExam || searchFilters.examSession
    });
  };
  
  // タグを削除
  const handleRemoveTag = (tag: string) => {
    const newUnits = selectedUnits.filter(u => u !== tag);
    setSelectedUnits(newUnits);
    
    onSearchFiltersChange?.({
      ...searchFilters,
      units: newUnits,
      year: selectedYear || searchFilters.year,
      examSession: selectedExam || searchFilters.examSession
    });
  };
  
  // タグ検索入力の変更
  const handleTagSearchChange = (value: string) => {
    setTagSearchInput(value);
    setShowSuggestions(value.trim().length > 0);
  };

  return (
    <div className={`mb-4 pb-4 border-b-2 border-gray-200 ${isOpen ? 'is-open' : ''}`}>
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
          {/* タグ検索セクション */}
          <div className="pt-6 px-6 pb-0">
            {/* 選択されたタグの表示エリア */}
            {selectedUnits.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-3">
                {selectedUnits.map((tag) => (
                  <div
                    key={tag}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-[13px] font-bold bg-blue-50 text-blue-500 animate-[fadeInTag_0.2s_ease-out]"
                  >
                    <span>{tag}</span>
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      className="flex items-center justify-center w-4 h-4 rounded-full bg-transparent border-none cursor-pointer text-blue-500 opacity-60 hover:opacity-100 hover:bg-blue-100 transition-all"
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            )}
            
            {/* タグ検索入力 */}
            <div className="relative w-full mb-0">
              <input
                type="text"
                value={tagSearchInput}
                onChange={(e) => handleTagSearchChange(e.target.value)}
                onFocus={() => tagSearchInput.trim() && setShowSuggestions(true)}
                onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
                placeholder="タグを検索 (例: 相似, 二次関数...)"
                className="w-full font-sans text-[15px] border border-gray-200 rounded-lg px-3.5 py-3 mb-0 focus:outline-none focus:border-blue-500 focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)] transition-all"
              />
              
              {/* サジェストリスト */}
              {showSuggestions && getFilteredSuggestions().length > 0 && (
                <div className="absolute top-full left-0 w-full bg-white border border-gray-200 rounded-lg shadow-[0_4px_12px_rgba(0,0,0,0.15)] mt-1.5 max-h-60 overflow-y-auto z-[1000]">
                  {getFilteredSuggestions().map((tag) => (
                    <div
                      key={tag}
                      onClick={() => handleSelectTag(tag)}
                      className="px-3.5 py-3 text-sm text-gray-800 cursor-pointer border-b border-gray-50 last:border-b-0 transition-colors hover:bg-gray-100"
                    >
                      {tag.split(new RegExp(`(${tagSearchInput})`, 'gi')).map((part, i) =>
                        part.toLowerCase() === tagSearchInput.toLowerCase() ? (
                          <span key={i} className="font-bold text-blue-500">{part}</span>
                        ) : (
                          <span key={i}>{part}</span>
                        )
                      )}
                    </div>
                  ))}
                </div>
              )}
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
                        // 年度・回数は保持したまま単元のみ更新
                        onSearchFiltersChange?.({
                          ...searchFilters,
                          units: newUnits,
                          year: selectedYear || searchFilters.year,
                          examSession: selectedExam || searchFilters.examSession
                        });
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
                            // 年度・回数は保持したまま単元のみ更新
                            onSearchFiltersChange?.({
                              ...searchFilters,
                              units: newUnits,
                              year: selectedYear || searchFilters.year,
                              examSession: selectedExam || searchFilters.examSession
                            });
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
                      onChange={(e) => setSelectedYear(e.target.value)}
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
                      onChange={(e) => setSelectedExam(e.target.value)}
                      className="w-full px-3.5 py-2.5 pr-9 font-sans text-sm text-gray-800 bg-white border border-gray-200 rounded-lg cursor-pointer appearance-none transition-all hover:bg-gray-50 focus:outline-none focus:border-blue-500 focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)]"
                      style={{backgroundImage: "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%236b7280' stroke-width='2'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' d='M19 9l-7 7-7-7'/%3E%3C/svg%3E\")", backgroundRepeat: 'no-repeat', backgroundPosition: 'right 12px center', backgroundSize: '16px'}}
                    >
                      <option value="">回数</option>
                      <option value="第1回模試">第1回模試</option>
                      <option value="第2回模試">第2回模試</option>
                      <option value="第3回模試">第3回模試</option>
                      <option value="第4回模試">第4回模試</option>
                      <option value="第5回模試">第5回模試</option>
                      <option value="第6回模試">第6回模試</option>
                      <option value="第7回模試">第7回模試</option>
                      <option value="第8回模試">第8回模試</option>
                      <option value="本番">本番</option>
                    </select>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={handleAddSource}
                  className="w-full flex items-center justify-center gap-2 px-2.5 py-2.5 text-sm font-bold text-white bg-gray-800 border-none rounded-lg cursor-pointer transition-all hover:bg-black hover:-translate-y-0.5 hover:shadow-[0_2px_5px_rgba(0,0,0,0.1)] active:translate-y-0.5"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <line x1="12" y1="5" x2="12" y2="19"></line>
                    <line x1="5" y1="12" x2="19" y2="12"></line>
                  </svg>
                  リストに追加
                </button>
                
                {/* 追加された出典リスト（チェックボックス） */}
                {sourceList.length > 0 && (
                  <div className="flex flex-col gap-3 mt-2">
                    {sourceList.slice(0, showMoreSources ? sourceList.length : 1).map((source) => (
                      <label key={source.id} className="flex items-center justify-between gap-2 text-[15px] cursor-pointer group">
                        <div className="flex items-center gap-2 flex-1">
                          <input
                            type="checkbox"
                            className="w-4 h-4"
                            style={{accentColor: 'var(--primary)'}}
                            checked={selectedYear === source.year && selectedExam === source.exam_session}
                            onChange={(e) => {
                              if (e.target.checked) {
                                setSelectedYear(source.year);
                                setSelectedExam(source.exam_session);
                                onSearchFiltersChange?.({ ...searchFilters, year: source.year, examSession: source.exam_session });
                              } else {
                                setSelectedYear('');
                                setSelectedExam('');
                                const { year, examSession, ...rest } = searchFilters;
                                onSearchFiltersChange?.(rest);
                              }
                            }}
                          />
                          <span>{source.year}年 {source.exam_session}</span>
                        </div>
                        <button
                          type="button"
                          onClick={(e) => {
                            e.preventDefault();
                            handleRemoveSource(source.id);
                          }}
                          className="text-gray-400 hover:text-red-500 font-bold text-lg leading-none opacity-0 group-hover:opacity-100 transition-opacity"
                        >
                          ×
                        </button>
                      </label>
                    ))}
                    {sourceList.length > 1 && (
                      <button
                        type="button"
                        onClick={() => setShowMoreSources(!showMoreSources)}
                        className={`flex items-center justify-center gap-1.5 w-full mt-1 px-0 py-2 bg-transparent border border-dashed border-gray-200 rounded-md text-gray-500 text-[13px] font-bold cursor-pointer transition-all hover:bg-gray-100 hover:text-blue-500 hover:border-blue-500 ${showMoreSources ? 'is-open' : ''}`}
                      >
                        <span>{showMoreSources ? '閉じる' : 'もっと見る'}</span>
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
                          className={`transition-transform duration-300 ${showMoreSources ? 'rotate-180' : ''}`}
                        >
                          <polyline points="6 9 12 15 18 9"></polyline>
                        </svg>
                      </button>
                    )}
                  </div>
                )}
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
                      // 年度・回数は保持したままステータスのみ更新
                      onSearchFiltersChange?.({
                        ...searchFilters,
                        isChecked: true,
                        year: selectedYear || searchFilters.year,
                        examSession: selectedExam || searchFilters.examSession
                      });
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
                      // 年度・回数は保持したままステータスのみ更新
                      onSearchFiltersChange?.({
                        ...searchFilters,
                        isChecked: false,
                        year: selectedYear || searchFilters.year,
                        examSession: selectedExam || searchFilters.examSession
                      });
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
                      // 年度・回数は保持したままステータスのみ更新
                      onSearchFiltersChange?.({
                        ...rest,
                        year: selectedYear || searchFilters.year,
                        examSession: selectedExam || searchFilters.examSession
                      });
                    }}
                  />
                  <span>すべて</span>
                </label>
              </div>
            </div>
          </div>

          {/* 保存・読み込み・リセットボタン */}
          <div className="p-6 border-t border-gray-200 flex gap-3">
            <button
              type="button"
              onClick={handleReset}
              className="flex items-center justify-center gap-2 px-4 py-2.5 text-sm font-bold text-gray-700 bg-white border border-gray-300 rounded-lg cursor-pointer transition-all hover:bg-gray-50"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"></path>
                <path d="M3 3v5h5"></path>
              </svg>
              リセット
            </button>
            <button
              type="button"
              onClick={() => setShowSaveModal(true)}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 text-sm font-bold text-white bg-blue-600 border-none rounded-lg cursor-pointer transition-all hover:bg-blue-700"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path>
                <polyline points="17 21 17 13 7 13 7 21"></polyline>
                <polyline points="7 3 7 8 15 8"></polyline>
              </svg>
              検索条件を保存
            </button>
            <button
              type="button"
              onClick={() => {
                setShowLoadModal(true);
                loadSavedFilters();
              }}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 text-sm font-bold text-gray-700 bg-white border border-gray-300 rounded-lg cursor-pointer transition-all hover:bg-gray-50"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                <polyline points="7 10 12 15 17 10"></polyline>
                <line x1="12" y1="15" x2="12" y2="3"></line>
              </svg>
              検索条件を読み込む
            </button>
          </div>
        </div>
      )}

      {/* 保存モーダル */}
      {showSaveModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={() => setShowSaveModal(false)}>
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold mb-4">検索条件を保存</h3>
            <input
              type="text"
              value={filterName}
              onChange={(e) => setFilterName(e.target.value)}
              placeholder="検索条件の名前を入力"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg mb-4"
            />
            <div className="flex gap-3">
              <button
                onClick={handleSaveFilter}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                保存
              </button>
              <button
                onClick={() => {
                  setShowSaveModal(false);
                  setFilterName('');
                }}
                className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300"
              >
                キャンセル
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 読み込みモーダル */}
      {showLoadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={() => setShowLoadModal(false)}>
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 max-h-[80vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold mb-4">保存された検索条件</h3>
            {savedFilters.length === 0 ? (
              <p className="text-gray-500 text-center py-8">保存された検索条件がありません</p>
            ) : (
              <div className="space-y-3">
                {savedFilters.map((filter) => (
                  <div key={filter.id} className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50">
                    <div className="flex items-start justify-between mb-2">
                      <h4 className="font-bold text-gray-800">{filter.name}</h4>
                      <button
                        onClick={() => handleDeleteFilter(filter.id)}
                        className="text-red-500 hover:text-red-700"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                        </svg>
                      </button>
                    </div>
                    <div className="text-sm text-gray-600 mb-3 space-y-1">
                      {filter.keyword && <div>キーワード: {filter.keyword}</div>}
                      {filter.subject && <div>科目: {filter.subject}</div>}
                      {filter.units && filter.units.length > 0 && (
                        <div>単元: {filter.units.join(', ')}</div>
                      )}
                      {filter.year && <div>年度: {filter.year}年</div>}
                      {filter.exam_session && <div>回数: {filter.exam_session}</div>}
                      {filter.is_checked !== null && (
                        <div>ステータス: {filter.is_checked ? 'チェック済み' : '未チェック'}</div>
                      )}
                    </div>
                    <button
                      onClick={() => handleLoadFilter(filter)}
                      className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-bold"
                    >
                      この条件を読み込む
                    </button>
                  </div>
                ))}
              </div>
            )}
            <button
              onClick={() => setShowLoadModal(false)}
              className="w-full mt-4 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300"
            >
              閉じる
            </button>
          </div>
        </div>
      )}
    </div>
  );
}