'use client';

import { useState } from 'react';

interface OpinionProfile {
  domain: number;
  skill_level: number;
  structure_complexity: [number, number];
  difficulty_score: number;
}

interface OpinionProfileSettingsProps {
  opinionProfile: OpinionProfile;
  onOpinionProfileChange: (profile: OpinionProfile) => void;
}

const domainOptions = [
  { value: 1, label: '関数', description: '座標平面上にグラフが描かれており、その性質（変域、変化の割合、面積、交点など）を問う問題。' },
  { value: 2, label: '平面図形', description: '円、三角形・四角形の相似や合同、三平方の定理などを主軸に構成された、二次元空間内の図形問題。' },
  { value: 3, label: '空間図形', description: '直方体、角錐、円錐、球などの立体を対象に、体積、表面積、線分の長さ、最短距離、切断などを問う問題。' },
  { value: 4, label: '確率・統計', description: '複雑なルール下での確率計算や、複数の資料（箱ひげ図など）の読み取りと比較を問う問題。' },
  { value: 5, label: '数と式', description: '整数問題、方程式の応用、規則性の発見などを主軸とし、図形や関数に依らない問題。' },
  { value: 6, label: '融合問題', description: '上記の2つ以上の分野が同等の比重で組み合わされている問題。（例：6 (1+3) -> 関数と空間図形の融合）' }
];

const skillLevelDescriptions = [
  { value: 1, label: '基本的知識', description: '最終問題としては異例だが、基本的な公式や定理を直接的に用いて解ける場合。' },
  { value: 2, label: '応用的知識', description: '特定の応用公式や定理（例：メネラウスの定理等）を知っているかどうかが、解答時間を大きく左右する場合。' },
  { value: 3, label: '手順の遂行能力', description: '解法に至るまでのステップ数が多く、複数の基本的な解法を順番に、正確に適用していく作業の正確さが問われる場合。' },
  { value: 4, label: '計算の実行精度', description: '解法の方針は比較的見えやすいが、計算過程が非常に煩雑で、最後までミスなく計算しきる能力が最も問われる場合。' },
  { value: 5, label: '標準的なモデル化能力', description: '会話文や状況設定がやや複雑で、それを数式や図に変換するプロセスが主な課題となる、ごく一般的な応用問題。' },
  { value: 6, label: '複雑な情報統制・モデル化能力', description: 'ストーリー仕立ての長文や、複数の図表から必要な情報を抽出し、統合して一つの数理モデル（方程式や図形）に落とし込むプロセスが最も困難な場合。' },
  { value: 7, label: '緻密な論理構築能力', description: '複数の定理や定義を連鎖的に適用する必要がある、または複雑な場合分けを伴うなど、解答までの道筋を矛盾なく、段階的に組み立てる純粋な論理力が試される場合。' },
  { value: 8, label: '高度な空間認識能力', description: '複雑な立体の切断面の形状を正確に想像したり、展開図上での点の動きを三次元的に再構成したりする能力が、他のどの能力よりも中心的に要求される場合。' },
  { value: 9, label: '独創的な着眼力', description: '問題の突破口が、非常に巧妙な補助線、図形の回転・等積変形、想定外の視点からのアプローチなど、定石から大きく外れた「ひらめき」に強く依存している場合。' },
  { value: 10, label: '高次元の発想力', description: '解法が、高校数学の範囲の考え方（例：ベクトル）を導入すると著しく容易になる、またはそれに準ずる最高レベルの着想（例：体積の2通りの表現による高さ算出）を必要とする場合。' }
];

const readingComplexityDescriptions = [
  { value: 1, label: '図と数式のみ', description: '図と数式のみで構成され、文章による設定がほぼ存在しない。' },
  { value: 2, label: '短い補足文', description: '1〜2行の短い補足文が図に添えられている。' },
  { value: 3, label: '標準的な状況設定', description: '1段落程度の文章で、標準的な問題の状況設定が説明されている。' },
  { value: 4, label: '複数条件の整理', description: '複数の条件や定義が箇条書きで与えられ、それらを整理する必要がある。' },
  { value: 5, label: 'やや長文', description: 'やや長文（2段落以上）で構成され、問題の場面を理解するのに少し時間がかかる。' },
  { value: 6, label: '会話文形式', description: '会話文形式（登場人物2人程度）で、やり取りの中から条件を抽出する必要がある。' },
  { value: 7, label: 'ストーリー・動点', description: 'ストーリー形式で数学と直接関係のない背景情報が含まれる、または動点が1つ含まれる。' },
  { value: 8, label: '複雑な動的設定', description: '複数の動点や複雑な移動ルールが絡むなど、設定自体が動的で極めて複雑。' },
  { value: 9, label: '非常に複雑なストーリー', description: '非常に長く複雑なストーリー、または法律の条文のような厳密な独自ルールの読解が必須。' },
  { value: 10, label: '最高レベルの読解', description: '複数の独自ルールが複雑に重なり、ルールブックを正確に読み解くような最高レベルの読解力が要求される。' }
];

const guidanceDescriptions = [
  { value: 1, label: '完全な無誘導', description: '大問全体が1つの設問で構成されているか、小問があっても互いに全く関連性がない（完全な無誘導）。' },
  { value: 2, label: '小問間の関連性薄', description: '小問間の関連性が非常に薄く、思考のつながりがほとんどない。' },
  { value: 3, label: '独立した思考', description: '小問は同じ図形や設定を共有しているが、思考プロセスはそれぞれで独立している。' },
  { value: 4, label: '状況理解の助け', description: '前問が、次問を解く上での状況理解の助けになる程度。直接的なヒントではない。' },
  { value: 5, label: '標準的な誘導', description: '前問の結果が、次問を解く上での複数のヒントの一つとして利用できる（標準的な誘導）。' },
  { value: 6, label: '重要な要素', description: '前問で証明した事実や導出した結果が、次問を解く上で重要な要素として機能する。' },
  { value: 7, label: '解法テンプレート', description: '前問の解法プロセスそのものが、次問の解法のテンプレートや主要な考え方となっている。' },
  { value: 8, label: '直接的利用', description: '前問の答え（数値や式）を、次問の計算に直接的に利用する必要がある。' },
  { value: 9, label: '明確な連鎖構造', description: '設問(1)→(2)→(3)と、前の答えがないと次の設問に着手できない、明確な連鎖構造になっている。' },
  { value: 10, label: '完全なレール形式', description: '(9)に加え、各設問が次の設問を解くためのほぼ唯一の道筋を示している（完全なレール形式）。' }
];

const difficultyDescriptions = [
  { value: 1, label: 'エラー', description: '採点対象外や作問ミスの可能性が疑われるレベル。' },
  { value: 2, label: '超基礎', description: '最終問題としては異例だが、計算問題レベル。' },
  { value: 3, label: '基礎', description: '教科書の例題レベル。解法が一つに定まる。' },
  { value: 4, label: '基礎定着', description: '教科書の練習問題レベル。基本的な公式や定理を正しく使えるか問う。' },
  { value: 5, label: '基礎＋', description: '基本的な公式や定理を使うが、わずかに捻りがある。' },
  { value: 6, label: '基礎応用', description: '異なる単元の基本知識を直接的に組み合わせる。' },
  { value: 7, label: 'やや易', description: '標準的な問題の小問(1)レベル。手順が少なく、計算も平易。' },
  { value: 8, label: '標準', description: '教科書の章末問題レベル。思考力が必要だが、解法は典型的。' },
  { value: 9, label: '標準＋', description: '基本的な解法を複数ステップ踏む必要がある。' },
  { value: 10, label: '応用', description: '複数の単元の知識を組み合わせる、ごく標準的な応用問題。' },
  { value: 11, label: '応用＋', description: '典型的な応用問題だが、計算量がやや多い、または少し工夫が必要。' },
  { value: 12, label: '発展', description: '思考のステップ数が多く、解法選択に迷う可能性がある。' },
  { value: 13, label: '上位校標準', description: '複数の知識を組み合わせる、標準的な応用問題。上位校合格には完答したい。' },
  { value: 14, label: '上位校応用', description: '複雑な計算を伴う応用問題。時間内に処理しきる正確性が求められる。' },
  { value: 15, label: '難関', description: 'トップ校の合否を分ける典型的な難問。思考の深さが試される。' },
  { value: 16, label: '難関＋', description: '複数の応用知識を組み合わせ、かつ計算も複雑。思考力と処理能力の双方が高いレベルで要求される。' },
  { value: 17, label: '超難関', description: '正答率が数%と想定される。高度な発想に加え、極めて複雑な手順・計算を要する。' },
  { value: 18, label: '全国最難関', description: '私立最難関校の入試問題と比較しても遜色ないレベル。思考の独創性が求められる。' },
  { value: 19, label: '捨て問', description: '高校範囲の知識が背景にあり、中学範囲の知識だけでは発想が極めて困難な問題。' },
  { value: 20, label: '超・捨て問', description: '複数の高校範囲の知識を示唆する、公立高校入試の枠を明らかに逸脱した問題。' }
];

interface AccordionItemProps {
  title: string;
  children: React.ReactNode;
  isOpen: boolean;
  onToggle: () => void;
}

const AccordionItem: React.FC<AccordionItemProps> = ({ title, children, isOpen, onToggle }) => {
  return (
    <div className="border border-gray-200 rounded-lg mb-3">
      <button
        className="w-full px-4 py-3 text-left bg-gray-50 hover:bg-gray-100 rounded-t-lg flex justify-between items-center transition-colors"
        onClick={onToggle}
      >
        <span className="font-medium text-gray-800">{title}</span>
        <span className={`transform transition-transform ${isOpen ? 'rotate-180' : ''}`}>
          ▼
        </span>
      </button>
      {isOpen && (
        <div className="px-4 py-4 bg-white rounded-b-lg">
          {children}
        </div>
      )}
    </div>
  );
};

interface SliderSectionProps {
  value: number;
  min: number;
  max: number;
  step?: number;
  onChange: (value: number) => void;
  descriptions: { value: number; label: string; description: string }[];
}

const SliderSection: React.FC<SliderSectionProps> = ({ value, min, max, step = 1, onChange, descriptions }) => {
  const currentDesc = descriptions.find(desc => desc.value === value);
  
  return (
    <div className="space-y-4">
      <div className="flex items-center space-x-4">
        <span className="text-sm font-medium text-gray-600 w-12">{min}</span>
        <div className="flex-1">
          <input
            type="range"
            min={min}
            max={max}
            step={step}
            value={value}
            onChange={(e) => onChange(parseInt(e.target.value))}
            className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer slider"
            style={{
              background: `linear-gradient(to right, #3b82f6 0%, #3b82f6 ${((value - min) / (max - min)) * 100}%, #e5e7eb ${((value - min) / (max - min)) * 100}%, #e5e7eb 100%)`
            }}
          />
        </div>
        <span className="text-sm font-medium text-gray-600 w-12">{max}</span>
        <div className="bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm font-medium min-w-[3rem] text-center">
          {value}
        </div>
      </div>
      {currentDesc && (
        <div className="bg-gray-50 p-3 rounded-lg border-l-4 border-blue-500">
          <div className="font-medium text-gray-800 mb-1">{currentDesc.label}</div>
          <div className="text-sm text-gray-600">{currentDesc.description}</div>
        </div>
      )}
    </div>
  );
};

export default function OpinionProfileSettings({ opinionProfile, onOpinionProfileChange }: OpinionProfileSettingsProps) {
  const [openSections, setOpenSections] = useState<Record<string, boolean>>({
    domain: true,
    skill: false,
    structure: false,
    difficulty: false
  });

  const toggleSection = (section: string) => {
    setOpenSections(prev => ({
      ...prev,
      [section]: !prev[section]
    }));
  };

  const updateProfile = (updates: Partial<OpinionProfile>) => {
    onOpinionProfileChange({ ...opinionProfile, ...updates });
  };

  return (
    <div className="space-y-4">
      <div className="mb-6">
        <h3 className="text-lg font-semibold text-gray-800 mb-2">
          高校入試数学・最終問題「県別スタイルプロファイル」評価基準 Ver. 4.0
        </h3>
        <p className="text-sm text-gray-600">
          47都道府県の公立高等学校入学者選抜学力検査問題（数学）の最終大問を分析し、その特性を定量的にプロファイリングします。
        </p>
      </div>

      <AccordionItem
        title="指標1：出題分野コード (1-6)"
        isOpen={openSections.domain}
        onToggle={() => toggleSection('domain')}
      >
        <div className="space-y-4">
          <div className="grid grid-cols-1 gap-2">
            {domainOptions.map(option => (
              <label key={option.value} className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-gray-50 cursor-pointer">
                <input
                  type="radio"
                  name="domain"
                  value={option.value}
                  checked={opinionProfile.domain === option.value}
                  onChange={() => updateProfile({ domain: option.value })}
                  className="mt-1 text-blue-600"
                />
                <div className="flex-1">
                  <div className="font-medium text-gray-800">{option.value}. {option.label}</div>
                  <div className="text-sm text-gray-600">{option.description}</div>
                </div>
              </label>
            ))}
          </div>
        </div>
      </AccordionItem>

      <AccordionItem
        title="指標2：コアスキル評価 (1-10)"
        isOpen={openSections.skill}
        onToggle={() => toggleSection('skill')}
      >
        <SliderSection
          value={opinionProfile.skill_level}
          min={1}
          max={10}
          onChange={(value) => updateProfile({ skill_level: value })}
          descriptions={skillLevelDescriptions}
        />
      </AccordionItem>

      <AccordionItem
        title="指標3：問題構造評価 [A, B] (各1-10)"
        isOpen={openSections.structure}
        onToggle={() => toggleSection('structure')}
      >
        <div className="space-y-6">
          <div>
            <h4 className="font-medium text-gray-800 mb-3">A: 読解・設定の複雑度</h4>
            <SliderSection
              value={opinionProfile.structure_complexity[0]}
              min={1}
              max={10}
              onChange={(value) => updateProfile({ 
                structure_complexity: [value, opinionProfile.structure_complexity[1]] 
              })}
              descriptions={readingComplexityDescriptions}
            />
          </div>
          <div>
            <h4 className="font-medium text-gray-800 mb-3">B: 設問の誘導性</h4>
            <SliderSection
              value={opinionProfile.structure_complexity[1]}
              min={1}
              max={10}
              onChange={(value) => updateProfile({ 
                structure_complexity: [opinionProfile.structure_complexity[0], value] 
              })}
              descriptions={guidanceDescriptions}
            />
          </div>
        </div>
      </AccordionItem>

      <AccordionItem
        title="指標4：総合難易度スコア (1-20)"
        isOpen={openSections.difficulty}
        onToggle={() => toggleSection('difficulty')}
      >
        <SliderSection
          value={opinionProfile.difficulty_score}
          min={1}
          max={20}
          onChange={(value) => updateProfile({ difficulty_score: value })}
          descriptions={difficultyDescriptions}
        />
      </AccordionItem>
    </div>
  );
}
