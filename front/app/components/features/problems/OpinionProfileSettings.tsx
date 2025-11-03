'use client';

import { useState } from 'react';

interface OpinionProfileV2 {
  // 1. 文章量・構成に関する指標
  problem_text_length: number;
  sub_problem_text_length: number;
  given_values_count: number;
  sub_problem_count: number;
  sub_problem_types: string[];
  solid_composition: string;
  
  // 2. 解答形式に関する指標
  answer_formats: string[];
  answer_units: string[];
  uses_auxiliary_points: boolean;
  
  // 3. 使用単元に関する指標
  setup_units: string[];
  solution_units: string[];
  
  // 4. 図形に関する指標
  total_vertices: number;
  has_moving_point: boolean;
  figure_values_count: number;
  
  // 5. 解法プロセスと認知負荷に関する指標
  solution_steps: number;
  has_logical_branching: boolean;
  theorem_count: number;
  requires_multi_unit_integration: boolean;
  has_irrelevant_info: boolean;
}

interface OpinionProfileSettingsProps {
  opinionProfile: OpinionProfileV2;
  onOpinionProfileChange: (profile: OpinionProfileV2) => void;
}

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

export default function OpinionProfileSettings({ opinionProfile, onOpinionProfileChange }: OpinionProfileSettingsProps) {
  const [openSections, setOpenSections] = useState<Record<string, boolean>>({
    textComposition: true,
    answerFormat: false,
    units: false,
    geometry: false,
    solutionProcess: false
  });

  const toggleSection = (section: string) => {
    setOpenSections(prev => ({
      ...prev,
      [section]: !prev[section]
    }));
  };

  const updateProfile = (updates: Partial<OpinionProfileV2>) => {
    onOpinionProfileChange({ ...opinionProfile, ...updates });
  };

  const toggleArrayItem = (array: string[], item: string) => {
    if (array.includes(item)) {
      return array.filter(i => i !== item);
    } else {
      return [...array, item];
    }
  };

  return (
    <div className="space-y-4">
      <div className="mb-6">
        <h3 className="text-lg font-semibold text-gray-800 mb-2">
          意見プロファイル指標一覧（Ver. 2.0）
        </h3>
        <p className="text-sm text-gray-600">
          空間図形問題の詳細な特性を定量的に設定します。
        </p>
      </div>

      {/* 1. 文章量・構成に関する指標 */}
      <AccordionItem
        title="1. 文章量・構成に関する指標"
        isOpen={openSections.textComposition}
        onToggle={() => toggleSection('textComposition')}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              大問の問題文文字数: {opinionProfile.problem_text_length}
            </label>
            <input
              type="range"
              min="0"
              max="500"
              step="10"
              value={opinionProfile.problem_text_length}
              onChange={(e) => updateProfile({ problem_text_length: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              小問の総文字数: {opinionProfile.sub_problem_text_length}
            </label>
            <input
              type="range"
              min="0"
              max="300"
              step="10"
              value={opinionProfile.sub_problem_text_length}
              onChange={(e) => updateProfile({ sub_problem_text_length: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              与えられる数値の個数: {opinionProfile.given_values_count}
            </label>
            <input
              type="range"
              min="0"
              max="10"
              value={opinionProfile.given_values_count}
              onChange={(e) => updateProfile({ given_values_count: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              小問の数: {opinionProfile.sub_problem_count}
            </label>
            <input
              type="range"
              min="1"
              max="5"
              value={opinionProfile.sub_problem_count}
              onChange={(e) => updateProfile({ sub_problem_count: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">小問ごとの要求種別</label>
            <div className="flex flex-wrap gap-2">
              {['長さを求める', '面積を求める', '体積を求める', '最短距離を求める', '角度を求める', '比を求める'].map(type => (
                <button
                  key={type}
                  onClick={() => updateProfile({ sub_problem_types: toggleArrayItem(opinionProfile.sub_problem_types, type) })}
                  className={`px-3 py-1 rounded-full text-sm ${
                    opinionProfile.sub_problem_types.includes(type)
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  {type}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">立体の構成</label>
            <select
              value={opinionProfile.solid_composition}
              onChange={(e) => updateProfile({ solid_composition: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            >
              <option value="">選択してください</option>
              <option value="単一の立体">単一の立体</option>
              <option value="複数の立体の組み合わせ">複数の立体の組み合わせ</option>
            </select>
          </div>
        </div>
      </AccordionItem>

      {/* 2. 解答形式に関する指標 */}
      <AccordionItem
        title="2. 解答形式に関する指標"
        isOpen={openSections.answerFormat}
        onToggle={() => toggleSection('answerFormat')}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">解答の形式</label>
            <div className="flex flex-wrap gap-2">
              {['整数', '既約分数', '無理数(√)'].map(format => (
                <button
                  key={format}
                  onClick={() => updateProfile({ answer_formats: toggleArrayItem(opinionProfile.answer_formats, format) })}
                  className={`px-3 py-1 rounded-full text-sm ${
                    opinionProfile.answer_formats.includes(format)
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  {format}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">小問ごとの要求単位</label>
            <div className="flex flex-wrap gap-2">
              {['cm', 'cm²', 'cm³', '度', '単位なし'].map(unit => (
                <button
                  key={unit}
                  onClick={() => updateProfile({ answer_units: toggleArrayItem(opinionProfile.answer_units, unit) })}
                  className={`px-3 py-1 rounded-full text-sm ${
                    opinionProfile.answer_units.includes(unit)
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  {unit}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={opinionProfile.uses_auxiliary_points}
                onChange={(e) => updateProfile({ uses_auxiliary_points: e.target.checked })}
                className="rounded"
              />
              <span className="text-sm text-gray-700">正答例での補助点の使用</span>
            </label>
          </div>
        </div>
      </AccordionItem>

      {/* 3. 使用単元に関する指標 */}
      <AccordionItem
        title="3. 使用単元に関する指標"
        isOpen={openSections.units}
        onToggle={() => toggleSection('units')}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">単元（設定）</label>
            <div className="flex flex-wrap gap-2">
              {['正方形', '長方形', '正三角形', '二等辺三角形', '直角三角形', '台形', '円', 
                '直方体', '立方体', '正四角すい', '三角すい', '円すい', '三角柱', '円柱',
                '平行', '垂直・垂線', '中点', '交点', '平面'].map(unit => (
                <button
                  key={unit}
                  onClick={() => updateProfile({ setup_units: toggleArrayItem(opinionProfile.setup_units, unit) })}
                  className={`px-3 py-1 rounded-full text-sm ${
                    opinionProfile.setup_units.includes(unit)
                      ? 'bg-green-500 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  {unit}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">単元（解法）</label>
            <div className="flex flex-wrap gap-2">
              {['三平方の定理', '相似', '中点連結定理', '円周角の定理',
                '面積の公式', '体積の公式',
                '平行と比', '相似比', '面積比', '体積比', '二等辺三角形の性質', '正三角形の性質',
                '展開図', '補助線'].map(unit => (
                <button
                  key={unit}
                  onClick={() => updateProfile({ solution_units: toggleArrayItem(opinionProfile.solution_units, unit) })}
                  className={`px-3 py-1 rounded-full text-sm ${
                    opinionProfile.solution_units.includes(unit)
                      ? 'bg-purple-500 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  {unit}
                </button>
              ))}
            </div>
          </div>
        </div>
      </AccordionItem>

      {/* 4. 図形に関する指標 */}
      <AccordionItem
        title="4. 図形に関する指標"
        isOpen={openSections.geometry}
        onToggle={() => toggleSection('geometry')}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              総頂点・点の数: {opinionProfile.total_vertices}
            </label>
            <input
              type="range"
              min="0"
              max="20"
              value={opinionProfile.total_vertices}
              onChange={(e) => updateProfile({ total_vertices: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>

          <div>
            <label className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={opinionProfile.has_moving_point}
                onChange={(e) => updateProfile({ has_moving_point: e.target.checked })}
                className="rounded"
              />
              <span className="text-sm text-gray-700">動点の有無</span>
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              図中に明記された数値の個数: {opinionProfile.figure_values_count}
            </label>
            <input
              type="range"
              min="0"
              max="15"
              value={opinionProfile.figure_values_count}
              onChange={(e) => updateProfile({ figure_values_count: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>
        </div>
      </AccordionItem>

      {/* 5. 解法プロセスと認知負荷に関する指標 */}
      <AccordionItem
        title="5. 解法プロセスと認知負荷に関する指標"
        isOpen={openSections.solutionProcess}
        onToggle={() => toggleSection('solutionProcess')}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              解法のステップ数: {opinionProfile.solution_steps}
            </label>
            <input
              type="range"
              min="1"
              max="10"
              value={opinionProfile.solution_steps}
              onChange={(e) => updateProfile({ solution_steps: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>

          <div>
            <label className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={opinionProfile.has_logical_branching}
                onChange={(e) => updateProfile({ has_logical_branching: e.target.checked })}
                className="rounded"
              />
              <span className="text-sm text-gray-700">論理的分岐の有無（場合分け）</span>
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              使用定理・公式の数: {opinionProfile.theorem_count}
            </label>
            <input
              type="range"
              min="0"
              max="8"
              value={opinionProfile.theorem_count}
              onChange={(e) => updateProfile({ theorem_count: parseInt(e.target.value) })}
              className="w-full"
            />
          </div>

          <div>
            <label className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={opinionProfile.requires_multi_unit_integration}
                onChange={(e) => updateProfile({ requires_multi_unit_integration: e.target.checked })}
                className="rounded"
              />
              <span className="text-sm text-gray-700">複数単元の知識統合の要否</span>
            </label>
          </div>

          <div>
            <label className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={opinionProfile.has_irrelevant_info}
                onChange={(e) => updateProfile({ has_irrelevant_info: e.target.checked })}
                className="rounded"
              />
              <span className="text-sm text-gray-700">無関係な情報の有無（外発的認知負荷）</span>
            </label>
          </div>
        </div>
      </AccordionItem>
    </div>
  );
}
