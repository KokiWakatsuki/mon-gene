'use client';

import React from 'react';

interface ProblemCardProps {
  id: string;
  title: string;
  content?: string;
  imageBase64?: string;
  onPreview: (id: string) => void;
  onPrint: (id: string) => void;
}

export default function ProblemCard({ id, title, content, imageBase64, onPreview, onPrint }: ProblemCardProps) {
  // 問題内容の最初の部分を取得（プレビュー用）
  const getPreviewContent = () => {
    if (!content) return title;
    
    // 改行で分割して最初の数行を取得
    const lines = content.split('\n').filter(line => line.trim() !== '');
    const previewLines = lines.slice(0, 8); // 最初の8行を表示
    let preview = previewLines.join('\n');
    
    // 文字数制限（約200文字）
    if (preview.length > 200) {
      preview = preview.substring(0, 200) + '...';
    } else if (lines.length > 8) {
      preview += '\n...';
    }
    
    return preview;
  };

  return (
    <article className="bg-white border border-mongene-border rounded-xl p-4 shadow-sm flex flex-col gap-4">
      <div className="w-full h-96 bg-white border-2 border-gray-300 rounded p-4 mx-auto shadow-sm overflow-hidden">
        <div className="h-full flex flex-col">
          <div className="text-sm font-semibold text-mongene-ink mb-2 border-b border-gray-200 pb-2">
            {title}
          </div>
          <div className="flex-1 overflow-hidden">
            {imageBase64 ? (
              <div className="h-full flex gap-3">
                <div className="flex-1 overflow-hidden">
                  <div className="text-xs text-mongene-ink leading-relaxed whitespace-pre-wrap">
                    {getPreviewContent()}
                  </div>
                </div>
                <div className="w-40 flex-shrink-0">
                  <img 
                    src={`data:image/png;base64,${imageBase64}`}
                    alt="問題図形"
                    className="w-full h-full object-contain border border-gray-200 rounded"
                  />
                </div>
              </div>
            ) : (
              <div className="text-xs text-mongene-ink leading-relaxed whitespace-pre-wrap">
                {getPreviewContent()}
              </div>
            )}
          </div>
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <button
          className="appearance-none border-0 rounded-lg px-3.5 py-2.5 font-bold cursor-pointer bg-mongene-blue text-white hover:brightness-95 focus:outline-none focus:ring-3 focus:ring-mongene-blue/25 focus:ring-offset-2"
          type="button"
          onClick={() => onPreview(id)}
        >
          プレビュー
        </button>
        <button
          className="appearance-none border-0 rounded-lg px-3.5 py-2.5 font-bold cursor-pointer bg-mongene-yellow text-mongene-ink hover:brightness-95 focus:outline-none focus:ring-3 focus:ring-mongene-yellow/25 focus:ring-offset-2"
          type="button"
          onClick={() => onPrint(id)}
        >
          印刷
        </button>
      </div>
    </article>
  );
}
