'use client';

import React from 'react';

interface ThreeProblemsDisplayProps {
  problems: {
    patternA?: { id: string; title: string; content: string; imageBase64?: string; solution?: string };
    patternB?: { id: string; title: string; content: string; imageBase64?: string; solution?: string };
    patternC?: { id: string; title: string; content: string; imageBase64?: string; solution?: string };
  };
  onPreview: (id: string) => void;
  onPrint: (id: string) => void;
  onDelete: (id: string) => void;
}

export default function ThreeProblemsDisplay({ problems, onPreview, onPrint, onDelete }: ThreeProblemsDisplayProps) {
  if (!problems.patternA && !problems.patternB && !problems.patternC) {
    return null;
  }

  const renderProblemCard = (
    problem: { id: string; title: string; content: string; imageBase64?: string; solution?: string } | undefined,
    patternLabel: string,
    patternDescription: string,
    colorClass: string
  ) => {
    if (!problem) return null;

    return (
      <div className={`bg-white border-2 ${colorClass} rounded-xl shadow-sm p-6`}>
        <div className="mb-4">
          <div className={`inline-block px-3 py-1 ${colorClass.replace('border', 'bg').replace('500', '100')} ${colorClass.replace('border', 'text')} rounded-full text-sm font-semibold mb-2`}>
            {patternLabel}
          </div>
          <h3 className="text-lg font-bold text-gray-800">{problem.title}</h3>
          <p className="text-sm text-gray-600 mt-1">{patternDescription}</p>
        </div>

        <div className="mb-4 p-4 bg-gray-50 rounded-lg max-h-40 overflow-y-auto">
          <p className="text-sm text-gray-700 line-clamp-6">
            {problem.content.substring(0, 200)}
            {problem.content.length > 200 ? '...' : ''}
          </p>
        </div>

        {problem.imageBase64 && (
          <div className="mb-4">
            <img
              src={`data:image/png;base64,${problem.imageBase64}`}
              alt="問題図形"
              className="w-full h-auto rounded-lg border border-gray-200"
            />
          </div>
        )}

        <div className="flex justify-between items-center gap-2">
          <div className="flex gap-2 flex-1">
            <button
              onClick={() => onPreview(problem.id)}
              className="flex-1 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors text-sm font-medium"
            >
              📄 プレビュー
            </button>
            <button
              onClick={() => onPrint(problem.id)}
              className="flex-1 px-4 py-2 bg-gray-500 text-white rounded-lg hover:bg-gray-600 transition-colors text-sm font-medium"
            >
              🖨️ 印刷
            </button>
          </div>
          <button
            onClick={() => onDelete(problem.id)}
            className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors text-sm font-medium flex items-center gap-1"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              <line x1="10" y1="11" x2="10" y2="17"></line>
              <line x1="14" y1="11" x2="14" y2="17"></line>
            </svg>
            削除
          </button>
        </div>
      </div>
    );
  };

  return (
    <div className="mt-8 mb-8">
      <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-xl">
        <h3 className="text-lg font-semibold text-green-800 mb-2">✅ 3問の生成が完了しました</h3>
        <p className="text-sm text-green-700">
          アップロードされた問題を基に、3つの類似パターンを生成しました。
          各問題をプレビューして内容を確認できます。
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {renderProblemCard(
          problems.patternA,
          'パターンA',
          '数値のみ変更',
          'border-blue-500'
        )}
        {renderProblemCard(
          problems.patternB,
          'パターンB',
          '文脈・設定変更',
          'border-purple-500'
        )}
        {renderProblemCard(
          problems.patternC,
          'パターンC',
          '構造・問われる部分変更',
          'border-orange-500'
        )}
      </div>
    </div>
  );
}