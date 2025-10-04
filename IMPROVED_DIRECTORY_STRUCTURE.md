# 改善されたディレクトリ構成提案

## frontコンテナ（Next.js）

```
front/
├── app/
│   ├── (auth)/                    # 認証関連のルートグループ
│   │   ├── login/
│   │   │   └── page.tsx
│   │   └── register/
│   │       └── page.tsx
│   ├── (dashboard)/               # メインアプリケーション
│   │   ├── problems/              # 問題関連ページ
│   │   │   ├── page.tsx          # 問題生成ページ
│   │   │   ├── [id]/
│   │   │   │   └── page.tsx      # 問題詳細ページ
│   │   │   └── history/
│   │   │       └── page.tsx      # 問題履歴ページ
│   │   └── settings/
│   │       └── page.tsx          # 設定ページ
│   ├── api/                       # API Routes（Next.js）
│   │   └── auth/
│   │       └── route.ts          # 認証関連のクライアントサイド処理
│   ├── components/
│   │   ├── ui/                    # 再利用可能なUIコンポーネント
│   │   │   ├── Button.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Input.tsx
│   │   │   └── LoadingSpinner.tsx
│   │   ├── forms/                 # フォーム関連コンポーネント
│   │   │   ├── LoginForm.tsx
│   │   │   ├── ProblemGenerationForm.tsx
│   │   │   └── FilterForm.tsx
│   │   ├── layout/                # レイアウト関連コンポーネント
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── BackgroundShapes.tsx
│   │   └── features/              # 機能別コンポーネント
│   │       ├── problems/
│   │       │   ├── ProblemCard.tsx
│   │       │   ├── ProblemPreviewModal.tsx
│   │       │   ├── ProblemList.tsx
│   │       │   └── Tabs.tsx
│   │       └── filters/
│   │           └── Filters.tsx
│   ├── hooks/                     # カスタムフック
│   │   ├── useAuth.ts
│   │   ├── useProblems.ts
│   │   └── useLocalStorage.ts
│   ├── lib/                       # ユーティリティ・設定
│   │   ├── api/                   # API クライアント
│   │   │   ├── client.ts         # 基本APIクライアント
│   │   │   ├── auth.ts           # 認証API
│   │   │   ├── problems.ts       # 問題生成API
│   │   │   └── types.ts          # API型定義
│   │   ├── utils/
│   │   │   ├── validation.ts     # バリデーション
│   │   │   ├── formatting.ts     # フォーマット処理
│   │   │   └── constants.ts      # 定数
│   │   └── config/
│   │       ├── api.ts            # API設定
│   │       └── app.ts            # アプリケーション設定
│   ├── types/                     # TypeScript型定義
│   │   ├── auth.ts
│   │   ├── problems.ts
│   │   └── api.ts
│   ├── styles/                    # スタイル関連
│   │   ├── globals.css
│   │   └── components.css
│   ├── layout.tsx
│   └── page.tsx
├── public/
├── package.json
└── next.config.ts
```

## backコンテナ（Go）

```
back/
├── cmd/
│   └── server/
│       └── main.go               # エントリーポイント
├── internal/
│   ├── api/                      # API層
│   │   ├── handlers/             # HTTPハンドラー
│   │   │   ├── auth.go          # 認証関連ハンドラー
│   │   │   ├── problems.go      # 問題生成関連ハンドラー
│   │   │   └── health.go        # ヘルスチェック
│   │   ├── middleware/           # ミドルウェア
│   │   │   ├── cors.go
│   │   │   ├── auth.go
│   │   │   └── logging.go
│   │   └── routes/               # ルーティング
│   │       └── router.go
│   ├── services/                 # ビジネスロジック層
│   │   ├── auth_service.go       # 認証サービス
│   │   ├── problem_service.go    # 問題生成サービス
│   │   ├── integration_service.go # 外部API統合サービス
│   │   └── email_service.go      # メール送信サービス
│   ├── repositories/             # データアクセス層
│   │   ├── user_repository.go
│   │   ├── problem_repository.go
│   │   └── session_repository.go
│   ├── models/                   # データモデル
│   │   ├── user.go
│   │   ├── problem.go
│   │   ├── session.go
│   │   └── api_requests.go
│   ├── clients/                  # 外部API クライアント
│   │   ├── claude_client.go      # Claude API クライアント
│   │   ├── core_client.go        # Core API クライアント
│   │   └── interfaces.go         # インターフェース定義
│   ├── config/                   # 設定管理
│   │   ├── config.go
│   │   └── database.go
│   └── utils/                    # ユーティリティ
│       ├── validation.go
│       ├── crypto.go
│       └── response.go
├── migrations/                   # データベースマイグレーション
│   ├── 001_create_users.sql
│   ├── 002_create_problems.sql
│   └── 003_create_sessions.sql
├── scripts/                      # スクリプト
│   ├── build.sh
│   └── migrate.sh
├── .env.example
├── go.mod
├── go.sum
└── Dockerfile
```

## coreコンテナ（Python）

```
core/
├── app/
│   ├── main.py                   # FastAPI エントリーポイント
│   ├── api/                      # API層
│   │   ├── __init__.py
│   │   ├── endpoints/            # エンドポイント
│   │   │   ├── __init__.py
│   │   │   ├── geometry.py       # 図形関連API
│   │   │   ├── analysis.py       # 問題解析API
│   │   │   ├── pdf.py           # PDF生成API
│   │   │   └── health.py        # ヘルスチェック
│   │   └── dependencies.py       # 依存性注入
│   ├── services/                 # ビジネスロジック層
│   │   ├── __init__.py
│   │   ├── geometry_service.py   # 図形生成サービス
│   │   ├── analysis_service.py   # 問題解析サービス
│   │   ├── pdf_service.py        # PDF生成サービス
│   │   └── ai_service.py         # AI関連サービス（将来用）
│   ├── models/                   # データモデル・スキーマ
│   │   ├── __init__.py
│   │   ├── geometry.py           # 図形関連モデル
│   │   ├── analysis.py           # 解析関連モデル
│   │   └── pdf.py               # PDF関連モデル
│   ├── core/                     # コア機能
│   │   ├── __init__.py
│   │   ├── geometry/             # 図形描画エンジン
│   │   │   ├── __init__.py
│   │   │   ├── shapes/           # 図形別描画クラス
│   │   │   │   ├── __init__.py
│   │   │   │   ├── basic_shapes.py    # 基本図形
│   │   │   │   ├── solid_shapes.py    # 立体図形
│   │   │   │   └── custom_shapes.py   # カスタム図形
│   │   │   ├── renderer.py       # 描画エンジン
│   │   │   └── analyzer.py       # 図形解析
│   │   ├── pdf/                  # PDF生成エンジン
│   │   │   ├── __init__.py
│   │   │   ├── generator.py      # PDF生成
│   │   │   └── templates.py      # テンプレート
│   │   └── ai/                   # AI処理（将来用）
│   │       ├── __init__.py
│   │       └── processors.py
│   ├── utils/                    # ユーティリティ
│   │   ├── __init__.py
│   │   ├── image_utils.py        # 画像処理
│   │   ├── math_utils.py         # 数学計算
│   │   └── validation.py         # バリデーション
│   └── config/                   # 設定
│       ├── __init__.py
│       └── settings.py
├── tests/                        # テスト
│   ├── __init__.py
│   ├── test_geometry.py
│   ├── test_analysis.py
│   └── test_pdf.py
├── requirements.txt
├── requirements-dev.txt
└── Dockerfile
```

## 共通設定・ドキュメント

```
project_root/
├── docker/                       # Docker設定
│   ├── Dockerfile.front
│   ├── Dockerfile.back
│   ├── Dockerfile.core
│   └── Dockerfile.db
├── docs/                         # ドキュメント
│   ├── api/                      # API仕様書
│   │   ├── back_api.md
│   │   └── core_api.md
│   ├── architecture.md           # アーキテクチャ設計
│   ├── deployment.md             # デプロイメント手順
│   └── development.md            # 開発手順
├── scripts/                      # 共通スクリプト
│   ├── setup.sh                  # 環境セットアップ
│   ├── deploy.sh                 # デプロイスクリプト
│   └── test.sh                   # テストスクリプト
├── docker-compose.yml
├── docker-compose.prod.yml
├── .gitignore
├── README.md
└── Makefile
```

## 主な改善点

### 1. 責任の明確化
- **front**: UI/UX、状態管理、APIクライアント
- **back**: 認証、API統合、ビジネスロジック
- **core**: AI処理、図形生成、PDF生成

### 2. レイヤードアーキテクチャの導入
- API層、サービス層、リポジトリ層の分離
- 依存性の逆転による疎結合化

### 3. 機能別ディレクトリ構成
- 関連する機能をまとめて配置
- 再利用性とメンテナンス性の向上

### 4. 設定管理の統一
- 環境別設定の分離
- セキュリティ設定の適切な管理

### 5. テスト構造の整備
- 各層でのテスト可能性の向上
- 統合テストとユニットテストの分離
