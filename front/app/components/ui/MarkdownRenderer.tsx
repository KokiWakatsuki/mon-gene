'use client';

import React from 'react';

interface MarkdownRendererProps {
  content: string;
  className?: string;
}

export default function MarkdownRenderer({ content, className = '' }: MarkdownRendererProps) {
  // 高度な数学記号の変換を行う関数（HTML版）
  const renderAdvancedMathSymbols = (text: string): string => {
    if (!text) return '';
    
    return text
      // LaTeX数式ブロック（$$...$$）の処理
      .replace(/\$\$([\s\S]*?)\$\$/g, (match, formula) => {
        return `<div class="math-block">${renderLatexToHtml(formula.trim())}</div>`;
      })
      // LaTeXインライン数式（$...$）の処理  
      .replace(/\$([^$\n]+)\$/g, (match, formula) => {
        return `<span class="math-inline">${renderLatexToHtml(formula.trim())}</span>`;
      })
      // ベクトル記号の改善（LaTeX形式 \overrightarrow{AB}）
      .replace(/\\overrightarrow\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
      // ベクトル記号（通常の矢印付き）
      .replace(/([A-Z]{1,3})⃗/g, '<span class="math-vector">$1→</span>')
      // ルート記号の変換
      .replace(/√(\d+)/g, '<span class="math-symbol">√$1</span>')
      .replace(/√\(([^)]+)\)/g, '<span class="math-symbol">√($1)</span>')
      .replace(/√([a-zA-Z]+)/g, '<span class="math-symbol">√$1</span>')
      // 上付き文字の変換（数学表記）
      .replace(/(\w+)²/g, '$1<sup>2</sup>')
      .replace(/(\w+)³/g, '$1<sup>3</sup>')
      .replace(/(\w+)⁴/g, '$1<sup>4</sup>')
      .replace(/(\w+)⁵/g, '$1<sup>5</sup>')
      // 角度記号の変換
      .replace(/∠([A-Z]+)/g, '<span class="math-symbol">∠$1</span>')
      // 分数の変換（分数表示）
      .replace(/(\d+)\/(\d+)/g, '<span class="math-fraction"><sup>$1</sup><span class="fraction-line">⁄</span><sub>$2</sub></span>')
      // その他の数学記号
      .replace(/×/g, '<span class="math-symbol">×</span>')
      .replace(/÷/g, '<span class="math-symbol">÷</span>')
      .replace(/°/g, '<span class="math-symbol">°</span>')
      .replace(/π/g, '<span class="math-symbol">π</span>')
      .replace(/∞/g, '<span class="math-symbol">∞</span>')
      .replace(/±/g, '<span class="math-symbol">±</span>')
      .replace(/≤/g, '<span class="math-symbol">≤</span>')
      .replace(/≥/g, '<span class="math-symbol">≥</span>')
      .replace(/≠/g, '<span class="math-symbol">≠</span>')
      .replace(/≈/g, '<span class="math-symbol">≈</span>')
      .replace(/≅/g, '<span class="math-symbol">≅</span>')
      .replace(/∽/g, '<span class="math-symbol">∽</span>')
      // 矢印記号
      .replace(/→/g, '<span class="math-symbol">→</span>')
      .replace(/←/g, '<span class="math-symbol">←</span>');
  };

  // LaTeX記法をHTMLに変換する関数
  const renderLatexToHtml = (latex: string): string => {
    return latex
      // 分数 \frac{a}{b}
      .replace(/\\frac\{([^}]+)\}\{([^}]+)\}/g, '<span class="math-fraction-block"><span class="numerator">$1</span><span class="fraction-line-block">─</span><span class="denominator">$2</span></span>')
      // 平方根 \sqrt{x}
      .replace(/\\sqrt\{([^}]+)\}/g, '<span class="math-symbol">√<span class="sqrt-content">$1</span></span>')
      // ベクトル \vec{a}
      .replace(/\\vec\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
      // 行列式 \begin{vmatrix}...\end{vmatrix}
      .replace(/\\begin\{vmatrix\}([\s\S]*?)\\end\{vmatrix\}/g, (match: string, content: string) => {
        const rows = content.trim().split('\\\\');
        const matrixRows = rows.map((row: string) => {
          const cells = row.split('&').map((cell: string) => cell.trim());
          return `<tr>${cells.map((cell: string) => `<td>${cell}</td>`).join('')}</tr>`;
        }).join('');
        return `<div class="matrix-container"><span class="matrix-bracket-left">|</span><table class="matrix">${matrixRows}</table><span class="matrix-bracket-right">|</span></div>`;
      })
      // 行列 \begin{pmatrix}...\end{pmatrix}
      .replace(/\\begin\{pmatrix\}([\s\S]*?)\\end\{pmatrix\}/g, (match: string, content: string) => {
        const rows = content.trim().split('\\\\');
        const matrixRows = rows.map((row: string) => {
          const cells = row.split('&').map((cell: string) => cell.trim());
          return `<tr>${cells.map((cell: string) => `<td>${cell}</td>`).join('')}</tr>`;
        }).join('');
        return `<div class="matrix-container"><span class="matrix-bracket-left">(</span><table class="matrix">${matrixRows}</table><span class="matrix-bracket-right">)</span></div>`;
      })
      // ベクトル記号
      .replace(/\\overrightarrow\{([^}]+)\}/g, '<span class="math-vector">$1→</span>')
      // 数学記号
      .replace(/\\times/g, '×')
      .replace(/\\cdot/g, '·')
      .replace(/\\pi/g, 'π')
      .replace(/\\infty/g, '∞')
      .replace(/\\pm/g, '±')
      .replace(/\\leq/g, '≤')
      .replace(/\\geq/g, '≥')
      .replace(/\\neq/g, '≠')
      .replace(/\\approx/g, '≈')
      .replace(/\\rightarrow/g, '→')
      .replace(/\\leftarrow/g, '←')
      // ギリシャ文字
      .replace(/\\alpha/g, 'α')
      .replace(/\\beta/g, 'β')
      .replace(/\\gamma/g, 'γ')
      .replace(/\\delta/g, 'δ')
      .replace(/\\theta/g, 'θ')
      .replace(/\\lambda/g, 'λ')
      .replace(/\\mu/g, 'μ')
      .replace(/\\phi/g, 'φ')
      .replace(/\\psi/g, 'ψ')
      .replace(/\\omega/g, 'ω');
  };

  // Markdownの基本的な変換
  const renderBasicMarkdown = (text: string): string => {
    return text
      // 見出し
      .replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold mb-2 text-gray-700">$1</h3>')
      .replace(/^## (.*$)/gim, '<h2 class="text-xl font-bold mb-3 text-gray-800">$1</h2>')
      .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mb-4 text-gray-900">$1</h1>')
      // 太字
      .replace(/\*\*(.*?)\*\*/g, '<strong class="font-bold text-gray-900">$1</strong>')
      // イタリック
      .replace(/\*(.*?)\*/g, '<em class="italic">$1</em>')
      // コード（インライン）
      .replace(/`([^`]+)`/g, '<code class="bg-gray-100 px-1 py-0.5 rounded text-sm font-mono">$1</code>')
      // 改行を<br>に変換
      .replace(/\n/g, '<br />');
  };

  const processedContent = renderBasicMarkdown(renderAdvancedMathSymbols(content));

  return (
    <>
      <style jsx>{`
        .math-symbol {
          font-family: 'Times New Roman', serif;
          font-weight: normal;
          color: #1f2937;
          font-size: 1.1em;
        }
        .math-vector {
          font-family: 'Times New Roman', serif;
          font-weight: bold;
          color: #1f2937;
          font-size: 1.05em;
        }
        .math-fraction {
          display: inline-block;
          vertical-align: middle;
          font-family: 'Times New Roman', serif;
          margin: 0 2px;
        }
        .fraction-line {
          font-size: 1.2em;
          color: #374151;
        }
        .math-fraction-block {
          display: inline-block;
          vertical-align: middle;
          text-align: center;
          font-family: 'Times New Roman', serif;
          margin: 0 4px;
        }
        .numerator {
          display: block;
          font-size: 0.9em;
          line-height: 1.2;
        }
        .fraction-line-block {
          display: block;
          border-top: 1px solid #374151;
          margin: 1px 0;
          font-size: 0.8em;
        }
        .denominator {
          display: block;
          font-size: 0.9em;
          line-height: 1.2;
        }
        .sqrt-content {
          border-top: 1px solid #374151;
          padding: 0 2px;
        }
        .math-block {
          display: block;
          text-align: center;
          margin: 12px 0;
          padding: 8px;
          background-color: #f9fafb;
          border: 1px solid #e5e7eb;
          border-radius: 4px;
          font-family: 'Times New Roman', serif;
          font-size: 1.1em;
        }
        .math-inline {
          font-family: 'Times New Roman', serif;
          font-size: 1.05em;
          color: #1f2937;
        }
        .matrix-container {
          display: inline-block;
          vertical-align: middle;
          margin: 0 4px;
          font-family: 'Times New Roman', serif;
        }
        .matrix-bracket-left, .matrix-bracket-right {
          font-size: 2em;
          vertical-align: middle;
          line-height: 1;
        }
        .matrix {
          display: inline-table;
          vertical-align: middle;
          border-collapse: collapse;
          margin: 0 2px;
        }
        .matrix td {
          text-align: center;
          padding: 4px 8px;
          font-size: 0.95em;
          border: none;
        }
        sup {
          font-size: 0.75em;
          vertical-align: super;
          line-height: 0;
        }
        sub {
          font-size: 0.75em;
          vertical-align: sub;
          line-height: 0;
        }
        .math-subscript {
          font-size: 0.8em;
          vertical-align: sub;
          line-height: 0;
        }
      `}</style>
      <div 
        className={`markdown-content leading-relaxed ${className}`}
        dangerouslySetInnerHTML={{ __html: processedContent }}
      />
    </>
  );
}
