'use client';

import React, { useEffect, useState, useRef } from 'react';

interface LoadingModalProps {
  isOpen: boolean;
  message?: string;
  showProgress?: boolean;
  estimatedDuration?: number; // 各ステージの推定時間（ミリ秒）
  currentStage?: number; // 外部から渡される現在のステージ
  onStageChange?: (stage: number) => void; // ステージ変更時のコールバック
}

export default function LoadingModal({
  isOpen,
  message = '問題を生成しています...',
  showProgress = false,
  estimatedDuration = 60000, // デフォルト60秒（各ステージ）
  currentStage: externalStage,
  onStageChange
}: LoadingModalProps) {
  const [progress, setProgress] = useState(0);
  const [currentStage, setCurrentStage] = useState(1);
  const stageStartTimeRef = useRef<number>(Date.now());
  const previousStageRef = useRef<number>(1);
  const stageTimerRef = useRef<NodeJS.Timeout | null>(null);

  // 外部からステージが渡された場合は同期
  useEffect(() => {
    if (externalStage !== undefined && externalStage !== currentStage) {
      setCurrentStage(externalStage);
    }
  }, [externalStage]);

  // 60秒ごとに自動的にステージを進める
  // 外部からステージが変更されたときの処理
  useEffect(() => {
    if (!isOpen || !showProgress) {
      setProgress(0);
      setCurrentStage(1);
      stageStartTimeRef.current = Date.now();
      previousStageRef.current = 1;
      return;
    }

    // 各ステージを20%ずつ進める（Stage 1: 0-19%, Stage 2: 20-39%, ...）
    const interval = setInterval(() => {
      const currentStageNum = currentStage;
      const stageBaseProgress = (currentStageNum - 1) * 20; // ステージの開始位置（0%, 20%, 40%, 60%, 80%）
      
      // ステージの最大値を設定（19%, 39%, 59%, 79%, 99%）
      let stageMaxProgress: number;
      if (currentStageNum === 5) {
        stageMaxProgress = 99; // Stage 5は99%まで
      } else {
        stageMaxProgress = currentStageNum * 20 - 1; // 他のステージは19%, 39%, 59%, 79%まで
      }
      
      const elapsed = Date.now() - stageStartTimeRef.current;
      const stageProgress = Math.min((elapsed / estimatedDuration) * 20, 20); // 各ステージで0-20%進む
      const calculatedProgress = Math.min(stageBaseProgress + stageProgress, stageMaxProgress);
      
      setProgress(calculatedProgress);
    }, 100);

    return () => clearInterval(interval);
  }, [isOpen, showProgress, estimatedDuration, currentStage]);

  // ステージが変わったら開始時刻をリセットし、そのステージの開始位置に強制移行
  useEffect(() => {
    if (currentStage !== previousStageRef.current && currentStage > 1) {
      // 前のステージの最大値（19%, 39%, 59%, 79%）に到達していない場合は強制移行
      const prevStageMax = (previousStageRef.current) * 20 - 1;
      if (progress < prevStageMax) {
        setProgress(prevStageMax);
      }
      
      // 新しいステージの開始位置に移行
      const newStageStart = (currentStage - 1) * 20;
      setProgress(newStageStart);
      stageStartTimeRef.current = Date.now();
      previousStageRef.current = currentStage;
      
      console.log(`📊 [LoadingModal] Stage ${previousStageRef.current} → ${currentStage}: ${prevStageMax}% → ${newStageStart}%`);
    }
  }, [currentStage, progress]);

  if (!isOpen) return null;

  const stageMessages = [
    '小問構成と解答プロセスを生成中...',
    'パラメータ設定と動的検証を実行中...',
    '問題文用の図形を描画中...',
    '完全な問題文を生成中...',
    '完全な解答・解説を生成中...'
  ];

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
            <div className="mt-4 mb-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-mongene-ink">Stage {currentStage}/5</span>
                <span className="text-sm text-mongene-muted">{progress.toFixed(0)}%</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div
                  className="bg-gradient-to-r from-blue-500 to-purple-600 h-2.5 rounded-full transition-all duration-300"
                  style={{ width: `${progress}%` }}
                ></div>
              </div>
              <p className="text-xs text-mongene-muted mt-2">
                {stageMessages[currentStage - 1]}
              </p>
            </div>
          )}
          
          <p className="text-sm text-mongene-muted">
            {showProgress ? '5段階生成プロセスを実行中です。しばらくお待ちください。' : 'Claude AIが問題を生成中です。しばらくお待ちください。'}
          </p>
        </div>
      </div>
    </div>
  );
}
