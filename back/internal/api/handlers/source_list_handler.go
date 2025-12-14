package handlers

import (
	"encoding/json"
	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/services"
	"github.com/mon-gene/back/internal/utils"
	"net/http"
	"strconv"
	"strings"
)

type SourceListHandler struct {
	service     *services.SourceListService
	authService services.AuthService
}

func NewSourceListHandler(service *services.SourceListService, authService services.AuthService) *SourceListHandler {
	return &SourceListHandler{
		service:     service,
		authService: authService,
	}
}

// CreateItem は出典リスト項目を作成
func (h *SourceListHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
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

	var req models.CreateSourceListItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item, err := h.service.AddItem(user.ID, req.Year, req.ExamSession)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, models.SourceListResponse{
		Success: true,
		Items:   []*models.SourceListItem{item},
	})
}

// GetUserItems はユーザーの出典リストを取得
func (h *SourceListHandler) GetUserItems(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.service.GetUserItems(user.ID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, models.SourceListResponse{
		Success: true,
		Items:   items,
	})
}

// DeleteItem は出典リスト項目を削除
func (h *SourceListHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
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

	// パスからIDを取得
	path := strings.TrimPrefix(r.URL.Path, "/api/source-list/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = h.service.DeleteItem(id, user.ID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, models.SourceListResponse{
		Success: true,
	})
}