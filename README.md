# Mongene - AI問題生成システム

生成AIを利用して、新たな問題を生成するシステムです。問題の特徴量を指定して、満足するまで再生成を繰り返し、最終的に印刷することができます。

## アーキテクチャ

### マイクロサービス構成
- **front**: ユーザーインターフェース（Next.js + TypeScript + Tailwind CSS）
- **back**: 認証・API統合・ビジネスロジック（Go）
- **core**: AI処理・図形生成・PDF生成（Python + FastAPI）
- **db**: データストレージ（MySQL）

### 新しいディレクトリ構造
```
front/app/
├── (auth)/login/          # 認証関連ページ
├── (dashboard)/problems/  # メインアプリケーション
├── components/
│   ├── layout/           # レイアウトコンポーネント
│   ├── ui/              # 再利用可能UIコンポーネント
│   ├── features/        # 機能別コンポーネント
│   └── forms/           # フォームコンポーネント
├── lib/config/          # 設定ファイル
└── hooks/               # カスタムフック

back/
├── cmd/server/          # エントリーポイント
├── internal/
│   ├── models/          # データモデル
│   ├── services/        # ビジネスロジック
│   ├── api/handlers/    # HTTPハンドラー
│   ├── api/middleware/  # ミドルウェア
│   ├── api/routes/      # ルーティング
│   ├── clients/         # 外部API クライアント
│   └── repositories/    # データアクセス層

core/app/
├── api/endpoints/       # API エンドポイント
├── services/           # ビジネスロジック
├── models/             # Pydanticモデル
├── core/geometry/      # 図形生成エンジン
└── utils/              # ユーティリティ
```

## 機能

### 実装済み機能
- ✅ 認証システム（ログイン、セッション管理）
- ✅ 問題生成（Claude API統合）
- ✅ 図形描画（matplotlib）
- ✅ PDF生成
- ✅ メール送信機能
- ✅ RESTful API
- ✅ CORS対応

### 主要ページ
- **ログインページ**: 塾単位での認証
- **問題生成ページ**: 特徴量入力による問題生成
  - 科目選択（国数英理社）
  - 特徴量設定（難易度、単元、計算量など）
- **問題確認ページ**: 生成された問題の確認・再生成・印刷

## 開発環境のセットアップ

### 前提条件
- Docker & Docker Compose
- Node.js 18+
- Go 1.21+
- Python 3.11+

### 環境変数の設定
```bash
# back/.env
CLAUDE_API_KEY=your-claude-api-key
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_FROM=your-email@gmail.com
SMTP_PASSWORD=your-app-password
CORE_API_URL=http://core:1234

# front/.env.local
NEXT_PUBLIC_USE_REAL_API=true
NEXT_PUBLIC_BACKEND_API_URL=http://localhost:8080/api/generate-problem
```

### 起動方法

#### 1. Docker Composeを使用（推奨）
```bash
# 全サービスを起動
docker-compose up --build

# バックグラウンドで起動
docker-compose up --build -d

# 停止
docker-compose down
```

#### 2. 統合テストの実行
```bash
# 統合テストスクリプトを実行
./scripts/test-integration.sh
```

#### 3. 個別起動（開発時）
```bash
# フロントエンド
cd front
npm install
npm run dev

# バックエンド
cd back
go mod download
go run cmd/server/main.go

# コア
cd core
pip install -r requirements.txt
python app/main.py
```

## アクセス先

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Core API**: http://localhost:1234
- **Core API Docs**: http://localhost:1234/docs (FastAPI自動生成)

## API エンドポイント

### Backend API (Port 8080)
- `POST /api/login` - ログイン
- `POST /api/forgot-password` - パスワードリセット
- `POST /api/generate-problem` - 問題生成
- `POST /api/generate-pdf` - PDF生成
- `GET /health` - ヘルスチェック

### Core API (Port 1234)
- `POST /draw-geometry` - 図形描画
- `POST /draw-custom-geometry` - カスタム図形描画
- `POST /analyze-problem` - 問題解析
- `POST /generate-pdf` - PDF生成
- `GET /health` - ヘルスチェック

## 技術スタック

### Frontend
- **Next.js 14** (App Router)
- **TypeScript**
- **Tailwind CSS**
- **React Hooks**

### Backend
- **Go 1.21**
- **レイヤードアーキテクチャ**
- **依存性注入**
- **CORS対応**

### Core
- **Python 3.11**
- **FastAPI**
- **Pydantic**
- **matplotlib**
- **ReportLab**

### Infrastructure
- **Docker & Docker Compose**
- **MySQL 8.0**

## 開発ガイドライン

詳細な実装ガイドラインは以下を参照してください：
- [IMPROVED_DIRECTORY_STRUCTURE.md](./IMPROVED_DIRECTORY_STRUCTURE.md)
- [IMPLEMENTATION_GUIDELINES.md](./IMPLEMENTATION_GUIDELINES.md)

## ライセンス

このプロジェクトは私的利用のためのものです。
