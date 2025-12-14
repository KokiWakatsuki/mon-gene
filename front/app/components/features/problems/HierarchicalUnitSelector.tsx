'use client';

import React, { useState, useEffect, useRef } from 'react';
import { UnitItem, getAllChildren } from '@/app/lib/data/units';

interface HierarchicalUnitSelectorProps {
  units: UnitItem[];
  selectedUnits: string[];
  onSelectionChange: (selected: string[]) => void;
}

export default function HierarchicalUnitSelector({
  units,
  selectedUnits,
  onSelectionChange,
}: HierarchicalUnitSelectorProps) {
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());

  // ノードの展開/折りたたみ
  const toggleNode = (nodeId: string) => {
    setExpandedNodes(prev => {
      const newSet = new Set(prev);
      if (newSet.has(nodeId)) {
        newSet.delete(nodeId);
      } else {
        newSet.add(nodeId);
      }
      return newSet;
    });
  };

  // チェックボックスの状態を判定
  const getCheckState = (item: UnitItem): 'checked' | 'unchecked' | 'indeterminate' => {
    if (!item.children || item.children.length === 0) {
      // 子がない場合は、選択されているかどうか
      return selectedUnits.includes(item.label) ? 'checked' : 'unchecked';
    }

    // 子がある場合は、すべての子孫の状態を確認
    const allDescendants = getAllChildren(item.id, units);
    const selectedDescendants = allDescendants.filter(label => selectedUnits.includes(label));

    if (selectedDescendants.length === 0) {
      return 'unchecked';
    } else if (selectedDescendants.length === allDescendants.length) {
      return 'checked';
    } else {
      return 'indeterminate';
    }
  };

  // チェックボックスの変更処理
  const handleCheckChange = (item: UnitItem) => {
    const currentState = getCheckState(item);
    const newSelected = new Set(selectedUnits);

    if (item.children && item.children.length > 0) {
      // 親ノードの場合：親自身とすべての子孫を取得
      const allDescendants = getAllChildren(item.id, units);
      
      if (currentState === 'checked') {
        // チェック済み → 親自身とすべての子孫のチェックを外す
        newSelected.delete(item.label);
        allDescendants.forEach(label => newSelected.delete(label));
      } else {
        // 未チェックまたは不確定 → 親自身とすべての子孫をチェック
        newSelected.add(item.label);
        allDescendants.forEach(label => newSelected.add(label));
      }
    } else {
      // 子ノードの場合：自分自身のみトグル
      if (newSelected.has(item.label)) {
        newSelected.delete(item.label);
      } else {
        newSelected.add(item.label);
      }
    }

    onSelectionChange(Array.from(newSelected));
  };

  // 再帰的にツリーをレンダリング
  const renderTree = (items: UnitItem[], level: number = 0) => {
    return items.map((item) => {
      const hasChildren = !!(item.children && item.children.length > 0);
      const isExpanded = expandedNodes.has(item.id);
      const checkState = getCheckState(item);

      return (
        <TreeNode
          key={item.id}
          item={item}
          level={level}
          hasChildren={hasChildren}
          isExpanded={isExpanded}
          checkState={checkState}
          onToggle={toggleNode}
          onCheckChange={handleCheckChange}
          renderTree={renderTree}
        />
      );
    });
  };

  return (
    <div className="space-y-1">
      {renderTree(units)}
    </div>
  );
}

// TreeNodeコンポーネント：チェックボックスの状態管理を分離
interface TreeNodeProps {
  item: UnitItem;
  level: number;
  hasChildren: boolean;
  isExpanded: boolean;
  checkState: 'checked' | 'unchecked' | 'indeterminate';
  onToggle: (nodeId: string) => void;
  onCheckChange: (item: UnitItem) => void;
  renderTree: (items: UnitItem[], level: number) => React.ReactNode[];
}

function TreeNode({
  item,
  level,
  hasChildren,
  isExpanded,
  checkState,
  onToggle,
  onCheckChange,
  renderTree,
}: TreeNodeProps) {
  const checkboxRef = useRef<HTMLInputElement>(null);

  // checkStateが変わるたびにスタイルを適用
  useEffect(() => {
    if (checkboxRef.current) {
      checkboxRef.current.indeterminate = checkState === 'indeterminate';
      
      if (checkState === 'indeterminate') {
        // !importantを使用して強制的にスタイルを適用
        checkboxRef.current.style.setProperty('background-color', '#93c5fd', 'important');
        checkboxRef.current.style.setProperty('border-color', '#60a5fa', 'important');
      } else {
        checkboxRef.current.style.removeProperty('background-color');
        checkboxRef.current.style.removeProperty('border-color');
      }
    }
  }, [checkState]);

  return (
    <div className="select-none">
      <div
        className={`flex items-center gap-2 py-2 px-3 rounded-lg transition-colors hover:bg-gray-50 ${
          level > 0 ? 'ml-' + (level * 4) : ''
        }`}
        style={{ paddingLeft: `${level * 16 + 12}px` }}
      >
        {/* 展開/折りたたみボタン */}
        {hasChildren ? (
          <button
            type="button"
            onClick={() => onToggle(item.id)}
            className="flex-shrink-0 w-5 h-5 flex items-center justify-center text-gray-500 hover:text-gray-700 transition-transform"
            style={{ transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)' }}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <polyline points="9 18 15 12 9 6"></polyline>
            </svg>
          </button>
        ) : (
          <div className="w-5" />
        )}

        {/* チェックボックス */}
        <label className="flex items-center gap-2 cursor-pointer flex-1">
          <div className="relative">
            <input
              ref={checkboxRef}
              type="checkbox"
              checked={checkState === 'checked' || checkState === 'indeterminate'}
              onChange={() => onCheckChange(item)}
              className="w-4 h-4 rounded border-gray-300 text-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-offset-0"
            />
          </div>
          <span
            className={`text-sm ${
              level === 0
                ? 'font-bold text-gray-800'
                : level === 1
                ? 'font-semibold text-gray-700'
                : 'font-normal text-gray-600'
            }`}
          >
            {item.label}
          </span>
        </label>
      </div>

      {/* 子要素 */}
      {hasChildren && isExpanded && (
        <div className="mt-1">
          {renderTree(item.children!, level + 1)}
        </div>
      )}
    </div>
  );
}