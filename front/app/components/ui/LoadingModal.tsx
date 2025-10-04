'use client';

import React from 'react';

interface LoadingModalProps {
  isOpen: boolean;
  message?: string;
}

export default function LoadingModal({ isOpen, message = '問題を生成しています...' }: LoadingModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-8 max-w-sm w-full mx-4">
        <div className="text-center">
          <div className="mb-4">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-mongene-green"></div>
          </div>
          <h3 className="text-lg font-semibold text-mongene-ink mb-2">
            {message}
          </h3>
          <p className="text-sm text-mongene-muted">
            Claude AIが問題を生成中です。しばらくお待ちください。
          </p>
        </div>
      </div>
    </div>
  );
}
