'use client';

import React, { useState, useEffect } from 'react';

interface CheckFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  problemId: string;
  problemTitle: string;
  initialCheckInfo?: CheckInfo;
  onSave: (checkInfo: CheckInfo) => Promise<void>;
}

export interface CheckInfo {
  problem_text_ok: boolean;
  solution_ok: boolean;
  figure_ok: boolean;
  units: string[];
  year: string;
  exam_session: string;
}

const AVAILABLE_UNITS = [
  '多項式（展開・因数分解）',
  '平方根',
  '二次方程式',
  '関数 y=ax²',
  '図形の相似',
  '円の性質（円周角）',
  '三平方の定理',
  '標本調査',
];

const YEARS = ['2020', '2021', '2022', '2023', '2024', '2025'];
const EXAM_SESSIONS = ['第1回', '第2回', '第3回', 'プレ', '追試'];

export default function CheckFormModal({
  isOpen,
  onClose,
  problemId,
  problemTitle,
  initialCheckInfo,
  onSave,
}: CheckFormModalProps) {
  const [checkInfo, setCheckInfo] = useState<CheckInfo>({
    problem_text_ok: false,
    solution_ok: false,
    figure_ok: false,
    units: [],
    year: '',
    exam_session: '',
  });
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && initialCheckInfo) {
      setCheckInfo(initialCheckInfo);
    } else if (isOpen) {
      // リセット
      setCheckInfo({
        problem_text_ok: false,
        solution_ok: false,
        figure_ok: false,
        units: [],
        year: '',
        exam_session: '',
      });
    }
    setError(null);
  }, [isOpen, initialCheckInfo]);

  const handleCheckboxChange = (field: 'problem_text_ok' | 'solution_ok' | 'figure_ok') => {
    setCheckInfo((prev) => ({
      ...prev,
      [field]: !prev[field],
    }));
  };

  const handleUnitToggle = (unit: string) => {
    setCheckInfo((prev) => ({
      ...prev,
      units: prev.units.includes(unit)
        ? prev.units.filter((u) => u !== unit)
        : [...prev.units, unit],
    }));
  };

  const handleYearChange = (year: string) => {
    setCheckInfo((prev) => ({
      ...prev,
      year,
    }));
  };

  const handleExamSessionChange = (session: string) => {
    setCheckInfo((prev) => ({
      ...prev,
      exam_session: session,
    }));
  };

  const isComplete = () => {
    return (
      checkInfo.problem_text_ok &&
      checkInfo.solution_ok &&
      checkInfo.figure_ok &&
      checkInfo.units.length > 0 &&
      checkInfo.year !== '' &&
      checkInfo.exam_session !== ''
    );
  };

  const handleSave = async () => {
    setIsSaving(true);
    setError(null);
    try {
      await onSave(checkInfo);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'チェック情報の保存に失敗しました');
    } finally {
      setIsSaving(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl max-w-2xl w-full max-h-[90vh] overflow-auto">
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-bold text-gray-800">問題チェック - {problemTitle}</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-800 text-2xl font-bold w-8 h-8 flex items-center justify-center"
            >
              ×
            </button>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-300 text-red-700 rounded-lg">
              {error}
            </div>
          )}

          <div className="space-y-6">
            {/* 基本チェック項目 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">基本チェック</h3>
              <div className="space-y-2">
                <label className="flex items-center space-x-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={checkInfo.problem_text_ok}
                    onChange={() => handleCheckboxChange('problem_text_ok')}
                    className="w-5 h-5 text-green-500 rounded focus:ring-2 focus:ring-green-500"
                  />
                  <span className="text-gray-700">1. 問題文が適切か</span>
                </label>
                <label className="flex items-center space-x-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={checkInfo.solution_ok}
                    onChange={() => handleCheckboxChange('solution_ok')}
                    className="w-5 h-5 text-green-500 rounded focus:ring-2 focus:ring-green-500"
                  />
                  <span className="text-gray-700">2. 解答・解説が適切か</span>
                </label>
                <label className="flex items-center space-x-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={checkInfo.figure_ok}
                    onChange={() => handleCheckboxChange('figure_ok')}
                    className="w-5 h-5 text-green-500 rounded focus:ring-2 focus:ring-green-500"
                  />
                  <span className="text-gray-700">3. 図が適切か</span>
                </label>
              </div>
            </div>

            {/* 使用単元 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">4. 使用されている単元・公式・定理（複数選択可）</h3>
              <div className="grid grid-cols-2 gap-2">
                {AVAILABLE_UNITS.map((unit) => (
                  <label key={unit} className="flex items-center space-x-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={checkInfo.units.includes(unit)}
                      onChange={() => handleUnitToggle(unit)}
                      className="w-4 h-4 text-blue-500 rounded focus:ring-2 focus:ring-blue-500"
                    />
                    <span className="text-sm text-gray-700">{unit}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* 年度 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">5. 年度</h3>
              <div className="flex flex-wrap gap-2">
                {YEARS.map((year) => (
                  <button
                    key={year}
                    onClick={() => handleYearChange(year)}
                    className={`px-4 py-2 rounded-lg font-medium transition-all ${
                      checkInfo.year === year
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {year}年
                  </button>
                ))}
              </div>
            </div>

            {/* 回数 */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-gray-800">6. 回数</h3>
              <div className="flex flex-wrap gap-2">
                {EXAM_SESSIONS.map((session) => (
                  <button
                    key={session}
                    onClick={() => handleExamSessionChange(session)}
                    className={`px-4 py-2 rounded-lg font-medium transition-all ${
                      checkInfo.exam_session === session
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {session}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-3 mt-6 pt-6 border-t">
            <button
              onClick={onClose}
              disabled={isSaving}
              className="px-4 py-2 border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              キャンセル
            </button>
            <button
              onClick={handleSave}
              disabled={isSaving || !isComplete()}
              className={`px-4 py-2 rounded-lg font-semibold transition-all ${
                isComplete() && !isSaving
                  ? 'bg-green-500 text-white hover:bg-green-600'
                  : 'bg-gray-300 text-gray-500 cursor-not-allowed'
              }`}
            >
              {isSaving ? '保存中...' : 'チェック完了'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}