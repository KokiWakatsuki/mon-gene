'use client';

import React, { useState, useEffect } from 'react';
import Header from '../../components/layout/Header';
import Tabs from '../../components/features/problems/Tabs';
import MainTabs from '../../components/features/problems/MainTabs';
import OpinionProfileSettings from '../../components/features/problems/OpinionProfileSettings';
import ProblemCard from '../../components/features/problems/ProblemCard';
import BackgroundShapes from '../../components/layout/BackgroundShapes';
import ProblemPreviewModal from '../../components/features/problems/ProblemPreviewModal';
import LoadingModal from '../../components/ui/LoadingModal';
import FileUpload from '../../components/features/problems/FileUpload';
import SearchOptions from '../../components/features/problems/SearchOptions';
import ThreeProblemsDisplay from '../../components/features/problems/ThreeProblemsDisplay';
import CheckFormModal from '../../components/features/problems/CheckFormModal';
import { API_CONFIG } from '../../lib/config/api';

export default function Home() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);
  const [activeSubject] = useState('数学'); // 数学のみに固定
  const [activeMainTab, setActiveMainTab] = useState<'list' | 'generate'>('list'); // メインタブの状態
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
  
  // チェック情報の型定義
  interface CheckInfo {
    problem_text_ok: boolean;
    solution_ok: boolean;
    figure_ok: boolean;
    units: string[];
    year: string;
    exam_session: string;
  }
  
  const [problems, setProblems] = useState<Array<{
    id: string;
    title: string;
    content: string;
    imageBase64?: string;
    solution?: string;
    checkInfo?: CheckInfo;
  }>>([]);
  
  const [userInfo, setUserInfo] = useState<{
    school_code: string;
    email: string;
    problem_generation_limit: number;
    problem_generation_count: number;
    preview_limit: number;
    preview_count: number;
  } | null>(null);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [isSearchMode, setIsSearchMode] = useState(false);
  const [searchResults, setSearchResults] = useState<Array<{
    id: string;
    title: string;
    content: string;
    imageBase64?: string;
    solution?: string;
    checkInfo?: CheckInfo;
  }>>([]);
  const [searchMatchType, setSearchMatchType] = useState<'exact' | 'partial'>('partial');
  
  // チェックフォームモーダルの状態
  const [checkFormModal, setCheckFormModal] = useState<{
    isOpen: boolean;
    problemId: string;
    problemTitle: string;
    initialData?: CheckInfo;
  }>({
    isOpen: false,
    problemId: '',
    problemTitle: '',
    initialData: undefined,
  });
  
  // 生成システム用の状態（3問生成のみ）
  const [generationMode] = useState<'three-problems'>('three-problems');
  
  // ファイルアップロード用の状態
  const [uploadedFiles, setUploadedFiles] = useState<File[]>([]);
  const [uploadedSolutionFiles, setUploadedSolutionFiles] = useState<File[]>([]);
  const [showFilePreview, setShowFilePreview] = useState(false);
  const [filePreviewContent, setFilePreviewContent] = useState<string>('');
  
  // まだ習っていない単元の状態
  const [excludedUnits, setExcludedUnits] = useState<string[]>([]);
  
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
  const [currentStage, setCurrentStage] = useState<number>(0); // 0=未開始, 1-5=各段階（5段階）, 1-15=各段階（3問生成）
  const [stageProgress, setStageProgress] = useState<number>(0); // 進捗率 0-100
  
  // 3問生成システム用の状態
  const [threeProblemsResults, setThreeProblemsResults] = useState<{
    patternA?: { id: string; title: string; content: string; solution: string; imageBase64?: string };
    patternB?: { id: string; title: string; content: string; solution: string; imageBase64?: string };
    patternC?: { id: string; title: string; content: string; solution: string; imageBase64?: string };
  }>({});

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
          checkInfo: problem.check_info,
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
      <div className="relative min-h-screen overflow-hidden">
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

  const subjects = ['数学']; // 数学のみ

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

  // 科目は「数学」に固定されているため、handleSubjectChangeは不要

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

  // チェック済みかどうかを判定する関数
  const isChecked = (checkInfo?: CheckInfo): boolean => {
    if (!checkInfo) return false;
    
    // 1-3: チェックが入っている
    const basicChecksOk = checkInfo.problem_text_ok && checkInfo.solution_ok && checkInfo.figure_ok;
    
    // 4-6: 何かが選択されている
    const unitsSelected = checkInfo.units && checkInfo.units.length > 0;
    const yearSelected = !!checkInfo.year && checkInfo.year !== '';
    const examSessionSelected = !!checkInfo.exam_session && checkInfo.exam_session !== '';
    
    return basicChecksOk && unitsSelected && yearSelected && examSessionSelected;
  };

  // チェックボタンのハンドラー
  const handleCheck = (id: string) => {
    const problem = (isSearchMode ? searchResults : problems).find(p => p.id === id);
    if (problem) {
      setCheckFormModal({
        isOpen: true,
        problemId: id,
        problemTitle: problem.title,
        initialData: problem.checkInfo,
      });
    }
  };

  // チェックフォームの保存ハンドラー
  const handleCheckSave = async (checkInfo: CheckInfo) => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        alert('認証トークンが見つかりません');
        return;
      }

      const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/problems/update-check-info`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          id: parseInt(checkFormModal.problemId),
          check_info: checkInfo,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'チェック情報の保存に失敗しました');
      }

      // 問題リストを更新
      setProblems(prev => prev.map(problem =>
        problem.id === checkFormModal.problemId
          ? { ...problem, checkInfo }
          : problem
      ));

      // 検索結果も更新
      if (isSearchMode) {
        setSearchResults(prev => prev.map(problem =>
          problem.id === checkFormModal.problemId
            ? { ...problem, checkInfo }
            : problem
        ));
      }

      // モーダルを閉じる
      setCheckFormModal({
        isOpen: false,
        problemId: '',
        problemTitle: '',
        initialData: undefined,
      });

      alert('チェック情報を保存しました');
    } catch (error) {
      console.error('チェック情報の保存エラー:', error);
      alert(`チェック情報の保存に失敗しました: ${(error as Error).message}`);
    }
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
          ? `<div class="image-container">
               <img src="data:image/png;base64,${problem.imageBase64}"
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
              .problem-layout {
                display: flex;
                gap: 20px;
                align-items: flex-start;
              }
              .image-container {
                flex: 0 0 50%;
                max-width: 50%;
              }
              .image-container img {
                width: 100%;
                height: auto;
                border: 1px solid #ddd;
              }
              .content-container {
                flex: 1;
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
            ${problem.imageBase64
              ? `<div class="problem-layout">
                   <div class="content-container">
                     <div class="content">${processedContent}</div>
                   </div>
                   ${imageHtml}
                 </div>`
              : `<div class="content">${processedContent}</div>`
            }
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

  // 3問生成システムの関数（SSE使用）
  const handleGenerateThreeProblems = async () => {
    // 上限チェック
    if (isGenerationLimitReached()) {
      alert(`問題生成回数の上限（${userInfo?.problem_generation_limit}回）に達しました。これ以上問題を生成することはできません。`);
      return;
    }

    // ファイルアップロードチェック（3問生成では必須）
    if (uploadedFiles.length === 0) {
      alert('3問生成モードでは問題ファイルのアップロードが必須です。参考となる問題ファイルをアップロードしてください。');
      return;
    }

    // 解答ファイルアップロードチェック（3問生成では必須）
    if (!uploadedSolutionFiles || uploadedSolutionFiles.length === 0) {
      alert('3問生成モードでは解答ファイルのアップロードが必須です。解答ファイルをアップロードしてください。');
      return;
    }

    // opinion profile v2の必須項目チェック
    if (opinionProfileV2.sub_problem_count === 0) {
      alert('小問の数を設定してください');
      return;
    }
    
    setIsLoading(true);
    setThreeProblemsResults({});
    setCurrentStage(1);
    setStageProgress(0);
    
    try {
      const prompt = createPromptFromFilters();
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('認証トークンが見つかりません。再度ログインしてください。');
      }

      console.log('🚀 [ThreeProblems] 3問生成プロセス開始（SSE使用）');
      
      // PDFファイルがあるかチェック
      const hasPDF = uploadedFiles.some(file => file.type === 'application/pdf');
      
      let response: Response;
      
      if (hasPDF) {
        // PDFファイルがある場合：multipart/form-dataで送信
        console.log('📄 [ThreeProblems] PDF file detected, using multipart/form-data');
        
        const formData = new FormData();
        formData.append('subject', activeSubject);
        
        // 除外単元をJSON文字列として追加
        if (excludedUnits.length > 0) {
          formData.append('excluded_units', JSON.stringify(excludedUnits));
          console.log('📎 [ThreeProblems] Excluded units:', excludedUnits);
        }
        
        // 最初のPDFファイルのみを送信（複数ある場合は最初のもの）
        const pdfFile = uploadedFiles.find(file => file.type === 'application/pdf');
        if (pdfFile) {
          formData.append('file', pdfFile);
          console.log('📎 [ThreeProblems] Attached PDF file:', pdfFile.name, 'size:', pdfFile.size);
        }
        
        // fetchでSSE接続（multipart/form-data）
        response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-three-problems-sse`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            // Content-Typeは自動設定されるため指定しない
          },
          body: formData
        });
      } else {
        // PDFファイルがない場合：従来通りテキストで送信
        console.log('📝 [ThreeProblems] No PDF file, using text content');
        
        // ファイル内容を読み込んで1つの文字列に結合
        const fileContents = await Promise.all(
          uploadedFiles.map(async (file) => {
            const text = await file.text();
            return `【ファイル名: ${file.name}】\n${text}`;
          })
        );
        
        const uploadedProblemContent = fileContents.join('\n\n---\n\n');
        console.log('📄 [ThreeProblems] Uploaded problem content length:', uploadedProblemContent.length);

        // SSEを使用してリアルタイム進捗を取得
        const requestBody = JSON.stringify({
          uploaded_problem_content: uploadedProblemContent,
          subject: activeSubject,
          excluded_units: excludedUnits, // 除外単元を追加
        });

        // fetchでSSE接続（JSON）
        response = await fetch(`${API_CONFIG.API_BASE_URL}/api/generate-three-problems-sse`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: requestBody
        });
      }

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
                setCurrentStage(15);
              }
            } catch (e) {
              console.error('❌ [SSE] Parse error:', e);
            }
          }
        }
      }

      if (!finalResult) {
        throw new Error('3問生成の結果を取得できませんでした');
      }

      console.log('📦 [ThreeProblems] Final result:', finalResult);

      // 結果を問題リストに追加（3問すべて）
      const newProblems = [
        {
          id: String(problems.length + 1),
          title: `パターンA: 数値変更 ${problems.length + 1}`,
          content: finalResult.pattern_a?.content || '',
          solution: finalResult.pattern_a?.solution || '',
          imageBase64: finalResult.pattern_a?.image_base64 || undefined,
        },
        {
          id: String(problems.length + 2),
          title: `パターンB: 文脈変更 ${problems.length + 2}`,
          content: finalResult.pattern_b?.content || '',
          solution: finalResult.pattern_b?.solution || '',
          imageBase64: finalResult.pattern_b?.image_base64 || undefined,
        },
        {
          id: String(problems.length + 3),
          title: `パターンC: 構造変更 ${problems.length + 3}`,
          content: finalResult.pattern_c?.content || '',
          solution: finalResult.pattern_c?.solution || '',
          imageBase64: finalResult.pattern_c?.image_base64 || undefined,
        }
      ];

      setProblems(prev => [...prev, ...newProblems]);
      setThreeProblemsResults({
        patternA: newProblems[0],
        patternB: newProblems[1],
        patternC: newProblems[2],
      });

      // ユーザー情報を更新
      await fetchUserInfo();

      setIsLoading(false);

      // 最初の問題のプレビューモーダルを表示
      setPreviewModal({
        isOpen: true,
        problemId: newProblems[0].id,
        problemTitle: newProblems[0].title,
        problemContent: newProblems[0].content,
        imageBase64: newProblems[0].imageBase64,
        solutionText: newProblems[0].solution,
      });

      console.log('✅ [ThreeProblems] 3問生成プロセス完全完了');
      
    } catch (error) {
      setIsLoading(false);
      setCurrentStage(0);
      setStageProgress(0);
      console.error('3問生成エラー:', error);
      await handleGenerationError(error);
    }
  };

  const handleGenerate = async () => {
    if (generationMode === 'three-problems') {
      await handleGenerateThreeProblems();
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
    
    // まだ習っていない単元を除外
    if (excludedUnits.length > 0) {
      filterTexts.push(`\n【除外する単元】`);
      filterTexts.push(`以下の単元はまだ習っていないため、問題に含めないでください:`);
      excludedUnits.forEach(unit => {
        filterTexts.push(`- ${unit}`);
      });
    }
    
    // アップロードされたファイルがある場合は参考資料として追加
    if (uploadedFiles.length > 0) {
      filterTexts.push(`\n【参考資料】`);
      filterTexts.push(`- アップロードされたファイル数: ${uploadedFiles.length}件`);
      filterTexts.push(`- ファイル名: ${uploadedFiles.map(f => f.name).join(', ')}`);
      filterTexts.push('※これらのファイルを参考に問題を生成してください');
    }
    
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

  // キーワード検索する関数（キーワードなしでもタグ検索可能）
  const searchProblems = async () => {
    // キーワードがない場合はタグ検索を実行
    if (!searchKeyword.trim()) {
      await searchProblemsByFilters();
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
          checkInfo: problem.check_info,
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
          checkInfo: problem.check_info,
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
          checkInfo: problem.check_info,
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
    <div className="relative min-h-screen overflow-hidden">
      <BackgroundShapes />
      
      <div className="relative z-10">
        <Header />
        
        <div className="max-w-6xl mx-auto">
          <Tabs
            subjects={subjects}
            activeSubject={activeSubject}
            onSubjectChange={() => {}} // 数学のみなので何もしない
          />
          
          <MainTabs
            activeTab={activeMainTab}
            onTabChange={setActiveMainTab}
          />
          
          <div className="px-4 pb-12">
            {/* 問題一覧タブ */}
            {activeMainTab === 'list' && (
              <>
                {/* キーワード検索バー */}
                <div className="mb-4">
                  <div className="relative flex gap-2">
                    <input
                      type="text"
                      id="keywordInput"
                      placeholder="キーワード検索 (例: 面積, 太郎, 最大値...)"
                      value={searchKeyword}
                      onChange={(e) => setSearchKeyword(e.target.value)}
                      className="flex-1 px-4 py-3 pr-12 border border-gray-200 rounded-lg text-base shadow-[0_2px_5px_rgba(0,0,0,0.03)] transition-all focus:outline-none focus:border-blue-500 focus:shadow-[0_0_0_3px_rgba(59,130,246,0.1)]"
                      onKeyDown={(e) => e.key === 'Enter' && searchProblems()}
                    />
                    <button
                      onClick={searchProblems}
                      className="absolute right-1.5 top-1/2 -translate-y-1/2 w-9 h-9 bg-blue-500 text-white border-none rounded-md cursor-pointer flex items-center justify-center transition-colors hover:bg-blue-400"
                      aria-label="検索"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <circle cx="11" cy="11" r="8"></circle>
                        <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                      </svg>
                    </button>
                  </div>
                </div>
            
            {/* 検索オプション（アコーディオン） */}
            <SearchOptions
              opinionProfile={opinionProfileV2}
              onOpinionProfileChange={setOpinionProfileV2}
            />
            
            {/* 検索モード時の「一覧に戻る」ボタン */}
            {isSearchMode && (
              <div className="mb-6">
                <button
                  onClick={() => {
                    setIsSearchMode(false);
                    setSearchKeyword('');
                    fetchProblemHistory();
                  }}
                  className="px-6 py-3 bg-gray-500 text-white rounded-lg font-bold hover:brightness-110 transition-all flex items-center gap-2"
                >
                  <span>←</span>
                  <span>一覧に戻る</span>
                </button>
              </div>
            )}
            
                <section className="grid grid-cols-1 lg:grid-cols-2 gap-8 mt-6" aria-label="問題一覧">
                  {(isSearchMode ? searchResults : problems).map((problem) => (
                    <ProblemCard
                      key={problem.id}
                      id={problem.id}
                      title={problem.title}
                      content={problem.content}
                      imageBase64={problem.imageBase64}
                      isChecked={isChecked(problem.checkInfo)}
                      onPreview={handlePreview}
                      onPrint={handlePrint}
                    />
                  ))}
                </section>
              </>
            )}
        
        {/* 問題生成タブ */}
        {activeMainTab === 'generate' && (
          <>
            {/* ステップ1: 画像をアップロード */}
            <div className="mb-10">
              <div className="flex items-center gap-3 mb-3">
                <span className="w-6 h-6 bg-gray-800 text-white rounded-full flex items-center justify-center font-extrabold text-base pb-0.5">1</span>
                <h3 className="text-xl text-gray-800 font-semibold m-0">画像をアップロード</h3>
              </div>
              <p className="text-sm text-gray-500 ml-9 mb-4">問題文の画像と、あれば解答の画像をアップロードしてください。</p>
              
              <FileUpload
              uploadedFiles={uploadedFiles}
              onFilesChange={setUploadedFiles}
              uploadedSolutionFiles={uploadedSolutionFiles}
              onSolutionFilesChange={setUploadedSolutionFiles}
              />
            </div>

            {/* ステップ2: まだ習っていない単元を選択 */}
            <div className="mb-10">
              <div className="flex items-center gap-3 mb-3">
                <span className="w-6 h-6 bg-gray-800 text-white rounded-full flex items-center justify-center font-extrabold text-base pb-0.5">2</span>
                <h3 className="text-xl text-gray-800 font-semibold m-0">まだ習っていない単元を選択</h3>
              </div>
              <p className="text-sm text-gray-500 ml-9 mb-4">習っていない単元をタップして除外してください</p>
              
              <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-[0_2px_4px_rgba(0,0,0,0.03)]">
                <div className="flex flex-wrap gap-2.5">
                  {[
                    '多項式（展開・因数分解）',
                    '平方根',
                    '二次方程式',
                    '関数 y=ax²',
                    '図形の相似',
                    '円の性質（円周角）',
                    '三平方の定理',
                    '標本調査'
                  ].map((unit) => {
                    const isExcluded = excludedUnits.includes(unit);
                    return (
                      <label key={unit} className="relative cursor-pointer">
                        <input
                          type="checkbox"
                          name="exclude_unit"
                          checked={isExcluded}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setExcludedUnits([...excludedUnits, unit]);
                            } else {
                              setExcludedUnits(excludedUnits.filter(u => u !== unit));
                            }
                          }}
                          className="absolute opacity-0 w-0 h-0"
                        />
                        <span className={`inline-block px-4 py-2 rounded-full text-sm font-semibold transition-all shadow-[0_1px_2px_rgba(0,0,0,0.05)] ${
                          isExcluded
                            ? 'bg-gray-100 border-transparent text-gray-500 line-through opacity-60 shadow-none'
                            : 'bg-white border border-gray-200 text-gray-800 hover:border-blue-500 hover:text-blue-500'
                        }`}>
                          {unit}
                        </span>
                      </label>
                    );
                  })}
                </div>
              </div>
            </div>

            {/* ステップ3: 生成開始！ */}
            <div className="mb-10">
              <div className="flex items-center gap-3 mb-3">
                <span className="w-6 h-6 bg-gray-800 text-white rounded-full flex items-center justify-center font-extrabold text-base pb-0.5">3</span>
                <h3 className="text-xl text-gray-800 font-semibold m-0">生成開始！</h3>
              </div>
              
              <div className="flex flex-col items-start gap-4 max-w-[300px]">
                {/* 生成回数表示ボックス */}
                {userInfo && (
                  <div className="w-full bg-white border border-gray-200 rounded-xl p-3 px-4 shadow-[0_2px_6px_rgba(0,0,0,0.02)]">
                    <div className="flex justify-between items-center mb-2 text-[13px] font-bold text-gray-500">
                      <span className="flex items-center gap-1">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{marginRight: '4px', transform: 'translateY(1px)'}}>
                          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path>
                        </svg>
                        残り生成回数
                      </span>
                      <span className="text-gray-800">
                        <strong className="text-blue-500 text-base">{userInfo.problem_generation_limit === -1 ? '∞' : Math.max(0, userInfo.problem_generation_limit - userInfo.problem_generation_count)}</strong>
                        {' / '}
                        {userInfo.problem_generation_limit === -1 ? '∞' : userInfo.problem_generation_limit}回
                      </span>
                    </div>
                    <div className="h-1.5 w-full bg-gray-100 rounded-full overflow-hidden mb-1.5">
                      <div
                        className="h-full bg-gradient-to-r from-blue-400 to-blue-500 rounded-full transition-[width] duration-300"
                        style={{
                          width: userInfo.problem_generation_limit === -1
                            ? '100%'
                            : `${Math.min(100, ((userInfo.problem_generation_limit - userInfo.problem_generation_count) / userInfo.problem_generation_limit) * 100)}%`
                        }}
                      ></div>
                    </div>
                    <p className="text-[11px] text-gray-500 text-right m-0">次回リセット: 2026/01/01</p>
                  </div>
                )}
                
                <button
                  className={`w-full justify-center text-base font-bold px-7 py-3.5 rounded-xl transition-all shadow-[0_4px_15px_rgba(141,219,57,0.4)] border-none cursor-pointer ${
                    isGenerationLimitReached()
                      ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                      : 'bg-mongene-green text-gray-800 hover:brightness-105 hover:-translate-y-0.5'
                  }`}
                  type="button"
                  onClick={handleGenerate}
                  disabled={isGenerationLimitReached()}
                >
                  生成する
                </button>
              </div>
            </div>

            {/* 3問生成結果の表示 */}
            {Object.keys(threeProblemsResults).length > 0 && (
              <ThreeProblemsDisplay
                problems={threeProblemsResults}
                onPreview={handlePreview}
                onPrint={handlePrint}
              />
            )}
            
            <div className="text-center mt-8">
              <div className="flex items-center justify-center gap-4">
                
                {/* アップロードした問題の概要表示ボタン */}
                {(uploadedFiles.length > 0 || uploadedSolutionFiles.length > 0) && (
                  <button
                    className={`text-base font-bold px-6 py-3.5 rounded-xl transition-all shadow-[0_4px_15px_rgba(59,130,246,0.4)] ${
                      userInfo && userInfo.preview_limit !== -1 && userInfo.preview_count >= userInfo.preview_limit
                        ? 'bg-gray-400 text-gray-600 cursor-not-allowed'
                        : 'bg-blue-500 text-white hover:brightness-110 hover:-translate-y-0.5'
                    }`}
                    type="button"
                    disabled={userInfo ? (userInfo.preview_limit !== -1 && userInfo.preview_count >= userInfo.preview_limit) : false}
                    onClick={async () => {
                      // 使用制限チェック
                      if (userInfo && userInfo.preview_limit !== -1 && userInfo.preview_count >= userInfo.preview_limit) {
                        alert(`プレビュー回数の上限（${userInfo.preview_limit}回）に達しました。これ以上プレビューを表示することはできません。`);
                        return;
                      }

                      try {
                        console.log('📄 Loading file contents...', uploadedFiles, uploadedSolutionFiles);
                        setIsLoading(true);
                        
                        // 問題ファイルの内容を読み込む
                        const problemContents = await Promise.all(
                          uploadedFiles.map(async (file) => {
                            console.log('Processing problem file:', file.name, 'type:', file.type, 'size:', file.size);
                            
                            if (file.type === 'application/pdf') {
                              // PDFファイルの場合：バックエンドAPIを呼び出してテキストを抽出
                              console.log('📄 Extracting PDF content via API...');
                              
                              const token = localStorage.getItem('token');
                              if (!token) {
                                throw new Error('認証トークンが見つかりません');
                              }
                              
                              const formData = new FormData();
                              formData.append('file', file);
                              
                              const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/preview-pdf`, {
                                method: 'POST',
                                headers: {
                                  'Authorization': `Bearer ${token}`,
                                },
                                body: formData
                              });
                              
                              if (!response.ok) {
                                throw new Error(`PDF抽出エラー: ${response.status}`);
                              }
                              
                              const data = await response.json();
                              console.log('✅ PDF content extracted:', data.content.substring(0, 200));
                              
                              return `【問題PDFファイル】\nファイル名: ${file.name}\nサイズ: ${(file.size / 1024).toFixed(2)} KB\n\n【抽出された内容】\n${data.content}`;
                            } else {
                              const text = await file.text();
                              console.log('Text file content length:', text.length);
                              const preview = text.substring(0, 1000);
                              return `【問題ファイル名: ${file.name}】\nサイズ: ${(file.size / 1024).toFixed(2)} KB\n\n${preview}${text.length > 1000 ? '\n\n... (以下省略)' : ''}`;
                            }
                          })
                        );
                        
                        // 解答ファイルの内容を読み込む
                        const solutionContents = await Promise.all(
                          uploadedSolutionFiles.map(async (file) => {
                            console.log('Processing solution file:', file.name, 'type:', file.type, 'size:', file.size);
                            
                            if (file.type === 'application/pdf') {
                              // PDFファイルの場合：バックエンドAPIを呼び出してテキストを抽出
                              console.log('📄 Extracting solution PDF content via API...');
                              
                              const token = localStorage.getItem('token');
                              if (!token) {
                                throw new Error('認証トークンが見つかりません');
                              }
                              
                              const formData = new FormData();
                              formData.append('file', file);
                              
                              const response = await fetch(`${API_CONFIG.API_BASE_URL}/api/preview-pdf`, {
                                method: 'POST',
                                headers: {
                                  'Authorization': `Bearer ${token}`,
                                },
                                body: formData
                              });
                              
                              if (!response.ok) {
                                throw new Error(`解答PDF抽出エラー: ${response.status}`);
                              }
                              
                              const data = await response.json();
                              console.log('✅ Solution PDF content extracted:', data.content.substring(0, 200));
                              
                              return `【解答PDFファイル】\nファイル名: ${file.name}\nサイズ: ${(file.size / 1024).toFixed(2)} KB\n\n【抽出された内容】\n${data.content}`;
                            } else {
                              const text = await file.text();
                              console.log('Solution text file content length:', text.length);
                              const preview = text.substring(0, 1000);
                              return `【解答ファイル名: ${file.name}】\nサイズ: ${(file.size / 1024).toFixed(2)} KB\n\n${preview}${text.length > 1000 ? '\n\n... (以下省略)' : ''}`;
                            }
                          })
                        );
                        
                        // 問題と解答を結合
                        const allContents = [
                          ...problemContents,
                          ...(solutionContents.length > 0 ? ['\n\n' + '='.repeat(50) + '\n【解答・解説】\n' + '='.repeat(50) + '\n\n', ...solutionContents] : [])
                        ];
                        
                        const summary = allContents.join('\n\n' + '='.repeat(50) + '\n\n');
                        console.log('File summary generated:', summary.substring(0, 200));
                        
                        setFilePreviewContent(summary);
                        setShowFilePreview(true);
                        setIsLoading(false);
                        
                        // ユーザー情報を更新（プレビュー回数をインクリメント）
                        await fetchUserInfo();
                      } catch (error) {
                        console.error('ファイル読み込みエラー:', error);
                        setIsLoading(false);
                        alert('ファイルの読み込みに失敗しました: ' + (error as Error).message);
                      }
                    }}
                  >
                    📄 問題概要を表示
                    {userInfo && userInfo.preview_limit !== -1 && (
                      <span className="ml-2 text-sm">
                        (残り {userInfo.preview_limit - userInfo.preview_count}回)
                      </span>
                    )}
                  </button>
                )}
              </div>
            </div>
          </>
        )}
          </div>
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
        initialCheckInfo={(isSearchMode ? searchResults : problems).find(p => p.id === previewModal.problemId)?.checkInfo}
        onCheck={handleCheck}
        onCheckSave={handleCheckSave}
        onUpdate={(updatedData) => {
          // 問題リストを更新
          setProblems(prev => prev.map(problem =>
            problem.id === previewModal.problemId
              ? {
                  ...problem,
                  content: updatedData.content,
                  solution: updatedData.solution,
                  imageBase64: updatedData.imageBase64,
                  checkInfo: updatedData.checkInfo || problem.checkInfo // チェック情報を更新
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
                    imageBase64: updatedData.imageBase64,
                    checkInfo: updatedData.checkInfo || problem.checkInfo // チェック情報を更新
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
        message='📚 解き直しプロセスを実行中...'
        showProgress={true}
        estimatedDuration={60000}
        currentStage={currentStage}
        maxStages={15}
        onStageChange={(stage) => {
          setCurrentStage(stage);
          console.log(`📊 [Frontend] Stage ${stage} に移行`);
        }}
      />

      {/* チェックフォームモーダル */}
      <CheckFormModal
        isOpen={checkFormModal.isOpen}
        onClose={() => setCheckFormModal({ isOpen: false, problemId: '', problemTitle: '', initialData: undefined })}
        onSave={handleCheckSave}
        problemId={checkFormModal.problemId}
        problemTitle={checkFormModal.problemTitle}
        initialCheckInfo={checkFormModal.initialData}
      />

      {/* ファイルプレビューモーダル */}
      {showFilePreview && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-2xl max-w-4xl w-full max-h-[80vh] flex flex-col">
            <div className="p-6 border-b border-gray-200 flex items-center justify-between">
              <h2 className="text-2xl font-bold text-gray-800">📄 アップロードした問題の概要</h2>
              <button
                onClick={() => setShowFilePreview(false)}
                className="text-gray-500 hover:text-gray-700 text-2xl font-bold"
              >
                ×
              </button>
            </div>
            <div className="p-6 overflow-y-auto flex-1">
              <pre className="whitespace-pre-wrap font-mono text-sm text-gray-700 bg-gray-50 p-4 rounded-lg border border-gray-200">
                {filePreviewContent}
              </pre>
            </div>
            <div className="p-6 border-t border-gray-200 flex justify-end">
              <button
                onClick={() => setShowFilePreview(false)}
                className="px-6 py-2 bg-blue-500 text-white rounded-lg font-bold hover:brightness-110 transition-all"
              >
                閉じる
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}
