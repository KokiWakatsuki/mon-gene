# トラブルシューティングガイド

## 問題生成が失敗する場合

### 症状
- adminのチャット機能は正常に動作する
- しかし、問題生成機能が失敗する

### 原因と解決方法

#### 1. 環境変数が設定されていない

**確認方法:**
```bash
# Dockerコンテナ内で環境変数を確認
docker-compose exec back env | grep API_KEY
```

**解決方法:**
1. `back/.env`ファイルが存在することを確認
2. 以下の環境変数が設定されていることを確認:
   - `CLAUDE_API_KEY`
   - `OPENAI_API_KEY`
   - `GOOGLE_API_KEY`
   - `CORE_API_URL`

3. Docker Composeを再起動:
```bash
docker-compose down
docker-compose up -d
```

#### 2. ユーザーのAI設定が不完全

**確認方法:**
データベースでユーザー設定を確認:
```sql
SELECT school_code, preferred_api, preferred_model 
FROM users 
WHERE school_code = 'admin';
```

**解決方法:**
1. フロントエンドの「設定」ページにアクセス
2. 「AI設定」セクションで以下を設定:
   - API: `google`, `openai`, または `claude`
   - モデル: 選択したAPIに対応するモデル

#### 3. Core APIサービスが起動していない

**確認方法:**
```bash
# コンテナの状態を確認
docker-compose ps

# Coreサービスのログを確認
docker-compose logs core
```

**解決方法:**
```bash
# Coreサービスを再起動
docker-compose restart core

# または全体を再起動
docker-compose down
docker-compose up -d
```

#### 4. APIキーが無効または期限切れ

**確認方法:**
バックエンドのログを確認:
```bash
docker-compose logs back | grep "API error"
```

**解決方法:**
1. `back/.env`ファイルのAPIキーを更新
2. 各APIプロバイダーのダッシュボードでキーの有効性を確認:
   - OpenAI: https://platform.openai.com/api-keys
   - Google: https://makersuite.google.com/app/apikey
   - Anthropic (Claude): https://console.anthropic.com/

---

## PDF抽出エラーが発生する場合

### 症状
- 「問題概要を表示」ボタンをクリックするとエラーが発生
- "PDF抽出エラー" というメッセージが表示される

### 原因と解決方法

#### 1. Google API キーが設定されていない

**確認方法:**
```bash
docker-compose exec back env | grep GOOGLE_API_KEY
```

**解決方法:**
1. `back/.env`ファイルに`GOOGLE_API_KEY`を追加
2. Docker Composeを再起動:
```bash
docker-compose down
docker-compose up -d
```

#### 2. PDFファイルが大きすぎる

**制限:**
- 最大ファイルサイズ: 20MB

**解決方法:**
- より小さいPDFファイルを使用
- PDFを分割して複数のファイルとしてアップロード

#### 3. Google APIの利用制限に達した

**確認方法:**
バックエンドのログを確認:
```bash
docker-compose logs back | grep "rate limit\|quota"
```

**解決方法:**
- しばらく待ってから再試行
- Google Cloud Consoleで割り当てを確認・増加

#### 4. PDFファイルの形式が対応していない

**対応形式:**
- 標準的なPDF形式（PDF 1.4以降）
- テキストが埋め込まれているPDF（画像のみのPDFは非対応）

**解決方法:**
- PDFにテキストが含まれていることを確認
- 必要に応じてOCR処理を行ってからアップロード

---

## デバッグ手順

### 1. ログの確認

**バックエンドログ:**
```bash
# リアルタイムでログを表示
docker-compose logs -f back

# 最新100行を表示
docker-compose logs --tail=100 back
```

**Coreサービスログ:**
```bash
docker-compose logs -f core
```

**フロントエンドログ:**
```bash
docker-compose logs -f front
```

### 2. コンテナの状態確認

```bash
# すべてのコンテナの状態を確認
docker-compose ps

# 特定のコンテナの詳細情報
docker inspect mon-gene-back
```

### 3. ネットワーク接続の確認

```bash
# バックエンドからCoreへの接続テスト
docker-compose exec back curl http://core:1234/health

# バックエンドからデータベースへの接続テスト
docker-compose exec back nc -zv db 3306
```

### 4. データベースの確認

```bash
# データベースに接続
docker-compose exec db mysql -u user -ppassword develop

# ユーザー情報を確認
SELECT * FROM users WHERE school_code = 'admin';

# 問題生成履歴を確認
SELECT id, user_id, subject, created_at FROM problems ORDER BY created_at DESC LIMIT 10;
```

---

## よくある質問

### Q: 「AI設定が不完全です」というエラーが出る

**A:** 設定ページでAPIとモデルを選択してください。
1. ログイン後、右上のアカウントメニューから「設定」を選択
2. 「AI設定」セクションでAPIとモデルを選択
3. 「保存」ボタンをクリック

### Q: 問題生成回数の上限に達した

**A:** データベースで制限を確認・変更できます:
```sql
-- 現在の制限を確認
SELECT school_code, problem_generation_count, problem_generation_limit 
FROM users 
WHERE school_code = 'admin';

-- 制限を増やす（例: 100回に設定）
UPDATE users 
SET problem_generation_limit = 100 
WHERE school_code = 'admin';

-- 無制限にする
UPDATE users 
SET problem_generation_limit = -1 
WHERE school_code = 'admin';
```

### Q: Core APIに接続できない

**A:** 以下を確認してください:
1. Coreコンテナが起動しているか: `docker-compose ps`
2. ポート1234が使用可能か: `lsof -i :1234`
3. 環境変数`CORE_API_URL`が正しく設定されているか

---

## 緊急時の対応

### すべてをリセットする

```bash
# すべてのコンテナを停止・削除
docker-compose down

# ボリュームも削除（データベースもリセット）
docker-compose down -v

# イメージを再ビルド
docker-compose build --no-cache

# 起動
docker-compose up -d

# ログを確認
docker-compose logs -f
```

### データベースのみリセット

```bash
# データベースコンテナを停止
docker-compose stop db

# データベースボリュームを削除
docker volume rm mon-gene_db-data

# マイグレーションを再実行
docker-compose up -d db
sleep 10
docker-compose exec back sh /app/scripts/migrate.sh
```

---

## サポート

問題が解決しない場合は、以下の情報を含めて報告してください:

1. エラーメッセージの全文
2. バックエンドログ（`docker-compose logs back`）
3. 実行した操作の手順
4. 使用している環境（OS、Dockerバージョンなど）

```bash
# システム情報を収集
echo "=== Docker Version ===" > debug_info.txt
docker --version >> debug_info.txt
echo "\n=== Docker Compose Version ===" >> debug_info.txt
docker-compose --version >> debug_info.txt
echo "\n=== Container Status ===" >> debug_info.txt
docker-compose ps >> debug_info.txt
echo "\n=== Backend Logs (last 100 lines) ===" >> debug_info.txt
docker-compose logs --tail=100 back >> debug_info.txt
echo "\n=== Core Logs (last 100 lines) ===" >> debug_info.txt
docker-compose logs --tail=100 core >> debug_info.txt
```
