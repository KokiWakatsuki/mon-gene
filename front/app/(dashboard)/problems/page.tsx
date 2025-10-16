'use client';

import React, { useState, useEffect } from 'react';
import Header from '../../components/layout/Header';
import Tabs from '../../components/features/problems/Tabs';
import Filters from '../../components/features/filters/Filters';
import ProblemCard from '../../components/features/problems/ProblemCard';
import BackgroundShapes from '../../components/layout/BackgroundShapes';
import ProblemPreviewModal from '../../components/features/problems/ProblemPreviewModal';
import LoadingModal from '../../components/ui/LoadingModal';
import { API_CONFIG } from '../../lib/config/api';

export default function Home() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);
  const [activeSubject, setActiveSubject] = useState('数学');
  const [selectedFilters, setSelectedFilters] = useState<Record<string, string[]>>({});
  const [previewModal, setPreviewModal] = useState<{ 
    isOpen: boolean; 
    problemId: string; 
    problemTitle: string; 
    problemContent?: string; 
    imageBase64?: string; 
    solutionText?: string;
    // 2段階生成システム用の追加プロパティ
    solutionSteps?: string;
    calculationProgram?: string;
    calculationResults?: string;
    finalSolution?: string;
    generationLogs?: string;
  }>({
    isOpen: false,
    problemId: '',
    problemTitle: '',
    problemContent: '',
    imageBase64: undefined,
    solutionText: undefined,
  });
  const [isLoading, setIsLoading] = useState(false);
  const [problems, setProblems] = useState<Array<{ id: string; title: string; content: string; imageBase64?: string; solution?: string }>>([]);
  const [userInfo, setUserInfo] = useState<{
    school_code: string;
    email: string;
    problem_generation_limit: number;
    problem_generation_count: number;
  } | null>(null);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [isSearchMode, setIsSearchMode] = useState(false);
  const [searchResults, setSearchResults] = useState<Array<{ id: string; title: string; content: string; imageBase64?: string; solution?: string }>>([]);
  const [searchMatchType, setSearchMatchType] = useState<'exact' | 'partial'>('partial');
  
  // 5段階生成システム用の状態
  const [generationMode, setGenerationMode] = useState<'single' | 'two-stage' | 'five-stage'>('single');
  const [firstStageResult, setFirstStageResult] = useState<{
    problemText: string;
    imageBase64?: string;
    geometryCode?: string;
    log: string;
  } | null>(null);
  const [secondStageResult, setSecondStageResult] = useState<{
    solutionSteps: string;
    finalSolution: string;
    calculationResults: string;
    calculationProgram?: string;
    log: string;
  } | null>(null);
  const [generationLogs, setGenerationLogs] = useState<string>('');
  const [showLogs, setShowLogs] = useState<boolean>(false);
  const [isFirstStageComplete, setIsFirstStageComplete] = useState<boolean>(false);
  
  // 5段階生成システム専用の状態
  const [fiveStageResults, setFiveStageResults] = useState<{
    stage1?: { problemText: string; log: string };
    stage2?: { geometryCode: string; imageBase64: string; log: string };
    stage3?: { solutionSteps: string; log: string };
    stage4?: { calculationProgram: string; calculationResults: string; log: string };
    stage5?: { finalExplanation: string; log: string };
  }>({});
  const [currentStage, setCurrentStage] = useState<number>(0); // 0=未開始, 1-5=各段階
  const [stageProgress, setStageProgress] = useState<number>(0); // 進捗率 0-100

  // ユーザー情報を取得する関数
  const fetchUserInfo = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(API_CONFIG.USER_INFO_API_URL, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setUserInfo(data);
      }
    } catch (error) {
      console.error('ユーザー情報の取得に失敗しました:', error);
    }
  };

  // 問題履歴を取得する関数
  const fetchProblemHistory = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/history`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        const historyProblems = data.problems?.map((problem: any, index: number) => ({
          id: problem.id || String(index + 1),
          title: `問題 ${problem.id || index + 1}`,
          content: problem.content || problem.problem || '',
          imageBase64: problem.image_base64 || problem.ImageBase64,
          solution: problem.solution || problem.Solution,
        })) || [];
        
        setProblems(historyProblems);
        setIsSearchMode(false);
        console.log('問題履歴を取得しました:', historyProblems.length, '件');
      }
    } catch (error) {
      console.error('問題履歴の取得に失敗しました:', error);
    }
  };

  // 認証チェック
  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('token');
      if (!token) {
        window.location.href = '/login';
        return;
      }
      
      setIsAuthenticated(true);
      setIsCheckingAuth(false);
      
      // ユーザー情報を取得
      await fetchUserInfo();
      
      // 問題履歴を読み込む
      await fetchProblemHistory();
    };

    checkAuth();
  }, []);

  // 認証チェック中の表示
  if (isCheckingAuth) {
    return (
      <div className="relative min-h-screen overflow-hidden bg-mongene-bg">
        <BackgroundShapes />
        <div className="relative z-10 flex items-center justify-center min-h-screen">
          <div className="text-center">
            <div className="w-8 h-8 bg-mongene-blue rounded-lg mx-auto mb-4"></div>
            <div className="font-extrabold text-mongene-blue mb-2">Mongene</div>
            <div className="text-mongene-muted">認証を確認しています...</div>
          </div>
        </div>
      </div>
    );
  }

  // 認証されていない場合（念のため）
  if (!isAuthenticated) {
    return null;
  }

  const subjects = ['数学', '英語', '国語'];

  // 科目別の単元データ
  const subjectUnits = {
    '数学': [
      { label: '式の計算', value: 'calculation' },
      { label: '図形', value: 'geometry' },
      { label: '空間図形', value: 'spatial_geometry' },
      { label: '2次不等式', value: 'quadratic' },
      { label: '関数', value: 'function' },
      { label: '確率', value: 'probability' },
    ],
    '英語': [
      { label: '文法', value: 'grammar' },
      { label: '読解', value: 'reading' },
      { label: '語彙', value: 'vocabulary' },
      { label: 'リスニング', value: 'listening' },
    ],
    '国語': [
      { label: '現代文', value: 'modern' },
      { label: '古文', value: 'classical' },
      { label: '漢文', value: 'chinese' },
      { label: '文法', value: 'grammar' },
    ],
  };

  const getFilterGroups = () => [
    {
      label: '学年',
      options: [
        { label: '中1', value: 'grade1' },
        { label: '中2', value: 'grade2' },
        { label: '中3', value: 'grade3' },
      ],
      allowMultiple: false,
    },
    {
      label: '単元',
      options: subjectUnits[activeSubject as keyof typeof subjectUnits] || [],
      allowMultiple: true,
    },
    {
      label: '難易度',
      options: [
        { label: 'Lv1', value: 'level1' },
        { label: 'Lv2', value: 'level2' },
        { label: 'Lv3', value: 'level3' },
        { label: 'Lv4', value: 'level4' },
        { label: 'Lv5', value: 'level5' },
      ],
      allowMultiple: false,
    },
    {
      label: '必要な公式数',
      options: [
        { label: '1個', value: 'formula1' },
        { label: '2個', value: 'formula2' },
        { label: '3個', value: 'formula3' },
        { label: '4個以上', value: 'formula4plus' },
      ],
      allowMultiple: false,
    },
    {
      label: '計算量',
      options: [
        { label: '簡単', value: 'simple' },
        { label: '普通', value: 'medium' },
        { label: '複雑', value: 'complex' },
      ],
      allowMultiple: false,
    },
    {
      label: '数値の複雑性',
      options: [
        { label: '整数のみ', value: 'integer' },
        { label: '小数を含む', value: 'decimal' },
        { label: '分数を含む', value: 'fraction' },
      ],
      allowMultiple: false,
    },
    {
      label: '問題文の文章量',
      options: [
        { label: '短い', value: 'short' },
        { label: '普通', value: 'medium' },
        { label: '長い', value: 'long' },
      ],
      allowMultiple: false,
    },
  ];

  const handleSubjectChange = (subject: string) => {
    setActiveSubject(subject);
    // 科目が変わったら単元の選択をリセット
    setSelectedFilters(prev => {
      const newFilters = { ...prev };
      delete newFilters['単元'];
      return newFilters;
    });
  };

  const handleFilterChange = (groupLabel: string, value: string, allowMultiple: boolean) => {
    setSelectedFilters(prev => {
      const currentFilters = prev[groupLabel] || [];
      const isSelected = currentFilters.includes(value);
      
      if (allowMultiple) {
        // 複数選択可能な場合
        if (isSelected) {
          return {
            ...prev,
            [groupLabel]: currentFilters.filter(f => f !== value),
          };
        } else {
          return {
            ...prev,
            [groupLabel]: [...currentFilters, value],
          };
        }
      } else {
        // 単一選択の場合
        if (isSelected) {
          return {
            ...prev,
            [groupLabel]: [],
          };
        } else {
          return {
            ...prev,
            [groupLabel]: [value],
          };
        }
      }
    });
  };

  const handlePreview = (id: string) => {
    const problem = problems.find(p => p.id === id);
    if (problem) {
      setPreviewModal({
        isOpen: true,
        problemId: id,
        problemTitle: problem.title,
        problemContent: problem.content,
        imageBase64: problem.imageBase64,
        solutionText: problem.solution,
      });
    }
  };

  const handlePrint = (id: string) => {
    const problem = problems.find(p => p.id === id);
    if (problem) {
      // 印刷用の新しいウィンドウを開く
      const printWindow = window.open('', '_blank');
      if (printWindow) {
        const imageHtml = problem.imageBase64 
          ? `<div style="text-align: center; margin: 20px 0;">
               <img src="data:image/png;base64,${problem.imageBase64}" 
                    style="max-width: 100%; height: auto; border: 1px solid #ddd;" 
                    alt="問題図形" />
             </div>`
          : '';
        
        // 解答・解説がある場合は別ページに追加
        const solutionHtml = problem.solution 
          ? `<div style="page-break-before: always;">
               <h1>解答・解説</h1>
               <div class="content">${problem.solution}</div>
             </div>`
          : '';
        
        printWindow.document.write(`
          <!DOCTYPE html>
          <html>
          <head>
            <title>${problem.title}</title>
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
            <h1>${problem.title}</h1>
            <div class="content">${problem.content || ''}</div>
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
    }
  };

  // エラーハンドリング関数
  const handleGenerationError = async (error: unknown) => {
    let errorMessage = '不明なエラーが発生しました';
    let isTokenLimitError = false;
    let suggestions: string[] = [];
    
    if (error instanceof Response) {
      // HTTPレスポンスエラーの場合
      try {
        const errorData = await error.json();
        if (errorData.error) {
          errorMessage = errorData.error;
          
          // トークン関連のエラーかチェック
          if (errorMessage.includes('トークン数が上限を超えています') || 
              errorMessage.includes('入力テキストが長すぎます') ||
              errorMessage.includes('生成されるレスポンスが長すぎます')) {
            isTokenLimitError = true;
            suggestions = [
              '・問題文の文章量を「短い」に設定してください',
              '・必要な公式数を少なくしてください',
              '・計算量を「簡単」に設定してください',
              '・より具体的で短い条件を指定してください'
            ];
          }
        }
      } catch (parseError) {
        errorMessage = `HTTP Error ${error.status}: ${error.statusText}`;
      }
    } else if (error instanceof Error) {
      errorMessage = error.message;
      
      // エラーメッセージからトークン関連エラーを検出
      if (errorMessage.includes('トークン数が上限を超えています') || 
          errorMessage.includes('入力テキストが長すぎます') ||
          errorMessage.includes('生成されるレスポンスが長すぎます') ||
          errorMessage.includes('context_length_exceeded') ||
          errorMessage.includes('max_tokens_exceeded') ||
          errorMessage.includes('maximum context length') ||
          errorMessage.includes('too many tokens')) {
        isTokenLimitError = true;
        suggestions = [
          '・問題文の文章量を「短い」に設定してください',
          '・必要な公式数を少なくしてください',
          '・計算量を「簡単」に設定してください',
          '・より具体的で短い条件を指定してください'
        ];
      }
    }
    
    // エラーメッセージを表示
    if (isTokenLimitError) {
      const suggestionText = suggestions.length > 0 ? '\n\n対処法:\n' + suggestions.join('\n') : '';
      alert(`🚫 トークン数制限エラー\n\n${errorMessage}${suggestionText}`);
    } else {
      alert(`❌ 問題生成に失敗しました\n\n${errorMessage}`);
    }
  };

  // 上限チェック機能
  const isGenerationLimitReached = () => {
    if (!userInfo) return false;
    if (userInfo.problem_generation_limit === -1) return false; // 制限なし
    return userInfo.problem_generation_count >= userInfo.problem_generation_limit;
  };

  // 2段階生成システムの関数群
  
  // 1回目のAPI呼び出し（問題文・図形生成）
  const handleGenerateFirstStage = async () => {
    // 上限チェック
    if (isGenerationLimitReached()) {
      alert(`問題生成回数の上限（${userInfo?.problem_generation_limit}回）に達しました。これ以上問題を生成することはできません。`);
      return;
    }

    // 必須フィルターのチェック
    const requiredFilters = ['学年', '単元', '難易度', '必要な公式数', '計算量', '数値の複雑性', '問題文の文章量'];
    const missingFilters = requiredFilters.filter(filter => 
      !selectedFilters[filter] || selectedFilters[filter].length === 0
    );
    
    if (missingFilters.length > 0) {
      alert(`以下の項目を選択してください: ${missingFilters.join(', ')}`);
      return;
    }
    
    setIsLoading(true);
    setGenerationLogs('');
    setFirstStageResult(null);
    setSecondStageResult(null);
    setIsFirstStageComplete(false);
    
    try {
      const prompt = createPromptFromFilters();
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません。再度ログインしてください。');
      }

      console.log('🚀 [FirstStage] 1回目のAPI呼び出しを開始');
      
      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-first-stage`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          prompt: prompt,
          subject: activeSubject,
          filters: selectedFilters
        })
      });
      
      if (!response.ok) {
        throw new Error(`1回目API呼び出しエラー: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      
      if (!data.success) {
        throw new Error(data.error || '1回目のAPI呼び出しに失敗しました');
      }
      
      const result = {
        problemText: data.problem_text || data.problemText || '',
        imageBase64: data.image_base64 || data.imageBase64,
        geometryCode: data.geometry_code || data.geometryCode,
        log: data.log || ''
      };
      
      setFirstStageResult(result);
      setGenerationLogs(result.log);
      setIsFirstStageComplete(true);
      setIsLoading(false);
      
      console.log('✅ [FirstStage] 1回目のAPI呼び出し完了:', result);
      
    } catch (error) {
      setIsLoading(false);
      console.error('1回目API呼び出しエラー:', error);
      await handleGenerationError(error);
    }
  };

  // 2回目のAPI呼び出し（解答手順・数値計算）
  const handleGenerateSecondStage = async () => {
    if (!firstStageResult) {
      alert('先に1回目の生成を完了してください');
      return;
    }
    
    setIsLoading(true);
    
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません。再度ログインしてください。');
      }

      console.log('🚀 [SecondStage] 2回目のAPI呼び出しを開始');
      
      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-second-stage`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          problem_text: firstStageResult.problemText,
          geometry_code: firstStageResult.geometryCode || ''
        })
      });
      
      if (!response.ok) {
        throw new Error(`2回目API呼び出しエラー: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      
      if (!data.success) {
        throw new Error(data.error || '2回目のAPI呼び出しに失敗しました');
      }
      
      const result = {
        solutionSteps: data.solution_steps || data.solutionSteps || '',
        finalSolution: data.final_solution || data.finalSolution || '',
        calculationResults: data.calculation_results || data.calculationResults || '',
        calculationProgram: data.calculation_program || data.calculationProgram,
        log: data.log || ''
      };
      
      setSecondStageResult(result);
      setGenerationLogs(prev => prev + '\n' + result.log);
      setIsLoading(false);
      
      console.log('✅ [SecondStage] 2回目のAPI呼び出し完了:', result);
      
      // 完成した問題を問題リストに追加
      const problemTitle = `2段階生成問題 ${problems.length + 1}`;
      const newProblemId = String(problems.length + 1);
      const finalSolution = `${result.solutionSteps}\n\n${result.finalSolution}`;
      
      const newProblem = {
        id: newProblemId,
        title: problemTitle,
        content: firstStageResult.problemText,
        solution: finalSolution,
        imageBase64: firstStageResult.imageBase64,
      };
      
      setProblems(prev => [...prev, newProblem]);
      
      // ユーザー情報を更新（生成回数をインクリメント）
      await fetchUserInfo();
      
      // プレビューモーダルを表示（2段階生成の詳細データを含む）
      setPreviewModal({
        isOpen: true,
        problemId: newProblemId,
        problemTitle: problemTitle,
        problemContent: firstStageResult.problemText,
        imageBase64: firstStageResult.imageBase64,
        solutionText: finalSolution, // 従来形式の互換性維持
        // 2段階生成システム用の追加データ
        solutionSteps: result.solutionSteps,
        calculationProgram: result.calculationProgram,
        calculationResults: result.calculationResults,
        finalSolution: result.finalSolution,
        generationLogs: firstStageResult.log + '\n' + result.log,
      });
      
    } catch (error) {
      setIsLoading(false);
      console.error('2回目API呼び出しエラー:', error);
      await handleGenerationError(error);
    }
  };

  // 全体の2段階生成プロセス（一気に実行）
  const handleGenerateTwoStage = async () => {
    // 上限チェック
    if (isGenerationLimitReached()) {
      alert(`問題生成回数の上限（${userInfo?.problem_generation_limit}回）に達しました。これ以上問題を生成することはできません。`);
      return;
    }

    // 必須フィルターのチェック
    const requiredFilters = ['学年', '単元', '難易度', '必要な公式数', '計算量', '数値の複雑性', '問題文の文章量'];
    const missingFilters = requiredFilters.filter(filter => 
      !selectedFilters[filter] || selectedFilters[filter].length === 0
    );
    
    if (missingFilters.length > 0) {
      alert(`以下の項目を選択してください: ${missingFilters.join(', ')}`);
      return;
    }
    
    setIsLoading(true);
    setGenerationLogs('');
    setFirstStageResult(null);
    setSecondStageResult(null);
    setIsFirstStageComplete(false);
    
    try {
      const prompt = createPromptFromFilters();
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません。再度ログインしてください。');
      }

      console.log('🚀 [TwoStage] 2段階生成プロセス全体を開始');
      
      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-problem-two-stage`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          prompt: prompt,
          subject: activeSubject,
          filters: selectedFilters
        })
      });
      
      if (!response.ok) {
        throw new Error(`2段階生成エラー: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      
      if (!data.success) {
        throw new Error(data.error || '2段階生成に失敗しました');
      }
      
      // ログを設定
      const combinedLogs = (data.first_stage_log || data.firstStageLog || '') + 
                          '\n' + 
                          (data.second_stage_log || data.secondStageLog || '');
      setGenerationLogs(combinedLogs);
      
      // 結果を問題リストに追加
      const problemTitle = `2段階生成問題 ${problems.length + 1}`;
      const newProblemId = String(problems.length + 1);
      const finalSolution = `${data.solution_steps || data.solutionSteps || ''}\n\n${data.final_solution || data.finalSolution || ''}`;
      
      const newProblem = {
        id: newProblemId,
        title: problemTitle,
        content: data.problem_text || data.problemText || '',
        solution: finalSolution,
        imageBase64: data.image_base64 || data.imageBase64,
      };
      
      setProblems(prev => [...prev, newProblem]);
      
      // ユーザー情報を更新
      await fetchUserInfo();
      
      setIsLoading(false);
      
      // プレビューモーダルを表示
      setPreviewModal({
        isOpen: true,
        problemId: newProblemId,
        problemTitle: problemTitle,
        problemContent: newProblem.content,
        imageBase64: newProblem.imageBase64,
        solutionText: finalSolution,
      });
      
      console.log('✅ [TwoStage] 2段階生成プロセス完了');
      
    } catch (error) {
      setIsLoading(false);
      console.error('2段階生成エラー:', error);
      await handleGenerationError(error);
    }
  };

  // 5段階生成システムの関数（リアルタイム進捗付き）
  const handleGenerateFiveStage = async () => {
    // 上限チェック
    if (isGenerationLimitReached()) {
      alert(`問題生成回数の上限（${userInfo?.problem_generation_limit}回）に達しました。これ以上問題を生成することはできません。`);
      return;
    }

    // 必須フィルターのチェック
    const requiredFilters = ['学年', '単元', '難易度', '必要な公式数', '計算量', '数値の複雑性', '問題文の文章量'];
    const missingFilters = requiredFilters.filter(filter => 
      !selectedFilters[filter] || selectedFilters[filter].length === 0
    );
    
    if (missingFilters.length > 0) {
      alert(`以下の項目を選択してください: ${missingFilters.join(', ')}`);
      return;
    }
    
    setIsLoading(true);
    setGenerationLogs('');
    setFiveStageResults({});
    setCurrentStage(0);
    setStageProgress(0);
    
    const allLogs: string[] = [];
    
    try {
      const prompt = createPromptFromFilters();
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません。再度ログインしてください。');
      }

      console.log('🚀 [FiveStage] 5段階生成プロセス開始（リアルタイム進捗付き）');
      
      // Stage 1: 問題文生成
      setCurrentStage(1);
      setStageProgress(10);
      console.log('🚀 [Stage1] 問題文生成開始');
      
      const stage1Response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-stage1`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          prompt: prompt,
          subject: activeSubject,
          filters: selectedFilters
        })
      });
      
      if (!stage1Response.ok) {
        throw new Error(`Stage1エラー: ${stage1Response.status} ${stage1Response.statusText}`);
      }
      
      const stage1Data = await stage1Response.json();
      if (!stage1Data.success) {
        throw new Error(stage1Data.error || 'Stage1に失敗しました');
      }
      
      const stage1Result = {
        problemText: stage1Data.problem_text || '',
        log: stage1Data.log || ''
      };
      
      setFiveStageResults(prev => ({ ...prev, stage1: stage1Result }));
      allLogs.push(`=== Stage 1: 問題文生成 ===\n${stage1Result.log}`);
      setGenerationLogs(allLogs.join('\n\n'));
      setStageProgress(20);
      
      console.log('✅ [Stage1] 完了');
      
      // Stage 2: 図形生成
      setCurrentStage(2);
      setStageProgress(30);
      console.log('🚀 [Stage2] 図形生成開始');
      
      const stage2Response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-stage2`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          problem_text: stage1Result.problemText
        })
      });
      
      if (!stage2Response.ok) {
        throw new Error(`Stage2エラー: ${stage2Response.status} ${stage2Response.statusText}`);
      }
      
      const stage2Data = await stage2Response.json();
      if (!stage2Data.success) {
        throw new Error(stage2Data.error || 'Stage2に失敗しました');
      }
      
      const stage2Result = {
        geometryCode: stage2Data.geometry_code || '',
        imageBase64: stage2Data.image_base64 || '',
        log: stage2Data.log || ''
      };
      
      setFiveStageResults(prev => ({ ...prev, stage2: stage2Result }));
      allLogs.push(`=== Stage 2: 図形生成 ===\n${stage2Result.log}`);
      setGenerationLogs(allLogs.join('\n\n'));
      setStageProgress(40);
      
      console.log('✅ [Stage2] 完了');
      
      // Stage 3: 解答手順生成
      setCurrentStage(3);
      setStageProgress(50);
      console.log('🚀 [Stage3] 解答手順生成開始');
      
      const stage3Response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-stage3`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          problem_text: stage1Result.problemText,
          geometry_code: stage2Result.geometryCode,
          image_base64: stage2Result.imageBase64
        })
      });
      
      if (!stage3Response.ok) {
        throw new Error(`Stage3エラー: ${stage3Response.status} ${stage3Response.statusText}`);
      }
      
      const stage3Data = await stage3Response.json();
      if (!stage3Data.success) {
        throw new Error(stage3Data.error || 'Stage3に失敗しました');
      }
      
      const stage3Result = {
        solutionSteps: stage3Data.solution_steps || '',
        log: stage3Data.log || ''
      };
      
      setFiveStageResults(prev => ({ ...prev, stage3: stage3Result }));
      allLogs.push(`=== Stage 3: 解答手順生成 ===\n${stage3Result.log}`);
      setGenerationLogs(allLogs.join('\n\n'));
      setStageProgress(60);
      
      console.log('✅ [Stage3] 完了');
      
      // Stage 4: 数値計算プログラム生成・実行
      setCurrentStage(4);
      setStageProgress(70);
      console.log('🚀 [Stage4] 数値計算開始');
      
      const stage4Response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-stage4`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          problem_text: stage1Result.problemText,
          solution_steps: stage3Result.solutionSteps
        })
      });
      
      if (!stage4Response.ok) {
        throw new Error(`Stage4エラー: ${stage4Response.status} ${stage4Response.statusText}`);
      }
      
      const stage4Data = await stage4Response.json();
      if (!stage4Data.success) {
        throw new Error(stage4Data.error || 'Stage4に失敗しました');
      }
      
      const stage4Result = {
        calculationProgram: stage4Data.calculation_program || '',
        calculationResults: stage4Data.calculation_results || '',
        log: stage4Data.log || ''
      };
      
      setFiveStageResults(prev => ({ ...prev, stage4: stage4Result }));
      allLogs.push(`=== Stage 4: 数値計算 ===\n${stage4Result.log}`);
      setGenerationLogs(allLogs.join('\n\n'));
      setStageProgress(80);
      
      console.log('✅ [Stage4] 完了');
      
      // Stage 5: 最終解説生成
      setCurrentStage(5);
      setStageProgress(90);
      console.log('🚀 [Stage5] 最終解説生成開始');
      
      const stage5Response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-stage5`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          problem_text: stage1Result.problemText,
          solution_steps: stage3Result.solutionSteps,
          calculation_results: stage4Result.calculationResults
        })
      });
      
      if (!stage5Response.ok) {
        throw new Error(`Stage5エラー: ${stage5Response.status} ${stage5Response.statusText}`);
      }
      
      const stage5Data = await stage5Response.json();
      if (!stage5Data.success) {
        throw new Error(stage5Data.error || 'Stage5に失敗しました');
      }
      
      const stage5Result = {
        finalExplanation: stage5Data.final_explanation || '',
        log: stage5Data.log || ''
      };
      
      setFiveStageResults(prev => ({ ...prev, stage5: stage5Result }));
      allLogs.push(`=== Stage 5: 最終解説生成 ===\n${stage5Result.log}`);
      setGenerationLogs(allLogs.join('\n\n'));
      setStageProgress(100);
      
      console.log('✅ [Stage5] 完了');
      
      // 結果を問題リストに追加
      const problemTitle = `5段階生成問題 ${problems.length + 1}`;
      const newProblemId = String(problems.length + 1);
      const finalSolution = stage5Result.finalExplanation;
      
      const newProblem = {
        id: newProblemId,
        title: problemTitle,
        content: stage1Result.problemText,
        solution: finalSolution,
        imageBase64: stage2Result.imageBase64 || undefined,
      };
      
      setProblems(prev => [...prev, newProblem]);
      
      // ユーザー情報を更新
      await fetchUserInfo();
      
      setIsLoading(false);
      
      // プレビューモーダルを表示
      setPreviewModal({
        isOpen: true,
        problemId: newProblemId,
        problemTitle: problemTitle,
        problemContent: newProblem.content,
        imageBase64: newProblem.imageBase64,
        solutionText: finalSolution,
      });
      
      console.log('✅ [FiveStage] 5段階生成プロセス完全完了');
      
    } catch (error) {
      setIsLoading(false);
      setCurrentStage(0);
      setStageProgress(0);
      console.error('5段階生成エラー:', error);
      await handleGenerationError(error);
    }
  };

  const handleGenerate = async () => {
    if (generationMode === 'two-stage') {
      await handleGenerateTwoStage();
    } else if (generationMode === 'five-stage') {
      await handleGenerateFiveStage();
    } else {
      await handleGenerateSingle();
    }
  };

  // 従来の1段階生成（元のhandleGenerateの内容）
  const handleGenerateSingle = async () => {
    // 上限チェック
    if (isGenerationLimitReached()) {
      alert(`問題生成回数の上限（${userInfo?.problem_generation_limit}回）に達しました。これ以上問題を生成することはできません。`);
      return;
    }

    // 必須フィルターのチェック
    const requiredFilters = ['学年', '単元', '難易度', '必要な公式数', '計算量', '数値の複雑性', '問題文の文章量'];
    const missingFilters = requiredFilters.filter(filter => 
      !selectedFilters[filter] || selectedFilters[filter].length === 0
    );
    
    if (missingFilters.length > 0) {
      alert(`以下の項目を選択してください: ${missingFilters.join(', ')}`);
      return;
    }
    
    setIsLoading(true);
    
    try {
      // 選択されたフィルターから問題生成のプロンプトを作成
      const prompt = createPromptFromFilters();
      
      console.log('問題生成プロンプト:', prompt);
      console.log('選択されたフィルター:', selectedFilters);
      console.log('選択された科目:', activeSubject);
      console.log('API使用モード:', API_CONFIG.USE_REAL_API ? '実際のAPI' : 'テスト版');
      
      let generatedContent = '';
      let problemTitle = '';
      let newProblemId = String(problems.length + 1);
      
      if (API_CONFIG.USE_REAL_API) {
        // バックエンドサーバー経由でClaude APIを呼び出す
        console.log('バックエンドサーバー経由でClaude APIを呼び出しています...');
        
        // 認証トークンを取得
        const token = localStorage.getItem('token');
        if (!token) {
          throw new Error('認証トークンが見つかりません。再度ログインしてください。');
        }

        const response = await fetch(API_CONFIG.BACKEND_API_URL, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({
            prompt: prompt,
            subject: activeSubject,
            filters: selectedFilters
          })
        });
        
        if (!response.ok) {
          throw new Error(`バックエンドAPI呼び出しエラー: ${response.status} ${response.statusText}`);
        }
        
        const data = await response.json();
        generatedContent = data.content || data.problem || 'エラー: 問題の生成に失敗しました';
        problemTitle = `AI生成問題 ${problems.length + 1}`;
        
        console.log('🔍 バックエンドAPIレスポンス:', data);
        console.log('🔍 data.content:', data.content);
        console.log('🔍 data.solution:', data.solution);
        console.log('🔍 data.Solution:', data.Solution);
        console.log('🔍 data.ImageBase64:', data.ImageBase64);
        console.log('🔍 data.image_base64:', data.image_base64);
        console.log('🔍 ImageBase64 exists:', !!(data.ImageBase64 || data.image_base64));
        console.log('🔍 ImageBase64 length:', (data.ImageBase64 || data.image_base64 || '').length);
        console.log('🔍 Solution exists:', !!(data.solution || data.Solution));
        console.log('🔍 Solution length:', (data.solution || data.Solution || '').length);
        
        // 画像データの処理
        const imageBase64 = data.ImageBase64 || data.image_base64;
        const finalImageBase64 = (imageBase64 && imageBase64.length > 0) ? imageBase64 : undefined;
        
        console.log('🔍 Final imageBase64 for problem:', !!finalImageBase64);
        console.log('🔍 Final imageBase64 length:', finalImageBase64?.length || 0);
        
        // 解答・解説データの処理
        const solutionText = data.solution || data.Solution || '';
        console.log('🔍 Final solutionText:', solutionText);
        console.log('🔍 Final solutionText length:', solutionText.length);
        
        // 新しい問題を追加（画像データと解答・解説を含む）
        const newProblemId = String(problems.length + 1);
        const newProblem = {
          id: newProblemId,
          title: problemTitle,
          content: generatedContent,
          solution: solutionText,
          imageBase64: finalImageBase64,
        };
        
        setProblems(prev => [...prev, newProblem]);
        
        // ユーザー情報を更新（生成回数をインクリメント）
        await fetchUserInfo();
        
        // ローディングを終了
        setIsLoading(false);
        
        // 生成された問題のプレビューを自動的に表示（画像データを含む）
        setPreviewModal({
          isOpen: true,
          problemId: newProblemId,
          problemTitle: problemTitle,
          problemContent: generatedContent,
          imageBase64: finalImageBase64,
          solutionText: solutionText,
        });
        
      } else {
        // テスト版（ダミーデータ）
        console.log('テスト版を使用しています');
        generatedContent = `これはテスト用の問題です。\n\n選択された条件:\n${prompt}\n\n実際のAPI版では、ここにClaude AIが生成した問題が表示されます。`;
        problemTitle = `テスト問題 ${problems.length + 1}`;
        
        // 新しい問題を追加（テスト版でも画像データを含む）
        const newProblemId = String(problems.length + 1);
        const newProblem = {
          id: newProblemId,
          title: problemTitle,
          content: generatedContent,
          imageBase64: undefined,
        };
        
        setProblems(prev => [...prev, newProblem]);
        
        // ローディングを終了
        setIsLoading(false);
        
        // 生成された問題のプレビューを自動的に表示
        setPreviewModal({
          isOpen: true,
          problemId: newProblemId,
          problemTitle: problemTitle,
          problemContent: generatedContent,
          imageBase64: undefined,
          solutionText: undefined,
        });
      }
      
    } catch (error) {
      setIsLoading(false);
      console.error('問題生成エラー:', error);
      
      // エラーレスポンスを解析して詳細なメッセージを表示
      await handleGenerationError(error);
    }
  };

  const createPromptFromFilters = () => {
    const filterTexts = [];
    
    filterTexts.push(`科目: ${activeSubject}`);
    
    Object.entries(selectedFilters).forEach(([key, values]) => {
      if (values.length > 0) {
        filterTexts.push(`${key}: ${values.join(', ')}`);
      }
    });
    
    return `以下の条件で${activeSubject}の問題を生成してください:\n${filterTexts.join('\n')}`;
  };

  // キーワード検索する関数
  const searchProblems = async () => {
    if (!searchKeyword.trim()) {
      alert('検索キーワードを入力してください');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/search?keyword=${encodeURIComponent(searchKeyword)}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        const foundProblems = data.problems?.map((problem: any, index: number) => ({
          id: problem.id || String(index + 1),
          title: `検索結果 ${problem.id || index + 1}`,
          content: problem.content || problem.problem || '',
          imageBase64: problem.image_base64 || problem.ImageBase64,
          solution: problem.solution || problem.Solution,
        })) || [];
        
        setSearchResults(foundProblems);
        setIsSearchMode(true);
        console.log('検索結果:', foundProblems.length, '件');
      }
    } catch (error) {
      console.error('検索に失敗しました:', error);
      alert('検索に失敗しました');
    }
  };

  // パラメータ検索する関数
  const searchProblemsByFilters = async () => {
    // 検索条件をチェック
    const hasSubject = activeSubject !== '';
    const hasFilters = Object.keys(selectedFilters).some(key => 
      selectedFilters[key] && selectedFilters[key].length > 0
    );

    if (!hasSubject && !hasFilters) {
      alert('科目を選択するか、フィルター条件を設定してください');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/search-by-filters`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          subject: activeSubject,
          filters: selectedFilters,
          matchType: searchMatchType,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        const foundProblems = data.problems?.map((problem: any, index: number) => ({
          id: problem.id || String(index + 1),
          title: `パラメータ検索結果 ${problem.id || index + 1}`,
          content: problem.content || problem.problem || '',
          imageBase64: problem.image_base64 || problem.ImageBase64,
          solution: problem.solution || problem.Solution,
        })) || [];
        
        setSearchResults(foundProblems);
        setIsSearchMode(true);
        console.log('パラメータ検索結果:', foundProblems.length, '件');
      } else {
        const errorData = await response.json();
        alert(`検索に失敗しました: ${errorData.error || 'サーバーエラー'}`);
      }
    } catch (error) {
      console.error('パラメータ検索に失敗しました:', error);
      alert('パラメータ検索に失敗しました');
    }
  };

  // キーワード + 条件の組み合わせ検索する関数
  const searchProblemsByKeywordAndFilters = async () => {
    // 検索条件をチェック
    const hasKeyword = searchKeyword.trim() !== '';
    const hasSubject = activeSubject !== '';
    const hasFilters = Object.keys(selectedFilters).some(key => 
      selectedFilters[key] && selectedFilters[key].length > 0
    );

    if (!hasKeyword && !hasSubject && !hasFilters) {
      alert('キーワードを入力するか、科目・フィルター条件を設定してください');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/search-combined`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          keyword: searchKeyword.trim() || undefined,
          subject: activeSubject || undefined,
          filters: Object.keys(selectedFilters).length > 0 ? selectedFilters : undefined,
          matchType: searchMatchType,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        const foundProblems = data.problems?.map((problem: any, index: number) => ({
          id: problem.id || String(index + 1),
          title: `組み合わせ検索結果 ${problem.id || index + 1}`,
          content: problem.content || problem.problem || '',
          imageBase64: problem.image_base64 || problem.ImageBase64,
          solution: problem.solution || problem.Solution,
        })) || [];
        
        setSearchResults(foundProblems);
        setIsSearchMode(true);
        console.log('キーワード+条件検索結果:', foundProblems.length, '件');
      } else {
        const errorData = await response.json();
        alert(`検索に失敗しました: ${errorData.error || 'サーバーエラー'}`);
      }
    } catch (error) {
      console.error('キーワード+条件検索に失敗しました:', error);
      alert('キーワード+条件検索に失敗しました');
    }
  };

  return (
    <div className="relative min-h-screen overflow-hidden bg-mongene-bg">
      <BackgroundShapes />
      
      <div className="relative z-10 max-w-6xl mx-auto p-6">
        <Header />
        
        <Tabs 
          subjects={subjects}
          activeSubject={activeSubject}
          onSubjectChange={handleSubjectChange}
        />
        
        <Filters 
          filterGroups={getFilterGroups()}
          selectedFilters={selectedFilters}
          onFilterChange={handleFilterChange}
        />
        
        {/* 検索・履歴機能UI */}
        <div className="mb-6 p-4 bg-white/10 backdrop-blur-sm rounded-xl border border-white/20">
          <h3 className="text-lg font-bold text-mongene-ink mb-4">🔍 問題検索・履歴</h3>
          
          {/* キーワード検索 */}
          <div className="flex flex-col sm:flex-row gap-3 mb-4">
            <div className="flex-1">
              <input
                type="text"
                placeholder="キーワードを入力（例：図形、関数、確率...）"
                value={searchKeyword}
                onChange={(e) => setSearchKeyword(e.target.value)}
                className="w-full px-4 py-2 rounded-lg border border-white/20 bg-white/10 text-mongene-ink placeholder-mongene-muted focus:outline-none focus:ring-2 focus:ring-mongene-blue"
                onKeyDown={(e) => e.key === 'Enter' && searchProblems()}
              />
            </div>
            <button
              onClick={searchProblems}
              className="px-4 py-2 bg-mongene-blue text-white rounded-lg hover:brightness-110 transition-all"
            >
              キーワード検索
            </button>
          </div>

          {/* 検索タイプ選択 */}
          <div className="mb-3">
            <div className="flex items-center gap-4">
              <span className="text-sm font-medium text-mongene-ink">検索タイプ:</span>
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  name="searchMatchType"
                  value="partial"
                  checked={searchMatchType === 'partial'}
                  onChange={(e) => setSearchMatchType(e.target.value as 'exact' | 'partial')}
                  className="text-mongene-blue"
                />
                <span className="text-sm text-mongene-ink">部分一致（おすすめ）</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  name="searchMatchType"
                  value="exact"
                  checked={searchMatchType === 'exact'}
                  onChange={(e) => setSearchMatchType(e.target.value as 'exact' | 'partial')}
                  className="text-mongene-blue"
                />
                <span className="text-sm text-mongene-ink">完全一致</span>
              </label>
            </div>
            <div className="text-xs text-mongene-muted mt-1">
              {searchMatchType === 'partial' 
                ? '条件の一部でも一致すれば検索結果に表示されます' 
                : 'すべての条件が完全に一致する場合のみ検索結果に表示されます'
              }
            </div>
          </div>

          {/* パラメータ検索・履歴ボタン */}
          <div className="flex flex-col sm:flex-row gap-3 mb-4">
            <button
              onClick={searchProblemsByFilters}
              className="px-4 py-2 bg-mongene-green text-white rounded-lg hover:brightness-110 transition-all"
            >
              📊 現在の条件で検索 ({searchMatchType === 'partial' ? '部分一致' : '完全一致'})
            </button>
            <button
              onClick={searchProblemsByKeywordAndFilters}
              className="px-4 py-2 bg-purple-500 text-white rounded-lg hover:brightness-110 transition-all"
            >
              🔍📊 キーワード+条件で検索
            </button>
            <button
              onClick={fetchProblemHistory}
              className="px-4 py-2 bg-mongene-muted text-white rounded-lg hover:brightness-110 transition-all"
            >
              📚 履歴表示
            </button>
          </div>
          
          {/* 現在の表示モード */}
          <div className="text-sm text-mongene-muted">
            {isSearchMode ? (
              <div className="flex items-center gap-2">
                <span>🔍 検索結果: "{searchKeyword}" ({searchResults.length}件)</span>
                <button 
                  onClick={() => {
                    setIsSearchMode(false);
                    setSearchKeyword('');
                    fetchProblemHistory();
                  }}
                  className="text-mongene-blue hover:underline"
                >
                  履歴に戻る
                </button>
              </div>
            ) : (
              <span>📚 問題履歴 ({problems.length}件)</span>
            )}
          </div>
        </div>
        
        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-7" aria-label="問題一覧">
          {/* 検索モードの場合は検索結果を表示、そうでなければ履歴を表示 */}
          {(isSearchMode ? searchResults : problems).map((problem) => (
            <ProblemCard
              key={problem.id}
              id={problem.id}
              title={problem.title}
              content={problem.content}
              imageBase64={problem.imageBase64}
              onPreview={handlePreview}
              onPrint={handlePrint}
            />
          ))}
        </section>
        
        {/* ユーザー情報表示 */}
        {userInfo && (
          <div className="mb-6 p-4 bg-white/10 backdrop-blur-sm rounded-xl border border-white/20">
            <div className="flex items-center justify-between">
              <div className="text-mongene-ink">
                <span className="font-medium">塾コード: {userInfo.school_code}</span>
                <span className="ml-4">
                  問題生成回数: {userInfo.problem_generation_count}/
                  {userInfo.problem_generation_limit === -1 ? '無制限' : userInfo.problem_generation_limit}
                </span>
              </div>
              {isGenerationLimitReached() && (
                <div className="text-red-600 font-bold">
                  ⚠️ 生成上限に達しました
                </div>
              )}
            </div>
          </div>
        )}

        {/* 5段階生成システムの選択UI */}
        <div className="mb-6 p-4 bg-white/10 backdrop-blur-sm rounded-xl border border-white/20">
          <h3 className="text-lg font-bold text-mongene-ink mb-4">🚀 問題生成方式</h3>
          
          {/* 生成モード選択 */}
          <div className="mb-4">
            <div className="flex flex-col gap-3 mb-3">
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  name="generationMode"
                  value="single"
                  checked={generationMode === 'single'}
                  onChange={(e) => setGenerationMode(e.target.value as 'single' | 'two-stage' | 'five-stage')}
                  className="text-mongene-blue"
                />
                <span className="text-sm font-medium text-mongene-ink">従来方式（1回のAPI呼び出し）</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  name="generationMode"
                  value="two-stage"
                  checked={generationMode === 'two-stage'}
                  onChange={(e) => setGenerationMode(e.target.value as 'single' | 'two-stage' | 'five-stage')}
                  className="text-mongene-blue"
                />
                <span className="text-sm font-medium text-mongene-ink">2段階生成（高精度）</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  name="generationMode"
                  value="five-stage"
                  checked={generationMode === 'five-stage'}
                  onChange={(e) => setGenerationMode(e.target.value as 'single' | 'two-stage' | 'five-stage')}
                  className="text-mongene-blue"
                />
                <span className="text-sm font-medium text-mongene-ink">🆕 5段階生成（最高精度）</span>
              </label>
            </div>
            <div className="text-xs text-mongene-muted">
              {generationMode === 'single' 
                ? '問題文と解答を1回のAPI呼び出しで生成します（従来の方式）'
                : generationMode === 'two-stage'
                  ? '1回目で問題文・図形を生成し、2回目で解答手順・数値計算を生成します（トークン制限対応）'
                  : '5段階に分けて最高精度で生成します：①問題文→②図形→③解答手順→④数値計算→⑤最終解説'
              }
            </div>
          </div>

          {/* 2段階生成の場合の詳細UI */}
          {generationMode === 'two-stage' && (
            <div className="border-t border-white/20 pt-4">
              <h4 className="font-bold text-mongene-ink mb-3">📋 2段階生成プロセス</h4>
              
              {/* 現在の状態表示 */}
              <div className="mb-4 p-3 bg-white/5 rounded-lg">
                <div className="flex items-center gap-4 mb-2">
                  <div className={`flex items-center gap-2 ${isFirstStageComplete ? 'text-green-600' : 'text-mongene-muted'}`}>
                    <span>{isFirstStageComplete ? '✅' : '⏸️'}</span>
                    <span className="text-sm font-medium">1回目: 問題文・図形生成</span>
                  </div>
                  <div className={`flex items-center gap-2 ${secondStageResult ? 'text-green-600' : 'text-mongene-muted'}`}>
                    <span>{secondStageResult ? '✅' : '⏸️'}</span>
                    <span className="text-sm font-medium">2回目: 解答手順・数値計算</span>
                  </div>
                </div>
                
                {firstStageResult && (
                  <div className="text-xs text-mongene-ink mb-2">
                    📝 問題文生成済み ({firstStageResult.problemText.length}文字)
                    {firstStageResult.imageBase64 && ' | 🖼️ 図形生成済み'}
                    {firstStageResult.geometryCode && ' | 🐍 図形コード生成済み'}
                  </div>
                )}
                
                {secondStageResult && (
                  <div className="text-xs text-mongene-ink">
                    📚 解答手順生成済み ({secondStageResult.solutionSteps.length}文字)
                    {secondStageResult.calculationProgram && ' | 🧮 計算プログラム生成済み'}
                  </div>
                )}
              </div>

              {/* 2段階生成の個別実行ボタン */}
              <div className="flex flex-col sm:flex-row gap-3 mb-4">
                <button
                  onClick={handleGenerateFirstStage}
                  disabled={isLoading || isGenerationLimitReached()}
                  className={`px-4 py-2 rounded-lg font-medium transition-all ${
                    isLoading || isGenerationLimitReached()
                      ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                      : 'bg-blue-500 text-white hover:brightness-110'
                  }`}
                >
                  1️⃣ 問題文・図形を生成
                </button>
                
                <button
                  onClick={handleGenerateSecondStage}
                  disabled={!isFirstStageComplete || isLoading}
                  className={`px-4 py-2 rounded-lg font-medium transition-all ${
                    !isFirstStageComplete || isLoading
                      ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                      : 'bg-purple-500 text-white hover:brightness-110'
                  }`}
                >
                  2️⃣ 解答手順・計算を生成
                </button>
                
                <button
                  onClick={() => {
                    setFirstStageResult(null);
                    setSecondStageResult(null);
                    setIsFirstStageComplete(false);
                    setGenerationLogs('');
                  }}
                  className="px-4 py-2 bg-gray-500 text-white rounded-lg hover:brightness-110 transition-all"
                >
                  🔄 リセット
                </button>
              </div>
              
              {/* ログ表示 */}
              {generationLogs && (
                <div className="mb-4">
                  <button
                    onClick={() => setShowLogs(!showLogs)}
                    className="flex items-center gap-2 text-mongene-blue hover:underline text-sm font-medium mb-2"
                  >
                    <span>{showLogs ? '🔽' : '▶️'}</span>
                    生成ログを{showLogs ? '非表示' : '表示'}
                  </button>
                  
                  {showLogs && (
                    <div className="p-3 bg-gray-900 text-green-400 rounded-lg text-xs font-mono max-h-60 overflow-y-auto">
                      <pre className="whitespace-pre-wrap">{generationLogs}</pre>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}

          {/* 5段階生成の場合の詳細UI */}
          {generationMode === 'five-stage' && (
            <div className="border-t border-white/20 pt-4">
              <h4 className="font-bold text-mongene-ink mb-3">🔥 5段階生成プロセス（最高精度）</h4>
              
              {/* 進捗バー */}
              <div className="mb-4 p-3 bg-white/5 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-mongene-ink">進捗: Stage {currentStage}/5</span>
                  <span className="text-sm text-mongene-muted">{stageProgress.toFixed(0)}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-gradient-to-r from-blue-500 to-purple-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${stageProgress}%` }}
                  ></div>
                </div>
              </div>

              {/* 各段階の状態表示 */}
              <div className="mb-4 p-3 bg-white/5 rounded-lg">
                <div className="grid grid-cols-1 md:grid-cols-5 gap-2 mb-3">
                  <div className={`flex items-center gap-1 text-xs ${fiveStageResults.stage1 ? 'text-green-600' : currentStage === 1 ? 'text-blue-500' : 'text-mongene-muted'}`}>
                    <span>{fiveStageResults.stage1 ? '✅' : currentStage === 1 ? '⏳' : '⏸️'}</span>
                    <span>Stage1: 問題文</span>
                  </div>
                  <div className={`flex items-center gap-1 text-xs ${fiveStageResults.stage2 ? 'text-green-600' : currentStage === 2 ? 'text-blue-500' : 'text-mongene-muted'}`}>
                    <span>{fiveStageResults.stage2 ? '✅' : currentStage === 2 ? '⏳' : '⏸️'}</span>
                    <span>Stage2: 図形</span>
                  </div>
                  <div className={`flex items-center gap-1 text-xs ${fiveStageResults.stage3 ? 'text-green-600' : currentStage === 3 ? 'text-blue-500' : 'text-mongene-muted'}`}>
                    <span>{fiveStageResults.stage3 ? '✅' : currentStage === 3 ? '⏳' : '⏸️'}</span>
                    <span>Stage3: 解答手順</span>
                  </div>
                  <div className={`flex items-center gap-1 text-xs ${fiveStageResults.stage4 ? 'text-green-600' : currentStage === 4 ? 'text-blue-500' : 'text-mongene-muted'}`}>
                    <span>{fiveStageResults.stage4 ? '✅' : currentStage === 4 ? '⏳' : '⏸️'}</span>
                    <span>Stage4: 数値計算</span>
                  </div>
                  <div className={`flex items-center gap-1 text-xs ${fiveStageResults.stage5 ? 'text-green-600' : currentStage === 5 ? 'text-blue-500' : 'text-mongene-muted'}`}>
                    <span>{fiveStageResults.stage5 ? '✅' : currentStage === 5 ? '⏳' : '⏸️'}</span>
                    <span>Stage5: 最終解説</span>
                  </div>
                </div>
                
                {/* 各段階の詳細情報 */}
                {fiveStageResults.stage1 && (
                  <div className="text-xs text-mongene-ink mb-1">
                    📝 問題文生成完了 ({fiveStageResults.stage1.problemText.length}文字)
                  </div>
                )}
                {fiveStageResults.stage2 && (
                  <div className="text-xs text-mongene-ink mb-1">
                    🖼️ 図形生成完了 {fiveStageResults.stage2.geometryCode && '| 🐍 図形コード生成済み'}
                  </div>
                )}
                {fiveStageResults.stage3 && (
                  <div className="text-xs text-mongene-ink mb-1">
                    📚 解答手順生成完了 ({fiveStageResults.stage3.solutionSteps.length}文字)
                  </div>
                )}
                {fiveStageResults.stage4 && (
                  <div className="text-xs text-mongene-ink mb-1">
                    🧮 数値計算完了 {fiveStageResults.stage4.calculationProgram && '| 🐍 計算プログラム実行済み'}
                  </div>
                )}
                {fiveStageResults.stage5 && (
                  <div className="text-xs text-mongene-ink">
                    ✨ 最終解説生成完了 ({fiveStageResults.stage5.finalExplanation.length}文字)
                  </div>
                )}
              </div>

              {/* 5段階生成の実行ボタン */}
              <div className="flex flex-col sm:flex-row gap-3 mb-4">
                <button
                  onClick={() => handleGenerateFiveStage()}
                  disabled={isLoading || isGenerationLimitReached()}
                  className={`px-4 py-2 rounded-lg font-medium transition-all ${
                    isLoading || isGenerationLimitReached()
                      ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                      : 'bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:from-blue-600 hover:to-purple-700'
                  }`}
                >
                  🔥 5段階生成を実行
                </button>
                
                <button
                  onClick={() => {
                    setFiveStageResults({});
                    setCurrentStage(0);
                    setStageProgress(0);
                    setGenerationLogs('');
                  }}
                  className="px-4 py-2 bg-gray-500 text-white rounded-lg hover:brightness-110 transition-all"
                >
                  🔄 リセット
                </button>
              </div>
              
              {/* ログ表示 */}
              {generationLogs && (
                <div className="mb-4">
                  <button
                    onClick={() => setShowLogs(!showLogs)}
                    className="flex items-center gap-2 text-mongene-blue hover:underline text-sm font-medium mb-2"
                  >
                    <span>{showLogs ? '🔽' : '▶️'}</span>
                    5段階生成ログを{showLogs ? '非表示' : '表示'}
                  </button>
                  
                  {showLogs && (
                    <div className="p-3 bg-gray-900 text-green-400 rounded-lg text-xs font-mono max-h-80 overflow-y-auto">
                      <pre className="whitespace-pre-wrap">{generationLogs}</pre>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        <div className="flex flex-col items-center">
          {/* 上限に達した場合の専用メッセージ */}
          {isGenerationLimitReached() && (
            <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-center max-w-md">
              <div className="font-bold mb-2">🚫 問題生成上限に達しました</div>
              <div className="text-sm">
                問題生成回数の上限（{userInfo?.problem_generation_limit}回）に達したため、
                これ以上問題を生成することはできません。
              </div>
            </div>
          )}
          
          <button
            className={`appearance-none border-0 rounded-xl px-5 py-3 font-bold transition-all focus:outline-none focus:ring-3 focus:ring-offset-2 ${
              isGenerationLimitReached()
                ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                : 'bg-mongene-green text-mongene-ink shadow-lg hover:brightness-98 hover:-translate-y-0.5 cursor-pointer focus:ring-mongene-green/25'
            }`}
            type="button"
            onClick={handleGenerate}
            disabled={isGenerationLimitReached()}
          >
            {isGenerationLimitReached() 
              ? '生成上限に達しました' 
              : generationMode === 'two-stage' 
                ? '🚀 2段階生成を実行'
                : generationMode === 'five-stage'
                  ? '🔥 5段階生成を実行'
                  : '問題を新しく生成'
            }
          </button>
        </div>
      </div>

      <ProblemPreviewModal
        isOpen={previewModal.isOpen}
        onClose={() => setPreviewModal({ isOpen: false, problemId: '', problemTitle: '', problemContent: '', imageBase64: undefined, solutionText: undefined })}
        problemId={previewModal.problemId}
        problemTitle={previewModal.problemTitle}
        problemContent={previewModal.problemContent}
        imageBase64={previewModal.imageBase64}
        solutionText={previewModal.solutionText}
        onUpdate={(updatedData) => {
          // 問題リストを更新
          setProblems(prev => prev.map(problem => 
            problem.id === previewModal.problemId 
              ? { 
                  ...problem, 
                  content: updatedData.content, 
                  solution: updatedData.solution,
                  imageBase64: updatedData.imageBase64 
                }
              : problem
          ));

          // 検索結果も更新
          if (isSearchMode) {
            setSearchResults(prev => prev.map(problem => 
              problem.id === previewModal.problemId 
                ? { 
                    ...problem, 
                    content: updatedData.content, 
                    solution: updatedData.solution,
                    imageBase64: updatedData.imageBase64 
                  }
                : problem
            ));
          }

          // プレビューモーダルの状態も更新
          setPreviewModal(prev => ({
            ...prev,
            problemContent: updatedData.content,
            solutionText: updatedData.solution,
            imageBase64: updatedData.imageBase64,
          }));

          console.log('✅ Frontend state updated with:', updatedData);
        }}
      />

      <LoadingModal
        isOpen={isLoading}
        message={generationMode === 'five-stage' ? '5段階生成を実行中...' : '問題を生成しています...'}
        showProgress={generationMode === 'five-stage'}
        currentStage={currentStage}
        maxStages={5}
        stageProgress={stageProgress}
        stageMessage={
          generationMode === 'five-stage' ? 
            currentStage === 1 ? '📝 問題文を生成中...' :
            currentStage === 2 ? '🖼️ 図形を生成中...' :
            currentStage === 3 ? '📚 解答手順を生成中...' :
            currentStage === 4 ? '🧮 数値計算を実行中...' :
            currentStage === 5 ? '✨ 最終解説を生成中...' :
            '' : undefined
        }
      />

    </div>
  );
}
