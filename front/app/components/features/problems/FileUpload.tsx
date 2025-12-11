'use client';

import React, { useState, useRef } from 'react';

interface FileUploadProps {
  uploadedFiles: File[];
  onFilesChange: (files: File[]) => void;
  uploadedSolutionFiles?: File[];
  onSolutionFilesChange?: (files: File[]) => void;
  selectedUnits?: string[];
  onSelectedUnitsChange?: (units: string[]) => void;
  label?: string;
  description?: string;
}

export default function FileUpload({
  uploadedFiles,
  onFilesChange,
  uploadedSolutionFiles,
  onSolutionFilesChange,
  selectedUnits = [],
  onSelectedUnitsChange,
  label = '画像アップロード',
  description = '対応形式: JPG, PNG, PDF, HEIC (複数選択可)'
}: FileUploadProps) {
  const [isDragOver, setIsDragOver] = useState(false);
  const [isSolutionDragOver, setIsSolutionDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const solutionFileInputRef = useRef<HTMLInputElement>(null);

  // 中学1~2年生の単元（自動的に選択済み、UIには表示しない）
  const grade1And2Units = [
    '正負の数',
    '文字と式',
    '方程式',
    '比例と反比例',
    '平面図形',
    '空間図形',
    'データの活用',
    '式の計算',
    '連立方程式',
    '一次関数',
    '図形の性質と合同',
    '図形の性質と証明',
    '確率'
  ];

  // 中学3年生の単元リスト（UIに表示）
  const grade3Units = [
    '式の展開・因数分解',
    '平方根',
    '二次方程式',
    '二次関数',
    '相似',
    '三平方の定理',
    '円',
    '標本調査'
  ];

  const toggleUnit = (unit: string) => {
    if (!onSelectedUnitsChange) return;
    
    // 使用する単元として選択/解除
    // 1~2年生の単元は常に含める
    const baseUnits = [...grade1And2Units];
    
    if (selectedUnits.includes(unit)) {
      // 選択解除：1~2年生の単元 + 選択解除後の3年生単元
      const newGrade3Units = selectedUnits.filter(u => u !== unit && !grade1And2Units.includes(u));
      onSelectedUnitsChange([...baseUnits, ...newGrade3Units]);
    } else {
      // 選択追加：1~2年生の単元 + 既存の3年生単元 + 新規選択単元
      const existingGrade3Units = selectedUnits.filter(u => !grade1And2Units.includes(u));
      onSelectedUnitsChange([...baseUnits, ...existingGrade3Units, unit]);
    }
  };

  // 3年生の単元のみを抽出（表示用）
  const selectedGrade3Units = selectedUnits.filter(u => grade3Units.includes(u));

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);

    const files = Array.from(e.dataTransfer.files);
    addFiles(files);
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      addFiles(files);
      e.target.value = ''; // リセット
    }
  };

  const addFiles = (files: File[]) => {
    if (files.length === 0) return;
    onFilesChange([...uploadedFiles, ...files]);
  };

  const removeFile = (index: number) => {
    const newFiles = uploadedFiles.filter((_, i) => i !== index);
    onFilesChange(newFiles);
  };

  // 解答ファイル用のハンドラー
  const handleSolutionDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsSolutionDragOver(true);
  };

  const handleSolutionDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsSolutionDragOver(false);
  };

  const handleSolutionDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleSolutionDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsSolutionDragOver(false);

    const files = Array.from(e.dataTransfer.files);
    addSolutionFiles(files);
  };

  const handleSolutionFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      addSolutionFiles(files);
      e.target.value = ''; // リセット
    }
  };

  const addSolutionFiles = (files: File[]) => {
    if (files.length === 0 || !onSolutionFilesChange || !uploadedSolutionFiles) return;
    onSolutionFilesChange([...uploadedSolutionFiles, ...files]);
  };

  const removeSolutionFile = (index: number) => {
    if (!onSolutionFilesChange || !uploadedSolutionFiles) return;
    const newFiles = uploadedSolutionFiles.filter((_, i) => i !== index);
    onSolutionFilesChange(newFiles);
  };

  const getFileIcon = (file: File) => {
    if (file.type.startsWith('image/')) {
      return URL.createObjectURL(file);
    } else if (file.type === 'application/pdf') {
      return 'PDF';
    } else if (file.name.toLowerCase().endsWith('.heic') || file.name.toLowerCase().endsWith('.heif')) {
      return 'HEIC';
    }
    return 'FILE';
  };

  return (
    <div className="space-y-6">
      {/* 単元選択UI（オプション） */}
      {onSelectedUnitsChange && (
        <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
          <h3 className="flex items-center gap-2 text-lg font-semibold text-gray-800 mb-4">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
            </svg>
            📚 使用する単元を選択（中学3年生）
          </h3>
          <p className="text-sm text-gray-600 mb-4">
            問題生成時に使用する中学3年生の単元を選択してください。中学1~2年生の単元は自動的に含まれます。
          </p>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2">
            {grade3Units.map((unit) => (
              <button
                key={unit}
                type="button"
                onClick={() => toggleUnit(unit)}
                className={`px-3 py-2 rounded-lg text-sm font-medium transition-all ${
                  selectedGrade3Units.includes(unit)
                    ? 'bg-blue-500 text-white shadow-md'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {unit}
              </button>
            ))}
          </div>
          {selectedGrade3Units.length > 0 && (
            <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
              <p className="text-sm text-blue-800">
                <strong>選択中の3年生単元:</strong> {selectedGrade3Units.join('、')}
              </p>
              <p className="text-xs text-gray-600 mt-1">
                ※ 中学1~2年生の全単元も自動的に含まれます
              </p>
            </div>
          )}
        </div>
      )}

      <div className={onSolutionFilesChange && uploadedSolutionFiles !== undefined ? "grid grid-cols-1 md:grid-cols-2 gap-6" : "space-y-6"}>
        {/* 問題ファイルアップロード */}
        <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
        <h3 className="flex items-center gap-2 text-lg font-semibold text-gray-800 mb-4">
          <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 16.5V9.75m0 0l3 3m-3-3l-3 3M6.75 19.5a4.5 4.5 0 01-1.41-8.775 5.25 5.25 0 0110.233-2.33 3 3 0 013.758 3.848A3.752 3.752 0 0118 19.5H6.75z" />
          </svg>
          📝 問題ファイルアップロード
        </h3>

      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-all ${
          isDragOver
            ? 'border-blue-500 bg-blue-50'
            : 'border-gray-200 bg-gray-50 hover:border-blue-400 hover:bg-blue-50'
        }`}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept=".jpg,.jpeg,.png,.pdf,.heic,.heif"
          multiple
          className="hidden"
          onChange={handleFileInputChange}
        />

        <div className="flex flex-col items-center">
          <svg width="40" height="40" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24" className="text-gray-400 mb-3">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 16.5V9.75m0 0l3 3m-3-3l-3 3M6.75 19.5a4.5 4.5 0 01-1.41-8.775 5.25 5.25 0 0110.233-2.33 3 3 0 013.758 3.848A3.752 3.752 0 0118 19.5H6.75z" />
          </svg>
          <p className="text-base font-semibold text-gray-800 mb-2">
            ここにファイルをドラッグ＆ドロップ<br />
            または <span className="text-blue-500 underline">ファイルを選択</span>
          </p>
          <p className="text-xs text-gray-500">
            {description}
          </p>
        </div>
      </div>

        {uploadedFiles.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3 mt-4">
            {uploadedFiles.map((file, index) => {
              const isImage = file.type.startsWith('image/') && !file.name.toLowerCase().endsWith('.heic') && !file.name.toLowerCase().endsWith('.heif');
              const fileIcon = getFileIcon(file);

              return (
                <div key={index} className="relative border border-gray-200 rounded-lg bg-white overflow-hidden shadow-sm aspect-square">
                  <div className="w-full h-full flex items-center justify-center bg-gray-50">
                    {isImage ? (
                      <img src={fileIcon} alt={file.name} className="w-full h-full object-cover" />
                    ) : (
                      <span className={`text-xs font-bold text-white px-2 py-1 rounded ${
                        fileIcon === 'PDF' ? 'bg-red-500' : fileIcon === 'HEIC' ? 'bg-purple-500' : 'bg-gray-500'
                      }`}>
                        {fileIcon}
                      </span>
                    )}
                  </div>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      removeFile(index);
                    }}
                    className="absolute top-1 right-1 w-6 h-6 rounded-full bg-black/60 text-white flex items-center justify-center hover:bg-black/80 transition-colors text-base leading-none z-10"
                  >
                    ×
                  </button>
                  <div className="absolute bottom-0 left-0 right-0 bg-white border-t border-gray-200 px-2 py-1.5">
                    <p className="text-[10px] text-gray-700 truncate text-center" title={file.name}>
                      {file.name}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* 解答ファイルアップロード（オプション） */}
      {onSolutionFilesChange && uploadedSolutionFiles !== undefined && (
        <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
          <h3 className="flex items-center gap-2 text-lg font-semibold text-gray-800 mb-4">
            <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            ✅ 解答ファイルアップロード（任意）
          </h3>

          <div
            className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-all ${
              isSolutionDragOver
                ? 'border-green-500 bg-green-50'
                : 'border-gray-200 bg-gray-50 hover:border-green-400 hover:bg-green-50'
            }`}
            onDragEnter={handleSolutionDragEnter}
            onDragLeave={handleSolutionDragLeave}
            onDragOver={handleSolutionDragOver}
            onDrop={handleSolutionDrop}
            onClick={() => solutionFileInputRef.current?.click()}
          >
            <input
              ref={solutionFileInputRef}
              type="file"
              accept=".jpg,.jpeg,.png,.pdf,.heic,.heif"
              multiple
              className="hidden"
              onChange={handleSolutionFileInputChange}
            />

            <div className="flex flex-col items-center">
              <svg width="40" height="40" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24" className="text-gray-400 mb-3">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p className="text-base font-semibold text-gray-800 mb-2">
                解答ファイルをドラッグ＆ドロップ<br />
                または <span className="text-green-500 underline">ファイルを選択</span>
              </p>
              <p className="text-xs text-gray-500">
                対応形式: JPG, PNG, PDF, HEIC (複数選択可)
              </p>
            </div>
          </div>

          {uploadedSolutionFiles.length > 0 && (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3 mt-4">
              {uploadedSolutionFiles.map((file, index) => {
                const isImage = file.type.startsWith('image/') && !file.name.toLowerCase().endsWith('.heic') && !file.name.toLowerCase().endsWith('.heif');
                const fileIcon = getFileIcon(file);

                return (
                  <div key={index} className="relative border border-green-200 rounded-lg bg-white overflow-hidden shadow-sm aspect-square">
                    <div className="w-full h-full flex items-center justify-center bg-green-50">
                      {isImage ? (
                        <img src={fileIcon} alt={file.name} className="w-full h-full object-cover" />
                      ) : (
                        <span className={`text-xs font-bold text-white px-2 py-1 rounded ${
                          fileIcon === 'PDF' ? 'bg-green-600' : fileIcon === 'HEIC' ? 'bg-green-500' : 'bg-green-400'
                        }`}>
                          {fileIcon}
                        </span>
                      )}
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        removeSolutionFile(index);
                      }}
                      className="absolute top-1 right-1 w-6 h-6 rounded-full bg-black/60 text-white flex items-center justify-center hover:bg-black/80 transition-colors text-base leading-none z-10"
                    >
                      ×
                    </button>
                    <div className="absolute bottom-0 left-0 right-0 bg-white border-t border-green-200 px-2 py-1.5">
                      <p className="text-[10px] text-gray-700 truncate text-center" title={file.name}>
                        {file.name}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
          </div>
        )}
      </div>
    </div>
  );
}