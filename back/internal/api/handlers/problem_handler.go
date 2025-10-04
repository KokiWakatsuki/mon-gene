package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/services"
	"github.com/mon-gene/back/internal/utils"
)

type ProblemHandler struct {
	problemService services.ProblemService
}

func NewProblemHandler(problemService services.ProblemService) *ProblemHandler {
	return &ProblemHandler{
		problemService: problemService,
	}
}

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
	if req.Subject == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "科目は必須です")
		return
	}

	problem, err := h.problemService.GenerateProblem(r.Context(), req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// デバッグログを追加
	println("🔍 [DEBUG] Generated problem:")
	println("  Content length:", len(problem.Content))
	println("  Solution length:", len(problem.Solution))
	println("  ImageBase64 length:", len(problem.ImageBase64))
	if len(problem.Solution) > 0 {
		println("  Solution preview:", problem.Solution[:min(100, len(problem.Solution))])
	} else {
		println("  Solution preview: (empty)")
	}

	// レスポンス形式に変換
	response := models.GenerateProblemResponse{
		Content:     problem.Content,
		Success:     true,
		ImageBase64: problem.ImageBase64,
		Solution:    problem.Solution,
	}

	println("🔍 [DEBUG] Response solution length:", len(response.Solution))

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *ProblemHandler) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	var req models.PDFGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.ProblemText == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "問題文は必須です")
		return
	}

	pdfBase64, err := h.problemService.GeneratePDF(r.Context(), req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.PDFGenerateResponse{
		Success:   true,
		PDFBase64: pdfBase64,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}
