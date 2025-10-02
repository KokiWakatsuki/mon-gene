'use client';

import React, { useState } from 'react';
import Header from './components/Header';
import Tabs from './components/Tabs';
import Filters from './components/Filters';
import ProblemCard from './components/ProblemCard';
import BackgroundShapes from './components/BackgroundShapes';
import ProblemPreviewModal from './components/ProblemPreviewModal';
import LoadingModal from './components/LoadingModal';
import { API_CONFIG } from '../config/api';

export default function Home() {
  const [activeSubject, setActiveSubject] = useState('数学');
  const [selectedFilters, setSelectedFilters] = useState<Record<string, string[]>>({});
  const [previewModal, setPreviewModal] = useState<{ isOpen: boolean; problemId: string; problemTitle: string; problemContent?: string }>({
    isOpen: false,
    problemId: '',
    problemTitle: '',
    problemContent: '',
  });
  const [isLoading, setIsLoading] = useState(false);
  const [problems, setProblems] = useState([
    { id: '1', title: 'A4 プレビュー 1', content: 'これはサンプル問題です。' },
    { id: '2', title: 'A4 プレビュー 2', content: 'これはサンプル問題です。' },
    { id: '3', title: 'A4 プレビュー 3', content: 'これはサンプル問題です。' },
    { id: '4', title: 'A4 プレビュー 4', content: 'これはサンプル問題です。' },
  ]);

  const subjects = ['数学', '英語', '国語'];

  // 科目別の単元データ
  const subjectUnits = {
    '数学': [
      { label: '式の計算', value: 'calculation' },
      { label: '図形', value: 'geometry' },
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
      });
    }
  };

  const handlePrint = (id: string) => {
    const problem = problems.find(p => p.id === id);
    if (problem) {
      // 印刷用の新しいウィンドウを開く
      const printWindow = window.open('', '_blank');
      if (printWindow) {
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
              }
              @media print {
                body { margin: 0; }
                h1 { page-break-after: avoid; }
              }
            </style>
          </head>
          <body>
            <h1>${problem.title}</h1>
            <div class="content">${problem.content || ''}</div>
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
        
        console.log('バックエンドAPIレスポンス:', data);
      } else {
        // テスト版（ダミーデータ）
        console.log('テスト版を使用しています');
        generatedContent = `これはテスト用の問題です。\n\n選択された条件:\n${prompt}\n\n実際のAPI版では、ここにClaude AIが生成した問題が表示されます。`;
        problemTitle = `テスト問題 ${problems.length + 1}`;
      }
      
      // 新しい問題を追加
      const newProblemId = String(problems.length + 1);
      const newProblem = {
        id: newProblemId,
        title: problemTitle,
        content: generatedContent,
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
      });
      
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
        onClose={() => setPreviewModal({ isOpen: false, problemId: '', problemTitle: '', problemContent: '' })}
        problemId={previewModal.problemId}
        problemTitle={previewModal.problemTitle}
        problemContent={previewModal.problemContent}
      />

      <LoadingModal
        isOpen={isLoading}
        message="問題を生成しています..."
      />

    </div>
  );
}
