# 3問生成システム 詳細設計書

## 概要

アップロードされた問題を参考に、3種類の対策問題を生成するシステム。
各問題は5段階のプロセスで生成され、合計15段階（3パターン × 5段階）のプロセスを実行する。

## システム構成

### 1. プロセスフロー

```
大ステップ会話（MainConversationHistory）
│
├─ パターンA生成（小ステップ会話: PatternA.ConversationHistory）
│  ├─ Stage 1: 骨組み設計
│  ├─ Stage 2: パラメータ設定と動的検証
│  ├─ Stage 3: 図形描画
│  ├─ Stage 4: 完全な問題文生成
│  └─ Stage 5: 完全な解答・解説生成
│
├─ パターンB生成（小ステップ会話: PatternB.ConversationHistory）
│  ├─ Stage 1: 骨組み設計
│  ├─ Stage 2: パラメータ設定と動的検証
│  ├─ Stage 3: 図形描画
│  ├─ Stage 4: 完全な問題文生成
│  └─ Stage 5: 完全な解答・解説生成
│
└─ パターンC生成（小ステップ会話: PatternC.ConversationHistory）
   ├─ Stage 1: 骨組み設計
   ├─ Stage 2: パラメータ設定と動的検証
   ├─ Stage 3: 図形描画
   ├─ Stage 4: 完全な問題文生成
   └─ Stage 5: 完全な解答・解説生成
```

### 2. データ構造

#### ThreeProblemGenerationRequest
```go
type ThreeProblemGenerationRequest struct {
    UploadedProblemContent string // アップロードされた問題の内容
    Subject                string // 科目（"数学"）
}
```

#### ThreeProblemGenerationResponse
```go
type ThreeProblemGenerationResponse struct {
    Success bool
    Error   string
    
    PatternA PatternResult // パターンA（数値だけ違う）
    PatternB PatternResult // パターンB（公式は同じだが問題は違う）
    PatternC PatternResult // パターンC（構成は同じだが問われる部分が違う）
    
    MainConversationHistory *ConversationHistory // 大ステップ用会話履歴
    Log string
}
```

#### PatternResult
```go
type PatternResult struct {
    Success bool
    Error   string
    
    // 5段階の結果
    Stage1Result string // 骨組み設計
    Stage2Result string // パラメータ設定と動的検証
    Stage3Result string // 図形描画コード
    Stage4Result string // 完全な問題文
    Stage5Result string // 完全な解答・解説
    
    // 最終的な問題データ
    Content     string // 問題文（Stage4の結果）
    Solution    string // 解答・解説（Stage5の結果）
    ImageBase64 string // 図形画像（Stage3から生成）
    
    ConversationHistory *ConversationHistory // 小ステップ用会話履歴
    
    // 各段階のログ
    Stage1Log string
    Stage2Log string
    Stage3Log string
    Stage4Log string
    Stage5Log string
}
```

### 3. サービス層メソッド

#### GenerateThreeProblemsWithProgress
```go
func (s *problemService) GenerateThreeProblemsWithProgress(
    ctx context.Context,
    req models.ThreeProblemGenerationRequest,
    userSchoolCode string,
    progressCallback func(stage int, message string),
) (*models.ThreeProblemGenerationResponse, error)
```

**処理フロー:**

1. **初期化**
   - ユーザー情報取得
   - 生成回数チェック（3問で1回カウント）
   - 生成回数を更新
   - 大ステップ会話履歴を初期化

2. **パターンA生成（Stage 1-5）**
   - 小ステップ会話履歴を初期化
   - Stage 1-5を順次実行
   - 進捗コールバック: stage 1-5
   - 結果をPatternAに格納

3. **パターンB生成（Stage 6-10）**
   - 小ステップ会話履歴を初期化
   - Stage 1-5を順次実行
   - 進捗コールバック: stage 6-10
   - 結果をPatternBに格納

4. **パターンC生成（Stage 11-15）**
   - 小ステップ会話履歴を初期化
   - Stage 1-5を順次実行
   - 進捗コールバック: stage 11-15
   - 結果をPatternCに格納

5. **データベース保存**
   - 3問をそれぞれproblemsテーブルに保存
   - 各問題に会話履歴を含める

#### generateSinglePattern
```go
func (s *problemService) generateSinglePattern(
    ctx context.Context,
    patternName string,
    uploadedProblemContent string,
    userSchoolCode string,
    mainHistory *models.ConversationHistory,
    progressCallback func(stage int, message string),
    baseStage int, // 1, 6, 11
) (*models.PatternResult, error)
```

**処理フロー:**

1. **小ステップ会話履歴を初期化**
   ```go
   patternHistory := &models.ConversationHistory{
       Messages: make([]models.ConversationMessage, 0),
   }
   ```

2. **初期プロンプトを作成**
   ```go
   initialPrompt := s.promptLoader.LoadThreeProblemGenerationPrompt(
       uploadedProblemContent,
       patternName, // "パターンA: 数値だけ違う"
   )
   ```

3. **Stage 1-5を順次実行**
   - 各ステージで`generateStageWithHistory`を呼び出し
   - 結果を抽出してPatternResultに格納
   - 進捗コールバックを呼び出し

4. **図形生成（Stage 3の後）**
   - Stage 3で生成されたPythonコードを抽出
   - coreClientで図形を生成
   - ImageBase64に格納

5. **結果を返す**

#### generateStageWithHistory
```go
func (s *problemService) generateStageWithHistory(
    ctx context.Context,
    stage int,
    history *models.ConversationHistory,
    userSchoolCode string,
) (string, error)
```

**処理フロー:**

1. **トリガーを送信**（Stage 2-5の場合）
   ```go
   if stage > 1 {
       trigger := s.promptLoader.LoadStageTrigger()
       s.buildConversationHistory(history, trigger, "", stage)
   }
   ```

2. **AI呼び出し**
   ```go
   user, _ := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
   clientMessages := s.convertToClientMessages(history)
   
   var content string
   switch user.PreferredAPI {
   case "claude":
       client := clients.NewClaudeClient(user.PreferredModel)
       content, err = client.GenerateWithHistory(ctx, clientMessages)
   // ... 他のAPI
   }
   ```

3. **結果を会話履歴に追加**
   ```go
   s.buildConversationHistory(history, "", content, stage)
   ```

4. **結果を返す**

### 4. ハンドラー層

#### GenerateThreeProblemsSSE
```go
func (h *ProblemHandler) GenerateThreeProblemsSSE(w http.ResponseWriter, r *http.Request)
```

**処理フロー:**

1. **SSE接続を確立**
   ```go
   w.Header().Set("Content-Type", "text/event-stream")
   w.Header().Set("Cache-Control", "no-cache")
   w.Header().Set("Connection", "keep-alive")
   flusher, _ := w.(http.Flusher)
   ```

2. **リクエストを解析**
   ```go
   var req models.ThreeProblemGenerationRequest
   json.NewDecoder(r.Body).Decode(&req)
   ```

3. **進捗コールバックを定義**
   ```go
   progressCallback := func(stage int, message string) {
       event := map[string]interface{}{
           "type":    "stage_progress",
           "stage":   stage,
           "message": message,
       }
       data, _ := json.Marshal(event)
       fmt.Fprintf(w, "data: %s\n\n", data)
       flusher.Flush()
   }
   ```

4. **3問生成を実行**
   ```go
   response, err := h.problemService.GenerateThreeProblemsWithProgress(
       r.Context(),
       req,
       user.SchoolCode,
       progressCallback,
   )
   ```

5. **完了イベントを送信**
   ```go
   completeEvent := map[string]interface{}{
       "type":    "complete",
       "message": json.Marshal(response),
   }
   data, _ := json.Marshal(completeEvent)
   fmt.Fprintf(w, "data: %s\n\n", data)
   flusher.Flush()
   ```

### 5. ルーティング

```go
// router.go
router.HandleFunc("/api/generate-three-problems-sse", 
    problemHandler.GenerateThreeProblemsSSE).Methods("POST")
```

### 6. フロントエンド連携

#### リクエスト形式
```typescript
const response = await fetch(`${API_BASE_URL}/api/generate-three-problems-sse`, {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
        uploaded_problem_content: fileContent,
        subject: '数学',
    })
});
```

#### SSEイベント処理
```typescript
const reader = response.body?.getReader();
const decoder = new TextDecoder();

while (reader) {
    const { done, value } = await reader.read();
    if (done) break;
    
    const data = decoder.decode(value);
    const event = JSON.parse(data.slice(6)); // "data: " を除去
    
    if (event.type === 'stage_progress') {
        setCurrentStage(event.stage);
        setProgressMessage(event.message);
    } else if (event.type === 'complete') {
        const result = JSON.parse(event.message);
        // 3問を表示
    }
}
```

### 7. 進捗表示

#### ステージマッピング
```
Stage 1-5:   パターンA生成（数値だけ違う）
Stage 6-10:  パターンB生成（公式は同じだが問題は違う）
Stage 11-15: パターンC生成（構成は同じだが問われる部分が違う）
```

#### 進捗率計算
```typescript
const progress = (currentStage / 15) * 100;
```

### 8. データベース保存

各パターンを個別のレコードとして保存：

```go
// パターンA
problemA := &models.Problem{
    UserID:              user.ID,
    Subject:             req.Subject,
    Prompt:              "パターンA: 数値だけ違う",
    Content:             response.PatternA.Content,
    Solution:            response.PatternA.Solution,
    ImageBase64:         response.PatternA.ImageBase64,
    ConversationHistory: response.PatternA.ConversationHistory,
    CreatedAt:           time.Now(),
    UpdatedAt:           time.Now(),
}
s.problemRepo.Create(ctx, problemA)

// パターンB, Cも同様
```

### 9. エラーハンドリング

1. **ユーザー認証エラー**
   - 401 Unauthorized

2. **生成回数上限エラー**
   - エラーメッセージ: "問題生成回数の上限に達しました"

3. **ファイルアップロードなしエラー**
   - エラーメッセージ: "問題ファイルをアップロードしてください"

4. **AI呼び出しエラー**
   - 各ステージでリトライ
   - 最大3回まで

5. **図形生成エラー**
   - 図形なしで続行
   - ImageBase64を空文字列に設定

### 10. 実装の優先順位

1. **Phase 1: サービス層の基本実装**
   - GenerateThreeProblemsWithProgress
   - generateSinglePattern
   - generateStageWithHistory

2. **Phase 2: ハンドラー層の実装**
   - GenerateThreeProblemsSSE
   - ルーティング追加

3. **Phase 3: フロントエンド実装**
   - 生成モード選択UI
   - SSE接続とイベント処理
   - 進捗表示
   - 3問表示UI

4. **Phase 4: テストと調整**
   - 統合テスト
   - エラーハンドリングの確認
   - パフォーマンス最適化

## 実装時の注意点

1. **会話履歴の管理**
   - 大ステップと小ステップの会話履歴を明確に分離
   - 各パターンの会話履歴は独立

2. **進捗コールバック**
   - 各ステージの開始時と完了時に呼び出し
   - ステージ番号は1-15の範囲

3. **生成回数のカウント**
   - 3問生成で1回とカウント
   - 処理開始前にカウントを更新

4. **エラー時の処理**
   - 途中でエラーが発生した場合、それまでの結果を保持
   - 部分的な成功も許容

5. **タイムアウト**
   - 各AI呼び出しに適切なタイムアウトを設定
   - 全体で最大10分程度を想定