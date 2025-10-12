'use client';

import React, { useState, useEffect } from 'react';
import { API_CONFIG } from '@/app/lib/config/api';

interface ProblemPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  problemId: string;
  problemTitle: string;
  problemContent?: string;
  imageBase64?: string;
  solutionText?: string;
  onUpdate?: (updatedData: { content: string; solution: string; imageBase64?: string }) => void;
}

export default function ProblemPreviewModal({ isOpen, onClose, problemId, problemTitle, problemContent, imageBase64, solutionText, onUpdate }: ProblemPreviewModalProps) {
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedContent, setEditedContent] = useState('');
  const [editedSolution, setEditedSolution] = useState('');
  const [currentImageBase64, setCurrentImageBase64] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // プロパティが変更されたときに編集状態をリセット
  useEffect(() => {
    setEditedContent(problemContent || '');
    setEditedSolution(solutionText || '');
    setCurrentImageBase64(imageBase64 || '');
    setIsEditMode(false);
    setError(null);
  }, [problemContent, solutionText, imageBase64, isOpen]);

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
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl max-w-4xl w-full max-h-[90vh] overflow-auto">
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-bold text-mongene-ink">問題プレビュー - {problemTitle}</h2>
            <button
              onClick={onClose}
              className="text-mongene-muted hover:text-mongene-ink text-2xl font-bold w-8 h-8 flex items-center justify-center"
            >
              ×
            </button>
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
                  /* 編集モード */
                  <div className="space-y-6">
                    {/* 問題文編集 */}
                    <div>
                      <h3 className="text-lg font-semibold mb-3 text-mongene-ink">問題文</h3>
                      <textarea
                        value={editedContent}
                        onChange={(e) => setEditedContent(e.target.value)}
                        className="w-full h-40 p-3 border border-mongene-border rounded-lg resize-vertical focus:outline-none focus:ring-2 focus:ring-mongene-yellow"
                        placeholder="問題文を入力してください..."
                      />
                    </div>

                    {/* 図形部分 */}
                    {currentImageBase64 && (
                      <div>
                        <div className="flex items-center justify-between mb-3">
                          <h3 className="text-lg font-semibold text-mongene-ink">図形</h3>
                          <button
                            onClick={handleRegenerateGeometry}
                            disabled={isLoading}
                            className="px-3 py-1 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                          >
                            {isLoading ? '再生成中...' : '図形を再生成'}
                          </button>
                        </div>
                        <div className="w-80 mx-auto">
                          <img 
                            src={`data:image/png;base64,${currentImageBase64}`}
                            alt="問題図形"
                            className="w-full h-auto border border-gray-200 rounded"
                          />
                        </div>
                      </div>
                    )}

                    {/* 解答・解説編集 */}
                    <div>
                      <h3 className="text-lg font-semibold mb-3 text-mongene-ink">解答・解説</h3>
                      <textarea
                        value={editedSolution}
                        onChange={(e) => setEditedSolution(e.target.value)}
                        className="w-full h-40 p-3 border border-mongene-border rounded-lg resize-vertical focus:outline-none focus:ring-2 focus:ring-mongene-yellow"
                        placeholder="解答・解説を入力してください..."
                      />
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
                          <div className="flex-1 whitespace-pre-wrap leading-relaxed">
                            {editedContent || problemContent}
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
                        <div className="whitespace-pre-wrap leading-relaxed">
                          {editedContent || problemContent}
                        </div>
                      )}
                    </div>
                    
                    {/* 解答・解説ページ（改ページ） */}
                    {(editedSolution || solutionText) && (
                      <div className="solution-page" style={{ pageBreakBefore: 'always', marginTop: '40px', paddingTop: '40px', borderTop: '2px solid #e5e7eb' }}>
                        <h3 className="text-xl font-bold mb-4">解答・解説</h3>
                        <div className="whitespace-pre-wrap leading-relaxed">
                          {editedSolution || solutionText}
                        </div>
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
            <div>
              {isEditMode ? (
                <div className="flex gap-3">
                  <button
                    onClick={handleCancelEdit}
                    disabled={isLoading}
                    className="px-4 py-2 border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    キャンセル
                  </button>
                  <button
                    onClick={handleSaveChanges}
                    disabled={isLoading}
                    className="px-4 py-2 bg-green-500 text-white rounded-lg font-semibold hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                  >
                    {isLoading ? '保存中...' : '変更を保存'}
                  </button>
                </div>
              ) : (
                <button
                  onClick={handleStartEdit}
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 transition-all"
                >
                  編集
                </button>
              )}
            </div>

            {/* 右側のボタン */}
            <div className="flex gap-3">
              <button
                onClick={onClose}
                className="px-4 py-2 border border-mongene-border rounded-lg text-mongene-ink hover:bg-gray-50 transition-colors"
              >
                閉じる
              </button>
              {!isEditMode && (
                <button
                  onClick={() => {
                    // 印刷用の新しいウィンドウを開く
                    const printWindow = window.open('', '_blank');
                    if (printWindow) {
                      const imageHtml = (currentImageBase64 || imageBase64)
                        ? `<div style="text-align: center; margin: 20px 0;">
                             <img src="data:image/png;base64,${currentImageBase64 || imageBase64}" 
                                  style="max-width: 100%; height: auto; border: 1px solid #ddd;" 
                                  alt="問題図形" />
                           </div>`
                        : '';
                      
                      // 解答・解説がある場合は別ページに追加
                      const solutionHtml = (editedSolution || solutionText)
                        ? `<div style="page-break-before: always;">
                             <h1>解答・解説</h1>
                             <div class="content">${editedSolution || solutionText}</div>
                           </div>`
                        : '';
                      
                      printWindow.document.write(`
                        <!DOCTYPE html>
                        <html>
                        <head>
                          <title>${problemTitle}</title>
                          <style>
                            body {
                              font-family: Arial, sans-serif;
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
                              white-space: pre-wrap;
                              font-size: 14px;
                              margin-bottom: 20px;
                            }
                            .image-container {
                              text-align: center;
                              margin: 20px 0;
                            }
                            .image-container img {
                              max-width: 100%;
                              height: auto;
                              border: 1px solid #ddd;
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
                          <div class="content">${editedContent || problemContent || ''}</div>
                          ${imageHtml}
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
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
