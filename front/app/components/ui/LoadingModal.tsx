'use client';

import React, { useEffect, useState, useRef } from 'react';

interface LoadingModalProps {
  isOpen: boolean;
  message?: string;
  showProgress?: boolean;
  estimatedDuration?: number;
  currentStage?: number;
  maxStages?: number; // 最大ステージ数（5 or 15）
  onStageChange?: (stage: number) => void;
}

export default function LoadingModal({
  isOpen,
  message = '問題を生成しています...',
  showProgress = false,
  estimatedDuration = 60000,
  currentStage: externalStage,
  maxStages = 5,
  onStageChange
}: LoadingModalProps) {
  const [progress, setProgress] = useState(0);
  const [currentStage, setCurrentStage] = useState(1);
  const stageStartTimeRef = useRef<number>(Date.now());
  const previousStageRef = useRef<number>(1);

  // 外部からステージが渡された場合は同期
  useEffect(() => {
    if (externalStage !== undefined && externalStage !== currentStage) {
      setCurrentStage(externalStage);
    }
  }, [externalStage]);

  // 進捗バー表示ロジック（5段階 or 15段階対応）
  useEffect(() => {
    if (!isOpen || !showProgress) {
      setProgress(0);
      setCurrentStage(1);
      stageStartTimeRef.current = Date.now();
      previousStageRef.current = 1;
      return;
    }

    // 各ステージの進捗率を計算（maxStagesに応じて調整）
    const progressPerStage = 100 / maxStages;
    
    const interval = setInterval(() => {
      const currentStageNum = currentStage;
      const stageBaseProgress = (currentStageNum - 1) * progressPerStage;
      
      let stageMaxProgress: number;
      if (currentStageNum === maxStages) {
        stageMaxProgress = 99;
      } else {
        stageMaxProgress = currentStageNum * progressPerStage - 1;
      }
      
      const elapsed = Date.now() - stageStartTimeRef.current;
      const stageProgress = Math.min((elapsed / estimatedDuration) * progressPerStage, progressPerStage);
      const calculatedProgress = Math.min(stageBaseProgress + stageProgress, stageMaxProgress);
      
      setProgress(calculatedProgress);
    }, 100);

    return () => clearInterval(interval);
  }, [isOpen, showProgress, estimatedDuration, currentStage, maxStages]);

  // ステージが変わったら開始時刻をリセット
  useEffect(() => {
    if (currentStage !== previousStageRef.current && currentStage > 1) {
      const progressPerStage = 100 / maxStages;
      const prevStageMax = (previousStageRef.current) * progressPerStage - 1;
      if (progress < prevStageMax) {
        setProgress(prevStageMax);
      }
      
      const newStageStart = (currentStage - 1) * progressPerStage;
      setProgress(newStageStart);
      stageStartTimeRef.current = Date.now();
      previousStageRef.current = currentStage;
      
      console.log(`📊 [LoadingModal] Stage ${previousStageRef.current} → ${currentStage}: ${prevStageMax}% → ${newStageStart}%`);
    }
  }, [currentStage, progress, maxStages]);

  if (!isOpen) return null;

  // ステージメッセージ（5段階 or 15段階）
  const stageMessages5 = [
    '小問構成と解答プロセスを生成中...',
    'パラメータ設定と動的検証を実行中...',
    '問題文用の図形を描画中...',
    '完全な問題文を生成中...',
    '完全な解答・解説を生成中...'
  ];

  const stageMessages15 = [
    'パターンA: 骨組み設計中...',
    'パターンA: パラメータ設定中...',
    'パターンA: 図形描画中...',
    'パターンA: 問題文生成中...',
    'パターンA: 解答生成中...',
    'パターンB: 骨組み設計中...',
    'パターンB: パラメータ設定中...',
    'パターンB: 図形描画中...',
    'パターンB: 問題文生成中...',
    'パターンB: 解答生成中...',
    'パターンC: 骨組み設計中...',
    'パターンC: パラメータ設定中...',
    'パターンC: 図形描画中...',
    'パターンC: 問題文生成中...',
    'パターンC: 解答生成中...'
  ];

  const stageMessages = maxStages === 15 ? stageMessages15 : stageMessages5;

  return (
    <div 
      className="fixed top-0 left-0 w-full h-full bg-white/80 backdrop-blur-[5px] z-[9999] flex items-center justify-center opacity-0 animate-[fadeIn_0.3s_forwards]"
      style={{
        animation: 'fadeIn 0.3s forwards'
      }}
    >
      <style jsx>{`
        @keyframes fadeIn {
          to { opacity: 1; }
        }
        @keyframes spin {
          100% { transform: rotate(360deg); }
        }
      `}</style>
      
      <div className="bg-white px-10 py-10 rounded-[20px] shadow-[0_10px_40px_rgba(0,0,0,0.1)] border border-gray-200 text-center w-[90%] max-w-[400px]">
        <div className="mb-4">
          <svg 
            width="48" 
            height="48" 
            fill="none" 
            viewBox="0 0 24 24" 
            strokeWidth="2" 
            stroke="currentColor"
            className="inline-block text-blue-500 animate-[spin_1.5s_linear_infinite]"
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
          </svg>
        </div>
        
        <h3 className="m-0 mb-5 text-lg text-gray-800">
          {showProgress ? '準備中...' : message}
        </h3>
        
        {showProgress && (
          <>
            <div className="w-full h-2 bg-gray-100 rounded mb-3 overflow-hidden">
              <div
                className="h-full bg-mongene-green rounded transition-[width] duration-500 ease-out"
                style={{ width: `${progress}%` }}
              ></div>
            </div>
            
            <p className="text-[13px] text-gray-500 m-0">
              {stageMessages[currentStage - 1] || 'AIが思考しています'}
            </p>
          </>
        )}
        
        {!showProgress && (
          <p className="text-[13px] text-gray-500 m-0">
            AIが思考しています
          </p>
        )}
      </div>
    </div>
  );
}
