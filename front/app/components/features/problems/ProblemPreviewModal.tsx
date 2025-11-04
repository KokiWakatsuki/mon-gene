'use client';

import React, { useState, useEffect } from 'react';
import { API_CONFIG } from '@/app/lib/config/api';
import MarkdownRenderer from '../../ui/MarkdownRenderer';

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

interface UserInfo {
  school_code: string;
  email: string;
  problem_generation_limit: number;
  problem_generation_count: number;
  figure_regeneration_limit: number;
  figure_regeneration_count: number;
}

export default function ProblemPreviewModal({ 
  isOpen, 
  onClose, 
  problemId, 
  problemTitle, 
  problemContent, 
  imageBase64, 
  solutionText, 
  onUpdate 
}: ProblemPreviewModalProps) {
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedContent, setEditedContent] = useState('');
  const [editedSolution, setEditedSolution] = useState('');
  const [currentImageBase64, setCurrentImageBase64] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);

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
    if (userInfo.figure_regeneration_limit === undefined || userInfo.figure_regeneration_limit === null) return true;
    if (userInfo.figure_regeneration_count === undefined || userInfo.figure_regeneration_count === null) return true;
    if (userInfo.figure_regeneration_limit === -1) return false; // 制限なし
    return userInfo.figure_regeneration_count >= userInfo.figure_regeneration_limit;
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
                          <div>
                            <h3 className="text-lg font-semibold text-mongene-ink">図形</h3>
                            {userInfo && (
                              <div className="text-xs text-mongene-muted mt-1">
                                図形再生成回数: {userInfo.figure_regeneration_count ?? 0}/
                                {userInfo.figure_regeneration_limit === -1 ? '無制限' : (userInfo.figure_regeneration_limit ?? 0)}
                                {isFigureRegenerationLimitReached() && (
                                  <span className="text-red-600 font-bold ml-2">⚠️ 上限到達</span>
                                )}
                              </div>
                            )}
                          </div>
                          <div className="flex flex-col items-end gap-1">
                            {isFigureRegenerationLimitReached() && (
                              <div className="text-xs text-red-600 font-bold">
                                再生成上限に達しました
                              </div>
                            )}
                            <button
                              onClick={handleRegenerateGeometry}
                              disabled={isLoading || isFigureRegenerationLimitReached()}
                              className={`px-3 py-1 rounded-lg text-sm transition-all ${
                                isFigureRegenerationLimitReached()
                                  ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                                  : 'bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed'
                              }`}
                            >
                              {isLoading ? '再生成中...' : 
                               isFigureRegenerationLimitReached() ? '再生成不可' : '図形を再生成'}
                            </button>
                          </div>
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
                        ? `<div style="text-align: center; margin: 20px 0;">
                             <img src="data:image/png;base64,${currentImageBase64 || imageBase64}"
                                  style="max-width: 100%; height: auto; border: 1px solid #ddd;"
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
                          <div class="content">${processedContent}</div>
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
