'use client';

import React from 'react';

interface ProblemPreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  problemId: string;
  problemTitle: string;
  problemContent?: string;
}

export default function ProblemPreviewModal({ isOpen, onClose, problemId, problemTitle, problemContent }: ProblemPreviewModalProps) {
  if (!isOpen) return null;

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
          
          <div className="border-2 border-mongene-border rounded-lg p-8 bg-white min-h-[600px]">
            {problemContent ? (
              <div className="text-mongene-ink whitespace-pre-wrap leading-relaxed">
                {problemContent}
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
          
          <div className="flex justify-end gap-3 mt-6">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-mongene-border rounded-lg text-mongene-ink hover:bg-gray-50 transition-colors"
            >
              閉じる
            </button>
            <button
              onClick={() => {
                window.print();
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
