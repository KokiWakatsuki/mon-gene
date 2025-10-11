# データベースセットアップガイド

## 概要

このドキュメントでは、問題保存機能のためのデータベースセットアップ方法を説明します。

## 実装した機能

1. **問題のDB保存**: 生成した問題を塾コード（user）に紐づけてMySQLに保存
   - 問題文（content）
   - 解答（solution）
   - 図（image_base64）
   - 生成パラメータ（prompt, filters）

2. **パラメータ検索**: 同じパラメータで過去に生成した問題があれば再利用
   - 新しい問題を生成せずに済む
   - API使用量の削減

3. **フリーワード検索**: 問題文、解答、プロンプト、科目から検索可能

## データベーススキーマ

```sql
CREATE TABLE problems (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    subject VARCHAR(100) NOT NULL COMMENT '科目（数学、物理など）',
    prompt TEXT NOT NULL COMMENT '生成時のプロンプト',
    content TEXT NOT NULL COMMENT '問題文',
    solution TEXT COMMENT '解答',
    image_base64 LONGTEXT COMMENT '図（Base64エンコード）',
    filters JSON COMMENT '生成パラメータ（フィルタ条件）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_subject (subject),
    INDEX idx_created_at (created_at),
    FULLTEXT INDEX idx_fulltext_search (content, solution, prompt, subject)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

## セットアップ手順

### 1. 環境変数の設定

`back/.env` ファイルにデータベース接続情報が設定されています：

```bash
DB_HOST=db
DB_PORT=3306
DB_USER=user
DB_PASSWORD=password
DB_NAME=develop
```

### 2. Dockerでデータベースを起動

```bash
# プロジェクトルートで実行
docker-compose up -d db
```

### 3. マイグレーションの実行

```bash
# back/scriptsディレクトリで実行
cd back/scripts
./migrate.sh
```

または、Dockerコンテナ内から実行：

```bash
docker-compose exec back /bin/sh
cd scripts
./migrate.sh
```

### 4. アプリケーションの起動

```bash
# プロジェクトルートで実行
docker-compose up
```

## API エンドポイント

### 1. 問題生成（自動保存）

```bash
POST /api/generate-problem
Authorization: Bearer <token>

{
  "prompt": "二次方程式の問題を作成してください",
  "subject": "数学",
  "filters": {
    "difficulty": "medium",
    "grade": 3
  }
}
```

**機能**:
- 同じパラメータで過去に生成された問題があれば、それを返す（API呼び出しなし）
- 新しい問題の場合は生成してDBに自動保存

### 2. 問題履歴の取得

```bash
GET /api/problems/history
Authorization: Bearer <token>
```

**レスポンス例**:
```json
{
  "success": true,
  "problems": [
    {
      "id": 1,
      "user_id": 1,
      "subject": "数学",
      "prompt": "二次方程式の問題を作成してください",
      "content": "【問題】次の二次方程式を解け...",
      "solution": "【解答】x = 2, -3",
      "image_base64": "data:image/png;base64,...",
      "filters": {
        "difficulty": "medium",
        "grade": 3
      },
      "created_at": "2025-10-11T01:00:00Z",
      "updated_at": "2025-10-11T01:00:00Z"
    }
  ],
  "count": 1
}
```

### 3. キーワード検索

```bash
GET /api/problems/search?keyword=二次方程式
Authorization: Bearer <token>
```

**機能**:
- 問題文、解答、プロンプト、科目から部分一致で検索
- 全文検索インデックスを使用した高速検索

## アーキテクチャ

### リポジトリ層

- `MySQLProblemRepository`: MySQL用の問題リポジトリ
  - `Create`: 問題の保存
  - `GetByID`: IDで問題を取得
  - `GetByUserID`: ユーザーの問題一覧を取得
  - `SearchByParameters`: パラメータで完全一致検索
  - `SearchByKeyword`: キーワードで部分一致検索
  - `Delete`: 問題の削除

### サービス層

- `ProblemService`
  - `GenerateProblem`: 問題生成（パラメータ検索 → キャッシュヒット時は再利用）
  - `SearchProblemsByParameters`: パラメータ検索
  - `SearchProblemsByKeyword`: キーワード検索
  - `GetUserProblems`: 問題履歴取得

### ハンドラー層

- `ProblemHandler`
  - `GenerateProblem`: 問題生成API
  - `SearchProblems`: 検索API
  - `GetUserProblems`: 履歴取得API

## データフロー

### 問題生成時

```
1. クライアント → GenerateProblem API
2. パラメータで既存問題を検索
3. 見つかった場合: 既存問題を返す（終了）
4. 見つからない場合:
   - AI APIを呼び出して新しい問題を生成
   - DBに保存
   - クライアントに返す
```

### 検索時

```
1. クライアント → SearchProblems API
2. キーワードでDB検索（全文検索）
3. 結果をクライアントに返す
```

## トラブルシューティング

### データベースに接続できない

```bash
# データベースコンテナのログを確認
docker-compose logs db

# データベースコンテナに直接接続
docker-compose exec db mysql -u user -p
# パスワード: password
```

### マイグレーションが失敗する

```bash
# MySQLクライアントで直接実行
docker-compose exec db mysql -u user -p develop < back/migrations/001_create_problems_table.sql
```

### 問題が保存されない

- アプリケーションログを確認
- データベース接続が正常かチェック
- `back/.env` の設定を確認

## パフォーマンス最適化

1. **インデックス**: user_id, subject, created_atにインデックス設定済み
2. **全文検索**: MySQLの全文検索インデックスを使用
3. **接続プール**: 最大25接続、アイドル5接続で設定済み
4. **キャッシング**: 同じパラメータの問題は再利用

## セキュリティ

1. **認証**: すべてのAPIでBearerトークン認証が必須
2. **データ分離**: ユーザーごとにデータが分離
3. **SQLインジェクション対策**: プリペアドステートメントを使用
