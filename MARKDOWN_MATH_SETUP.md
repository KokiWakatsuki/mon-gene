# 数学記号・Markdown対応セットアップ手順

## 概要
5段階生成システムで生成される数学問題の表示を改善するため、数学記号とMarkdown表記に対応しました。

## 実装済み機能

### ✅ 簡易版 MarkdownRenderer
- 外部ライブラリに依存しない軽量実装
- 数学記号の適切な表示
- 基本的なMarkdown記法対応

### ✅ 対応する数学記号
- **ルート記号**: √22 → √22（適切なフォント表示）
- **上付き文字**: x² → x<sup>2</sup>
- **分数表示**: 3/2 → <sup>3</sup>⁄<sub>2</sub>
- **ベクトル記号**: OM⃗ → OM⃗（太字表示）
- **角度記号**: ∠ABC → ∠ABC
- **数学演算子**: ×、÷、±、≤、≥、≠、≈、≅、∽
- **矢印記号**: →、←
- **特殊記号**: π、°、∞

### ✅ 対応するMarkdown記法
- **見出し**: # ## ###
- **太字**: **text**
- **イタリック**: *text*
- **インラインコード**: `code`
- **改行**: 自動変換

## Docker環境でのフル機能追加手順

現在は簡易版を使用していますが、より高度な数学表記が必要な場合は以下の手順でフル機能版に更新できます。

### 1. package.jsonは既に更新済み
```json
{
  "dependencies": {
    "react-markdown": "^9.0.1",
    "remark-math": "^6.0.0",
    "rehype-katex": "^7.0.0",
    "katex": "^0.16.9"
  }
}
```

### 2. Dockerコンテナの再ビルド
```bash
# フロントエンドコンテナを再ビルド
docker-compose build front

# サービスを再起動
docker-compose up -d
```

### 3. フル機能版MarkdownRendererに切り替え
`front/app/components/ui/MarkdownRenderer.tsx`で、コメントアウトされているReactMarkdown + KaTeX版のコードを有効にします。

```typescript
// 簡易版（現在使用中）
import React from 'react';

// フル機能版（Docker再ビルド後に使用可能）
// import ReactMarkdown from 'react-markdown';
// import remarkMath from 'remark-math';
// import rehypeKatex from 'rehype-katex';
// import 'katex/dist/katex.min.css';
```

### 4. フル機能版の利点
- **LaTeX数式記法**: $\sqrt{22}$、$\frac{3}{2}$、$x^2$ など
- **高度な数式**: 積分、微分、行列、化学式等
- **完全なMarkdown対応**: テーブル、リスト、コードブロック等

## 現在の表示例

### 問題文での数学記号表示
```
底面が正方形の四角錐O-ABCDがある。
線分OMの長さ：3√13 cm
二面角αの正接の値：3/2
三角形OMNの面積：9√22 cm²
```

↓ 表示結果 ↓

```html
底面が正方形の四角錐O-ABCDがある。
線分OMの長さ：3<span class="math-symbol">√<span class="math-subscript">13</span></span> cm
二面角αの正接の値：<span class="math-fraction"><sup>3</sup>⁄<sub>2</sub></span>
三角形OMNの面積：9<span class="math-symbol">√22</span> cm<sup>2</sup>
```

## トラブルシューティング

### TypeScriptエラーが出る場合
Docker再ビルドが完了していない状態でフル機能版のimportを有効にするとエラーが発生します。
簡易版を使用するか、Dockerの再ビルドを完了してください。

### 数学記号が正しく表示されない場合
1. ブラウザのフォント設定を確認
2. Times New Roman フォントが利用可能か確認
3. CSSスタイルが正しく適用されているか確認

### 印刷時に数学記号が崩れる場合
印刷用CSSで数学記号のフォント指定を追加してください：

```css
@media print {
  .math-symbol, .math-vector, .math-fraction {
    font-family: 'Times New Roman', serif !important;
  }
}
```

## 今後の改善点

1. **数式エディター**: ユーザーが直接数式を編集できる機能
2. **プレビュー機能**: 編集中の数式をリアルタイムプレビュー
3. **テンプレート**: よく使う数学記号のテンプレート集
4. **エクスポート**: LaTeX形式でのエクスポート機能

---

この実装により、5段階生成システムで生成される高度な数学問題も、適切な数学記号表示で表示できるようになりました。
