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
}

export default function ThreeProblemsDisplay({ problems, onPreview, onPrint }: ThreeProblemsDisplayProps) {
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

        <div className="flex gap-2">
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