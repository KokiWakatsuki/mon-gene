package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/services"
	"github.com/mon-gene/back/internal/utils"
)

type ProblemHandler struct {
	problemService services.ProblemService
	authService    services.AuthService
}

func NewProblemHandler(problemService services.ProblemService, authService services.AuthService) *ProblemHandler {
	return &ProblemHandler{
		problemService: problemService,
		authService:    authService,
	}
}

func (h *ProblemHandler) GenerateProblem(w http.ResponseWriter, r *http.Request) {
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

	// デバッグ: リクエストの内容を確認
	println("📋 [DEBUG] GenerateProblem request:")
	println("  OpinionProfile (v1):", req.OpinionProfile != nil)
	println("  OpinionProfileV2 (v2):", req.OpinionProfileV2 != nil)

	// opinion_ver1.md基準を使用している場合の追加検証（レガシー）
	if req.OpinionProfile != nil {
		if req.OpinionProfile.Domain < 1 || req.OpinionProfile.Domain > 6 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "出題分野コードは1-6の範囲で指定してください")
			return
		}
		if req.OpinionProfile.SkillLevel < 1 || req.OpinionProfile.SkillLevel > 10 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "コアスキルレベルは1-10の範囲で指定してください")
			return
		}
		if req.OpinionProfile.DifficultyScore < 1 || req.OpinionProfile.DifficultyScore > 20 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "総合難易度スコアは1-20の範囲で指定してください")
			return
		}
		if req.OpinionProfile.StructureComplexity[0] < 1 || req.OpinionProfile.StructureComplexity[0] > 10 ||
		   req.OpinionProfile.StructureComplexity[1] < 1 || req.OpinionProfile.StructureComplexity[1] > 10 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "問題構造評価は各軸とも1-10の範囲で指定してください")
			return
		}
		
		println("📋 [DEBUG] Opinion ver1 criteria detected (legacy):")
		println("  Domain:", req.OpinionProfile.Domain)
		println("  SkillLevel:", req.OpinionProfile.SkillLevel)
		println("  StructureComplexity: [", req.OpinionProfile.StructureComplexity[0], ",", req.OpinionProfile.StructureComplexity[1], "]")
		println("  DifficultyScore:", req.OpinionProfile.DifficultyScore)
	}

	// opinion_ver2.md基準を使用している場合の追加検証
	if req.OpinionProfileV2 != nil {
		// 基本的な数値範囲チェック
		if req.OpinionProfileV2.ProblemTextLength < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "大問の問題文文字数は0以上である必要があります")
			return
		}
		if req.OpinionProfileV2.SubProblemTextLength < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "小問の総文字数は0以上である必要があります")
			return
		}
		if req.OpinionProfileV2.GivenValuesCount < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "与えられる数値の個数は0以上である必要があります")
			return
		}
		if req.OpinionProfileV2.SubProblemCount < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "小問の数は0以上である必要があります")
			return
		}
		if req.OpinionProfileV2.TotalVertices < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "総頂点・点の数は0以上である必要があります")
			return
		}
		if req.OpinionProfileV2.FigureValuesCount < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "図中に明記された数値の個数は0以上である必要があります")
			return
		}
		if req.OpinionProfileV2.SolutionSteps < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "解法のステップ数は0以上である必要があります")
			return
		}
		if req.OpinionProfileV2.TheoremCount < 0 {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "使用定理・公式の数は0以上である必要があります")
			return
		}
		
		println("📋 [DEBUG] Opinion ver2 criteria detected:")
		println("  ProblemTextLength:", req.OpinionProfileV2.ProblemTextLength)
		println("  SubProblemCount:", req.OpinionProfileV2.SubProblemCount)
		println("  SolutionSteps:", req.OpinionProfileV2.SolutionSteps)
		println("  HasMovingPoint:", req.OpinionProfileV2.HasMovingPoint)
		println("  TheoremCount:", req.OpinionProfileV2.TheoremCount)
	}

	// ユーザーのSchoolCodeを渡して問題を生成
	problem, err := h.problemService.GenerateProblem(r.Context(), req, user.SchoolCode)
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

// SearchProblems キーワードで問題を検索
func (h *ProblemHandler) SearchProblems(w http.ResponseWriter, r *http.Request) {
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

	// クエリパラメータから検索キーワードを取得
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "検索キーワードは必須です")
		return
	}

	// ページネーション
	limit := 20
	offset := 0

	problems, err := h.problemService.SearchProblemsByKeyword(r.Context(), user.ID, keyword, limit, offset)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"problems": problems,
		"count":    len(problems),
	})
}

// 5段階生成システムのハンドラー（高精度）

// GenerateProblemFiveStage 5段階生成プロセス全体を実行
func (h *ProblemHandler) GenerateProblemFiveStage(w http.ResponseWriter, r *http.Request) {
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

	var req models.FiveStageGenerationRequest
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

	// デバッグ: リクエストの内容を確認
	println("📋 [DEBUG] FiveStageGeneration request:")
	println("  OpinionProfile (v1):", req.OpinionProfile != nil)
	println("  OpinionProfileV2 (v2):", req.OpinionProfileV2 != nil)

	// 5段階生成プロセスを実行
	response, err := h.problemService.GenerateProblemFiveStage(r.Context(), req, user.SchoolCode)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GenerateStage1 1段階目：小問構成と解答プロセス生成（新しいプロセス）
func (h *ProblemHandler) GenerateStage1(w http.ResponseWriter, r *http.Request) {
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

	var req models.Stage1Request
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

	// 1段階目を実行（小問構成と解答プロセス生成）
	response, err := h.problemService.GenerateStage1(r.Context(), req, user.SchoolCode)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GenerateStage2 2段階目：完全な問題生成（新しいプロセス）
func (h *ProblemHandler) GenerateStage2(w http.ResponseWriter, r *http.Request) {
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

	var req models.Stage2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.SubProblemsAndProcess == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "小問構成と解答プロセスは必須です")
		return
	}

	// 2段階目を実行（完全な問題生成）
	response, err := h.problemService.GenerateStage2(r.Context(), req, user.SchoolCode)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GenerateStage3 3段階目：数値計算プログラム生成・実行（新しいプロセス）
func (h *ProblemHandler) GenerateStage3(w http.ResponseWriter, r *http.Request) {
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

	var req models.Stage3Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.SubProblemsAndProcess == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "小問構成と解答プロセスは必須です")
		return
	}
	if req.CompleteProblem == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "完全な問題は必須です")
		return
	}

	// 3段階目を実行（数値計算プログラム生成・実行）
	response, err := h.problemService.GenerateStage3(r.Context(), req, user.SchoolCode)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GenerateStage4 4段階目：完全な解答・解説生成（新しいプロセス）
func (h *ProblemHandler) GenerateStage4(w http.ResponseWriter, r *http.Request) {
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

	var req models.Stage4Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.SubProblemsAndProcess == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "小問構成と解答プロセスは必須です")
		return
	}
	if req.CompleteProblem == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "完全な問題は必須です")
		return
	}
	if req.CalculationResults == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "計算結果は必須です")
		return
	}

	// 4段階目を実行（完全な解答・解説生成）
	response, err := h.problemService.GenerateStage4(r.Context(), req, user.SchoolCode)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GenerateStage5 5段階目：図形描画プログラム生成（新しいプロセス）
func (h *ProblemHandler) GenerateStage5(w http.ResponseWriter, r *http.Request) {
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

	var req models.Stage5Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.CompleteProblem == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "完全な問題は必須です")
		return
	}

	// 5段階目を実行（図形描画プログラム生成）
	response, err := h.problemService.GenerateStage5(r.Context(), req, user.SchoolCode)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Stage5完了後、5段階生成の結果をproblemsテーブルに保存
	// （オプション：フロントエンドがFiveStageDataを送信した場合のみ）
	if response.Success && req.FiveStageData != nil {
		println("💾 [Stage5Handler] Attempting to save complete 5-stage problem to database")
		println("🔍 [Stage5Handler] Problem data prepared:")
		println("  Subject:", req.FiveStageData.Subject)
		println("  Content length:", len(req.CompleteProblem))
		println("  Solution length:", len(req.FiveStageData.FinalExplanation))
		println("  Has image:", len(response.ImageBase64) > 0)
		println("  OpinionProfileV2 (v2):", req.FiveStageData.OpinionProfileV2 != nil)
		
		// 実際のDB保存処理を実行
		problem := &models.Problem{
			UserID:           user.ID,
			Subject:          req.FiveStageData.Subject,
			Prompt:           req.FiveStageData.Prompt,
			Content:          req.CompleteProblem,
			Solution:         req.FiveStageData.FinalExplanation, // Stage4で生成された解答を保存
			ImageBase64:      response.ImageBase64,
			OpinionProfileV2: req.FiveStageData.OpinionProfileV2, // Ver.2のみ使用
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		
		// GenerateProblemサービスのDB保存ロジックを参考に、直接リポジトリに保存
		if saveErr := h.problemService.SaveDirectProblem(r.Context(), problem); saveErr != nil {
			println("⚠️ [Stage5Handler] Failed to save problem to database:", saveErr.Error())
		} else {
			println("✅ [Stage5Handler] Problem saved to database successfully with ID:", problem.ID)
		}
	} else {
		println("ℹ️ [Stage5Handler] FiveStageData not provided, skipping database save")
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}


// SearchProblemsCombined キーワードとフィルターの組み合わせで問題を検索
func (h *ProblemHandler) SearchProblemsCombined(w http.ResponseWriter, r *http.Request) {
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

	// リクエストボディから検索条件を取得
	var searchRequest struct {
		Keyword   string                 `json:"keyword,omitempty"`
		Subject   string                 `json:"subject,omitempty"`
		Filters   map[string]interface{} `json:"filters,omitempty"`
		MatchType string                 `json:"matchType,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 少なくとも1つの検索条件が指定されている必要がある
	if searchRequest.Keyword == "" && searchRequest.Subject == "" && (searchRequest.Filters == nil || len(searchRequest.Filters) == 0) {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "キーワード、科目、またはフィルター条件のいずれかを指定してください")
		return
	}

	// ページネーション
	limit := 20
	offset := 0

	// デフォルトは部分一致
	matchType := searchRequest.MatchType
	if matchType == "" {
		matchType = "partial"
	}

	problems, err := h.problemService.SearchProblemsCombined(r.Context(), user.ID, searchRequest.Keyword, searchRequest.Subject, searchRequest.Filters, matchType, limit, offset)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"problems": problems,
		"count":    len(problems),
	})
}

// GetUserProblems ユーザーの問題履歴を取得
func (h *ProblemHandler) GetUserProblems(w http.ResponseWriter, r *http.Request) {
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

	// ページネーション
	limit := 20
	offset := 0

	problems, err := h.problemService.GetUserProblems(r.Context(), user.ID, limit, offset)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"problems": problems,
		"count":    len(problems),
	})
}

// UpdateProblem 問題の内容を更新
func (h *ProblemHandler) UpdateProblem(w http.ResponseWriter, r *http.Request) {
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

	var req models.UpdateProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.ID <= 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "問題IDは必須です")
		return
	}
	if req.Content == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "問題文は必須です")
		return
	}

	// 問題を更新
	updatedProblem, err := h.problemService.UpdateProblem(r.Context(), req, user.ID)
	if err != nil {
		if err.Error() == "problem not found or access denied" {
			utils.WriteErrorResponse(w, http.StatusForbidden, "問題が見つからないか、アクセス権限がありません")
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.UpdateProblemResponse{
		Success: true,
		Problem: updatedProblem,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// RegenerateGeometry 問題の図形を再生成
func (h *ProblemHandler) RegenerateGeometry(w http.ResponseWriter, r *http.Request) {
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

	var req models.RegenerateGeometryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// バリデーション
	if req.ID <= 0 {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "問題IDは必須です")
		return
	}

	// 図形を再生成
	imageBase64, err := h.problemService.RegenerateGeometry(r.Context(), req, user.ID)
	if err != nil {
		if err.Error() == "problem not found or access denied" {
			utils.WriteErrorResponse(w, http.StatusForbidden, "問題が見つからないか、アクセス権限がありません")
			return
		}
		if err.Error() == "no geometry needed for this problem" {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "この問題には図形は不要です")
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := models.RegenerateGeometryResponse{
		Success:     true,
		ImageBase64: imageBase64,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// SearchProblemsByFilters パラメータ（フィルター）で問題を検索
func (h *ProblemHandler) SearchProblemsByFilters(w http.ResponseWriter, r *http.Request) {
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

	// リクエストボディから検索条件を取得
	var searchRequest struct {
		Subject   string                 `json:"subject,omitempty"`
		Filters   map[string]interface{} `json:"filters,omitempty"`
		MatchType string                 `json:"matchType,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 少なくとも科目またはフィルターのいずれかが指定されている必要がある
	if searchRequest.Subject == "" && (searchRequest.Filters == nil || len(searchRequest.Filters) == 0) {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "科目またはフィルター条件を指定してください")
		return
	}

	// ページネーション
	limit := 20
	offset := 0

	// デフォルトは部分一致
	matchType := searchRequest.MatchType
	if matchType == "" {
		matchType = "partial"
	}

	problems, err := h.problemService.SearchProblemsByFilters(r.Context(), user.ID, searchRequest.Subject, searchRequest.Filters, matchType, limit, offset)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"problems": problems,
		"count":    len(problems),
	})
}
