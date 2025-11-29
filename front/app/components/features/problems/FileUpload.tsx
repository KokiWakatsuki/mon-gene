'use client';

import React, { useState, useRef } from 'react';

interface FileUploadProps {
  uploadedFiles: File[];
  onFilesChange: (files: File[]) => void;
  uploadedSolutionFiles?: File[];
  onSolutionFilesChange?: (files: File[]) => void;
  label?: string;
  description?: string;
}

export default function FileUpload({
  uploadedFiles,
  onFilesChange,
  uploadedSolutionFiles,
  onSolutionFilesChange,
  label = '画像アップロード',
  description = '対応形式: JPG, PNG, PDF, HEIC (複数選択可)'
}: FileUploadProps) {
  const [isDragOver, setIsDragOver] = useState(false);
  const [isSolutionDragOver, setIsSolutionDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const solutionFileInputRef = useRef<HTMLInputElement>(null);

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
  );
}