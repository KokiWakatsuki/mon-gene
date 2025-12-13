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
    <div className={onSolutionFilesChange && uploadedSolutionFiles !== undefined ? "grid grid-cols-1 md:grid-cols-2 gap-6" : ""}>
      {/* 問題画像アップロード */}
      <section className="bg-white border border-gray-200 rounded-xl p-6 shadow-[0_2px_4px_rgba(0,0,0,0.03)] mb-0">
        <h3 className="flex items-center gap-2 text-lg font-semibold mb-4 m-0">
          <svg className="icon-mr" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#fe2020">
            <path d="M478-240q21 0 35.5-14.5T528-290q0-21-14.5-35.5T478-340q-21 0-35.5 14.5T428-290q0 21 14.5 35.5T478-240Zm-36-154h74q0-33 7.5-52t42.5-52q26-26 41-49.5t15-56.5q0-56-41-86t-97-30q-57 0-92.5 30T342-618l66 26q5-18 22.5-39t53.5-21q32 0 48 17.5t16 38.5q0 20-12 37.5T506-526q-44 39-54 59t-10 73Zm38 314q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z"/>
          </svg>
          問題画像
        </h3>
        
        <div
          className={`border-2 border-dashed rounded-lg p-8 text-center bg-gray-100 cursor-pointer transition-all mb-4 ${
            isDragOver ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-blue-500 hover:bg-blue-50'
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
            <div className="text-gray-500 mb-3">
              <svg width="40" height="40" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5Zm10.5-11.25h.008v.008h-.008V8.25Zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0Z" />
              </svg>
            </div>
            <p className="text-base font-semibold text-gray-800 mb-2">
              画像をドラッグ＆ドロップ<br />
              <span className="text-blue-500 underline cursor-pointer">ファイルを選択</span>
            </p>
          </div>
        </div>

        {uploadedFiles.length > 0 && (
          <div className="grid grid-cols-4 gap-3">
            {uploadedFiles.map((file, index) => {
              const isImage = file.type.startsWith('image/') && !file.name.toLowerCase().endsWith('.heic') && !file.name.toLowerCase().endsWith('.heif');
              const fileIcon = getFileIcon(file);

              return (
                <div key={index} className="relative border border-gray-200 rounded-lg bg-white overflow-hidden shadow-[0_1px_3px_rgba(0,0,0,0.05)] aspect-square">
                  <div className="w-full h-full flex items-center justify-center bg-gray-100">
                    {isImage ? (
                      <img src={fileIcon} alt={file.name} className="w-full h-full object-cover" />
                    ) : (
                      <span className={`text-xs font-bold text-white px-2 py-1 rounded ${
                        fileIcon === 'PDF' ? 'bg-red-500' : 'bg-gray-500'
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
                    className="absolute top-1 right-1 w-6 h-6 rounded-full bg-black/60 text-white flex items-center justify-center hover:bg-black/80 transition-colors"
                  >
                    ×
                  </button>
                  <div className="absolute bottom-0 left-0 right-0 bg-white border-t border-gray-200 px-1.5 py-1.5">
                    <p className="text-[11px] text-center truncate" title={file.name}>
                      {file.name}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>

      {/* 解答画像アップロード（任意） */}
      {onSolutionFilesChange && uploadedSolutionFiles !== undefined && (
        <section className="bg-white border border-gray-200 rounded-xl p-6 shadow-[0_2px_4px_rgba(0,0,0,0.03)] mb-0">
          <h3 className="flex items-center gap-2 text-lg font-semibold mb-4 m-0">
            <svg className="icon-mr" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#3DD200">
              <path d="M276-280h76l40-112h176l40 112h76L520-720h-80L276-280Zm138-176 64-182h4l64 182H414Zm66 376q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-400Zm0 320q133 0 226.5-93.5T800-480q0-133-93.5-226.5T480-800q-133 0-226.5 93.5T160-480q0 133 93.5 226.5T480-160Z"/>
            </svg>
            解答画像 (任意)
          </h3>

          <div
            className={`border-2 border-dashed rounded-lg p-8 text-center bg-gray-100 cursor-pointer transition-all mb-4 ${
              isSolutionDragOver ? 'border-green-500 bg-green-50' : 'border-gray-200 hover:border-green-500 hover:bg-green-50'
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
              <div className="text-gray-500 mb-3">
                <svg width="40" height="40" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9Z" />
                </svg>
              </div>
              <p className="text-base font-semibold text-gray-800 mb-2">
                画像をドラッグ＆ドロップ<br />
                <span className="text-green-500 underline cursor-pointer">ファイルを選択</span>
              </p>
            </div>
          </div>

          {uploadedSolutionFiles.length > 0 && (
            <div className="grid grid-cols-4 gap-3">
              {uploadedSolutionFiles.map((file, index) => {
                const isImage = file.type.startsWith('image/') && !file.name.toLowerCase().endsWith('.heic') && !file.name.toLowerCase().endsWith('.heif');
                const fileIcon = getFileIcon(file);

                return (
                  <div key={index} className="relative border border-gray-200 rounded-lg bg-white overflow-hidden shadow-[0_1px_3px_rgba(0,0,0,0.05)] aspect-square">
                    <div className="w-full h-full flex items-center justify-center bg-gray-100">
                      {isImage ? (
                        <img src={fileIcon} alt={file.name} className="w-full h-full object-cover" />
                      ) : (
                        <span className={`text-xs font-bold text-white px-2 py-1 rounded ${
                          fileIcon === 'PDF' ? 'bg-red-500' : 'bg-gray-500'
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
                      className="absolute top-1 right-1 w-6 h-6 rounded-full bg-black/60 text-white flex items-center justify-center hover:bg-black/80 transition-colors"
                    >
                      ×
                    </button>
                    <div className="absolute bottom-0 left-0 right-0 bg-white border-t border-gray-200 px-1.5 py-1.5">
                      <p className="text-[11px] text-center truncate" title={file.name}>
                        {file.name}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>
      )}
    </div>
  );
}