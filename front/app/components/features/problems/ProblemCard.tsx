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
  // デバッグ用コンソール出力
  React.useEffect(() => {
    console.log(`🔍 ProblemCard Debug - ID: ${id}`);
    console.log(`📝 Title: ${title}`);
    console.log(`📄 Content length: ${content?.length || 0}`);
    console.log(`🖼️ ImageBase64 exists: ${!!imageBase64}`);
    console.log(`🖼️ ImageBase64 length: ${imageBase64?.length || 0}`);
    if (imageBase64) {
      console.log(`🖼️ ImageBase64 preview: ${imageBase64.substring(0, 50)}...`);
    }
  }, [id, title, content, imageBase64]);

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
    <article className="bg-white border border-gray-200 rounded-xl shadow-[0_2px_4px_rgba(0,0,0,0.03)] flex flex-col min-h-[220px]">
      <div className="p-6 flex-1">
        <h3 className="m-0 text-gray-800 text-xl font-bold mb-3">{title}</h3>
        <div className="text-base text-gray-800 max-h-[120px] overflow-hidden text-ellipsis whitespace-pre-wrap leading-relaxed">
          {getPreviewContent()}
        </div>
      </div>
      <div className="border-t border-gray-200 px-6 py-4 flex gap-3 bg-gray-100 rounded-b-xl mt-auto">
        <button
          className="flex-1 appearance-none border-0 rounded-lg px-4 py-2 text-sm font-bold cursor-pointer bg-blue-500 text-white hover:brightness-110 transition-all"
          type="button"
          onClick={() => onPreview(id)}
        >
          プレビュー
        </button>
        <button
          className="flex-1 appearance-none border-0 rounded-lg px-4 py-2 text-sm font-bold cursor-pointer bg-gray-100 text-gray-800 hover:bg-gray-200 transition-all"
          type="button"
          onClick={() => onPrint(id)}
        >
          印刷
        </button>
      </div>
    </article>
  );
}
