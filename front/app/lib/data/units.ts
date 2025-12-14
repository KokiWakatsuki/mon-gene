// 単元の階層構造データ

export interface UnitItem {
  id: string;
  label: string;
  children?: UnitItem[];
}

// 中学数学の単元階層構造
export const UNITS_HIERARCHY: UnitItem[] = [
  {
    id: 'number',
    label: '数と式',
    children: [
      { id: 'number-1', label: '正の数・負の数' },
      { id: 'number-2', label: '文字と式' },
      { id: 'number-3', label: '一次方程式' },
      { id: 'number-4', label: '連立方程式' },
      { id: 'number-5', label: '多項式（展開・因数分解）' },
      { id: 'number-6', label: '平方根' },
      { id: 'number-7', label: '二次方程式' },
    ],
  },
  {
    id: 'function',
    label: '関数',
    children: [
      { id: 'function-1', label: '比例・反比例' },
      { id: 'function-2', label: '一次関数' },
      { id: 'function-3', label: '関数 y=ax²' },
    ],
  },
  {
    id: 'geometry',
    label: '図形',
    children: [
      { id: 'geometry-1', label: '平面図形' },
      { id: 'geometry-2', label: '空間図形' },
      { id: 'geometry-3', label: '図形の性質と合同' },
      { id: 'geometry-4', label: '図形の相似' },
      { id: 'geometry-5', label: '円の性質（円周角）' },
      { id: 'geometry-6', label: '三平方の定理' },
    ],
  },
  {
    id: 'data',
    label: 'データの活用',
    children: [
      { id: 'data-1', label: 'データの整理と活用' },
      { id: 'data-2', label: '確率' },
      { id: 'data-3', label: '標本調査' },
    ],
  },
];

// すべての単元をフラットなリストとして取得
export const getAllUnitsFlat = (): string[] => {
  const result: string[] = [];
  
  const traverse = (items: UnitItem[]) => {
    items.forEach(item => {
      result.push(item.label);
      if (item.children) {
        traverse(item.children);
      }
    });
  };
  
  traverse(UNITS_HIERARCHY);
  return result;
};

// 指定されたIDのすべての子孫を取得
export const getAllChildren = (id: string, items: UnitItem[]): string[] => {
  const result: string[] = [];
  
  const findAndTraverse = (items: UnitItem[]): boolean => {
    for (const item of items) {
      if (item.id === id) {
        if (item.children) {
          const traverse = (children: UnitItem[]) => {
            children.forEach(child => {
              result.push(child.label);
              if (child.children) {
                traverse(child.children);
              }
            });
          };
          traverse(item.children);
        }
        return true;
      }
      if (item.children && findAndTraverse(item.children)) {
        return true;
      }
    }
    return false;
  };
  
  findAndTraverse(items);
  return result;
};

// ラベルからIDを取得
export const getIdByLabel = (label: string, items: UnitItem[]): string | null => {
  for (const item of items) {
    if (item.label === label) {
      return item.id;
    }
    if (item.children) {
      const childId = getIdByLabel(label, item.children);
      if (childId) {
        return childId;
      }
    }
  }
  return null;
};