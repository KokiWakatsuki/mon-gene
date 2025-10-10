package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/services"
	"github.com/mon-gene/back/internal/utils"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.SchoolCode == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "塾コードは必須です")
		return
	}
	if req.Password == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "パスワードは必須です")
		return
	}

	response, err := h.authService.Login(r.Context(), req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.SchoolCode == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "塾コードは必須です")
		return
	}

	response, err := h.authService.ForgotPassword(r.Context(), req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	utils.EnableCORS(w)
	
	// OPTIONSリクエスト（プリフライト）の処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	// ユーザー情報のレスポンス（パスワードハッシュは除外）
	response := map[string]interface{}{
		"success":                   true,
		"school_code":              user.SchoolCode,
		"email":                    user.Email,
		"problem_generation_limit": user.ProblemGenerationLimit,
		"problem_generation_count": user.ProblemGenerationCount,
		"role":                     user.Role,
		"preferred_api":            user.PreferredAPI,
		"preferred_model":          user.PreferredModel,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	utils.EnableCORS(w)
	
	// OPTIONSリクエスト（プリフライト）の処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	// ユーザープロファイルのレスポンス
	response := map[string]interface{}{
		"id":                       user.ID,
		"school_code":              user.SchoolCode,
		"email":                    user.Email,
		"role":                     user.Role,
		"preferred_api":            user.PreferredAPI,
		"preferred_model":          user.PreferredModel,
		"problem_generation_limit": user.ProblemGenerationLimit,
		"problem_generation_count": user.ProblemGenerationCount,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	utils.EnableCORS(w)
	
	// OPTIONSリクエスト（プリフライト）の処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	// リクエストボディを解析
	var req struct {
		PreferredAPI   string `json:"preferred_api"`
		PreferredModel string `json:"preferred_model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	validAPIs := map[string]bool{"chatgpt": true, "claude": true, "gemini": true}
	if !validAPIs[req.PreferredAPI] {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なAPIが指定されました")
		return
	}

	if req.PreferredModel == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "モデル名は必須です")
		return
	}

	// ユーザー設定を更新
	err = h.authService.UpdateUserSettings(r.Context(), user.SchoolCode, req.PreferredAPI, req.PreferredModel)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "設定の更新に失敗しました")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "設定を更新しました",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if err := h.authService.Logout(r.Context(), token); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "ログアウトしました",
	})
}
