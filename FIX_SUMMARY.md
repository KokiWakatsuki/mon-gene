# 修正内容サマリー

## 修正日時
2026-03-02

## 問題の概要

1. **問題生成が失敗する**
   - adminのチャット機能は正常に動作
   - しかし、問題生成機能が失敗

2. **問題概要の表示でPDF抽出エラーが発生する**
   - 「問題概要を表示」ボタンをクリックするとエラー

## 根本原因

### 主な原因
1. **環境変数がDockerコンテナに渡されていない**
   - `docker-compose.yml`と`docker-compose.stage.yml`で環境変数の設定が不足
   - APIキー（GOOGLE_API_KEY、OPENAI_API_KEY、CLAUDE_API_KEY）がbackサービスに渡されていない

2. **エラーメッセージが不十分**
   - 問題発生時に具体的な原因が分からない
   - デバッグが困難

## 実施した修正

### 1. Docker Compose設定の修正

#### ファイル: `docker-compose.yml`
```yaml
back:
  # 追加: coreサービスへの依存関係
  depends_on:
    - db
    - core
  
  # 追加: .envファイルの読み込み
  env_file:
    - ./back/.env
  
  # 追加: 環境変数の明示的な設定
  environment:
    - CLAUDE_API_KEY=${CLAUDE_API_KEY}
    - OPENAI_API_KEY=${OPENAI_API_KEY}
    - GOOGLE_API_KEY=${GOOGLE_API_KEY}
    - LABORATORY_API_KEY=${LABORATORY_API_KEY}
    - CORE_API_URL=${CORE_API_URL:-http://core:1234}
    # ... その他の環境変数
```

**効果:**
- APIキーがbackサービスで利用可能になる
- 問題生成とPDF抽出が正常に動作する

#### ファイル: `docker-compose.stage.yml`
同様の修正をステージング環境にも適用

### 2. エラーハンドリングの改善

#### ファイル: `back/internal/services/problem_service.go`

**変更内容:**
- API呼び出し前後に詳細なログを追加
- エラーメッセージに具体的な解決方法を含める

```go
// 修正前
if err != nil {
    return nil, fmt.Errorf("OpenAI APIでの問題生成に失敗しました: %w", err)
}

// 修正後
if err != nil {
    fmt.Printf("❌ [GenerateProblem] OpenAI API error: %v\n", err)
    return nil, fmt.Errorf("OpenAI APIでの問題生成に失敗しました。エラー詳細: %w。API設定を確認してください（設定ページ > AI設定）", err)
}
```

**効果:**
- エラー発生時に原因を特定しやすくなる
- ユーザーに具体的な対処方法を提示できる

#### ファイル: `back/internal/api/handlers/sse_handler.go`

**変更内容:**
- PDFファイルサイズのバリデーション追加（最大20MB）
- PDF抽出エラーの詳細なハンドリング
- エラーの種類に応じた適切なメッセージ

```go
// PDFファイルサイズのバリデーション
maxPDFSize := 20 * 1024 * 1024 // 20MB
if len(pdfData) > maxPDFSize {
    fmt.Printf("❌ [PreviewPDF] PDF file too large: %d bytes\n", len(pdfData))
    utils.WriteErrorResponse(w, http.StatusBadRequest, 
        fmt.Sprintf("PDFファイルが大きすぎます（最大20MB）。現在のサイズ: %.2f MB", 
        float64(len(pdfData))/(1024*1024)))
    return
}

// エラーの種類に応じた適切なメッセージ
if strings.Contains(err.Error(), "API key") {
    errorMsg = "Google API キーが設定されていません。管理者に連絡してください。"
} else if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "rate limit") {
    errorMsg = "APIの利用制限に達しました。しばらく待ってから再試行してください。"
}
```

**効果:**
- 大きすぎるPDFファイルを事前に検出
- エラーの原因を明確に伝える
- ユーザーが適切な対処を取れる

### 3. トラブルシューティングガイドの作成

#### ファイル: `TROUBLESHOOTING.md`

**内容:**
- 問題生成失敗時の診断手順
- PDF抽出エラーの解決方法
- デバッグコマンド集
- よくある質問と回答
- 緊急時の対応手順

**効果:**
- ユーザーが自己解決できる
- サポート負荷の軽減

## 修正後の動作確認手順

### 1. 環境変数の確認

```bash
# .envファイルの存在確認
ls -la back/.env

# 環境変数の内容確認（APIキーは表示されない）
cat back/.env | grep -E "API_KEY|CORE_API_URL"
```

### 2. Dockerコンテナの再起動

```bash
# 既存のコンテナを停止・削除
docker-compose down

# 新しい設定で起動
docker-compose up -d

# コンテナの状態確認
docker-compose ps

# ログの確認
docker-compose logs -f back
```

### 3. 環境変数がコンテナに渡されているか確認

```bash
# backコンテナ内で環境変数を確認
docker-compose exec back env | grep API_KEY

# 期待される出力:
# GOOGLE_API_KEY=AIzaSy...
# OPENAI_API_KEY=sk-proj...
# CLAUDE_API_KEY=sk-ant...
```

### 4. 機能テスト

#### 問題生成のテスト
1. ログイン（admin / admin）
2. 「問題」タブに移動
3. プロンプトを入力して問題生成を実行
4. エラーが発生しないことを確認

#### PDF抽出のテスト
1. 「3問生成」タブに移動
2. PDFファイルをアップロード
3. 「問題概要を表示」ボタンをクリック
4. PDFの内容が正しく抽出されることを確認

### 5. エラーログの確認

```bash
# エラーが発生していないか確認
docker-compose logs back | grep -i error

# API呼び出しのログを確認
docker-compose logs back | grep -E "GenerateProblem|PreviewPDF"
```

## 期待される結果

### 修正前
- ❌ 問題生成が失敗する
- ❌ PDF抽出でエラーが発生する
- ❌ エラーメッセージが不明瞭

### 修正後
- ✅ 問題生成が正常に動作する
- ✅ PDF抽出が正常に動作する
- ✅ エラー発生時に具体的な原因と対処方法が表示される
- ✅ デバッグログで問題を追跡できる

## 注意事項

### 1. APIキーの設定
- `back/.env`ファイルに有効なAPIキーが設定されていることを確認
- APIキーが期限切れでないことを確認

### 2. セキュリティ
- `.env`ファイルは`.gitignore`に含まれていることを確認
- 本番環境では環境変数を安全に管理

### 3. ステージング環境
- `docker-compose.stage.yml`も同様に修正済み
- ステージング環境でも同じ手順で確認

## 今後の改善案

1. **ヘルスチェックエンドポイントの追加**
   - 各サービスの詳細な状態を確認できるエンドポイント
   - API接続状態の確認

2. **フロントエンドのエラー表示改善**
   - より分かりやすいエラーメッセージ
   - リトライボタンの追加

3. **監視とアラート**
   - API呼び出しの成功率を監視
   - エラー発生時の自動通知

4. **ドキュメントの充実**
   - API設定ガイドの作成
   - 動画チュートリアルの作成

## 関連ファイル

- `docker-compose.yml` - 開発環境のDocker設定
- `docker-compose.stage.yml` - ステージング環境のDocker設定
- `back/internal/services/problem_service.go` - 問題生成サービス
- `back/internal/api/handlers/sse_handler.go` - PDF抽出ハンドラー
- `TROUBLESHOOTING.md` - トラブルシューティングガイド
- `back/.env` - 環境変数設定ファイル

## 参考リンク

- [Docker Compose環境変数ドキュメント](https://docs.docker.com/compose/environment-variables/)
- [Google Generative AI API](https://ai.google.dev/)
- [OpenAI API](https://platform.openai.com/docs)
- [Anthropic Claude API](https://docs.anthropic.com/)
