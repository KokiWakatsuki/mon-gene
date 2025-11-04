'use client';

import React, { useState, useEffect } from 'react';
import Header from '../../components/layout/Header';
import Tabs from '../../components/features/problems/Tabs';
import OpinionProfileSettings from '../../components/features/problems/OpinionProfileSettings';
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
  
  // 生成システム用の状態（デフォルトをインクリメンタルモデルに変更）
  const [generationMode, setGenerationMode] = useState<'single' | 'five-stage'>('five-stage');
  
  // opinion.md基準での問題生成モード（Ver.2に移行）
  const [useOpinionCriteria] = useState<boolean>(true);
  // デフォルト値を定義（検索時の比較用）
  const defaultOpinionProfileV2 = {
    problem_text_length: 100,
    sub_problem_text_length: 50,
    given_values_count: 3,
    sub_problem_count: 2,
    sub_problem_types: ['長さを求める'],
    solid_composition: '単一の立体',
    answer_formats: ['整数'],
    answer_units: ['cm'],
    uses_auxiliary_points: false,
    setup_units: ['直方体'],
    solution_units: ['三平方の定理'],
    total_vertices: 8,
    has_moving_point: false,
    figure_values_count: 3,
    solution_steps: 3,
    has_logical_branching: false,
    theorem_count: 2,
    requires_multi_unit_integration: false,
    has_irrelevant_info: false,
  };

  const [opinionProfileV2, setOpinionProfileV2] = useState<{
    problem_text_length: number;
    sub_problem_text_length: number;
    given_values_count: number;
    sub_problem_count: number;
    sub_problem_types: string[];
    solid_composition: string;
    answer_formats: string[];
    answer_units: string[];
    uses_auxiliary_points: boolean;
    setup_units: string[];
    solution_units: string[];
    total_vertices: number;
    has_moving_point: boolean;
    figure_values_count: number;
    solution_steps: number;
    has_logical_branching: boolean;
    theorem_count: number;
    requires_multi_unit_integration: boolean;
    has_irrelevant_info: boolean;
  }>(defaultOpinionProfileV2);
  
  // 5段階生成システム専用の状態（新しいプロセスに対応）
  const [fiveStageResults, setFiveStageResults] = useState<{
    stage1?: { solutionProcess: string; log: string };
    stage2?: { completeProblem: string; log: string };
    stage3?: { calculationProgram: string; calculationResults: string; log: string };
    stage4?: { finalExplanation: string; log: string };
    stage5?: { geometryCode: string; imageBase64: string; log: string };
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

  const getFilterGroups = () => {
    // opinion.md基準を使用する場合は異なるフィルターを表示
    if (useOpinionCriteria) {
      return [
        {
          label: '出題分野コード',
          options: [
            { label: '1: 関数', value: '1' },
            { label: '2: 平面図形', value: '2' },
            { label: '3: 空間図形', value: '3' },
            { label: '4: 確率・統計', value: '4' },
            { label: '5: 数と式', value: '5' },
            { label: '6: 融合問題', value: '6' },
          ],
          allowMultiple: false,
        },
        {
          label: 'コアスキルレベル',
          options: [
            { label: 'Lv1: 基本的知識', value: '1' },
            { label: 'Lv2: 応用的知識', value: '2' },
            { label: 'Lv3: 手順の遂行能力', value: '3' },
            { label: 'Lv4: 計算の実行精度', value: '4' },
            { label: 'Lv5: 標準的なモデル化能力', value: '5' },
            { label: 'Lv6: 複雑な情報統制・モデル化能力', value: '6' },
            { label: 'Lv7: 緻密な論理構築能力', value: '7' },
            { label: 'Lv8: 高度な空間認識能力', value: '8' },
            { label: 'Lv9: 独創的な着眼力', value: '9' },
            { label: 'Lv10: 高次元の発想力', value: '10' },
          ],
          allowMultiple: false,
        },
        {
          label: '読解・設定の複雑度',
          options: [
            { label: 'Lv1: 図と数式のみ', value: '1' },
            { label: 'Lv2: 短い補足文', value: '2' },
            { label: 'Lv3: 1段落程度の文章', value: '3' },
            { label: 'Lv4: 複数条件の整理', value: '4' },
            { label: 'Lv5: やや長文', value: '5' },
            { label: 'Lv6: 会話文形式', value: '6' },
            { label: 'Lv7: ストーリー形式', value: '7' },
            { label: 'Lv8: 動的で複雑な設定', value: '8' },
            { label: 'Lv9: 複雑なストーリー・独自ルール', value: '9' },
            { label: 'Lv10: 極めて複雑な独自ルール', value: '10' },
          ],
          allowMultiple: false,
        },
        {
          label: '設問の誘導性',
          options: [
            { label: 'Lv1: 完全な無誘導', value: '1' },
            { label: 'Lv2: 関連性が薄い小問', value: '2' },
            { label: 'Lv3: 独立した思考プロセス', value: '3' },
            { label: 'Lv4: 状況理解の助け程度', value: '4' },
            { label: 'Lv5: 標準的な誘導', value: '5' },
            { label: 'Lv6: 重要な要素として機能', value: '6' },
            { label: 'Lv7: 解法のテンプレート', value: '7' },
            { label: 'Lv8: 直接的な利用', value: '8' },
            { label: 'Lv9: 明確な連鎖構造', value: '9' },
            { label: 'Lv10: 完全なレール形式', value: '10' },
          ],
          allowMultiple: false,
        },
        {
          label: '総合難易度スコア',
          options: [
            { label: 'Lv1-4: 基礎応用', value: '1-4' },
            { label: 'Lv5-8: 標準的難問', value: '5-8' },
            { label: 'Lv9-12: 上位校レベル', value: '9-12' },
            { label: 'Lv13-16: 最難関校レベル', value: '13-16' },
            { label: 'Lv17-18: 超難関', value: '17-18' },
            { label: 'Lv19: 全国レベル', value: '19' },
            { label: 'Lv20: 捨て問', value: '20' },
          ],
          allowMultiple: false,
        },
      ];
    }

    // 従来のフィルター
    return [
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
  };

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
      // LaTeX数式レンダリング関数（ProblemPreviewModalと同じ）
      const renderLatexToHtml = (latex: string): string => {
        return latex
          .replace(/\\frac\{([^}]+)\}\{([^}]+)\}/g, '<span class="math-fraction-block"><span class="numerator">$1</span><span class="fraction-line-block"></span><span class="denominator">$2</span></span>')
          .replace(/\\sqrt\{([^}]+)\}/g, '<span class="math-symbol">√<span class="sqrt-content">$1</span></span>')
          .replace(/\\vec\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
          .replace(/\\overrightarrow\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
          .replace(/\\times/g, '×')
          .replace(/\\cdot/g, '·')
          .replace(/\\pi/g, 'π')
          .replace(/\\infty/g, '∞')
          .replace(/\\pm/g, '±')
          .replace(/\\leq/g, '≤')
          .replace(/\\geq/g, '≥')
          .replace(/\\neq/g, '≠')
          .replace(/\\approx/g, '≈')
          .replace(/\\rightarrow/g, '→')
          .replace(/\\leftarrow/g, '←');
      };

      const renderAdvancedMathSymbols = (text: string): string => {
        if (!text) return '';
        
        const mathPlaceholders: string[] = [];
        let placeholderIndex = 0;
        
        let processedText = text
          .replace(/\$\$([\s\S]*?)\$\$/g, (match, formula) => {
            const placeholder = `__MATH_BLOCK_${placeholderIndex}__`;
            mathPlaceholders[placeholderIndex] = `<div class="math-block">${renderLatexToHtml(formula.trim())}</div>`;
            placeholderIndex++;
            return placeholder;
          })
          .replace(/\$([^$\n]+)\$/g, (match, formula) => {
            const placeholder = `__MATH_INLINE_${placeholderIndex}__`;
            mathPlaceholders[placeholderIndex] = `<span class="math-inline">${renderLatexToHtml(formula.trim())}</span>`;
            placeholderIndex++;
            return placeholder;
          });
        
        processedText = processedText
          .replace(/\\overrightarrow\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
          .replace(/([A-Z]{1,3})⃗/g, '<span class="math-vector">$1→</span>')
          .replace(/√(\d+)/g, '<span class="math-symbol">√$1</span>')
          .replace(/√\(([^)]+)\)/g, '<span class="math-symbol">√($1)</span>')
          .replace(/√([a-zA-Z]+)/g, '<span class="math-symbol">√$1</span>')
          .replace(/(\w+)²/g, '$1<sup>2</sup>')
          .replace(/(\w+)³/g, '$1<sup>3</sup>')
          .replace(/(\w+)⁴/g, '$1<sup>4</sup>')
          .replace(/(\w+)⁵/g, '$1<sup>5</sup>')
          .replace(/∠([A-Z]+)/g, '<span class="math-symbol">∠$1</span>')
          .replace(/(\d+)\/(\d+)/g, '<span class="math-fraction"><sup>$1</sup>/<sub>$2</sub></span>')
          .replace(/×/g, '<span class="math-symbol">×</span>')
          .replace(/÷/g, '<span class="math-symbol">÷</span>')
          .replace(/°/g, '<span class="math-symbol">°</span>')
          .replace(/π/g, '<span class="math-symbol">π</span>')
          .replace(/∞/g, '<span class="math-symbol">∞</span>')
          .replace(/±/g, '<span class="math-symbol">±</span>')
          .replace(/≤/g, '<span class="math-symbol">≤</span>')
          .replace(/≥/g, '<span class="math-symbol">≥</span>')
          .replace(/≠/g, '<span class="math-symbol">≠</span>')
          .replace(/≈/g, '<span class="math-symbol">≈</span>')
          .replace(/≅/g, '<span class="math-symbol">≅</span>')
          .replace(/∽/g, '<span class="math-symbol">∽</span>')
          .replace(/→/g, '<span class="math-symbol">→</span>')
          .replace(/←/g, '<span class="math-symbol">←</span>');
        
        mathPlaceholders.forEach((replacement, index) => {
          processedText = processedText.replace(`__MATH_BLOCK_${index}__`, replacement);
          processedText = processedText.replace(`__MATH_INLINE_${index}__`, replacement);
        });
        
        return processedText;
      };

      const renderBasicMarkdown = (text: string): string => {
        return text
          .replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold mb-2 text-gray-700">$1</h3>')
          .replace(/^## (.*$)/gim, '<h2 class="text-xl font-bold mb-3 text-gray-800">$1</h2>')
          .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mb-4 text-gray-900">$1</h1>')
          .replace(/\*\*(.*?)\*\*/g, '<strong class="font-bold text-gray-900">$1</strong>')
          .replace(/\*(.*?)\*/g, '<em class="italic">$1</em>')
          .replace(/`([^`]+)`/g, '<code class="bg-gray-100 px-1 py-0.5 rounded text-sm font-mono">$1</code>')
          .replace(/\n/g, '<br />');
      };

      const processContent = (content: string): string => {
        return renderBasicMarkdown(renderAdvancedMathSymbols(content));
      };

      // 印刷用の新しいウィンドウを開く
      const printWindow = window.open('', '_blank');
      if (printWindow) {
        const processedContent = processContent(problem.content || '');
        const processedSolution = problem.solution ? processContent(problem.solution) : '';
        
        const imageHtml = problem.imageBase64
          ? `<div style="text-align: center; margin: 20px 0;">
               <img src="data:image/png;base64,${problem.imageBase64}"
                    style="max-width: 100%; height: auto; border: 1px solid #ddd;"
                    alt="問題図形" />
             </div>`
          : '';
        
        // 解答・解説がある場合は別ページに追加
        const solutionHtml = processedSolution
          ? `<div style="page-break-before: always;">
               <h1>解答・解説</h1>
               <div class="content">${processedSolution}</div>
             </div>`
          : '';
        
        printWindow.document.write(`
          <!DOCTYPE html>
          <html>
          <head>
            <title>${problem.title}</title>
            <style>
              body {
                font-family: 'Times New Roman', Arial, sans-serif;
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
                font-size: 14px;
                margin-bottom: 20px;
              }
              .math-symbol {
                font-family: 'Times New Roman', serif;
                font-weight: normal;
                color: #1f2937;
                font-size: 1.1em;
              }
              .math-vector {
                font-family: 'Times New Roman', serif;
                font-weight: bold;
                color: #1f2937;
                font-size: 1.05em;
              }
              .math-fraction {
                display: inline-block;
                vertical-align: middle;
                font-family: 'Times New Roman', serif;
                margin: 0 2px;
              }
              .fraction-line {
                font-size: 1.2em;
                color: #374151;
              }
              .math-fraction-block {
                display: inline-flex;
                flex-direction: column;
                vertical-align: middle;
                text-align: center;
                font-family: 'Times New Roman', serif;
                margin: 0 4px;
                align-items: center;
              }
              .numerator {
                display: block;
                font-size: 0.9em;
                line-height: 1;
                padding: 0 2px;
              }
              .fraction-line-block {
                display: block;
                border-top: 1.5px solid #374151;
                margin: 1px 0;
                width: 100%;
                min-width: 20px;
                height: 0;
                line-height: 0;
              }
              .fraction-line-block::before {
                content: '';
                display: block;
              }
              .denominator {
                display: block;
                font-size: 0.9em;
                line-height: 1;
                padding: 0 2px;
              }
              .sqrt-content {
                border-top: 1px solid #374151;
                padding: 0 2px;
              }
              .math-block {
                display: block;
                text-align: center;
                margin: 12px 0;
                padding: 8px;
                background-color: #f9fafb;
                border: 1px solid #e5e7eb;
                border-radius: 4px;
                font-family: 'Times New Roman', serif;
                font-size: 1.1em;
              }
              .math-inline {
                font-family: 'Times New Roman', serif;
                font-size: 1.05em;
                color: #1f2937;
              }
              sup {
                font-size: 0.75em;
                vertical-align: super;
                line-height: 0;
              }
              sub {
                font-size: 0.75em;
                vertical-align: sub;
                line-height: 0;
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
            <div class="content">${processedContent}</div>
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


  // 5段階生成システムの関数（SSE使用）
  const handleGenerateFiveStage = async () => {
    // 上限チェック
    if (isGenerationLimitReached()) {
      alert(`問題生成回数の上限（${userInfo?.problem_generation_limit}回）に達しました。これ以上問題を生成することはできません。`);
      return;
    }

    // opinion profile v2の必須項目チェック
    if (opinionProfileV2.sub_problem_count === 0) {
      alert('小問の数を設定してください');
      return;
    }
    
    setIsLoading(true);
    setFiveStageResults({});
    setCurrentStage(1);
    setStageProgress(0);
    
    try {
      const prompt = createPromptFromFilters();
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません。再度ログインしてください。');
      }

      console.log('🚀 [FiveStage] 5段階生成プロセス開始（SSE使用）');
      
      // SSEを使用してリアルタイム進捗を取得
      const requestBody = JSON.stringify({
        prompt: prompt,
        subject: activeSubject,
        filters: selectedFilters,
        opinion_profile_v2: createOpinionProfileFromFilters()
      });

      // fetchでSSE接続（EventSourceはPOSTをサポートしていないため）
      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-problem-five-stage-sse`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: requestBody
      });

      if (!response.ok) {
        throw new Error(`SSE接続エラー: ${response.status} ${response.statusText}`);
      }

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let finalResult: any = null;

      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            try {
              const event = JSON.parse(data);
              console.log('📨 [SSE] Event:', event);

              if (event.type === 'stage_complete') {
                // ステージ完了：次のステージの開始位置に強制移行
                console.log(`✅ [SSE] Stage ${event.stage} 完了`);
                setCurrentStage(event.stage + 1);
              } else if (event.type === 'stage_start') {
                console.log(`🚀 [SSE] Stage ${event.stage} 開始`);
                setCurrentStage(event.stage);
              } else if (event.type === 'error') {
                console.error('❌ [SSE] Error:', event.error);
                throw new Error(event.error);
              } else if (event.type === 'complete') {
                console.log('✅ [SSE] 完了');
                finalResult = JSON.parse(event.message);
                setStageProgress(99);
                setCurrentStage(5);
              }
            } catch (e) {
              console.error('❌ [SSE] Parse error:', e);
            }
          }
        }
      }

      if (!finalResult) {
        throw new Error('5段階生成の結果を取得できませんでした');
      }

      console.log('📦 [FiveStage] Final result:', finalResult);

      // 結果を問題リストに追加
      const problemTitle = `問題 ${problems.length + 1}`;
      const newProblemId = String(problems.length + 1);

      const newProblem = {
        id: newProblemId,
        title: problemTitle,
        content: finalResult.complete_problem || '',
        solution: finalResult.final_explanation || '',
        imageBase64: finalResult.image_base64 || undefined,
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
        solutionText: newProblem.solution,
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
    if (generationMode === 'five-stage') {
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

    // opinion profile v2の必須項目チェック
    if (opinionProfileV2.sub_problem_count === 0) {
      alert('小問の数を設定してください');
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
            filters: selectedFilters,
            opinion_profile_v2: createOpinionProfileFromFilters() // Ver.2に変更
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
    filterTexts.push('評価基準: opinion_ver2.md に基づく空間図形問題の詳細指標');
    
    // opinionProfileV2から詳細プロンプトを生成
    filterTexts.push(`\n【文章量・構成】`);
    filterTexts.push(`- 大問の問題文文字数: ${opinionProfileV2.problem_text_length}文字`);
    filterTexts.push(`- 小問の総文字数: ${opinionProfileV2.sub_problem_text_length}文字`);
    filterTexts.push(`- 与えられる数値の個数: ${opinionProfileV2.given_values_count}個`);
    filterTexts.push(`- 小問の数: ${opinionProfileV2.sub_problem_count}問`);
    if (opinionProfileV2.sub_problem_types.length > 0) {
      filterTexts.push(`- 小問の種別: ${opinionProfileV2.sub_problem_types.join(', ')}`);
    }
    filterTexts.push(`- 立体の構成: ${opinionProfileV2.solid_composition}`);
    
    filterTexts.push(`\n【解答形式】`);
    if (opinionProfileV2.answer_formats.length > 0) {
      filterTexts.push(`- 解答の形式: ${opinionProfileV2.answer_formats.join(', ')}`);
    }
    if (opinionProfileV2.answer_units.length > 0) {
      filterTexts.push(`- 要求単位: ${opinionProfileV2.answer_units.join(', ')}`);
    }
    filterTexts.push(`- 補助点の使用: ${opinionProfileV2.uses_auxiliary_points ? 'あり' : 'なし'}`);
    
    filterTexts.push(`\n【使用単元】`);
    if (opinionProfileV2.setup_units.length > 0) {
      filterTexts.push(`- 設定: ${opinionProfileV2.setup_units.join(', ')}`);
    }
    if (opinionProfileV2.solution_units.length > 0) {
      filterTexts.push(`- 解法: ${opinionProfileV2.solution_units.join(', ')}`);
    }
    
    filterTexts.push(`\n【図形】`);
    filterTexts.push(`- 総頂点・点の数: ${opinionProfileV2.total_vertices}個`);
    filterTexts.push(`- 動点: ${opinionProfileV2.has_moving_point ? 'あり' : 'なし'}`);
    filterTexts.push(`- 図中の数値: ${opinionProfileV2.figure_values_count}個`);
    
    filterTexts.push(`\n【解法プロセス】`);
    filterTexts.push(`- 解法のステップ数: ${opinionProfileV2.solution_steps}ステップ`);
    filterTexts.push(`- 論理的分岐: ${opinionProfileV2.has_logical_branching ? 'あり' : 'なし'}`);
    filterTexts.push(`- 使用定理・公式の数: ${opinionProfileV2.theorem_count}個`);
    filterTexts.push(`- 複数単元の統合: ${opinionProfileV2.requires_multi_unit_integration ? '必要' : '不要'}`);
    filterTexts.push(`- 無関係な情報: ${opinionProfileV2.has_irrelevant_info ? 'あり' : 'なし'}`);
    
    filterTexts.push('');
    filterTexts.push('※この基準に従って、空間図形問題を生成してください。');
    
    return `以下の条件で${activeSubject}の問題を生成してください:\n${filterTexts.join('\n')}`;
  };

  // opinionProfileV2からAPIリクエスト用のオブジェクトを作成
  const createOpinionProfileFromFilters = () => {
    return opinionProfileV2;
  };

  // 検索用：デフォルト値から変更されたフィールドのみを抽出
  const getModifiedFilters = () => {
    const modifiedFilters: any = {};
    
    // 数値フィールドの比較
    if (opinionProfileV2.problem_text_length !== defaultOpinionProfileV2.problem_text_length) {
      modifiedFilters.problem_text_length = opinionProfileV2.problem_text_length;
    }
    if (opinionProfileV2.sub_problem_text_length !== defaultOpinionProfileV2.sub_problem_text_length) {
      modifiedFilters.sub_problem_text_length = opinionProfileV2.sub_problem_text_length;
    }
    if (opinionProfileV2.given_values_count !== defaultOpinionProfileV2.given_values_count) {
      modifiedFilters.given_values_count = opinionProfileV2.given_values_count;
    }
    if (opinionProfileV2.sub_problem_count !== defaultOpinionProfileV2.sub_problem_count) {
      modifiedFilters.sub_problem_count = opinionProfileV2.sub_problem_count;
    }
    if (opinionProfileV2.total_vertices !== defaultOpinionProfileV2.total_vertices) {
      modifiedFilters.total_vertices = opinionProfileV2.total_vertices;
    }
    if (opinionProfileV2.figure_values_count !== defaultOpinionProfileV2.figure_values_count) {
      modifiedFilters.figure_values_count = opinionProfileV2.figure_values_count;
    }
    if (opinionProfileV2.solution_steps !== defaultOpinionProfileV2.solution_steps) {
      modifiedFilters.solution_steps = opinionProfileV2.solution_steps;
    }
    if (opinionProfileV2.theorem_count !== defaultOpinionProfileV2.theorem_count) {
      modifiedFilters.theorem_count = opinionProfileV2.theorem_count;
    }
    
    // ブール値フィールドの比較
    if (opinionProfileV2.uses_auxiliary_points !== defaultOpinionProfileV2.uses_auxiliary_points) {
      modifiedFilters.uses_auxiliary_points = opinionProfileV2.uses_auxiliary_points;
    }
    if (opinionProfileV2.has_moving_point !== defaultOpinionProfileV2.has_moving_point) {
      modifiedFilters.has_moving_point = opinionProfileV2.has_moving_point;
    }
    if (opinionProfileV2.has_logical_branching !== defaultOpinionProfileV2.has_logical_branching) {
      modifiedFilters.has_logical_branching = opinionProfileV2.has_logical_branching;
    }
    if (opinionProfileV2.requires_multi_unit_integration !== defaultOpinionProfileV2.requires_multi_unit_integration) {
      modifiedFilters.requires_multi_unit_integration = opinionProfileV2.requires_multi_unit_integration;
    }
    if (opinionProfileV2.has_irrelevant_info !== defaultOpinionProfileV2.has_irrelevant_info) {
      modifiedFilters.has_irrelevant_info = opinionProfileV2.has_irrelevant_info;
    }
    
    // 文字列フィールドの比較
    if (opinionProfileV2.solid_composition !== defaultOpinionProfileV2.solid_composition) {
      modifiedFilters.solid_composition = opinionProfileV2.solid_composition;
    }
    
    // 配列フィールドの比較（JSON文字列化して比較）
    if (JSON.stringify(opinionProfileV2.sub_problem_types.sort()) !== JSON.stringify(defaultOpinionProfileV2.sub_problem_types.sort())) {
      modifiedFilters.sub_problem_types = opinionProfileV2.sub_problem_types;
    }
    if (JSON.stringify(opinionProfileV2.answer_formats.sort()) !== JSON.stringify(defaultOpinionProfileV2.answer_formats.sort())) {
      modifiedFilters.answer_formats = opinionProfileV2.answer_formats;
    }
    if (JSON.stringify(opinionProfileV2.answer_units.sort()) !== JSON.stringify(defaultOpinionProfileV2.answer_units.sort())) {
      modifiedFilters.answer_units = opinionProfileV2.answer_units;
    }
    if (JSON.stringify(opinionProfileV2.setup_units.sort()) !== JSON.stringify(defaultOpinionProfileV2.setup_units.sort())) {
      modifiedFilters.setup_units = opinionProfileV2.setup_units;
    }
    if (JSON.stringify(opinionProfileV2.solution_units.sort()) !== JSON.stringify(defaultOpinionProfileV2.solution_units.sort())) {
      modifiedFilters.solution_units = opinionProfileV2.solution_units;
    }
    
    return modifiedFilters;
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

  // パラメータ検索する関数（OpinionProfileV2基準対応）
  const searchProblemsByFilters = async () => {
    console.log('🔍 [Frontend] opinionProfileV2:', opinionProfileV2);

    // 検索条件をチェック
    const hasSubject = activeSubject !== '';

    if (!hasSubject) {
      alert('科目を選択してください');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const requestBody = {
        subject: activeSubject,
        filters: opinionProfileV2, // 現在の設定をそのまま送信
        matchType: searchMatchType,
      };

      console.log('🔍 [Frontend] 検索リクエスト:', requestBody);

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/search-by-filters`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(requestBody),
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

  // キーワード + 条件の組み合わせ検索する関数（OpinionProfileV2基準対応）
  const searchProblemsByKeywordAndFilters = async () => {
    // 変更されたフィールドのみを抽出
    const modifiedFilters = getModifiedFilters();
    console.log('🔍 [Frontend] opinionProfileV2:', opinionProfileV2);
    console.log('🔍 [Frontend] modifiedFilters:', modifiedFilters);

    // 検索条件をチェック
    const hasKeyword = searchKeyword.trim() !== '';
    const hasSubject = activeSubject !== '';
    const hasModifiedFilters = Object.keys(modifiedFilters).length > 0;

    if (!hasKeyword && !hasSubject) {
      alert('キーワードを入力するか、科目を選択してください');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      // 完全一致の場合は全フィールド、部分一致の場合は変更されたフィールドのみ
      const filtersToSend = searchMatchType === 'exact' ? opinionProfileV2 : modifiedFilters;

      const requestBody = {
        keyword: searchKeyword.trim() || undefined,
        subject: activeSubject || undefined,
        filters: hasModifiedFilters || searchMatchType === 'exact' ? filtersToSend : undefined,
        matchType: searchMatchType,
      };

      console.log('🔍 [Frontend] 組み合わせ検索リクエスト:', requestBody);

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/search-combined`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(requestBody),
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
        
        {/* OpinionProfileSettings統合（Ver.4.0基準） */}
        <div className="mb-6">
          <OpinionProfileSettings
            opinionProfile={opinionProfileV2}
            onOpinionProfileChange={setOpinionProfileV2}
          />
        </div>
        
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
                  onChange={(e) => setGenerationMode(e.target.value as 'single' | 'five-stage')}
                  className="text-mongene-blue"
                />
                <span className="text-sm font-medium text-mongene-ink">ワンショットモデルクエリ（簡単な問題向け）</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="radio"
                  name="generationMode"
                  value="five-stage"
                  checked={generationMode === 'five-stage'}
                  onChange={(e) => setGenerationMode(e.target.value as 'single' | 'five-stage')}
                  className="text-mongene-blue"
                />
                <span className="text-sm font-medium text-mongene-ink">インクリメンタルモデルクエリ（複雑な問題向け）</span>
              </label>
            </div>
            <div className="text-xs text-mongene-muted">
              {generationMode === 'single'
                ? '問題文と解答を1回のAPI呼び出しで生成します\n※計算はLLMが行います'
                : '5段階に分けて生成します：①小問構成→②数値計算→③図形描画→④問題文→⑤解答解説\n※計算はプログラムで行います'
              }
            </div>
          </div>


          {/* 5段階生成の場合の説明 */}
          {generationMode === 'five-stage' && (
            <div className="border-t border-white/20 pt-4">
              <h4 className="font-bold text-mongene-ink mb-3">🔥 5段階生成プロセス（最高精度）</h4>
              <p className="text-sm text-mongene-muted mb-3">
                問題生成を5つのステージに分けて実行します：
              </p>
              <ol className="text-sm text-mongene-muted space-y-1 ml-4">
                <li>1️⃣ 小問構成と解答プロセスの設計</li>
                <li>2️⃣ パラメータ設定と動的検証（数値計算）</li>
                <li>3️⃣ 問題文用の図形描画</li>
                <li>4️⃣ 完全な問題文の生成</li>
                <li>5️⃣ 完全な解答・解説の生成</li>
              </ol>
              <p className="text-xs text-mongene-muted mt-3">
                ※進捗はローディング画面で確認できます
              </p>
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
        message={
          generationMode === 'five-stage'
            ? '🔥 5段階生成プロセスを実行中...'
            : 'AIが問題を生成しています...'
        }
        showProgress={generationMode === 'five-stage'}
        estimatedDuration={60000}
        currentStage={currentStage}
        onStageChange={(stage) => {
          setCurrentStage(stage);
          console.log(`📊 [Frontend] Stage ${stage} に移行`);
        }}
      />

    </div>
  );
}
