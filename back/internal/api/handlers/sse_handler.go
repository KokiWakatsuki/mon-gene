package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	
	// Stage 1開始
	h.sendSSEEvent(w, StageProgressEvent{
		Type:    "stage_start",
		Stage:   1,
		Message: "小問構成と解答プロセスを生成中...",
	})
	
	// 実際の5段階生成を実行（内部で各ステージの完了を検知）
	// この部分は後で実装します（現在のGenerateProblemFiveStageを分割）
	result, err := h.problemService.GenerateProblemFiveStageWithProgress(ctx, req, userSchoolCode, func(stage int, message string) {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:    "stage_complete",
			Stage:   stage,
			Message: message,
		})
		
		// 次のステージ開始通知
		if stage < 5 {
			nextStageMessages := []string{
				"",
				"パラメータ設定と動的検証を実行中...",
				"問題文用の図形を描画中...",
				"完全な問題文を生成中...",
				"完全な解答・解説を生成中...",
			}
			h.sendSSEEvent(w, StageProgressEvent{
				Type:    "stage_start",
				Stage:   stage + 1,
				Message: nextStageMessages[stage],
			})
		}
	})
	
	if err != nil {
		h.sendSSEEvent(w, StageProgressEvent{
			Type:  "error",
			Error: fmt.Sprintf("問題生成に失敗しました: %v", err),
		})
		return
	}
	
	// 完了イベントを送信
	resultJSON, _ := json.Marshal(result)
	h.sendSSEEvent(w, StageProgressEvent{
		Type:    "complete",
		Stage:   5,
		Message: string(resultJSON),
	})
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