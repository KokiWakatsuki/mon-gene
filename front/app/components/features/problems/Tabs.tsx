'use client';

import React from 'react';

interface TabsProps {
  subjects: string[];
  activeSubject: string;
  onSubjectChange: (subject: string) => void;
}

export default function Tabs({ subjects, activeSubject, onSubjectChange }: TabsProps) {
  return (
    <nav className="flex gap-6 pb-0 pt-6" aria-label="科目タブ">
      <button
        className="relative font-sans text-lg font-semibold text-gray-500 bg-transparent border-none px-1 py-2 cursor-pointer hover:text-gray-800 transition-colors data-[active=true]:text-gray-800 after:content-[''] after:absolute after:bottom-[-2px] after:left-0 after:right-0 after:h-1 after:bg-mongene-green after:rounded-sm after:opacity-0 data-[active=true]:after:opacity-100"
        data-active={activeSubject === '数学'}
        onClick={() => onSubjectChange('数学')}
        aria-current={activeSubject === '数学' ? 'page' : undefined}
      >
        数学
      </button>
    </nav>
  );
}
