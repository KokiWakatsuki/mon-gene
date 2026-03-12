package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mon-gene/back/internal/clients"
	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/services"
	"github.com/mon-gene/back/internal/utils"
)

type SSEHandler struct {
	problemService services.ProblemService
	authService    services.AuthService
}

func NewSSEHandler(problemService services.ProblemService, authService services.AuthService) *SSEHandler {
	return &SSEHandler{
		problemService: problemService,
		authService:    authService,
	}
}

// StageProgressEvent SSEで送信するステージ進捗イベント
type StageProgressEvent struct {
	Type    string `json:"type"`    // "stage_start", "stage_complete", "error", "complete"
	Stage   int    `json:"stage"`   // 1-5
	Message string `json:"message"` // ステージの説明
	Error   string `json:"error,omitempty"`
}

// GenerateProblemFiveStageSSE 5段階生成をSSEで実行
func (h *SSEHandler) GenerateProblemFiveStageSSE(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	utils.EnableCORS(w)
	
	// OPTIONSリクエストの処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// SSEヘッダーを設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// 認証チェック
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:  "error",
			Error: "認証トークンが見つかりません",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:  "error",
			Error: "認証に失敗しました",
		})
		return
	}

	// リクエストボディをパース
	var req models.FiveStageGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:  "error",
			Error: fmt.Sprintf("リクエストのパースに失敗しました: %v", err),
		})
		return
	}

	userSchoolCode := user.SchoolCode

	// 各ステージを順次実行し、進捗をSSEで送信
	ctx := r.Context()
	
	// イベントと完了通知用のチャネル
	eventChan := make(chan StageProgressEvent, 100)
	doneChan := make(chan struct{})
	
	// Stage 1開始
	eventChan <- StageProgressEvent{
		Type:    "stage_start",
		Stage:   1,
		Message: "小問構成と解答プロセスを生成中...",
	}
	
	// 実際の5段階生成を非同期で実行
	go func() {
		defer close(doneChan)
		result, err := h.problemService.GenerateProblemFiveStageWithProgress(ctx, req, userSchoolCode, func(stage int, message string) {
			eventChan <- StageProgressEvent{
				Type:    "stage_complete",
				Stage:   stage,
				Message: message,
			}
			
			// 次のステージ開始通知
			if stage < 5 {
				nextStageMessages := []string{
					"",
					"パラメータ設定と動的検証を実行中...",
					"問題文用の図形を描画中...",
					"完全な問題文を生成中...",
					"完全な解答・解説を生成中...",
				}
				eventChan <- StageProgressEvent{
					Type:    "stage_start",
					Stage:   stage + 1,
					Message: nextStageMessages[stage],
				}
			}
		})
		
		if err != nil {
			eventChan <- StageProgressEvent{
				Type:  "error",
				Error: fmt.Sprintf("問題生成に失敗しました: %v", err),
			}
			return
		}
		
		// 完了イベントを送信
		resultJSON, _ := json.Marshal(result)
		eventChan <- StageProgressEvent{
			Type:    "complete",
			Stage:   5,
			Message: string(resultJSON),
		}
	}()

	// Cloudflare等のタイムアウト（100秒）防止用Ticker（15秒間隔）
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventChan:
			h.sendSSEEvent(w, event)
			if event.Type == "error" || event.Type == "complete" {
				return
			}
		case <-ticker.C:
			// 接続維持のための空コメントを送信
			fmt.Fprintf(w, ": keepalive\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-doneChan:
			// goroutine終了時はループを抜ける
			return
		}
	}
}

// sendSSEEvent SSEイベントを送信
func (h *SSEHandler) sendSSEEvent(w http.ResponseWriter, event StageProgressEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"イベントのシリアライズに失敗しました\"}\n\n")
		w.(http.Flusher).Flush()
		return
	}
	
	fmt.Fprintf(w, "data: %s\n\n", string(data))
	w.(http.Flusher).Flush()
}

// GenerateThreeProblemsSSE 3問生成をSSEで実行（15段階プロセス）
func (h *SSEHandler) GenerateThreeProblemsSSE(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	utils.EnableCORS(w)
	
	// OPTIONSリクエストの処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// SSEヘッダーを設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// 認証チェック
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:  "error",
			Error: "認証トークンが見つかりません",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:  "error",
			Error: "認証に失敗しました",
		})
		return
	}

	// Content-Typeをチェックしてmultipart/form-dataかJSONかを判定
	contentType := r.Header.Get("Content-Type")
	var req models.ThreeProblemGenerationRequest
	
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// multipart/form-dataの場合：PDFファイルを処理
		fmt.Printf("📦 [SSE] Processing multipart/form-data request\n")
		
		// マルチパートフォームをパース（最大32MB）
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			h.sendSSEEvent(w, StageProgressEvent{
				Type:  "error",
				Error: fmt.Sprintf("マルチパートフォームのパースに失敗しました: %v", err),
			})
			return
		}
		
		// subjectフィールドを取得
		subject := r.FormValue("subject")
		if subject == "" {
			subject = "数学" // デフォルト値
		}
		req.Subject = subject
		
		// excluded_unitsフィールドを取得（JSON文字列として送信されている）
		excludedUnitsStr := r.FormValue("excluded_units")
		if excludedUnitsStr != "" {
			var excludedUnits []string
			if err := json.Unmarshal([]byte(excludedUnitsStr), &excludedUnits); err != nil {
				fmt.Printf("⚠️ [SSE] Failed to parse excluded_units: %v\n", err)
			} else {
				req.ExcludedUnits = excludedUnits
				fmt.Printf("📎 [SSE] Excluded units: %v\n", excludedUnits)
			}
		}
		
		// PDFファイルを取得
		file, header, err := r.FormFile("file")
		if err != nil {
			h.sendSSEEvent(w, StageProgressEvent{
				Type:  "error",
				Error: fmt.Sprintf("PDFファイルの取得に失敗しました: %v", err),
			})
			return
		}
		defer file.Close()
		
		fmt.Printf("📄 [SSE] Received PDF file: %s (size: %d bytes)\n", header.Filename, header.Size)
		
		// PDFファイルをバイト配列として読み込む
		pdfData, err := io.ReadAll(file)
		if err != nil {
			h.sendSSEEvent(w, StageProgressEvent{
				Type:  "error",
				Error: fmt.Sprintf("PDFファイルの読み込みに失敗しました: %v", err),
			})
			return
		}
		
		req.UploadedProblemPDF = pdfData
		fmt.Printf("✅ [SSE] PDF data loaded successfully (%d bytes)\n", len(pdfData))
		
	} else {
		// JSONの場合：従来通りテキストコンテンツを処理
		fmt.Printf("📝 [SSE] Processing JSON request\n")
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendSSEEvent(w, StageProgressEvent{
				Type:  "error",
				Error: fmt.Sprintf("リクエストのパースに失敗しました: %v", err),
			})
			return
		}
		
		// アップロードされた問題内容のチェック
		if req.UploadedProblemContent == "" {
			h.sendSSEEvent(w, StageProgressEvent{
				Type:  "error",
				Error: "問題ファイルをアップロードしてください",
			})
			return
		}
	}
	
	// PDFまたはテキストコンテンツのいずれかが必要
	if len(req.UploadedProblemPDF) == 0 && req.UploadedProblemContent == "" {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:  "error",
			Error: "問題ファイル（PDFまたはテキスト）をアップロードしてください",
		})
		return
	}

	userSchoolCode := user.SchoolCode
	ctx := r.Context()
	
	// 15段階のステージメッセージ
	stageMessages := map[int]string{
		1:  "パターンA - Stage 1: 骨組み設計中...",
		2:  "パターンA - Stage 2: パラメータ設定と動的検証中...",
		3:  "パターンA - Stage 3: 図形描画中...",
		4:  "パターンA - Stage 4: 完全な問題文生成中...",
		5:  "パターンA - Stage 5: 完全な解答・解説生成中...",
		6:  "パターンB - Stage 1: 骨組み設計中...",
		7:  "パターンB - Stage 2: パラメータ設定と動的検証中...",
		8:  "パターンB - Stage 3: 図形描画中...",
		9:  "パターンB - Stage 4: 完全な問題文生成中...",
		10: "パターンB - Stage 5: 完全な解答・解説生成中...",
		11: "パターンC - Stage 1: 骨組み設計中...",
		12: "パターンC - Stage 2: パラメータ設定と動的検証中...",
		13: "パターンC - Stage 3: 図形描画中...",
		14: "パターンC - Stage 4: 完全な問題文生成中...",
		15: "パターンC - Stage 5: 完全な解答・解説生成中...",
	}
	
	// イベントと完了通知用のチャネル
	eventChan := make(chan StageProgressEvent, 100)
	doneChan := make(chan struct{})

	// Stage 1開始
	eventChan <- StageProgressEvent{
		Type:    "stage_start",
		Stage:   1,
		Message: stageMessages[1],
	}
	
	// 3問生成を非同期で実行（進捗コールバック付き）
	go func() {
		defer close(doneChan)
		result, err := h.problemService.GenerateThreeProblemsWithProgress(ctx, req, userSchoolCode, func(stage int, message string) {
			// ステージ完了通知
			eventChan <- StageProgressEvent{
				Type:    "stage_complete",
				Stage:   stage,
				Message: message,
			}
			
			// 次のステージ開始通知
			if stage < 15 {
				nextStage := stage + 1
				eventChan <- StageProgressEvent{
					Type:    "stage_start",
					Stage:   nextStage,
					Message: stageMessages[nextStage],
				}
			}
		})
		
		if err != nil {
			eventChan <- StageProgressEvent{
				Type:  "error",
				Error: fmt.Sprintf("3問生成に失敗しました: %v", err),
			}
			return
		}
		
		if !result.Success {
			eventChan <- StageProgressEvent{
				Type:  "error",
				Error: result.Error,
			}
			return
		}
		
		// 完了イベントを送信
		resultJSON, _ := json.Marshal(result)
		eventChan <- StageProgressEvent{
			Type:    "complete",
			Stage:   15,
			Message: string(resultJSON),
		}
	}()

	// Cloudflare等のタイムアウト（100秒）防止用Ticker（15秒間隔）
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventChan:
			h.sendSSEEvent(w, event)
			if event.Type == "error" || event.Type == "complete" {
				return
			}
		case <-ticker.C:
			// 接続維持のための空コメントを送信
			fmt.Fprintf(w, ": keepalive\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-doneChan:
			// goroutine終了時はループを抜ける
			return
		}
	}
}

// PreviewPDFContent PDFファイルの内容をプレビュー用に抽出（回数制限付き）
func (h *SSEHandler) PreviewPDFContent(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	utils.EnableCORS(w)
	
	// OPTIONSリクエストの処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// 認証チェック
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが見つかりません")
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証に失敗しました")
		return
	}
	
	// プレビュー回数制限チェック
	if user.PreviewLimit != -1 && user.PreviewCount >= user.PreviewLimit {
		utils.WriteErrorResponse(w, http.StatusForbidden, fmt.Sprintf("問題概要表示の上限（%d回）に達しました", user.PreviewLimit))
		return
	}

	// マルチパートフォームをパース（最大32MB）
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("マルチパートフォームのパースに失敗しました: %v", err))
		return
	}
	
	// PDFファイルを取得
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("PDFファイルの取得に失敗しました: %v", err))
		return
	}
	defer file.Close()
	
	fmt.Printf("📄 [PreviewPDF] Received PDF file: %s (size: %d bytes)\n", header.Filename, header.Size)
	
	// PDFファイルをバイト配列として読み込む
	pdfData, err := io.ReadAll(file)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("PDFファイルの読み込みに失敗しました: %v", err))
		return
	}
	
	fmt.Printf("✅ [PreviewPDF] PDF data loaded successfully (%d bytes)\n", len(pdfData))
	
	// PDFファイルサイズのバリデーション（20MBまで）
	maxPDFSize := 20 * 1024 * 1024 // 20MB
	if len(pdfData) > maxPDFSize {
		fmt.Printf("❌ [PreviewPDF] PDF file too large: %d bytes (max: %d bytes)\n", len(pdfData), maxPDFSize)
		utils.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("PDFファイルが大きすぎます（最大20MB）。現在のサイズ: %.2f MB", float64(len(pdfData))/(1024*1024)))
		return
	}
	
	// Google Files APIを使用してPDFからテキストを抽出
	// 簡単な抽出プロンプトを使用
	extractPrompt := "このPDFファイルに含まれる問題文を正確に抽出してください。数式、図形の説明、問題番号などすべての情報を含めてください。"
	
	fmt.Printf("🔄 [PreviewPDF] Starting PDF text extraction with Google API...\n")
	fmt.Printf("👤 [PreviewPDF] User API: %s, Model: %s\n", user.PreferredAPI, user.PreferredModel)
	
	// ユーザーの設定に基づいてモデルを選択（Google APIのみ対応）
	// Google以外のAPIが設定されている場合は、デフォルトでgemini-1.5-flashを使用
	modelToUse := "gemini-1.5-flash" // デフォルト（安定版）
	if user.PreferredAPI == "google" || user.PreferredAPI == "gemini" {
		if user.PreferredModel != "" {
			modelToUse = user.PreferredModel
			fmt.Printf("🔧 [PreviewPDF] Using user's preferred model: %s\n", modelToUse)
		}
	} else {
		fmt.Printf("⚠️ [PreviewPDF] User's API (%s) doesn't support PDF, using default Google model: %s\n", user.PreferredAPI, modelToUse)
	}
	
	// Google Clientを使用
	googleClient := clients.NewGoogleClient(modelToUse)
	extractedText, err := googleClient.GenerateContentWithPDF(r.Context(), extractPrompt, pdfData)
	if err != nil {
		fmt.Printf("❌ [PreviewPDF] PDF extraction failed: %v\n", err)
		// エラーの種類に応じて適切なメッセージを返す
		errorMsg := fmt.Sprintf("PDFからのテキスト抽出に失敗しました: %v", err)
		if strings.Contains(err.Error(), "API key") {
			errorMsg = "Google API キーが設定されていません。管理者に連絡してください。"
		} else if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "rate limit") {
			errorMsg = "APIの利用制限に達しました。しばらく待ってから再試行してください。"
		} else if strings.Contains(err.Error(), "too large") || strings.Contains(err.Error(), "size") {
			errorMsg = "PDFファイルが大きすぎます。より小さいファイルを使用してください。"
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			errorMsg = fmt.Sprintf("指定されたモデル（%s）が利用できません。gemini-1.5-flashまたはgemini-1.5-proを使用してください。", modelToUse)
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, errorMsg)
		return
	}
	
	fmt.Printf("✅ [PreviewPDF] Successfully extracted text from PDF (length: %d)\n", len(extractedText))
	
	// プレビュー回数をインクリメント
	if err := h.authService.IncrementPreviewCount(r.Context(), user.ID); err != nil {
		fmt.Printf("⚠️ [PreviewPDF] Failed to increment preview count: %v\n", err)
		// エラーでも処理は続行（カウント失敗は致命的ではない）
	} else {
		fmt.Printf("✅ [PreviewPDF] Preview count incremented for user %d\n", user.ID)
	}
	
	// レスポンスを返す
	response := map[string]interface{}{
		"success":  true,
		"filename": header.Filename,
		"size":     header.Size,
		"content":  extractedText,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}