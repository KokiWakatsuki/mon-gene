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
  const [previewModal, setPreviewModal] = useState<{ isOpen: boolean; problemId: string; problemTitle: string; problemContent?: string; imageBase64?: string }>({
    isOpen: false,
    problemId: '',
    problemTitle: '',
    problemContent: '',
    imageBase64: undefined,
  });
  const [isLoading, setIsLoading] = useState(false);
  const [problems, setProblems] = useState<Array<{ id: string; title: string; content: string; imageBase64?: string }>>([]);

  // 認証チェック
  useEffect(() => {
    const checkAuth = () => {
      const token = localStorage.getItem('authToken');
      if (!token) {
        window.location.href = '/login';
        return;
      }
      
      // TODO: バックエンドでトークンの有効性を確認する場合
      // 現在はトークンの存在のみをチェック
      setIsAuthenticated(true);
      setIsCheckingAuth(false);
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
        const solutionHtml = (problem as any).solution 
          ? `<div style="page-break-before: always;">
               <h1>解答・解説</h1>
               <div class="content">${(problem as any).solution}</div>
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

  const handleGenerate = async () => {
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
        
        const response = await fetch(API_CONFIG.BACKEND_API_URL, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
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
        console.log('🔍 data.ImageBase64:', data.ImageBase64);
        console.log('🔍 data.image_base64:', data.image_base64);
        console.log('🔍 ImageBase64 exists:', !!(data.ImageBase64 || data.image_base64));
        console.log('🔍 ImageBase64 length:', (data.ImageBase64 || data.image_base64 || '').length);
        
        // 画像データの処理
        const imageBase64 = data.ImageBase64 || data.image_base64;
        const finalImageBase64 = (imageBase64 && imageBase64.length > 0) ? imageBase64 : undefined;
        
        console.log('🔍 Final imageBase64 for problem:', !!finalImageBase64);
        console.log('🔍 Final imageBase64 length:', finalImageBase64?.length || 0);
        
        // 解答・解説データの処理
        const solutionText = data.solution || '';
        
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
        
        // ローディングを終了
        setIsLoading(false);
        
        // 生成された問題のプレビューを自動的に表示（画像データを含む）
        setPreviewModal({
          isOpen: true,
          problemId: newProblemId,
          problemTitle: problemTitle,
          problemContent: generatedContent,
          imageBase64: finalImageBase64,
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
        });
      }
      
    } catch (error) {
      setIsLoading(false);
      console.error('問題生成エラー:', error);
      const errorMessage = error instanceof Error ? error.message : '不明なエラーが発生しました';
      alert(`問題生成に失敗しました: ${errorMessage}`);
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
        
        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-7" aria-label="問題一覧">
          {problems.map((problem) => (
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
        
        <div className="flex justify-center">
          <button
            className="appearance-none border-0 rounded-xl px-5 py-3 font-bold cursor-pointer bg-mongene-green text-mongene-ink shadow-lg hover:brightness-98 hover:-translate-y-0.5 transition-all focus:outline-none focus:ring-3 focus:ring-mongene-green/25 focus:ring-offset-2"
            type="button"
            onClick={handleGenerate}
          >
            問題を新しく生成
          </button>
        </div>
      </div>

      <ProblemPreviewModal
        isOpen={previewModal.isOpen}
        onClose={() => setPreviewModal({ isOpen: false, problemId: '', problemTitle: '', problemContent: '', imageBase64: undefined })}
        problemId={previewModal.problemId}
        problemTitle={previewModal.problemTitle}
        problemContent={previewModal.problemContent}
        imageBase64={previewModal.imageBase64}
      />

      <LoadingModal
        isOpen={isLoading}
        message="問題を生成しています..."
      />

    </div>
  );
}
