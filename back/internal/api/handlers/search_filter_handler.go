package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/services"
	"github.com/mon-gene/back/internal/utils"
)

type SearchFilterHandler struct {
	searchFilterService services.SearchFilterService
	authService         services.AuthService
}

func NewSearchFilterHandler(searchFilterService services.SearchFilterService, authService services.AuthService) *SearchFilterHandler {
	return &SearchFilterHandler{
		searchFilterService: searchFilterService,
		authService:         authService,
	}
}

// CreateSearchFilter 検索条件を作成
func (h *SearchFilterHandler) CreateSearchFilter(w http.ResponseWriter, r *http.Request) {
	// 認証トークンを取得
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// トークンからユーザー情報を取得
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	var req models.CreateSearchFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.Name == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "検索条件の名前は必須です")
		return
	}

	filter, err := h.searchFilterService.CreateSearchFilter(r.Context(), req, user.ID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.SearchFilterResponse{
		Success:      true,
		SearchFilter: filter,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetSearchFilter 検索条件を取得
func (h *SearchFilterHandler) GetSearchFilter(w http.ResponseWriter, r *http.Request) {
	// 認証トークンを取得
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// トークンからユーザー情報を取得
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	// パスパラメータからIDを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/search-filters/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なIDです")
		return
	}

	filter, err := h.searchFilterService.GetSearchFilter(r.Context(), id, user.ID)
	if err != nil {
		if err.Error() == "アクセス権限がありません" {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.SearchFilterResponse{
		Success:      true,
		SearchFilter: filter,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetUserSearchFilters ユーザーの検索条件一覧を取得
func (h *SearchFilterHandler) GetUserSearchFilters(w http.ResponseWriter, r *http.Request) {
	// 認証トークンを取得
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// トークンからユーザー情報を取得
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	filters, err := h.searchFilterService.GetUserSearchFilters(r.Context(), user.ID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.SearchFilterResponse{
		Success: true,
		Filters: filters,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdateSearchFilter 検索条件を更新
func (h *SearchFilterHandler) UpdateSearchFilter(w http.ResponseWriter, r *http.Request) {
	// 認証トークンを取得
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// トークンからユーザー情報を取得
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	var req models.UpdateSearchFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.ID <= 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "検索条件IDは必須です")
		return
	}
	if req.Name == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "検索条件の名前は必須です")
		return
	}

	filter, err := h.searchFilterService.UpdateSearchFilter(r.Context(), req, user.ID)
	if err != nil {
		if err.Error() == "アクセス権限がありません" {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.SearchFilterResponse{
		Success:      true,
		SearchFilter: filter,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeleteSearchFilter 検索条件を削除
func (h *SearchFilterHandler) DeleteSearchFilter(w http.ResponseWriter, r *http.Request) {
	// 認証トークンを取得
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "認証トークンが必要です")
		return
	}

	// "Bearer " プレフィックスを削除
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// トークンからユーザー情報を取得
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "無効な認証トークンです")
		return
	}

	// パスパラメータからIDを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/search-filters/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "無効なIDです")
		return
	}

	err = h.searchFilterService.DeleteSearchFilter(r.Context(), id, user.ID)
	if err != nil {
		if err.Error() == "アクセス権限がありません" {
			utils.WriteErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.SearchFilterResponse{
		Success: true,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}