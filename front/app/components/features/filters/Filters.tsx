'use client';

import React from 'react';

interface FilterOption {
  label: string;
  value: string;
}

interface FilterGroup {
  label: string;
  options: FilterOption[];
  allowMultiple?: boolean;
}

interface FiltersProps {
  filterGroups: FilterGroup[];
  selectedFilters: Record<string, string[]>;
  onFilterChange: (groupLabel: string, value: string, allowMultiple: boolean) => void;
}

export default function Filters({ filterGroups, selectedFilters, onFilterChange }: FiltersProps) {
  return (
    <section className="grid gap-2.5 mb-6" aria-label="検索フィルター">
      {filterGroups.map((group) => (
        <div key={group.label} className="flex items-center gap-2 flex-wrap">
          <div className="font-semibold text-mongene-ink min-w-16">
            {group.label}:
          </div>
          {group.options.map((option) => {
            const isSelected = selectedFilters[group.label]?.includes(option.value) || false;
            return (
              <button
                key={option.value}
                className={`
                  border px-2.5 py-1.5 rounded-full cursor-pointer text-sm transition-colors
                  ${isSelected 
                    ? 'border-mongene-blue bg-mongene-blue text-white' 
                    : 'border-mongene-border bg-white hover:border-gray-400'
                  }
                `}
                onClick={() => onFilterChange(group.label, option.value, group.allowMultiple || false)}
              >
                {option.label}
              </button>
            );
          })}
        </div>
      ))}
    </section>
  );
}
