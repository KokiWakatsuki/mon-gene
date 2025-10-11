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
                // 印刷用の新しいウィンドウを開く
                const printWindow = window.open('', '_blank');
                if (printWindow) {
                  const imageHtml = imageBase64 
                    ? `<div style="text-align: center; margin: 20px 0;">
                         <img src="data:image/png;base64,${imageBase64}" 
                              style="max-width: 100%; height: auto; border: 1px solid #ddd;" 
                              alt="問題図形" />
                       </div>`
                    : '';
                  
                  // 解答・解説がある場合は別ページに追加
                  const solutionHtml = solutionText 
                    ? `<div style="page-break-before: always;">
                         <h1>解答・解説</h1>
                         <div class="content">${solutionText}</div>
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
                      <div class="content">${problemContent || ''}</div>
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
          </div>
        </div>
      </div>
    </div>
  );
}
