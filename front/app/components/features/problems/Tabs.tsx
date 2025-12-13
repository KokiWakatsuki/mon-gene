'use client';

import React from 'react';

interface TabsProps {
  subjects: string[];
  activeSubject: string;
  onSubjectChange: (subject: string) => void;
}

export default function Tabs({ subjects, activeSubject, onSubjectChange }: TabsProps) {
  return (
    <nav className="flex gap-6 px-4 pt-6 pb-0 max-w-6xl mx-auto" aria-label="科目タブ">
      <button
        className={`relative font-sans text-lg font-semibold bg-transparent border-none px-1 py-2 cursor-pointer transition-colors ${
          activeSubject === '数学'
            ? 'text-gray-800 after:content-[""] after:absolute after:bottom-[-2px] after:left-0 after:right-0 after:h-1 after:bg-green-500 after:rounded-sm'
            : 'text-gray-500 hover:text-gray-800'
        }`}
        onClick={() => onSubjectChange('数学')}
        aria-current={activeSubject === '数学' ? 'page' : undefined}
      >
        数学
      </button>
    </nav>
  );
}
