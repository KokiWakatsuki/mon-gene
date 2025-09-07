'use client';

import React, { useState } from 'react';

interface TabsProps {
  subjects: string[];
  activeSubject: string;
  onSubjectChange: (subject: string) => void;
}

export default function Tabs({ subjects, activeSubject, onSubjectChange }: TabsProps) {
  return (
    <nav className="flex gap-3 pb-2.5 border-b border-mongene-border mb-4" aria-label="科目タブ">
      {subjects.map((subject) => (
        <button
          key={subject}
          className={`
            border-0 bg-transparent px-3 py-2 rounded-lg cursor-pointer font-semibold transition-colors
            ${activeSubject === subject 
              ? 'bg-mongene-green text-mongene-ink' 
              : 'text-mongene-ink hover:bg-gray-100'
            }
          `}
          onClick={() => onSubjectChange(subject)}
          aria-current={activeSubject === subject ? 'page' : undefined}
        >
          {subject}
        </button>
      ))}
    </nav>
  );
}
