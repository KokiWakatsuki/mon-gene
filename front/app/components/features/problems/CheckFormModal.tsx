'use client';

import React, { useState, useEffect } from 'react';
import { API_CONFIG } from '@/app/lib/config/api';
import { UNITS_HIERARCHY, getAllUnitsFlat, getAllChildren, getIdByLabel } from '@/app/lib/data/units';
import HierarchicalUnitSelector from './HierarchicalUnitSelector';

interface CheckFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  problemId: string;
  problemTitle: string;
  problemContent?: string;
  solutionText?: string;
  imageBase64?: string;
  initialCheckInfo?: CheckInfo;
  onSave: (checkInfo: CheckInfo) => Promise<void>;
  onUpdate?: (updatedData: { imageBase64?: string }) => void;
}

interface UserInfo {
  school_code: string;
  email: string;
  problem_generation_limit: number;
  problem_generation_count: number;
  figure_regeneration_limit: number;
  figure_regeneration_count: number;
}

export interface CheckInfo {
  problem_text_ok: boolean;
  solution_ok: boolean;
  figure_ok: boolean;
  units: string[];
  year: string;
  exam_session: string;
}

const YEARS = ['2020', '2021', '2022', '2023', '2024', '2025'];
const EXAM_SESSIONS = ['第1回模試', '第2回模試', '第3回模試', '第4回模試', '第5回模試', '第6回模試', '第7回模試', '第8回模試', '本番'];

export default function CheckFormModal({
  isOpen,
  onClose,
  problemId,
  problemTitle,
  problemContent,
  solutionText,
  imageBase64,
  initialCheckInfo,
  onSave,
  onUpdate,
}: CheckFormModalProps) {
  const [checkInfo, setCheckInfo] = useState<CheckInfo>({
    problem_text_ok: false,
    solution_ok: false,
    figure_ok: false,
    units: [],
    year: '',
    exam_session: '',
  });
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [currentImageBase64, setCurrentImageBase64] = useState('');
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  
  // タグ検索機能の状態
  const [tagSearchInput, setTagSearchInput] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);
  
  // アコーディオンの開閉状態
  const [isFigureOpen, setIsFigureOpen] = useState(false);
  
  // 利用可能なすべてのタグ（階層構造から取得）
  const allAvailableTags = getAllUnitsFlat();

  // ユーザー情報を取得する関数
  const fetchUserInfo = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(API_CONFIG.USER_INFO_API_URL, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setUserInfo(data);
      }
    } catch (error) {
      console.error('ユーザー情報の取得に失敗しました:', error);
    }
  };

  // 図形再生成制限チェック
  const isFigureRegenerationLimitReached = () => {
    if (!userInfo) return true;
    if (userInfo.figure_regeneration_limit === undefined || userInfo.figure_regeneration_limit === null) return true;
    if (userInfo.figure_regeneration_count === undefined || userInfo.figure_regeneration_count === null) return true;
    if (userInfo.figure_regeneration_limit === -1) return false;
    return userInfo.figure_regeneration_count >= userInfo.figure_regeneration_limit;
  };

  // モーダルが開かれたときにユーザー情報を取得
  useEffect(() => {
    if (isOpen) {
      fetchUserInfo();
      setCurrentImageBase64(imageBase64 || '');
    }
  }, [isOpen, imageBase64]);

  useEffect(() => {
    if (isOpen && initialCheckInfo) {
      setCheckInfo(initialCheckInfo);
    } else if (isOpen) {
      // リセット
      setCheckInfo({
        problem_text_ok: false,
        solution_ok: false,
        figure_ok: false,
        units: [],
        year: '',
        exam_session: '',
      });
    }
    setError(null);
  }, [isOpen, initialCheckInfo]);

  const handleCheckboxChange = (field: 'problem_text_ok' | 'solution_ok' | 'figure_ok') => {
    setCheckInfo((prev) => ({
      ...prev,
      [field]: !prev[field],
    }));
  };

  const handleYearChange = (year: string) => {
    setCheckInfo((prev) => ({
      ...prev,
      year,
    }));
  };

  const handleExamSessionChange = (session: string) => {
    setCheckInfo((prev) => ({
      ...prev,
      exam_session: session,
    }));
  };
  
  // タグ検索のサジェスト機能
  const getFilteredSuggestions = () => {
    if (!tagSearchInput.trim()) return [];
    
    return allAvailableTags.filter(tag =>
      tag.includes(tagSearchInput) && !checkInfo.units.includes(tag)
    );
  };
  
  // タグを選択
  const handleSelectTag = (tag: string) => {
    const tagId = allAvailableTags.indexOf(tag) >= 0 ? getIdByLabel(tag, UNITS_HIERARCHY) : null;
    
    let newUnits = [...checkInfo.units];
    
    // タグ自体を追加
    if (!newUnits.includes(tag)) {
      newUnits.push(tag);
    }
    
    // もしタグが親階層の場合、すべての子孫も追加
    if (tagId) {
      const children = getAllChildren(tagId, UNITS_HIERARCHY);
      children.forEach(child => {
        if (!newUnits.includes(child)) {
          newUnits.push(child);
        }
      });
    }
    
    setCheckInfo((prev) => ({
      ...prev,
      units: newUnits,
    }));
    setTagSearchInput('');
    setShowSuggestions(false);
  };
  
  // タグを削除
  const handleRemoveTag = (tag: string) => {
    setCheckInfo((prev) => ({
      ...prev,
      units: prev.units.filter(u => u !== tag),
    }));
  };
  
  // タグ検索入力の変更
  const handleTagSearchChange = (value: string) => {
    setTagSearchInput(value);
    setShowSuggestions(value.trim().length > 0);
  };

  // 図形を再生成
  const handleRegenerateGeometry = async () => {
    if (!problemId) return;

    // 制限チェック
    if (isFigureRegenerationLimitReached()) {
      alert(`図形再生成回数の上限（${userInfo?.figure_regeneration_limit}回）に達しました。これ以上図形を再生成することはできません。`);
      return;
    }

    const parsedId = parseInt(problemId);
    if (parsedId <= 0) {
      setError('無効な問題IDです');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません');
      }

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/regenerate-geometry`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          id: parsedId,
          content: problemContent,
        }),
      });

      if (!response.ok) {
        throw new Error('図形の再生成に失敗しました');
      }

      const data = await response.json();
      if (data.success) {
        setCurrentImageBase64(data.image_base64);
        if (onUpdate) {
          onUpdate({
            imageBase64: data.image_base64,
          });
        }
        await fetchUserInfo();
      } else {
        throw new Error(data.error || '図形の再生成に失敗しました');
      }
    } catch (err) {
      console.error('Error regenerating geometry:', err);
      setError(err instanceof Error ? err.message : '図形の再生成に失敗しました');
    } finally {
      setIsLoading(false);
    }
  };

  const isComplete = () => {
    return (
      checkInfo.problem_text_ok &&
      checkInfo.solution_ok &&
      checkInfo.figure_ok &&
      checkInfo.units.length > 0 &&
      checkInfo.year !== '' &&
      checkInfo.exam_session !== ''
    );
  };

  const handleSave = async () => {
    setIsSaving(true);
    setError(null);
    try {
      await onSave(checkInfo);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'チェック情報の保存に失敗しました');
    } finally {
      setIsSaving(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl max-w-2xl w-full max-h-[90vh] overflow-auto">
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-bold text-gray-800">問題チェック - {problemTitle}</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-800 text-2xl font-bold w-8 h-8 flex items-center justify-center"
            >
              ×
            </button>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-300 text-red-700 rounded-lg">
              {error}
            </div>
          )}

          <div className="space-y-6">
            {/* 基本チェック項目 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">基本チェック</h3>
              <div className="space-y-2">
                <label className="flex items-center space-x-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={checkInfo.problem_text_ok}
                    onChange={() => handleCheckboxChange('problem_text_ok')}
                    className="w-5 h-5 text-green-500 rounded focus:ring-2 focus:ring-green-500"
                  />
                  <span className="text-gray-700">1. 問題文が適切か</span>
                </label>
                <label className="flex items-center space-x-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={checkInfo.solution_ok}
                    onChange={() => handleCheckboxChange('solution_ok')}
                    className="w-5 h-5 text-green-500 rounded focus:ring-2 focus:ring-green-500"
                  />
                  <span className="text-gray-700">2. 解答・解説が適切か</span>
                </label>
              </div>
            </div>

            {/* 図形部分（アコーディオン形式、図がない場合も表示） */}
            <div className="border border-gray-200 rounded-lg">
              <div
                className="flex items-center justify-between p-3 cursor-pointer hover:bg-gray-50"
                onClick={() => setIsFigureOpen(!isFigureOpen)}
              >
                <div className="flex items-center gap-2">
                  <svg
                    className={`w-5 h-5 transition-transform ${isFigureOpen ? 'rotate-90' : ''}`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                  <div>
                    <h3 className="text-lg font-semibold text-gray-800">3. 図が適切か</h3>
                    {userInfo && (
                      <div className="text-xs text-gray-500 mt-1">
                        図形再生成回数: {userInfo.figure_regeneration_count ?? 0}/
                        {userInfo.figure_regeneration_limit === -1 ? '無制限' : (userInfo.figure_regeneration_limit ?? 0)}
                        {isFigureRegenerationLimitReached() && (
                          <span className="text-red-600 font-bold ml-2">⚠️ 上限到達</span>
                        )}
                      </div>
                    )}
                  </div>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCheckboxChange('figure_ok');
                  }}
                  className={`px-4 py-2 rounded-lg font-semibold transition-all ${
                    checkInfo.figure_ok
                      ? 'bg-green-500 text-white'
                      : 'bg-gray-300 text-gray-600'
                  }`}
                >
                  {checkInfo.figure_ok ? '✓ チェック済み' : '未チェック'}
                </button>
              </div>
              {isFigureOpen && (
                <div className="p-3 pt-0">
                  {currentImageBase64 ? (
                    <>
                      <div className="w-80 mx-auto mb-3">
                        <img
                          src={`data:image/png;base64,${currentImageBase64}`}
                          alt="問題図形"
                          className="w-full h-auto border border-gray-200 rounded"
                        />
                      </div>
                      <div className="flex justify-center">
                        <button
                          onClick={handleRegenerateGeometry}
                          disabled={isLoading || isFigureRegenerationLimitReached()}
                          className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all ${
                            isFigureRegenerationLimitReached()
                              ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                              : 'bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed'
                          }`}
                        >
                          {isLoading ? '再生成中...' :
                           isFigureRegenerationLimitReached() ? '再生成不可' : '図形を再生成'}
                        </button>
                      </div>
                    </>
                  ) : (
                    <div className="text-center py-4">
                      <p className="text-gray-500 mb-3">図形がありません</p>
                      <button
                        onClick={handleRegenerateGeometry}
                        disabled={isLoading || isFigureRegenerationLimitReached()}
                        className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all ${
                          isFigureRegenerationLimitReached()
                            ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                            : 'bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed'
                        }`}
                      >
                        {isLoading ? '生成中...' :
                         isFigureRegenerationLimitReached() ? '生成不可' : '図形を生成'}
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>

            {/* 使用単元 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">4. 使用されている単元・公式・定理（複数選択可）</h3>
              
              {/* 選択されたタグの表示エリア */}
              {checkInfo.units.length > 0 && (
                <div className="flex flex-wrap gap-2 mb-3">
                  {checkInfo.units.map((tag) => (
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
              <div className="relative w-full mb-3">
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
              
              <div className="max-h-[400px] overflow-y-auto border border-gray-200 rounded-lg p-4 bg-gray-50">
                <HierarchicalUnitSelector
                  units={UNITS_HIERARCHY}
                  selectedUnits={checkInfo.units}
                  onSelectionChange={(newUnits) => {
                    setCheckInfo((prev) => ({
                      ...prev,
                      units: newUnits,
                    }));
                  }}
                />
              </div>
            </div>

            {/* 年度 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">5. 年度</h3>
              <div className="flex flex-wrap gap-2">
                {YEARS.map((year) => (
                  <button
                    key={year}
                    onClick={() => handleYearChange(year)}
                    className={`px-4 py-2 rounded-lg font-medium transition-all ${
                      checkInfo.year === year
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {year}年
                  </button>
                ))}
              </div>
            </div>

            {/* 回数 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">6. 回数</h3>
              <div className="flex flex-wrap gap-2">
                {EXAM_SESSIONS.map((session) => (
                  <button
                    key={session}
                    onClick={() => handleExamSessionChange(session)}
                    className={`px-4 py-2 rounded-lg font-medium transition-all ${
                      checkInfo.exam_session === session
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {session}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-3 mt-6 pt-6 border-t">
            <button
              onClick={onClose}
              disabled={isSaving}
              className="px-4 py-2 border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              キャンセル
            </button>
            <button
              onClick={handleSave}
              disabled={isSaving || !isComplete()}
              className={`px-4 py-2 rounded-lg font-semibold transition-all ${
                isComplete() && !isSaving
                  ? 'bg-green-500 text-white hover:bg-green-600'
                  : 'bg-gray-300 text-gray-500 cursor-not-allowed'
              }`}
            >
              {isSaving ? '保存中...' : 'チェック完了'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}