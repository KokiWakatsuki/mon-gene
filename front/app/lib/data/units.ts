// 単元の階層構造データ

export interface UnitItem {
  id: string;
  label: string;
  children?: UnitItem[];
}

// 中学数学の単元階層構造（学年別）
export const UNITS_HIERARCHY: UnitItem[] = [
  {
    id: 'grade1',
    label: '中学1年',
    children: [
      {
        id: 'grade1-function',
        label: '関数',
        children: [
          { id: 'grade1-function-1', label: '比例式を求める・グラフ' },
          { id: 'grade1-function-2', label: '反比例の式・グラフ' },
        ],
      },
      {
        id: 'grade1-geometry',
        label: '図形',
        children: [
          {
            id: 'grade1-geometry-plane',
            label: '平面図形',
            children: [
              { id: 'grade1-geometry-plane-1', label: '直線と角' },
              { id: 'grade1-geometry-plane-2', label: '図形の移動' },
              { id: 'grade1-geometry-plane-3', label: '円' },
              { id: 'grade1-geometry-plane-4', label: 'おうぎ形' },
            ],
          },
          {
            id: 'grade1-geometry-space',
            label: '空間図形',
            children: [
              { id: 'grade1-geometry-space-1', label: '直線と平面の位置関係' },
              { id: 'grade1-geometry-space-2', label: '立体の表面積' },
              { id: 'grade1-geometry-space-3', label: '立体の体積' },
            ],
          },
        ],
      },
      {
        id: 'grade1-data',
        label: '資料の整理',
        children: [
          { id: 'grade1-data-1', label: '度数・相対度数' },
          { id: 'grade1-data-2', label: '代表値（平均・中央値・最頻値）' },
          { id: 'grade1-data-3', label: '真の値・有効数字' },
        ],
      },
    ],
  },
  {
    id: 'grade2',
    label: '中学2年',
    children: [
      { id: 'grade2-equation', label: '連立方程式' },
      {
        id: 'grade2-function',
        label: '一次関数',
        children: [
          { id: 'grade2-function-1', label: '変化の割合' },
          { id: 'grade2-function-2', label: 'グラフ・式の求め方' },
        ],
      },
      {
        id: 'grade2-geometry',
        label: '合同な図形',
        children: [
          { id: 'grade2-geometry-1', label: '図形の性質' },
          { id: 'grade2-geometry-2', label: '三角形と四角形' },
        ],
      },
      {
        id: 'grade2-probability',
        label: '確率',
        children: [
          { id: 'grade2-probability-1', label: '確率の求め方' },
          { id: 'grade2-probability-2', label: 'いろいろな確率' },
        ],
      },
    ],
  },
  {
    id: 'grade3',
    label: '中学3年',
    children: [
      { id: 'grade3-calculation', label: '式の計算' },
      { id: 'grade3-sqrt', label: '平方根' },
      { id: 'grade3-quadratic', label: '二次方程式' },
      {
        id: 'grade3-function',
        label: '二次関数（2乗に比例する関数）',
        children: [
          { id: 'grade3-function-1', label: '関数の決定・グラフ' },
          { id: 'grade3-function-2', label: '変域・値の変化' },
          { id: 'grade3-function-3', label: '放物線と直線' },
        ],
      },
      {
        id: 'grade3-similarity',
        label: '相似な図形',
        children: [
          { id: 'grade3-similarity-1', label: '拡大・縮小' },
          { id: 'grade3-similarity-2', label: '相似条件' },
          { id: 'grade3-similarity-3', label: '三角形と比' },
          { id: 'grade3-similarity-4', label: '中点連結定理・平行線と線分の比' },
          { id: 'grade3-similarity-5', label: '面積比・表面積比・体積比' },
        ],
      },
      { id: 'grade3-circle', label: '円周角の定理' },
      {
        id: 'grade3-pythagorean',
        label: '三平方の定理',
        children: [
          { id: 'grade3-pythagorean-1', label: '定理とその逆' },
          { id: 'grade3-pythagorean-2', label: '平面図形への利用' },
          { id: 'grade3-pythagorean-3', label: '空間図形への利用' },
        ],
      },
      {
        id: 'grade3-data',
        label: '資料の活用',
        children: [
          { id: 'grade3-data-1', label: '母集団と標本' },
          { id: 'grade3-data-2', label: '標本抽出と平均' },
          { id: 'grade3-data-3', label: '母集団の推定' },
        ],
      },
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