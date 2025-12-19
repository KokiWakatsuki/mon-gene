'use client';

import React, { useState, useEffect } from 'react';
import { API_CONFIG } from '@/app/lib/config/api';
import MarkdownRenderer from '../../ui/MarkdownRenderer';
import { UNITS_HIERARCHY, getAllUnitsFlat, getAllChildren, getIdByLabel } from '@/app/lib/data/units';
import HierarchicalUnitSelector from './HierarchicalUnitSelector';

interface ProblemPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  problemId: string;
  problemTitle: string;
  problemContent?: string;
  imageBase64?: string;
  solutionText?: string;
  initialCheckInfo?: CheckInfo;
  onCheck?: (id: string) => void;
  onUpdate?: (updatedData: { content: string; solution: string; imageBase64?: string; checkInfo?: CheckInfo }) => void;
  onCheckSave?: (checkInfo: CheckInfo) => Promise<void>;
  onDelete?: (id: string) => void;
}

interface UserInfo {
  school_code: string;
  email: string;
  problem_generation_limit: number;
  problem_generation_count: number;
  figure_regeneration_limit: number;
  figure_regeneration_count: number;
  role: string;
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
const EXAM_SESSIONS = ['第1回', '第2回', '第3回', 'プレ', '追試'];

export default function ProblemPreviewModal({
  isOpen,
  onClose,
  problemId,
  problemTitle,
  problemContent,
  imageBase64,
  solutionText,
  initialCheckInfo,
  onCheck,
  onUpdate,
  onCheckSave,
  onDelete
}: ProblemPreviewModalProps) {
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedContent, setEditedContent] = useState('');
  const [editedSolution, setEditedSolution] = useState('');
  const [currentImageBase64, setCurrentImageBase64] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  const [showResetConfirmModal, setShowResetConfirmModal] = useState(false);
  
  // タグ検索機能の状態
  const [tagSearchInput, setTagSearchInput] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);
  
  // 利用可能なすべてのタグ（階層構造から取得）
  const allAvailableTags = getAllUnitsFlat();
  
  // チェック情報の状態
  const [checkInfo, setCheckInfo] = useState<CheckInfo>({
    problem_text_ok: false,
    solution_ok: false,
    figure_ok: false,
    units: [],
    year: '',
    exam_session: '',
  });
  
  // アコーディオンの開閉状態（デフォルトは閉じた状態）
  const [isProblemTextOpen, setIsProblemTextOpen] = useState(false);
  const [isFigureOpen, setIsFigureOpen] = useState(false);
  const [isSolutionOpen, setIsSolutionOpen] = useState(false);

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
    if (!userInfo) return true; // ユーザー情報がない場合は安全のため制限扱い
    // デモモードの場合は常に制限扱い
    if (userInfo.role === 'demo') return true;
    if (userInfo.figure_regeneration_limit === undefined || userInfo.figure_regeneration_limit === null) return true;
    if (userInfo.figure_regeneration_count === undefined || userInfo.figure_regeneration_count === null) return true;
    if (userInfo.figure_regeneration_limit === -1) return false; // 制限なし
    return userInfo.figure_regeneration_count >= userInfo.figure_regeneration_limit;
  };
  
  // デモモードかどうかを判定
  const isDemoMode = () => {
    return userInfo?.role === 'demo';
  };

  // モーダルが開かれたときにユーザー情報を取得
  useEffect(() => {
    if (isOpen) {
      fetchUserInfo();
    }
  }, [isOpen]);

  // プロパティが変更されたときに編集状態をリセット
  useEffect(() => {
    setEditedContent(problemContent || '');
    setEditedSolution(solutionText || '');
    setCurrentImageBase64(imageBase64 || '');
    setIsEditMode(false);
    setError(null);
    
    // チェック情報の初期化
    if (initialCheckInfo) {
      setCheckInfo(initialCheckInfo);
    } else {
      setCheckInfo({
        problem_text_ok: false,
        solution_ok: false,
        figure_ok: false,
        units: [],
        year: '',
        exam_session: '',
      });
    }
  }, [problemContent, solutionText, imageBase64, isOpen, initialCheckInfo]);

  // 編集モードに入る
  const handleStartEdit = () => {
    setIsEditMode(true);
    setError(null);
  };

  // 編集をキャンセル
  const handleCancelEdit = () => {
    setEditedContent(problemContent || '');
    setEditedSolution(solutionText || '');
    setCurrentImageBase64(imageBase64 || '');
    setIsEditMode(false);
    setError(null);
  };

  // 変更を保存
  const handleSaveChanges = async () => {
    if (!problemId) return;

    setIsLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません');
      }

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/update`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          id: parseInt(problemId),
          content: editedContent,
          solution: editedSolution,
        }),
      });

      if (!response.ok) {
        throw new Error('問題の更新に失敗しました');
      }

      const data = await response.json();
      if (data.success) {
        // 更新成功
        setIsEditMode(false);
        if (onUpdate) {
          onUpdate({
            content: editedContent,
            solution: editedSolution,
            imageBase64: currentImageBase64,
          });
        }
      } else {
        throw new Error(data.error || '問題の更新に失敗しました');
      }
    } catch (err) {
      console.error('Error updating problem:', err);
      setError(err instanceof Error ? err.message : '問題の更新に失敗しました');
    } finally {
      setIsLoading(false);
    }
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
    console.log('🔍 [DEBUG] Regenerating geometry for problem:', {
      problemId,
      parsedId,
      isValid: parsedId > 0
    });

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
          content: editedContent, // 編集後の問題文を送信
        }),
      });

      if (!response.ok) {
        throw new Error('図形の再生成に失敗しました');
      }

      const data = await response.json();
      if (data.success) {
        // 図形更新成功
        setCurrentImageBase64(data.image_base64);
        if (onUpdate) {
          onUpdate({
            content: editedContent,
            solution: editedSolution,
            imageBase64: data.image_base64,
          });
        }
        // ユーザー情報を再取得してカウントを更新
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

  // チェック情報の更新ハンドラー
  const handleCheckToggle = (field: 'problem_text_ok' | 'solution_ok' | 'figure_ok') => {
    setCheckInfo((prev) => ({
      ...prev,
      [field]: !prev[field],
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

  const isCheckComplete = () => {
    return (
      checkInfo.problem_text_ok &&
      checkInfo.solution_ok &&
      checkInfo.figure_ok &&
      checkInfo.units.length > 0 &&
      checkInfo.year !== '' &&
      checkInfo.exam_session !== ''
    );
  };

  // 未チェックに戻す関数（確認モーダルを表示）
  const handleResetCheck = () => {
    setShowResetConfirmModal(true);
  };

  // 未チェックに戻すことを確定して保存
  const handleConfirmReset = async () => {
    if (!problemId) return;

    setIsLoading(true);
    setError(null);
    setShowResetConfirmModal(false);

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません');
      }

      // 未チェック状態のチェック情報
      const resetCheckInfo: CheckInfo = {
        problem_text_ok: false,
        solution_ok: false,
        figure_ok: false,
        units: [],
        year: '',
        exam_session: '',
      };

      // チェック情報を未チェック状態で保存
      const checkResponse = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/update-check-info`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          id: parseInt(problemId),
          check_info: resetCheckInfo,
        }),
      });

      if (!checkResponse.ok) {
        const errorData = await checkResponse.json();
        throw new Error(errorData.error || 'チェック情報のリセットに失敗しました');
      }

      // ローカル状態を更新
      setCheckInfo(resetCheckInfo);

      // 親コンポーネントの状態を更新
      if (onUpdate) {
        onUpdate({
          content: editedContent,
          solution: editedSolution,
          imageBase64: currentImageBase64,
          checkInfo: resetCheckInfo,
        });
      }

      // 編集モードを終了してモーダルを閉じる
      setIsEditMode(false);
      onClose();
    } catch (err) {
      console.error('Error resetting check info:', err);
      setError(err instanceof Error ? err.message : 'チェック情報のリセットに失敗しました');
    } finally {
      setIsLoading(false);
    }
  };

  // リセット確認をキャンセル
  const handleCancelReset = () => {
    setShowResetConfirmModal(false);
  };

  const handleSaveAll = async () => {
    if (!problemId) return;

    setIsLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません');
      }

      // 問題内容の更新
      const updateResponse = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/update`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          id: parseInt(problemId),
          content: editedContent,
          solution: editedSolution,
        }),
      });

      if (!updateResponse.ok) {
        throw new Error('問題の更新に失敗しました');
      }

      // チェック情報の保存
      const checkResponse = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/update-check-info`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          id: parseInt(problemId),
          check_info: checkInfo,
        }),
      });

      if (!checkResponse.ok) {
        const errorData = await checkResponse.json();
        throw new Error(errorData.error || 'チェック情報の保存に失敗しました');
      }

      // 更新成功
      setIsEditMode(false);
      if (onUpdate) {
        onUpdate({
          content: editedContent,
          solution: editedSolution,
          imageBase64: currentImageBase64,
          checkInfo: checkInfo, // チェック情報を追加
        });
      }
    } catch (err) {
      console.error('Error saving:', err);
      setError(err instanceof Error ? err.message : '保存に失敗しました');
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  // デバッグログを追加
  console.log('🔍 [ProblemPreviewModal] Props received:');
  console.log('  problemId:', problemId);
  console.log('  problemTitle:', problemTitle);
  console.log('  problemContent length:', problemContent?.length || 0);
  console.log('  imageBase64 exists:', !!imageBase64);
  console.log('  solutionText exists:', !!solutionText);
  console.log('  solutionText length:', solutionText?.length || 0);
  console.log('  solutionText preview:', solutionText?.substring(0, 100) || 'No solution');

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-xl max-w-4xl w-full max-h-[90vh] overflow-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-bold text-mongene-ink">問題プレビュー - {problemTitle}</h2>
          </div>
          
          <div className="border-2 border-mongene-border rounded-lg p-8 bg-white min-h-[600px] max-h-[70vh] overflow-y-auto">
            {error && (
              <div className="mb-4 p-3 bg-red-100 border border-red-300 text-red-700 rounded-lg">
                {error}
              </div>
            )}

            {problemContent ? (
              <div className="text-mongene-ink">
                {isEditMode ? (
                  /* 編集・チェックモード */
                  <div className="space-y-6">
                    {/* 問題文編集 */}
                    <div className="border border-gray-200 rounded-lg">
                      <div
                        className="flex items-center justify-between p-3 cursor-pointer hover:bg-gray-50"
                        onClick={() => setIsProblemTextOpen(!isProblemTextOpen)}
                      >
                        <div className="flex items-center gap-2">
                          <svg
                            className={`w-5 h-5 transition-transform ${isProblemTextOpen ? 'rotate-90' : ''}`}
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                          </svg>
                          <h3 className="text-lg font-semibold text-mongene-ink">問題文</h3>
                        </div>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCheckToggle('problem_text_ok');
                          }}
                          className={`px-4 py-2 rounded-lg font-semibold transition-all ${
                            checkInfo.problem_text_ok
                              ? 'bg-green-500 text-white'
                              : 'bg-gray-300 text-gray-600'
                          }`}
                        >
                          {checkInfo.problem_text_ok ? '✓ チェック済み' : '未チェック'}
                        </button>
                      </div>
                      {isProblemTextOpen && (
                        <div className="p-3 pt-0">
                          <textarea
                            value={editedContent}
                            onChange={(e) => setEditedContent(e.target.value)}
                            className="w-full p-3 border border-mongene-border rounded-lg resize-vertical focus:outline-none focus:ring-2 focus:ring-mongene-yellow"
                            placeholder="問題文を入力してください..."
                            rows={Math.max(10, (editedContent.match(/\n/g) || []).length + 3)}
                          />
                        </div>
                      )}
                    </div>

                    {/* 図形部分（図がない場合も表示） */}
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
                            <h3 className="text-lg font-semibold text-mongene-ink">図形</h3>
                            {userInfo && (
                              <div className="text-xs text-mongene-muted mt-1">
                                {isDemoMode() ? (
                                  <span className="text-orange-600 font-bold">デモモードでは図形の再生成はサポートされていません</span>
                                ) : (
                                  <>
                                    図形再生成回数: {userInfo.figure_regeneration_count ?? 0}/
                                    {userInfo.figure_regeneration_limit === -1 ? '無制限' : (userInfo.figure_regeneration_limit ?? 0)}
                                    {isFigureRegenerationLimitReached() && (
                                      <span className="text-red-600 font-bold ml-2">⚠️ 上限到達</span>
                                    )}
                                  </>
                                )}
                              </div>
                            )}
                          </div>
                        </div>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCheckToggle('figure_ok');
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
                                  title={isDemoMode() ? 'デモモードでは図形の再生成はサポートされていません' : ''}
                                >
                                  {isLoading ? '再生成中...' :
                                   isDemoMode() ? 'デモモードでは利用不可' :
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
                                title={isDemoMode() ? 'デモモードでは図形の生成はサポートされていません' : ''}
                              >
                                {isLoading ? '生成中...' :
                                 isDemoMode() ? 'デモモードでは利用不可' :
                                 isFigureRegenerationLimitReached() ? '生成不可' : '図形を生成'}
                              </button>
                            </div>
                          )}
                        </div>
                      )}
                    </div>

                    {/* 解答・解説編集 */}
                    <div className="border border-gray-200 rounded-lg">
                      <div
                        className="flex items-center justify-between p-3 cursor-pointer hover:bg-gray-50"
                        onClick={() => setIsSolutionOpen(!isSolutionOpen)}
                      >
                        <div className="flex items-center gap-2">
                          <svg
                            className={`w-5 h-5 transition-transform ${isSolutionOpen ? 'rotate-90' : ''}`}
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                          </svg>
                          <h3 className="text-lg font-semibold text-mongene-ink">解答・解説</h3>
                        </div>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCheckToggle('solution_ok');
                          }}
                          className={`px-4 py-2 rounded-lg font-semibold transition-all ${
                            checkInfo.solution_ok
                              ? 'bg-green-500 text-white'
                              : 'bg-gray-300 text-gray-600'
                          }`}
                        >
                          {checkInfo.solution_ok ? '✓ チェック済み' : '未チェック'}
                        </button>
                      </div>
                      {isSolutionOpen && (
                        <div className="p-3 pt-0">
                          <textarea
                            value={editedSolution}
                            onChange={(e) => setEditedSolution(e.target.value)}
                            className="w-full p-3 border border-mongene-border rounded-lg resize-vertical focus:outline-none focus:ring-2 focus:ring-mongene-yellow"
                            placeholder="解答・解説を入力してください..."
                            rows={Math.max(10, (editedSolution.match(/\n/g) || []).length + 3)}
                          />
                        </div>
                      )}
                    </div>

                    {/* 使用単元 */}
                    <div>
                      <h3 className="text-lg font-semibold mb-3 text-mongene-ink">4. 使用されている単元・公式・定理（複数選択可）</h3>
                      
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
                      <h3 className="text-lg font-semibold mb-3 text-mongene-ink">5. 年度</h3>
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
                      <h3 className="text-lg font-semibold mb-3 text-mongene-ink">6. 回数</h3>
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
                ) : (
                  /* 表示モード */
                  <div className="print-content">
                    {/* 問題ページ */}
                    <div className="problem-page">
                      <h3 className="text-xl font-bold mb-4">{problemTitle}</h3>
                      {currentImageBase64 ? (
                        <div className="flex gap-6">
                          <div className="flex-1">
                            <MarkdownRenderer 
                              content={editedContent || problemContent || ''} 
                              className="leading-relaxed"
                            />
                          </div>
                          <div className="w-80 flex-shrink-0">
                            <img 
                              src={`data:image/png;base64,${currentImageBase64}`}
                              alt="問題図形"
                              className="w-full h-auto border border-gray-200 rounded"
                            />
                          </div>
                        </div>
                      ) : (
                        <MarkdownRenderer 
                          content={editedContent || problemContent || ''} 
                          className="leading-relaxed"
                        />
                      )}
                    </div>
                    
                    {/* 解答・解説表示 */}
                    {(editedSolution || solutionText) && (
                      <div className="solution-page" style={{ pageBreakBefore: 'always', marginTop: '40px', paddingTop: '40px', borderTop: '2px solid #e5e7eb' }}>
                        <h3 className="text-xl font-bold mb-4">解答・解説</h3>
                        <MarkdownRenderer 
                          content={editedSolution || solutionText || ''} 
                          className="leading-relaxed"
                        />
                      </div>
                    )}
                  </div>
                )}
              </div>
            ) : (
              <div className="flex items-center justify-center h-full">
                <div className="text-center text-mongene-muted">
                  <div className="text-lg mb-2">問題ID: {problemId}</div>
                  <div className="text-sm">問題内容が見つかりません</div>
                </div>
              </div>
            )}
          </div>
          
          <div className="flex justify-between items-center mt-6 no-print">
            {/* 左側のボタン */}
            <div className="flex gap-3">
              {isEditMode ? (
                <>
                  <button
                    onClick={handleCancelEdit}
                    disabled={isLoading}
                    className="px-4 py-2 border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    キャンセル
                  </button>
                  <button
                    onClick={handleSaveAll}
                    disabled={isLoading || !isCheckComplete()}
                    className={`px-4 py-2 rounded-lg font-semibold transition-all ${
                      isCheckComplete() && !isLoading
                        ? 'bg-green-500 text-white hover:bg-green-600'
                        : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                    }`}
                  >
                    {isLoading ? '保存中...' : 'チェック完了'}
                  </button>
                  <button
                    onClick={handleResetCheck}
                    className="px-4 py-2 bg-gray-500 text-white rounded-lg font-semibold hover:bg-gray-600 transition-all"
                  >
                    未チェックに戻す
                  </button>
                </>
              ) : (
                <>
                  <button
                    onClick={handleStartEdit}
                    className="px-4 py-2 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 transition-all"
                  >
                    編集・チェック
                  </button>
                  <button
                    onClick={() => {
                    // MarkdownRendererと同じ処理を使用してHTMLを生成
                    const renderLatexToHtml = (latex: string): string => {
                      return latex
                        .replace(/\\frac\{([^}]+)\}\{([^}]+)\}/g, '<span class="math-fraction-block"><span class="numerator">$1</span><span class="fraction-line-block">/</span><span class="denominator">$2</span></span>')
                        .replace(/\\sqrt\{([^}]+)\}/g, '<span class="math-symbol">√<span class="sqrt-content">$1</span></span>')
                        .replace(/\\vec\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
                        .replace(/\\overrightarrow\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
                        .replace(/\\times/g, '×')
                        .replace(/\\cdot/g, '·')
                        .replace(/\\pi/g, 'π')
                        .replace(/\\infty/g, '∞')
                        .replace(/\\pm/g, '±')
                        .replace(/\\leq/g, '≤')
                        .replace(/\\geq/g, '≥')
                        .replace(/\\neq/g, '≠')
                        .replace(/\\approx/g, '≈')
                        .replace(/\\rightarrow/g, '→')
                        .replace(/\\leftarrow/g, '←');
                    };

                    const renderAdvancedMathSymbols = (text: string): string => {
                      if (!text) return '';
                      
                      const mathPlaceholders: string[] = [];
                      let placeholderIndex = 0;
                      
                      let processedText = text
                        .replace(/\$\$([\s\S]*?)\$\$/g, (match, formula) => {
                          const placeholder = `__MATH_BLOCK_${placeholderIndex}__`;
                          mathPlaceholders[placeholderIndex] = `<div class="math-block">${renderLatexToHtml(formula.trim())}</div>`;
                          placeholderIndex++;
                          return placeholder;
                        })
                        .replace(/\$([^$\n]+)\$/g, (match, formula) => {
                          const placeholder = `__MATH_INLINE_${placeholderIndex}__`;
                          mathPlaceholders[placeholderIndex] = `<span class="math-inline">${renderLatexToHtml(formula.trim())}</span>`;
                          placeholderIndex++;
                          return placeholder;
                        });
                      
                      processedText = processedText
                        .replace(/\\overrightarrow\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
                        .replace(/([A-Z]{1,3})⃗/g, '<span class="math-vector">$1→</span>')
                        .replace(/√(\d+)/g, '<span class="math-symbol">√$1</span>')
                        .replace(/√\(([^)]+)\)/g, '<span class="math-symbol">√($1)</span>')
                        .replace(/√([a-zA-Z]+)/g, '<span class="math-symbol">√$1</span>')
                        .replace(/(\w+)²/g, '$1<sup>2</sup>')
                        .replace(/(\w+)³/g, '$1<sup>3</sup>')
                        .replace(/(\w+)⁴/g, '$1<sup>4</sup>')
                        .replace(/(\w+)⁵/g, '$1<sup>5</sup>')
                        .replace(/∠([A-Z]+)/g, '<span class="math-symbol">∠$1</span>')
                        .replace(/(\d+)\/(\d+)/g, '<span class="math-fraction"><sup>$1</sup>/<sub>$2</sub></span>')
                        .replace(/×/g, '<span class="math-symbol">×</span>')
                        .replace(/÷/g, '<span class="math-symbol">÷</span>')
                        .replace(/°/g, '<span class="math-symbol">°</span>')
                        .replace(/π/g, '<span class="math-symbol">π</span>')
                        .replace(/∞/g, '<span class="math-symbol">∞</span>')
                        .replace(/±/g, '<span class="math-symbol">±</span>')
                        .replace(/≤/g, '<span class="math-symbol">≤</span>')
                        .replace(/≥/g, '<span class="math-symbol">≥</span>')
                        .replace(/≠/g, '<span class="math-symbol">≠</span>')
                        .replace(/≈/g, '<span class="math-symbol">≈</span>')
                        .replace(/≅/g, '<span class="math-symbol">≅</span>')
                        .replace(/∽/g, '<span class="math-symbol">∽</span>')
                        .replace(/→/g, '<span class="math-symbol">→</span>')
                        .replace(/←/g, '<span class="math-symbol">←</span>');
                      
                      mathPlaceholders.forEach((replacement, index) => {
                        processedText = processedText.replace(`__MATH_BLOCK_${index}__`, replacement);
                        processedText = processedText.replace(`__MATH_INLINE_${index}__`, replacement);
                      });
                      
                      return processedText;
                    };

                    const renderBasicMarkdown = (text: string): string => {
                      return text
                        .replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold mb-2 text-gray-700">$1</h3>')
                        .replace(/^## (.*$)/gim, '<h2 class="text-xl font-bold mb-3 text-gray-800">$1</h2>')
                        .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mb-4 text-gray-900">$1</h1>')
                        .replace(/\*\*(.*?)\*\*/g, '<strong class="font-bold text-gray-900">$1</strong>')
                        .replace(/\*(.*?)\*/g, '<em class="italic">$1</em>')
                        .replace(/`([^`]+)`/g, '<code class="bg-gray-100 px-1 py-0.5 rounded text-sm font-mono">$1</code>')
                        .replace(/\n/g, '<br />');
                    };

                    const processContent = (content: string): string => {
                      return renderBasicMarkdown(renderAdvancedMathSymbols(content));
                    };

                    // 印刷用の新しいウィンドウを開く
                    const printWindow = window.open('', '_blank');
                    if (printWindow) {
                      const rawContent = editedContent || problemContent || '';
                      const rawSolution = editedSolution || solutionText || '';
                      
                      console.log('🔍 [Print] Raw content:', rawContent.substring(0, 200));
                      console.log('🔍 [Print] Raw solution:', rawSolution.substring(0, 200));
                      
                      const processedContent = processContent(rawContent);
                      const processedSolution = rawSolution ? processContent(rawSolution) : '';
                      
                      console.log('🔍 [Print] Processed content:', processedContent.substring(0, 200));
                      console.log('🔍 [Print] Processed solution:', processedSolution.substring(0, 200));
                      
                      const imageHtml = (currentImageBase64 || imageBase64)
                        ? `<div class="image-container">
                             <img src="data:image/png;base64,${currentImageBase64 || imageBase64}"
                                  alt="問題図形" />
                           </div>`
                        : '';
                      
                      const solutionHtml = processedSolution
                        ? `<div style="page-break-before: always;">
                             <h1>解答・解説</h1>
                             <div class="content">${processedSolution}</div>
                           </div>`
                        : '';
                      
                      printWindow.document.write(`
                        <!DOCTYPE html>
                        <html>
                        <head>
                          <title>${problemTitle}</title>
                          <style>
                            body {
                              font-family: 'Times New Roman', Arial, sans-serif;
                              margin: 20px;
                              line-height: 1.6;
                            }
                            h1 {
                              font-size: 24px;
                              margin-bottom: 20px;
                              border-bottom: 2px solid #333;
                              padding-bottom: 10px;
                            }
                            .content {
                              font-size: 14px;
                              margin-bottom: 20px;
                            }
                            .math-symbol {
                              font-family: 'Times New Roman', serif;
                              font-weight: normal;
                              color: #1f2937;
                              font-size: 1.1em;
                            }
                            .math-vector {
                              font-family: 'Times New Roman', serif;
                              font-weight: bold;
                              color: #1f2937;
                              font-size: 1.05em;
                            }
                            .math-fraction {
                              display: inline-block;
                              vertical-align: middle;
                              font-family: 'Times New Roman', serif;
                              margin: 0 2px;
                            }
                            .fraction-line {
                              font-size: 1.2em;
                              color: #374151;
                            }
                            .math-fraction-block {
                              display: inline-flex;
                              flex-direction: column;
                              vertical-align: middle;
                              text-align: center;
                              font-family: 'Times New Roman', serif;
                              margin: 0 4px;
                              align-items: center;
                            }
                            .numerator {
                              display: block;
                              font-size: 0.9em;
                              line-height: 1;
                              padding: 0 2px;
                            }
                            .fraction-line-block {
                              display: block;
                              border-top: 1.5px solid #374151;
                              margin: 1px 0;
                              width: 100%;
                              min-width: 20px;
                              height: 0;
                              line-height: 0;
                            }
                            .fraction-line-block::before {
                              content: '';
                              display: block;
                            }
                            .denominator {
                              display: block;
                              font-size: 0.9em;
                              line-height: 1;
                              padding: 0 2px;
                            }
                            .sqrt-content {
                              border-top: 1px solid #374151;
                              padding: 0 2px;
                            }
                            .math-block {
                              display: block;
                              text-align: center;
                              margin: 12px 0;
                              padding: 8px;
                              background-color: #f9fafb;
                              border: 1px solid #e5e7eb;
                              border-radius: 4px;
                              font-family: 'Times New Roman', serif;
                              font-size: 1.1em;
                            }
                            .math-inline {
                              font-family: 'Times New Roman', serif;
                              font-size: 1.05em;
                              color: #1f2937;
                            }
                            sup {
                              font-size: 0.75em;
                              vertical-align: super;
                              line-height: 0;
                            }
                            sub {
                              font-size: 0.75em;
                              vertical-align: sub;
                              line-height: 0;
                            }
                            .problem-layout {
                              display: flex;
                              gap: 20px;
                              align-items: flex-start;
                            }
                            .image-container {
                              flex: 0 0 50%;
                              max-width: 50%;
                            }
                            .image-container img {
                              width: 100%;
                              height: auto;
                              border: 1px solid #ddd;
                            }
                            .content-container {
                              flex: 1;
                            }
                            @media print {
                              body { margin: 0; }
                              h1 { page-break-after: avoid; }
                              .image-container { page-break-inside: avoid; }
                            }
                          </style>
                        </head>
                        <body>
                          <h1>${problemTitle}</h1>
                          ${(currentImageBase64 || imageBase64)
                            ? `<div class="problem-layout">
                                 <div class="content-container">
                                   <div class="content">${processedContent}</div>
                                 </div>
                                 ${imageHtml}
                               </div>`
                            : `<div class="content">${processedContent}</div>`
                          }
                          ${solutionHtml}
                        </body>
                        </html>
                      `);
                      printWindow.document.close();
                      
                      // ページが読み込まれたら印刷ダイアログを表示
                      printWindow.onload = () => {
                        printWindow.print();
                        printWindow.close();
                      };
                    }
                    }}
                    className="px-4 py-2 bg-mongene-yellow text-mongene-ink rounded-lg font-semibold hover:brightness-95 transition-all"
                  >
                    印刷
                  </button>
                </>
              )}
            </div>

            {/* 右側のボタン */}
            <div>
              <button
                onClick={() => {
                  if (onDelete && confirm('この問題を削除してもよろしいですか？')) {
                    onDelete(problemId);
                    onClose();
                  }
                }}
                className="px-4 py-2 bg-red-500 text-white rounded-lg font-semibold hover:bg-red-600 transition-all flex items-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                削除
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* 未チェックに戻す確認モーダル */}
      {showResetConfirmModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[60]">
          <div className="bg-white rounded-xl p-6 max-w-md w-full mx-4 shadow-2xl">
            <h3 className="text-xl font-bold text-gray-800 mb-4">確認</h3>
            <p className="text-gray-600 mb-6">
              本当に未チェックに戻しますか？<br />
              すべてのチェック情報がリセットされ、保存されます。
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={handleCancelReset}
                className="px-6 py-2 border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50 transition-colors"
              >
                いいえ
              </button>
              <button
                onClick={handleConfirmReset}
                disabled={isLoading}
                className="px-6 py-2 bg-red-500 text-white rounded-lg font-semibold hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'リセット中...' : 'はい'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
