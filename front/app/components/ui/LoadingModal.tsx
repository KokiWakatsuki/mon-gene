'use client';

import React from 'react';

interface LoadingModalProps {
  isOpen: boolean;
  message?: string;
  showProgress?: boolean;
  currentStage?: number;
  maxStages?: number;
  stageProgress?: number;
  stageMessage?: string;
}

export default function LoadingModal({ 
  isOpen, 
  message = '問題を生成しています...',
  showProgress = false,
  currentStage = 0,
  maxStages = 5,
  stageProgress = 0,
  stageMessage = ''
}: LoadingModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-8 max-w-md w-full mx-4">
        <div className="text-center">
          <div className="mb-4">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-mongene-green"></div>
          </div>
          <h3 className="text-lg font-semibold text-mongene-ink mb-2">
            {message}
          </h3>
          
          {showProgress && (
            <div className="mb-4">
              {/* 段階表示 */}
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-mongene-ink">
                  Stage {currentStage}/{maxStages}
                </span>
                <span className="text-sm text-mongene-muted">
                  {stageProgress.toFixed(0)}%
                </span>
              </div>
              
              {/* 進捗バー */}
              <div className="w-full bg-gray-200 rounded-full h-3 mb-3">
                <div 
                  className="bg-gradient-to-r from-blue-500 to-purple-600 h-3 rounded-full transition-all duration-500"
                  style={{ width: `${stageProgress}%` }}
                ></div>
              </div>
              
              {/* 段階メッセージ */}
              {stageMessage && (
                <div className="text-sm text-mongene-muted mb-2">
                  {stageMessage}
                </div>
              )}
              
              {/* 段階インジケーター */}
              <div className="flex justify-center gap-2">
                {Array.from({ length: maxStages }, (_, i) => (
                  <div
                    key={i}
                    className={`w-2 h-2 rounded-full transition-colors ${
                      i < currentStage
                        ? 'bg-green-500'
                        : i === currentStage - 1
                          ? 'bg-blue-500'
                          : 'bg-gray-300'
                    }`}
                  />
                ))}
              </div>
            </div>
          )}
          
          <p className="text-sm text-mongene-muted">
            {showProgress 
              ? '各段階を順次実行しています。しばらくお待ちください。'
              : 'Claude AIが問題を生成中です。しばらくお待ちください。'
            }
          </p>
        </div>
      </div>
    </div>
  );
}
