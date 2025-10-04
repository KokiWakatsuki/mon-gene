'use client';

import React from 'react';

interface ProblemPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  problemId: string;
  problemTitle: string;
  problemContent?: string;
  imageBase64?: string;
  solutionText?: string;
}

export default function ProblemPreviewModal({ isOpen, onClose, problemId, problemTitle, problemContent, imageBase64, solutionText }: ProblemPreviewModalProps) {
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
            {problemContent ? (
              <div className="print-content text-mongene-ink">
                {/* 問題ページ */}
                <div className="problem-page">
                  <h3 className="text-xl font-bold mb-4">{problemTitle}</h3>
                  {imageBase64 ? (
                    <div className="flex gap-6">
                      <div className="flex-1 whitespace-pre-wrap leading-relaxed">
                        {problemContent}
                      </div>
                      <div className="w-80 flex-shrink-0">
                        <img 
                          src={`data:image/png;base64,${imageBase64}`}
                          alt="問題図形"
                          className="w-full h-auto border border-gray-200 rounded"
                        />
                      </div>
                    </div>
                  ) : (
                    <div className="whitespace-pre-wrap leading-relaxed">
                      {problemContent}
                    </div>
                  )}
                </div>
                
                {/* 解答・解説ページ（改ページ） */}
                {solutionText && (
                  <div className="solution-page" style={{ pageBreakBefore: 'always', marginTop: '40px', paddingTop: '40px', borderTop: '2px solid #e5e7eb' }}>
                    <h3 className="text-xl font-bold mb-4">解答・解説</h3>
                    <div className="whitespace-pre-wrap leading-relaxed">
                      {solutionText}
                    </div>
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
          
          <div className="flex justify-end gap-3 mt-6 no-print">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-mongene-border rounded-lg text-mongene-ink hover:bg-gray-50 transition-colors"
            >
              閉じる
            </button>
            <button
              onClick={() => {
                // 印刷用のスタイルを追加
                const printStyles = `
                  <style>
                    @media print {
                      body * { visibility: hidden; }
                      .print-content, .print-content * { visibility: visible; }
                      .print-content { position: absolute; left: 0; top: 0; width: 100%; }
                      .solution-page { page-break-before: always !important; }
                      .no-print { display: none !important; }
                    }
                  </style>
                `;
                
                // 一時的にスタイルを追加
                const styleElement = document.createElement('div');
                styleElement.innerHTML = printStyles;
                document.head.appendChild(styleElement);
                
                // 印刷実行
                window.print();
                
                // スタイルを削除
                setTimeout(() => {
                  document.head.removeChild(styleElement);
                }, 1000);
              }}
              className="px-4 py-2 bg-mongene-yellow text-mongene-ink rounded-lg font-semibold hover:brightness-95 transition-all"
            >
              印刷
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
