# 図形描画・PDF生成システム

このシステムは、数学問題に図形が含まれている場合に自動的に図形を描画し、問題文と図形を組み合わせたPDFを生成する機能を提供します。

## 🎯 機能概要

1. **問題分析**: 問題文から図形の必要性を自動判定
2. **図形描画**: 角柱、円柱、三角形、円などの図形を自動生成
3. **PDF生成**: 問題文と図形を組み合わせたPDFを作成
4. **システム統合**: 既存の問題生成システムに統合

## 📋 使用ライブラリ

### Core API (Python)
- **matplotlib**: 図形描画ライブラリ
- **numpy**: 数値計算ライブラリ
- **reportlab**: PDF生成ライブラリ
- **Pillow**: 画像処理ライブラリ
- **fastapi**: Web APIフレームワーク
- **uvicorn**: ASGIサーバー

### Backend API (Go)
- **net/http**: HTTP サーバー
- **encoding/json**: JSON処理
- **github.com/joho/godotenv**: 環境変数管理

## 🚀 セットアップ手順

### 1. 依存関係のインストール

```bash
# Core API の依存関係をインストール
cd core
pip install -r requirements.txt
```

### 2. サーバーの起動

```bash
# Core API サーバーを起動 (ターミナル1)
cd core
python main.py

# Backend API サーバーを起動 (ターミナル2)
cd back
go run main.go
```

### 3. 環境変数の設定

```bash
# back/.env ファイルに以下を追加
CORE_API_URL=http://localhost:1234
CLAUDE_API_KEY=your_claude_api_key_here
```

## 🧪 テスト実行

```bash
# テストスクリプトを実行
python test_geometry_system.py
```

## 📡 API エンドポイント

### Core API (Port 1234)

#### 1. 問題分析
```http
POST /analyze-problem
Content-Type: application/json

{
  "problem_text": "円柱の体積を求めなさい。底面の半径が3cm、高さが8cmの円柱があります。",
  "unit_parameters": {"geometry": true, "shapes": ["cylinder"]},
  "subject": "math"
}
```

#### 2. 図形描画
```http
POST /draw-geometry
Content-Type: application/json

{
  "shape_type": "cylinder",
  "parameters": {
    "radius": 3,
    "height": 8
  }
}
```

#### 3. PDF生成
```http
POST /generate-pdf
Content-Type: application/json

{
  "problem_text": "円柱の体積を求めなさい。",
  "image_base64": "base64_encoded_image_data"
}
```

### Backend API (Port 8080)

#### 1. 統合問題生成
```http
POST /api/generate-problem
Content-Type: application/json

{
  "prompt": "円柱の体積を求める問題を作成してください。",
  "subject": "math",
  "filters": {
    "geometry": true,
    "shapes": ["cylinder"],
    "difficulty": 3
  }
}
```

#### 2. 統合PDF生成
```http
POST /api/generate-pdf
Content-Type: application/json

{
  "problem_text": "円柱の体積を求めなさい。",
  "image_base64": "base64_encoded_image_data"
}
```

## 🎨 対応図形

### 1. 円柱 (Cylinder)
```json
{
  "shape_type": "cylinder",
  "parameters": {
    "radius": 2,
    "height": 4
  }
}
```

### 2. 角柱 (Prism)
```json
{
  "shape_type": "prism",
  "parameters": {
    "base_area": 10,
    "height": 5,
    "prism_type": "rectangular"
  }
}
```

### 3. 三角形 (Triangle)
```json
{
  "shape_type": "triangle",
  "parameters": {
    "vertices": [[0, 0], [3, 0], [1.5, 3]]
  }
}
```

### 4. 円 (Circle)
```json
{
  "shape_type": "circle",
  "parameters": {
    "radius": 2,
    "center": [0, 0]
  }
}
```

## 🔄 システムフロー

1. **問題生成リクエスト** → Backend API
2. **Claude API呼び出し** → 問題文生成
3. **問題分析** → Core API (図形の必要性判定)
4. **図形描画** → Core API (必要に応じて)
5. **レスポンス** → 問題文 + 図形画像 (Base64)
6. **PDF生成** → Core API (印刷時)

## 📁 ファイル構成

```
mon-gene/
├── core/
│   ├── main.py              # Core API サーバー
│   └── requirements.txt     # Python依存関係
├── back/
│   └── main.go             # Backend API サーバー
├── test_geometry_system.py # テストスクリプト
└── GEOMETRY_SYSTEM_README.md # このファイル
```

## 🐛 トラブルシューティング

### Core APIが起動しない場合
```bash
# 依存関係を再インストール
cd core
pip install --upgrade -r requirements.txt

# ポート8000が使用中の場合
lsof -ti:8000 | xargs kill -9
```

### Backend APIが起動しない場合
```bash
# Go モジュールを更新
cd back
go mod tidy

# ポート8080が使用中の場合
lsof -ti:8080 | xargs kill -9
```

### 図形が生成されない場合
- matplotlib のバックエンド設定を確認
- システムにフォントがインストールされているか確認

### PDF生成でフォントエラーが発生する場合
- macOS: システムフォントが自動検出されます
- Linux: 日本語フォントをインストールしてください
- Windows: フォントパスを調整してください

## 🔧 カスタマイズ

### 新しい図形の追加
1. `core/main.py` に描画関数を追加
2. `draw_geometry` エンドポイントに条件分岐を追加
3. `analyze_problem` の検出ロジックを更新

### PDF スタイルの変更
1. `core/main.py` の `generate_pdf` 関数を編集
2. ReportLabのスタイル設定を調整

## 📞 サポート

問題が発生した場合は、以下の情報を含めてお問い合わせください：
- エラーメッセージ
- 実行環境 (OS, Python/Goバージョン)
- テストスクリプトの実行結果

## 🎉 完了

システムが正常に動作している場合、以下のような流れで図形付き問題が生成されます：

1. 「円柱の体積を求める問題を作成して」とリクエスト
2. Claude APIが問題文を生成
3. システムが「円柱」キーワードを検出
4. 円柱の図形を自動描画
5. 問題文と図形を含むレスポンスを返却
6. 必要に応じてPDFを生成

これで図形を含む数学問題の自動生成システムが完成しました！
