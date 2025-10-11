package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mon-gene/back/internal/clients"
	"github.com/mon-gene/back/internal/services"
	"github.com/mon-gene/back/internal/utils"
)

type ChatHandler struct {
	authService services.AuthService
}

type ChatRequest struct {
	Message string           `json:"message"`
	Files   []ChatFileUpload `json:"files,omitempty"`
}

type ChatFileUpload struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Data     string `json:"data"`     // base64 encoded data
	MimeType string `json:"mimeType"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
	Model string `json:"model"`
	API   string `json:"api"`
}

func NewChatHandler(authService services.AuthService) *ChatHandler {
	return &ChatHandler{
		authService: authService,
	}
}

func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// JWTトークンから認証情報を取得
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Authorization header missing")
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format")
		return
	}

	user, err := h.authService.ValidateToken(r.Context(), tokenParts[1])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// リクエストボディを解析
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// ユーザーの設定されたAIクライアントを取得
	var aiClient clients.AIClient
	var clientError error

	switch user.PreferredAPI {
	case "claude":
		aiClient = clients.NewClaudeClient(user.PreferredModel)
	case "chatgpt":
		aiClient = clients.NewOpenAIClient(user.PreferredModel)
	case "gemini":
		aiClient = clients.NewGoogleClient(user.PreferredModel)
	case "laboratory":
		// laboratoryの場合はClaude clientを使用
		aiClient = clients.NewClaudeClient(user.PreferredModel)
	default:
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid API configuration. Please check your settings.")
		return
	}

	if clientError != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to initialize AI client: "+clientError.Error())
		return
	}

	// ファイルが含まれている場合はマルチモーダルAPIを使用
	var reply string

	if len(req.Files) > 0 {
		// ファイルをFileContent形式に変換
		var fileContents []clients.FileContent
		for _, file := range req.Files {
			fileContents = append(fileContents, clients.FileContent{
				Name:     file.Name,
				Type:     file.Type,
				Data:     file.Data,
				MimeType: file.MimeType,
			})
		}
		reply, err = aiClient.GenerateMultimodalContent(r.Context(), req.Message, fileContents)
	} else {
		reply, err = aiClient.GenerateContent(r.Context(), req.Message)
	}

	if err != nil {
		// エラーの種類に応じて適切な応答を返す
		if strings.Contains(err.Error(), "API key") {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "API key configuration error: "+err.Error())
			return
		}
		if strings.Contains(err.Error(), "model not specified") {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Model not configured: "+err.Error())
			return
		}
		if strings.Contains(err.Error(), "rate limit") {
			utils.WriteErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded: "+err.Error())
			return
		}
		if strings.Contains(err.Error(), "tokens") || strings.Contains(err.Error(), "length") {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Message too long: "+err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "AI service error: "+err.Error())
		return
	}

	// レスポンスを返す
	response := ChatResponse{
		Reply: reply,
		Model: user.PreferredModel,
		API:   user.PreferredAPI,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
