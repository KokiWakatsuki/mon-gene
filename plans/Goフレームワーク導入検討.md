# Go Webフレームワーク導入検討書

## 🎯 結論：**導入を推奨（Gin または Echo）**

現在の標準ライブラリ（`net/http`）から、軽量フレームワークへの移行を推奨します。

---

## 📊 現状の課題

### 1. ルーティングの冗長性
```go
// 現在のrouter.go（217行）
mux.HandleFunc("/api/problems/search", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET", "OPTIONS":
        problemHandler.SearchProblems(w, r)
    default:
        utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
    }
})

mux.HandleFunc("/api/problems/search-by-filters", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST", "OPTIONS":
        problemHandler.SearchProblemsByFilters(w, r)
    default:
        utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
    }
})
```

**問題点**:
- 各エンドポイントでメソッドチェックが必要
- HTTPメソッドごとの分岐が冗長
- エラーハンドリングが重複

### 2. ミドルウェアの実装コスト
```go
// 現在のCORSミドルウェア
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        // ... 多数のヘッダー設定
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**問題点**:
- 標準的な機能を自前実装
- テストが必要
- バグのリスク

### 3. リクエスト/レスポンス処理の手動実装
```go
// 現在のハンドラー
func (h *ProblemHandler) GenerateProblem(w http.ResponseWriter, r *http.Request) {
    var req models.GenerateProblemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
        return
    }
    
    // バリデーション
    if req.Prompt == "" {
        utils.WriteErrorResponse(w, http.StatusBadRequest, "プロンプトは必須です")
        return
    }
    // ...
}
```

**問題点**:
- JSONパース、バリデーション、エラーレスポンスを毎回手動実装
- ボイラープレートコードが多い

---

## 🎯 推奨フレームワーク比較

### 1. Gin（最推奨）⭐

#### 特徴
- **パフォーマンス**: 最速クラス（httprouterベース）
- **学習コスト**: 低い（シンプルなAPI）
- **エコシステム**: 豊富なミドルウェア
- **採用実績**: 非常に多い

#### コード例
```go
// ルーティング
r := gin.Default()

// グループ化
api := r.Group("/api")
{
    // 認証不要
    api.POST("/login", authHandler.Login)
    api.POST("/forgot-password", authHandler.ForgotPassword)
    
    // 認証必要
    authorized := api.Group("")
    authorized.Use(authMiddleware.RequireAuth())
    {
        authorized.POST("/generate-problem", problemHandler.GenerateProblem)
        authorized.GET("/problems/search", problemHandler.SearchProblems)
        authorized.POST("/problems/search-by-filters", problemHandler.SearchProblemsByFilters)
    }
}

// ハンドラー（簡潔に）
func (h *ProblemHandler) GenerateProblem(c *gin.Context) {
    var req models.GenerateProblemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // ビジネスロジック
    result, err := h.service.Generate(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}
```

#### メリット
✅ **コード量削減**: 約30-40%削減  
✅ **可読性向上**: ルーティングが直感的  
✅ **バリデーション**: 構造体タグで自動バリデーション  
✅ **ミドルウェア**: 豊富な既製品  
✅ **パフォーマンス**: 標準ライブラリより高速  

#### デメリット
⚠️ フレームワーク依存が増える  
⚠️ 移行コストがかかる  

---

### 2. Echo（次点）

#### 特徴
- **パフォーマンス**: Ginと同等
- **機能**: より多機能（WebSocket、HTTP/2等）
- **学習コスト**: やや高い
- **採用実績**: 多い

#### コード例
```go
e := echo.New()

// ミドルウェア
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(middleware.CORS())

// ルーティング
api := e.Group("/api")
api.POST("/login", authHandler.Login)

authorized := api.Group("")
authorized.Use(authMiddleware.RequireAuth)
authorized.POST("/generate-problem", problemHandler.GenerateProblem)

// ハンドラー
func (h *ProblemHandler) GenerateProblem(c echo.Context) error {
    var req models.GenerateProblemRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    result, err := h.service.Generate(c.Request().Context(), req)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
    
    return c.JSON(http.StatusOK, result)
}
```

#### メリット
✅ より多機能  
✅ エラーハンドリングが統一的  
✅ WebSocket対応（将来の拡張に有利）  

#### デメリット
⚠️ Ginより学習コストが高い  
⚠️ やや複雑  

---

### 3. Fiber（検討外）

#### 特徴
- Express.js風のAPI
- 非常に高速

#### 検討外の理由
❌ `net/http`と互換性がない（独自実装）  
❌ エコシステムが小さい  
❌ 移行コストが高い  

---

### 4. Chi（検討外）

#### 特徴
- 標準ライブラリベース
- 軽量

#### 検討外の理由
⚠️ 現状とあまり変わらない  
⚠️ バリデーション等は自前実装が必要  
⚠️ 改善効果が限定的  

---

## 📈 導入効果の試算

### コード量の削減

#### Before（現状）
```go
// router.go: 217行
mux.HandleFunc("/api/problems/search", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET", "OPTIONS":
        problemHandler.SearchProblems(w, r)
    default:
        utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
    }
})
// ... 30以上のエンドポイント
```

#### After（Gin導入後）
```go
// router.go: 約80行（60%削減）
r := gin.Default()
api := r.Group("/api")
{
    api.GET("/problems/search", problemHandler.SearchProblems)
    api.POST("/problems/search-by-filters", problemHandler.SearchProblemsByFilters)
    // ... シンプルな定義
}
```

### ハンドラーの簡潔化

#### Before
```go
func (h *Handler) Method(w http.ResponseWriter, r *http.Request) {
    // 1. JSONパース（5-10行）
    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
        return
    }
    
    // 2. バリデーション（10-20行）
    if req.Field1 == "" {
        utils.WriteErrorResponse(w, http.StatusBadRequest, "Field1 is required")
        return
    }
    // ...
    
    // 3. ビジネスロジック
    result, err := h.service.Do(r.Context(), req)
    if err != nil {
        utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    // 4. レスポンス
    utils.WriteJSONResponse(w, http.StatusOK, result)
}
```

#### After（Gin）
```go
func (h *Handler) Method(c *gin.Context) {
    var req Request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    result, err := h.service.Do(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}
```

**削減効果**: 約40%のコード削減

---

## 🔧 具体的な導入メリット

### 1. ルーティングの可読性向上

#### Before
```go
mux.HandleFunc("/api/generate-problem-five-stage-sse", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST", "OPTIONS":
        sseHandler.GenerateProblemFiveStageSSE(w, r)
    default:
        utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
    }
})
```

#### After（Gin）
```go
r.POST("/api/generate-problem-five-stage-sse", sseHandler.GenerateProblemFiveStageSSE)
```

### 2. ミドルウェアの簡潔化

#### Before（自前実装）
```go
// middleware/cors.go（約50行）
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 多数のヘッダー設定...
    })
}
```

#### After（Gin）
```go
import "github.com/gin-contrib/cors"

r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

### 3. バリデーションの自動化

#### Before（手動バリデーション）
```go
if req.Prompt == "" {
    utils.WriteErrorResponse(w, http.StatusBadRequest, "プロンプトは必須です")
    return
}
if req.Subject == "" {
    utils.WriteErrorResponse(w, http.StatusBadRequest, "科目は必須です")
    return
}
```

#### After（構造体タグ）
```go
type GenerateProblemRequest struct {
    Prompt  string `json:"prompt" binding:"required"`
    Subject string `json:"subject" binding:"required"`
}

// ハンドラーで自動バリデーション
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

### 4. グループ化による構造化

```go
r := gin.Default()

// 公開API
public := r.Group("/api")
{
    public.POST("/login", authHandler.Login)
    public.POST("/forgot-password", authHandler.ForgotPassword)
}

// 認証必須API
authorized := r.Group("/api")
authorized.Use(authMiddleware.RequireAuth())
{
    // 問題生成
    problems := authorized.Group("/problems")
    {
        problems.POST("/generate", problemHandler.Generate)
        problems.GET("/search", problemHandler.Search)
        problems.GET("/history", problemHandler.GetHistory)
    }
    
    // ユーザー設定
    user := authorized.Group("/user")
    {
        user.GET("/profile", authHandler.GetProfile)
        user.PUT("/settings", authHandler.UpdateSettings)
    }
}
```

---

## 🚀 移行計画

### Phase 1: 準備（1日）
1. Ginのインストール
   ```bash
   go get -u github.com/gin-gonic/gin
   go get -u github.com/gin-contrib/cors
   ```

2. 依存関係の整理

### Phase 2: ルーティング移行（2日）
1. `router.go`の書き換え
2. ハンドラーシグネチャの変更
   ```go
   // Before
   func (h *Handler) Method(w http.ResponseWriter, r *http.Request)
   
   // After
   func (h *Handler) Method(c *gin.Context)
   ```

### Phase 3: ミドルウェア移行（1日）
1. CORS: `gin-contrib/cors`を使用
2. 認証: カスタムミドルウェアの実装
3. ロギング: Ginの標準ロガーまたはzap統合

### Phase 4: ハンドラー最適化（2日）
1. JSONバインディングの活用
2. バリデーションタグの追加
3. エラーハンドリングの統一

### Phase 5: テスト（1日）
1. 既存のテストの修正
2. 統合テストの実行
3. パフォーマンステスト

**合計**: 約1週間

---

## 📊 導入しない場合のデメリット

### 1. 保守コストの増加
- ボイラープレートコードの維持
- 自前ミドルウェアのメンテナンス
- バグ修正コスト

### 2. 開発効率の低下
- 新機能追加時の冗長なコード
- 新メンバーのオンボーディング時間増加

### 3. 可読性の低下
- ルーティングが分散
- 一貫性のないエラーハンドリング

---

## ✅ 最終推奨

### **Gin を導入すべき理由**

1. **即効性**: 1週間で移行完了、即座に効果
2. **コスト対効果**: 移行コスト < 長期的な保守コスト削減
3. **可読性**: コード量30-40%削減
4. **標準化**: Go Webアプリケーションのデファクトスタンダード
5. **エコシステム**: 豊富なミドルウェア、活発なコミュニティ
6. **パフォーマンス**: 標準ライブラリより高速
7. **学習コスト**: 低い（1-2日で習得可能）

### 導入タイミング
- **今すぐ**: リファクタリングと同時に実施
- 改善アーキテクチャへの移行と並行して進められる

### 代替案
もし「フレームワーク依存を避けたい」場合:
- **Chi** を検討（標準ライブラリベース、軽量）
- ただし改善効果は限定的

---

## 📝 結論

**Gin の導入を強く推奨します。**

理由:
- ✅ 可読性が大幅に向上
- ✅ 保守コストが削減
- ✅ 開発効率が向上
- ✅ 移行コストは1週間程度で回収可能
- ✅ Go Webアプリケーションの標準的な選択

現在のプロジェクトの規模と複雑性を考えると、フレームワーク導入のメリットがデメリットを大きく上回ります。