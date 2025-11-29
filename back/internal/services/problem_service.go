package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mon-gene/back/internal/clients"
	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/repositories"
	"github.com/mon-gene/back/internal/utils"
)

type ProblemService interface {
	GenerateProblem(ctx context.Context, req models.GenerateProblemRequest, userSchoolCode string) (*models.Problem, error)
	GeneratePDF(ctx context.Context, req models.PDFGenerateRequest) (string, error)
	UpdateProblem(ctx context.Context, req models.UpdateProblemRequest, userID int64) (*models.Problem, error)
	RegenerateGeometry(ctx context.Context, req models.RegenerateGeometryRequest, userID int64) (string, error)
	SearchProblemsByFilters(ctx context.Context, userID int64, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error)
	SearchProblemsByKeyword(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*models.Problem, error)
	SearchProblemsCombined(ctx context.Context, userID int64, keyword string, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error)
	GetUserProblems(ctx context.Context, userID int64, limit, offset int) ([]*models.Problem, error)
	SaveDirectProblem(ctx context.Context, problem *models.Problem) error
	
	// 5段階生成メソッド（高精度）
	GenerateProblemFiveStage(ctx context.Context, req models.FiveStageGenerationRequest, userSchoolCode string) (*models.FiveStageGenerationResponse, error)
	GenerateProblemFiveStageWithProgress(ctx context.Context, req models.FiveStageGenerationRequest, userSchoolCode string, progressCallback func(stage int, message string)) (*models.FiveStageGenerationResponse, error)
	GenerateStage1(ctx context.Context, req models.Stage1Request, userSchoolCode string) (*models.Stage1Response, error)
	GenerateStage2(ctx context.Context, req models.Stage2Request, userSchoolCode string) (*models.Stage2Response, error)
	GenerateStage3(ctx context.Context, req models.Stage3Request, userSchoolCode string) (*models.Stage3Response, error)
	GenerateStage4(ctx context.Context, req models.Stage4Request, userSchoolCode string) (*models.Stage4Response, error)
	GenerateStage5(ctx context.Context, req models.Stage5Request, userSchoolCode string) (*models.Stage5Response, error)
	
	// 3問生成メソッド（15段階プロセス）
	GenerateThreeProblems(ctx context.Context, req models.ThreeProblemGenerationRequest, userSchoolCode string) (*models.ThreeProblemGenerationResponse, error)
	GenerateThreeProblemsWithProgress(ctx context.Context, req models.ThreeProblemGenerationRequest, userSchoolCode string, progressCallback func(stage int, message string)) (*models.ThreeProblemGenerationResponse, error)
}

type problemService struct {
	claudeClient  clients.ClaudeClient
	openaiClient  clients.OpenAIClient
	googleClient  clients.GoogleClient
	coreClient    clients.CoreClient
	problemRepo   repositories.ProblemRepository
	userRepo      repositories.UserRepository
	promptLoader  *utils.PromptLoader
}

func NewProblemService(
	claudeClient clients.ClaudeClient,
	openaiClient clients.OpenAIClient,
	googleClient clients.GoogleClient,
	coreClient clients.CoreClient,
	problemRepo repositories.ProblemRepository,
	userRepo repositories.UserRepository,
) ProblemService {
	// promptsディレクトリのパスを設定
	promptLoader := utils.NewPromptLoader("prompts")
	
	return &problemService{
		claudeClient:  claudeClient,
		openaiClient:  openaiClient,
		googleClient:  googleClient,
		coreClient:    coreClient,
		problemRepo:   problemRepo,
		userRepo:      userRepo,
		promptLoader:  promptLoader,
	}
}

func (s *problemService) GenerateProblem(ctx context.Context, req models.GenerateProblemRequest, userSchoolCode string) (*models.Problem, error) {
	// 1. ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Note: 既存問題の重複チェック機能は削除されました（不要な複雑性のため）
	
	// 3. ユーザーの問題生成回数制限をチェック
	
	// 制限チェック（-1は制限なし）
	if user.ProblemGenerationLimit >= 0 && user.ProblemGenerationCount >= user.ProblemGenerationLimit {
		return nil, fmt.Errorf("問題生成回数の上限（%d回）に達しました", user.ProblemGenerationLimit)
	}
	
	fmt.Printf("🔢 User %s: %d/%d problems generated\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit)
	
	// 問題生成成功時にユーザーの生成回数を更新（生成前に更新して制限をチェック）
	user.ProblemGenerationCount++
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		fmt.Printf("⚠️ Failed to update user generation count: %v\n", err)
		return nil, fmt.Errorf("問題生成カウントの更新に失敗しました: %w", err)
	} else {
		fmt.Printf("✅ 問題生成カウントを更新: %s = %d/%d\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit)
	}

	// ユーザーの設定に基づいてAI/モデル情報をconsoleに表示
	preferredAPI := user.PreferredAPI
	preferredModel := user.PreferredModel
	
	// 設定が空の場合はエラーを返す
	if preferredAPI == "" || preferredModel == "" {
		return nil, fmt.Errorf("AI設定が不完全です。設定ページでAPIとモデルを選択してください。現在の設定: API=%s, モデル=%s", preferredAPI, preferredModel)
	}
	
	fmt.Printf("🤖 AI設定 - API: %s, モデル: %s (ユーザー: %s)\n", preferredAPI, preferredModel, userSchoolCode)
	
	// 2. ユーザーの設定に基づいて適切なAIクライアントを選択
	// 注: GenerateProblemは旧方式のため、プロンプトをそのまま使用
	fmt.Printf("🔍 User prompt: %s\n", req.Prompt)
	
	var content string
	switch preferredAPI {
	case "openai", "chatgpt":
		// ユーザーの設定に基づいて新しいクライアントを作成
		dynamicClient := clients.NewOpenAIClient(preferredModel)
		content, err = dynamicClient.GenerateContent(ctx, req.Prompt)
		if err != nil {
			return nil, fmt.Errorf("OpenAI APIでの問題生成に失敗しました: %w", err)
		}
	case "google", "gemini":
		// ユーザーの設定に基づいて新しいクライアントを作成
		dynamicClient := clients.NewGoogleClient(preferredModel)
		content, err = dynamicClient.GenerateContent(ctx, req.Prompt)
		if err != nil {
			return nil, fmt.Errorf("Google APIでの問題生成に失敗しました: %w", err)
		}
	case "claude", "laboratory":
		// ユーザーの設定に基づいて新しいクライアントを作成
		// laboratoryもClaudeとして扱う
		dynamicClient := clients.NewClaudeClient(preferredModel)
		content, err = dynamicClient.GenerateContent(ctx, req.Prompt)
		if err != nil {
			return nil, fmt.Errorf("Claude APIでの問題生成に失敗しました: %w", err)
		}
	default:
		return nil, fmt.Errorf("サポートされていないAPI「%s」が指定されています。設定ページで正しいAPIを選択してください。サポートされているAPI: openai, google, claude", preferredAPI)
	}
	
	contentPreview := content
	if len(content) > 200 {
		contentPreview = content[:200] + "..."
	}
	fmt.Printf("✅ 問題生成完了 - 使用AI: %s, 使用モデル: %s\n", preferredAPI, preferredModel)
	fmt.Printf("📝 Generated content preview: %s\n", contentPreview)

	// 2. 問題文、Pythonコード、解答・解説を抽出
	problemText := s.extractProblemText(content)
	pythonCode := s.extractPythonCode(content)
	solutionText := s.extractSolutionText(content)
	
	fmt.Printf("🐍 Python code extracted: %t\n", pythonCode != "")
	fmt.Printf("📚 Solution extracted: %t\n", solutionText != "")
	
	cleanPreview := problemText
	if len(problemText) > 200 {
		cleanPreview = problemText[:200] + "..."
	}
	fmt.Printf("📝 Problem text: %s\n", cleanPreview)

	var imageBase64 string

	if pythonCode != "" {
		fmt.Printf("🎨 Generating custom geometry with Python code\n")
		// カスタムPythonコードで図形を生成
		imageBase64, err = s.coreClient.GenerateCustomGeometry(ctx, pythonCode, problemText)
		if err != nil {
			// エラーログを出力するが、処理は続行
			fmt.Printf("❌ Error generating custom geometry: %v\n", err)
		} else {
			fmt.Printf("✅ Custom geometry generated successfully\n")
		}
	} else {
		fmt.Printf("🔍 Analyzing problem for geometry needs\n")
		// 従来の方法で図形が必要かどうかを分析
		analysis, err := s.coreClient.AnalyzeProblem(ctx, problemText, nil)
		if err != nil {
			fmt.Printf("❌ Error analyzing problem: %v\n", err)
		} else {
			fmt.Printf("📊 Analysis result - needs_geometry: %t, detected_shapes: %v\n", 
				analysis.NeedsGeometry, analysis.DetectedShapes)
			
			if analysis.NeedsGeometry && len(analysis.DetectedShapes) > 0 {
				// 最初に検出された図形を描画
				shapeType := analysis.DetectedShapes[0]
				fmt.Printf("🎨 Generating geometry for shape: %s\n", shapeType)
				
				if params, exists := analysis.SuggestedParameters[shapeType]; exists {
					imageBase64, err = s.coreClient.GenerateGeometry(ctx, shapeType, params)
					if err != nil {
						fmt.Printf("❌ Error generating geometry: %v\n", err)
					} else {
						fmt.Printf("✅ Geometry generated successfully for %s\n", shapeType)
					}
				} else {
					fmt.Printf("⚠️ No parameters found for shape: %s\n", shapeType)
				}
			} else {
				fmt.Printf("ℹ️ No geometry needed for this problem\n")
			}
		}
	}
	
	fmt.Printf("🖼️ Final image base64 length: %d\n", len(imageBase64))

	// 3. 問題をデータベースに保存
	problem := &models.Problem{
		UserID:           user.ID,
		Subject:          req.Subject,
		Prompt:           req.Prompt,
		Content:          problemText,
		Solution:         solutionText,
		ImageBase64:      imageBase64,
		OpinionProfile:   req.OpinionProfile,   // レガシー（後方互換性）
		OpinionProfileV2: req.OpinionProfileV2, // 新基準（Ver.2）
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// リポジトリが実装されている場合のみ保存
	if s.problemRepo != nil {
		if err := s.problemRepo.Create(ctx, problem); err != nil {
			return nil, fmt.Errorf("failed to save problem: %w", err)
		}
		fmt.Printf("💾 Problem saved to database with ID: %d\n", problem.ID)
	}


	return problem, nil
}

// SaveDirectProblem 問題を直接データベースに保存
func (s *problemService) SaveDirectProblem(ctx context.Context, problem *models.Problem) error {
	if s.problemRepo == nil {
		return fmt.Errorf("problem repository is not initialized")
	}

	if err := s.problemRepo.Create(ctx, problem); err != nil {
		return fmt.Errorf("failed to save problem: %w", err)
	}

	fmt.Printf("💾 [SaveDirectProblem] Problem saved to database with ID: %d\n", problem.ID)
	return nil
}

func (s *problemService) GeneratePDF(ctx context.Context, req models.PDFGenerateRequest) (string, error) {
	pdfBase64, err := s.coreClient.GeneratePDF(ctx, req.ProblemText, req.ImageBase64, req.SolutionText)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}
	return pdfBase64, nil
}

// createGeometryRegenerationPrompt creates a prompt for regenerating geometry from existing problem text
func (s *problemService) createGeometryRegenerationPrompt(problemText string) string {
	prompt, err := s.promptLoader.LoadGeometryRegenerationPromptWithSamples(problemText)
	if err != nil {
		fmt.Printf("⚠️ Failed to load geometry regeneration prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		prompt, err = s.promptLoader.LoadGeometryRegenerationPrompt(problemText)
		if err != nil {
			return "図形生成プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return prompt
}


// createFiveStageInitialPrompt 5段階生成の初期プロンプトを作成（全ステージの指示を含む）
func (s *problemService) createFiveStageInitialPrompt(userPrompt, subject, opinionProfile string) string {
	promptText, err := s.promptLoader.LoadFiveStageInitialPrompt(userPrompt, subject, opinionProfile)
	if err != nil {
		fmt.Printf("⚠️ Failed to load five stage initial prompt: %v\n", err)
		return "5段階生成初期プロンプトの読み込みに失敗しました: " + err.Error()
	}
	return promptText
}

// loadStageTrigger ステージトリガーを読み込む
func (s *problemService) loadStageTrigger() string {
	promptText, err := s.promptLoader.LoadStageTrigger()
	if err != nil {
		fmt.Printf("⚠️ Failed to load stage trigger: %v\n", err)
		return "次のステージに進んでください。"
	}
	return promptText
}

// createGeometryPromptWithSamples 図形描画プロンプト（新Stage5用）
func (s *problemService) createGeometryPromptWithSamples(problemText string) string {
	promptText, err := s.promptLoader.LoadGeometryPromptWithSamples(problemText)
	if err != nil {
		fmt.Printf("⚠️ Failed to load geometry prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadGeometryRegenerationPrompt(problemText)
		if err != nil {
			return "図形描画プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// DEPRECATED: 古いプロンプトメソッドは削除済み（プロンプトファイルに移行）


// extractProblemText extracts problem text from the content
func (s *problemService) extractProblemText(content string) string {
	re := regexp.MustCompile(`(?s)---PROBLEM_START---(.*?)---PROBLEM_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	// フォールバック: 従来の方法で図形コードと解答を除去
	cleaned := s.removePythonCode(content)
	cleaned = s.removeSolutionText(cleaned)
	return strings.TrimSpace(cleaned)
}

// extractPythonCode extracts Python code from the content
func (s *problemService) extractPythonCode(content string) string {
	// パターン1: 旧形式のマーカー
	re := regexp.MustCompile(`(?s)---GEOMETRY_CODE_START---(.*?)---GEOMETRY_CODE_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		pythonCode := strings.TrimSpace(matches[1])
		// import文を除去
		pythonCode = s.removeImportStatements(pythonCode)
		return pythonCode
	}
	
	// パターン2: ```python```コードブロック（Stage 3の新形式）
	re = regexp.MustCompile("(?s)```python\\s*\\n(.*?)\\n```")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		pythonCode := strings.TrimSpace(matches[1])
		fmt.Printf("✅ [extractPythonCode] Extracted Python code from ```python``` block (length: %d)\n", len(pythonCode))
		// import文を除去
		pythonCode = s.removeImportStatements(pythonCode)
		return pythonCode
	}
	
	fmt.Printf("❌ [extractPythonCode] No Python code found in content (length: %d)\n", len(content))
	return ""
}

// removeImportStatements removes import statements from Python code
func (s *problemService) removeImportStatements(code string) string {
	lines := strings.Split(code, "\n")
	var cleanLines []string
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// import文やfrom文を除去
		if !strings.HasPrefix(trimmedLine, "import ") && 
		   !strings.HasPrefix(trimmedLine, "from ") {
			cleanLines = append(cleanLines, line)
		} else {
			fmt.Printf("🚫 Removed import statement: %s\n", trimmedLine)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

// extractSolutionText extracts solution text from the content
func (s *problemService) extractSolutionText(content string) string {
	fmt.Printf("🔍 [DEBUG] Extracting solution from content (length: %d)\n", len(content))
	
	re := regexp.MustCompile(`(?s)---SOLUTION_START---(.*?)---SOLUTION_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		solution := strings.TrimSpace(matches[1])
		fmt.Printf("✅ [DEBUG] Solution extracted successfully (length: %d)\n", len(solution))
		return solution
	}
	
	fmt.Printf("❌ [DEBUG] No solution markers found, checking for alternative patterns\n")
	
	// 代替パターン1: 【解答】や【解説】を含む部分を探す
	solutionPatterns := []string{
		`(?s)【解答・解説】(.*?)(?:---|\z)`,
		`(?s)【解答】(.*?)(?:【|---|\z)`,
		`(?s)【解説】(.*?)(?:【|---|\z)`,
		`(?s)解答・解説(.*?)(?:---|\z)`,
		`(?s)解答:(.*?)(?:解説|---|\z)`,
		`(?s)解説:(.*?)(?:---|\z)`,
	}
	
	for i, pattern := range solutionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			solution := strings.TrimSpace(matches[1])
			if len(solution) > 10 { // 最低限の長さチェック
				fmt.Printf("✅ [DEBUG] Solution found with pattern %d (length: %d)\n", i+1, len(solution))
				return solution
			}
		}
	}
	
	fmt.Printf("❌ [DEBUG] No solution found with any pattern\n")
	fmt.Printf("🔍 [DEBUG] Content preview (last 500 chars): %s\n", content[max(0, len(content)-500):])
	
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// removePythonCode removes Python code from the content
func (s *problemService) removePythonCode(content string) string {
	re := regexp.MustCompile(`(?s)---GEOMETRY_CODE_START---.*?---GEOMETRY_CODE_END---`)
	return strings.TrimSpace(re.ReplaceAllString(content, ""))
}

// removeSolutionText removes solution text from the content
func (s *problemService) removeSolutionText(content string) string {
	re := regexp.MustCompile(`(?s)---SOLUTION_START---.*?---SOLUTION_END---`)
	return strings.TrimSpace(re.ReplaceAllString(content, ""))
}


// SearchProblemsByFilters フィルター（パラメータ）で問題を検索
func (s *problemService) SearchProblemsByFilters(ctx context.Context, userID int64, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error) {
	if s.problemRepo == nil {
		return nil, fmt.Errorf("problem repository is not initialized")
	}
	
	problems, err := s.problemRepo.SearchByFilters(ctx, userID, subject, filters, matchType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by filters: %w", err)
	}
	
	return problems, nil
}

// SearchProblemsByKeyword キーワードで問題を検索
func (s *problemService) SearchProblemsByKeyword(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*models.Problem, error) {
	if s.problemRepo == nil {
		return nil, fmt.Errorf("problem repository is not initialized")
	}
	
	problems, err := s.problemRepo.SearchByKeyword(ctx, userID, keyword, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by keyword: %w", err)
	}
	
	return problems, nil
}

// SearchProblemsCombined キーワードとフィルターの組み合わせで問題を検索
func (s *problemService) SearchProblemsCombined(ctx context.Context, userID int64, keyword string, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error) {
	if s.problemRepo == nil {
		return nil, fmt.Errorf("problem repository is not initialized")
	}
	
	problems, err := s.problemRepo.SearchCombined(ctx, userID, keyword, subject, filters, matchType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by combined conditions: %w", err)
	}
	
	return problems, nil
}

// GetUserProblems ユーザーの問題一覧を取得
func (s *problemService) GetUserProblems(ctx context.Context, userID int64, limit, offset int) ([]*models.Problem, error) {
	if s.problemRepo == nil {
		return nil, fmt.Errorf("problem repository is not initialized")
	}
	
	problems, err := s.problemRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user problems: %w", err)
	}
	
	return problems, nil
}

// UpdateProblem 問題のテキスト内容を更新
func (s *problemService) UpdateProblem(ctx context.Context, req models.UpdateProblemRequest, userID int64) (*models.Problem, error) {
	if s.problemRepo == nil {
		return nil, fmt.Errorf("problem repository is not initialized")
	}

	// 問題の所有者確認
	existingProblem, err := s.problemRepo.GetByIDAndUserID(ctx, req.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	// 更新するフィールドをコピー
	updatedProblem := *existingProblem
	updatedProblem.Content = req.Content
	updatedProblem.Solution = req.Solution
	updatedProblem.UpdatedAt = time.Now()

	// データベースの更新
	if err := s.problemRepo.Update(ctx, &updatedProblem); err != nil {
		return nil, fmt.Errorf("failed to update problem: %w", err)
	}

	fmt.Printf("✅ Problem %d updated successfully\n", req.ID)
	return &updatedProblem, nil
}

// RegenerateGeometry 問題の図形を再生成（会話履歴を使用）
func (s *problemService) RegenerateGeometry(ctx context.Context, req models.RegenerateGeometryRequest, userID int64) (string, error) {
	if s.problemRepo == nil {
		return "", fmt.Errorf("problem repository is not initialized")
	}

	// 問題の所有者確認
	problem, err := s.problemRepo.GetByIDAndUserID(ctx, req.ID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get problem: %w", err)
	}

	// ユーザー情報を取得（制限チェックとAIクライアント選択のため）
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// 図形再生成回数の制限をチェック
	if user.FigureRegenerationLimit >= 0 && user.FigureRegenerationCount >= user.FigureRegenerationLimit {
		return "", fmt.Errorf("図形再生成回数の上限（%d回）に達しました", user.FigureRegenerationLimit)
	}

	fmt.Printf("🔢 [RegenerateGeometry] User %d: %d/%d figure regenerations used\n", userID, user.FigureRegenerationCount, user.FigureRegenerationLimit)
	fmt.Printf("🎨 [RegenerateGeometry] Starting geometry regeneration for problem ID: %d\n", req.ID)

	// 使用する問題文を決定（編集後の問題文がある場合はそれを使用）
	contentToAnalyze := problem.Content
	if req.Content != "" {
		contentToAnalyze = req.Content
		fmt.Printf("🔄 [RegenerateGeometry] Using edited content for geometry regeneration\n")
		fmt.Printf("📝 [RegenerateGeometry] Edited content preview: %s\n", contentToAnalyze[:min(200, len(contentToAnalyze))])
	} else {
		fmt.Printf("📝 [RegenerateGeometry] Using original content for geometry regeneration\n")
	}

	var imageBase64 string

	// 会話履歴がある場合は、それを使用して図形を再生成
	if problem.ConversationHistory != nil && len(problem.ConversationHistory.Messages) > 0 {
		fmt.Printf("💬 [RegenerateGeometry] Using conversation history (%d messages) for geometry regeneration\n", len(problem.ConversationHistory.Messages))
		
		// 会話履歴に図形再生成のトリガーを追加
		history := problem.ConversationHistory
		trigger := s.loadStageTrigger()
		s.buildConversationHistory(history, trigger, "", 5)
		
		// 会話履歴を使用してAPI呼び出し
		var content string
		clientMessages := s.convertToClientMessages(history)
		
		switch user.PreferredAPI {
		case "openai", "chatgpt":
			dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
			content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
		case "google", "gemini":
			dynamicClient := clients.NewGoogleClient(user.PreferredModel)
			content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
		case "claude", "laboratory":
			dynamicClient := clients.NewClaudeClient(user.PreferredModel)
			content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
		default:
			return "", fmt.Errorf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		}
		
		if err != nil {
			fmt.Printf("⚠️ [RegenerateGeometry] Failed to regenerate with conversation history: %v\n", err)
			fmt.Printf("🔄 [RegenerateGeometry] Falling back to standard method\n")
		} else {
			fmt.Printf("✅ [RegenerateGeometry] AI response generated from conversation history\n")
			
			// AIからPythonコードを抽出
			pythonCode := s.extractPythonCode(content)
			fmt.Printf("🐍 [RegenerateGeometry] Python code extracted: %t\n", pythonCode != "")
			
			if pythonCode != "" {
				fmt.Printf("🎨 [RegenerateGeometry] Generating custom geometry with Python code\n")
				imageBase64, err = s.coreClient.GenerateCustomGeometry(ctx, pythonCode, contentToAnalyze)
				if err != nil {
					fmt.Printf("❌ [RegenerateGeometry] Custom geometry generation failed: %v\n", err)
				} else {
					fmt.Printf("✅ [RegenerateGeometry] Custom geometry generated successfully from conversation history\n")
				}
			}
		}
	} else {
		fmt.Printf("ℹ️ [RegenerateGeometry] No conversation history found, using standard method\n")
	}

	// 会話履歴による図形生成が失敗した場合、または会話履歴がない場合は標準の方法を使用
	if imageBase64 == "" {
		fmt.Printf("🤖 [RegenerateGeometry] Generating matplotlib code with AI (standard method)\n")
		
		// 図形生成専用のプロンプトを構築
		geometryPrompt := s.createGeometryPromptWithSamples(contentToAnalyze)
		fmt.Printf("🔍 [RegenerateGeometry] Enhanced prompt created\n")
		
		// ユーザーの設定に基づいてAIクライアントを選択
		preferredAPI := user.PreferredAPI
		preferredModel := user.PreferredModel
	
	if preferredAPI == "" || preferredModel == "" {
		return "", fmt.Errorf("AI設定が不完全です。設定ページでAPIとモデルを選択してください")
	}
	
	fmt.Printf("🤖 [RegenerateGeometry] Using AI - API: %s, Model: %s\n", preferredAPI, preferredModel)
	
	var aiResponse string
	switch preferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(preferredModel)
		aiResponse, err = dynamicClient.GenerateContent(ctx, geometryPrompt)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(preferredModel)
		aiResponse, err = dynamicClient.GenerateContent(ctx, geometryPrompt)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(preferredModel)
		aiResponse, err = dynamicClient.GenerateContent(ctx, geometryPrompt)
	default:
		return "", fmt.Errorf("サポートされていないAPI「%s」が指定されています", preferredAPI)
	}
	
	if err != nil {
		fmt.Printf("❌ [RegenerateGeometry] AI failed, falling back to analysis: %v\n", err)
	} else {
		fmt.Printf("✅ [RegenerateGeometry] AI response generated\n")
		
		// AIからPythonコードを抽出
		pythonCode := s.extractPythonCode(aiResponse)
		fmt.Printf("🐍 [RegenerateGeometry] Python code extracted: %t\n", pythonCode != "")
		
		if pythonCode != "" {
			fmt.Printf("🎨 [RegenerateGeometry] Generating custom geometry with Python code\n")
			// カスタムPythonコードで図形を生成
			imageBase64, err = s.coreClient.GenerateCustomGeometry(ctx, pythonCode, contentToAnalyze)
			if err != nil {
				fmt.Printf("❌ [RegenerateGeometry] Custom geometry generation failed: %v\n", err)
			} else {
				fmt.Printf("✅ [RegenerateGeometry] Custom geometry generated successfully\n")
			}
		}
		}

		// AIによる図形生成が失敗した場合、従来の分析方法にフォールバック
		if imageBase64 == "" {
		fmt.Printf("🔍 [RegenerateGeometry] Falling back to problem analysis\n")
		
		analysis, err := s.coreClient.AnalyzeProblem(ctx, contentToAnalyze, nil)
		if err != nil {
			return "", fmt.Errorf("failed to analyze problem for geometry: %w", err)
		}

		fmt.Printf("📊 [RegenerateGeometry] Analysis result - needs_geometry: %t, detected_shapes: %v\n", 
			analysis.NeedsGeometry, analysis.DetectedShapes)

		if analysis.NeedsGeometry && len(analysis.DetectedShapes) > 0 {
			// 最初に検出された図形を描画
			shapeType := analysis.DetectedShapes[0]
			fmt.Printf("🎨 [RegenerateGeometry] Generating geometry for shape: %s\n", shapeType)
			
			if params, exists := analysis.SuggestedParameters[shapeType]; exists {
				imageBase64, err = s.coreClient.GenerateGeometry(ctx, shapeType, params)
				if err != nil {
					return "", fmt.Errorf("failed to generate geometry: %w", err)
				}
				fmt.Printf("✅ [RegenerateGeometry] Geometry generated successfully for %s\n", shapeType)
			} else {
				return "", fmt.Errorf("no parameters found for shape: %s", shapeType)
			}
		} else {
			return "", fmt.Errorf("no geometry needed for this problem")
		}
		}
	}

	// 図形が生成されなかった場合
	if imageBase64 == "" {
		return "", fmt.Errorf("failed to generate geometry for this problem")
	}

	// データベースの図形を更新
	if err := s.problemRepo.UpdateGeometry(ctx, req.ID, imageBase64); err != nil {
		return "", fmt.Errorf("failed to update geometry in database: %w", err)
	}

	// 図形再生成成功時にユーザーのカウントを更新
	user.FigureRegenerationCount++
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		// ログに記録するが、図形再生成は成功として扱う
		fmt.Printf("⚠️ [RegenerateGeometry] Failed to update figure regeneration count: %v\n", err)
	} else {
		fmt.Printf("✅ [RegenerateGeometry] Updated user %d figure regeneration count to %d\n", userID, user.FigureRegenerationCount)
	}

	fmt.Printf("✅ [RegenerateGeometry] Geometry for problem %d regenerated successfully\n", req.ID)
	return imageBase64, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 5段階生成システムの実装（高精度）


// GenerateStage4 4段階目：完全な解答・解説生成（新しいプロセス）
func (s *problemService) GenerateStage4(ctx context.Context, req models.Stage4Request, userSchoolCode string) (*models.Stage4Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage4] 4段階目を開始：完全な解答・解説生成 (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 4段階目は非推奨（会話形式を使用してください）
	logBuilder.WriteString("⚠️ 個別ステージの呼び出しは非推奨です\n")
	
	return &models.Stage4Response{
		Success: false,
		Error:   "個別ステージの呼び出しは非推奨です。GenerateProblemFiveStageを使用してください。",
		Log:     logBuilder.String(),
	}, fmt.Errorf("deprecated: use GenerateProblemFiveStage instead")
}


// GenerateStage5 5段階目：図形描画プログラム生成（非推奨）
func (s *problemService) GenerateStage5(ctx context.Context, req models.Stage5Request, userSchoolCode string) (*models.Stage5Response, error) {
	return &models.Stage5Response{
		Success: false,
		Error:   "個別ステージの呼び出しは非推奨です。GenerateProblemFiveStageを使用してください。",
		Log:     "⚠️ このAPIは非推奨です。5段階生成プロセス全体を実行するGenerateProblemFiveStageを使用してください。\n",
	}, fmt.Errorf("deprecated: use GenerateProblemFiveStage instead")
}
	
	


// extractSolutionSteps 解答手順を抽出
func (s *problemService) extractSolutionSteps(content string) string {
	re := regexp.MustCompile(`(?s)---SOLUTION_STEPS_START---(.*?)---SOLUTION_STEPS_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// フォールバック：【解答の手順】を探す
	re = regexp.MustCompile(`(?s)【解答の手順】(.*?)(?:---|\n\n|\z)`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}

// extractSolutionProcess 解答プロセスを抽出
func (s *problemService) extractSolutionProcess(content string) string {
	re := regexp.MustCompile(`(?s)---SOLUTION_PROCESS_START---(.*?)---SOLUTION_PROCESS_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// フォールバック：【解答プロセス】を探す
	re = regexp.MustCompile(`(?s)【解答プロセス】(.*?)(?:---|\n\n|\z)`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}

// extractSubProblemsAndProcess 小問構成と解答プロセスを抽出
func (s *problemService) extractSubProblemsAndProcess(content string) string {
	re := regexp.MustCompile(`(?s)---SUB_PROBLEMS_AND_PROCESS_START---(.*?)---SUB_PROBLEMS_AND_PROCESS_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// フォールバック：【小問構成と解答プロセス】を探す
	re = regexp.MustCompile(`(?s)【小問構成と解答プロセス】(.*?)(?:---|\n\n|\z)`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}

// extractCompleteProblem 完全な問題を抽出
func (s *problemService) extractCompleteProblem(content string) string {
	re := regexp.MustCompile(`(?s)---COMPLETE_PROBLEM_START---(.*?)---COMPLETE_PROBLEM_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// Stage 4のマーカーを探す
	re = regexp.MustCompile(`(?s)---STAGE4_START---(.*?)---STAGE4_END---`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// フォールバック：【完全な問題】を探す
	re = regexp.MustCompile(`(?s)【完全な問題】(.*?)(?:---|\n\n|\z)`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}

// extractCalculationProgram 数値計算プログラムを抽出
func (s *problemService) extractCalculationProgram(content string) string {
	fmt.Printf("🔍 [DEBUG] Extracting calculation program from content (length: %d)\n", len(content))
	
	// メインパターン：マーカーを使った抽出
	re := regexp.MustCompile(`(?s)---CALCULATION_PROGRAM_START---(.*?)---CALCULATION_PROGRAM_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		program := strings.TrimSpace(matches[1])
		fmt.Printf("✅ [DEBUG] Calculation program extracted with markers (length: %d)\n", len(program))
		// import文を除去
		cleanProgram := s.removeImportStatements(program)
		if len(cleanProgram) > 10 { // 最低限の長さチェック
			return cleanProgram
		}
	}
	
	fmt.Printf("❌ [DEBUG] No calculation program found with main markers\n")
	
	// フォールバック1：プログラムコードパターンを探す
	alternativePatterns := []string{
		`(?s)# 数値計算プログラム.*?\n(.*?)(?:\n---|\n#.*終了|\z)`,
		`(?s)print\("=== 数値計算結果 ===(.*?)(?:\n---|\z)`,
		`(?s)(import numpy as np.*?)(?:\n---|\z)`,
		`(?s)(# .*計算.*?\n.*?print.*?)(?:\n---|\z)`,
		`(?s)(.*?print.*?=.*?)(?:\n---|\z)`,
	}
	
	for i, pattern := range alternativePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			program := strings.TrimSpace(matches[1])
			// Pythonコードらしきものかチェック
			if strings.Contains(program, "print") || strings.Contains(program, "=") || strings.Contains(program, "import") {
				fmt.Printf("✅ [DEBUG] Calculation program found with pattern %d (length: %d)\n", i+1, len(program))
				cleanProgram := s.removeImportStatements(program)
				if len(cleanProgram) > 5 {
					return cleanProgram
				}
			}
		}
	}
	
	// フォールバック2：全体からPythonコードらしき部分を抽出
	lines := strings.Split(content, "\n")
	var programLines []string
	inCodeSection := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Pythonコードの開始を検出
		if strings.Contains(trimmed, "import numpy") || 
		   strings.Contains(trimmed, "print(") ||
		   strings.Contains(trimmed, "# 数値計算") ||
		   strings.Contains(trimmed, "=== 数値計算結果 ===") {
			inCodeSection = true
		}
		
		// コードセクション中の場合
		if inCodeSection {
			// セクション終了条件
			if strings.HasPrefix(trimmed, "---") && 
			   !strings.Contains(trimmed, "CALCULATION_PROGRAM") {
				break
			}
			
			// 明らかに計算関連の行を追加
			if strings.Contains(trimmed, "print") || 
			   strings.Contains(trimmed, "=") || 
			   strings.Contains(trimmed, "#") ||
			   strings.Contains(trimmed, "import") ||
			   strings.Contains(trimmed, "numpy") ||
			   strings.Contains(trimmed, "math") ||
			   trimmed == "" {
				programLines = append(programLines, line)
			}
		}
	}
	
	if len(programLines) > 0 {
		program := strings.Join(programLines, "\n")
		fmt.Printf("✅ [DEBUG] Fallback extraction found code (length: %d)\n", len(program))
		cleanProgram := s.removeImportStatements(program)
		if len(cleanProgram) > 5 {
			return cleanProgram
		}
	}
	
	fmt.Printf("❌ [DEBUG] No calculation program found with any method\n")
	fmt.Printf("🔍 [DEBUG] Content preview (last 1000 chars): %s\n", content[max(0, len(content)-1000):])
	
	return ""
}

// extractFinalSolution 最終解答を抽出
func (s *problemService) extractFinalSolution(content string) string {
	re := regexp.MustCompile(`(?s)---FINAL_SOLUTION_START---(.*?)---FINAL_SOLUTION_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// Stage 5のマーカーを探す
	re = regexp.MustCompile(`(?s)---STAGE5_START---(.*?)---STAGE5_END---`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// フォールバック：【最終解答】を探す
	re = regexp.MustCompile(`(?s)【最終解答】(.*?)(?:---|\n\n|\z)`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}


// executeCalculationProgram 数値計算プログラムを実行
func (s *problemService) executeCalculationProgram(ctx context.Context, program string) (string, error) {
	fmt.Printf("🧮 [ExecuteCalculation] Starting calculation program execution\n")
	
	// プログラムの前処理：numpy as np、math ライブラリを利用可能にする
	preprocessedProgram := `import numpy as np
import math

` + program
	
	fmt.Printf("🐍 [ExecuteCalculation] Preprocessed program (length: %d)\n", len(preprocessedProgram))
	
	// coreクライアントで実際にPythonプログラムを実行
	executionResult, err := s.coreClient.ExecutePython(ctx, preprocessedProgram)
	if err != nil {
		fmt.Printf("❌ [ExecuteCalculation] Python execution failed: %v\n", err)
		// エラー時は疑似結果を返す
		return fmt.Sprintf(`計算プログラム実行エラー: %v

実行しようとしたプログラム:
%s

注意: Python実行環境でエラーが発生しました。上記のプログラムを手動実行してください。`, err, preprocessedProgram), nil
	}
	
	fmt.Printf("✅ [ExecuteCalculation] Python program executed successfully\n")
	fmt.Printf("📊 [ExecuteCalculation] Execution output length: %d\n", len(executionResult))
	
	// 実行結果をフォーマット
	formattedResults := fmt.Sprintf(`=== 数値計算実行結果 ===

%s

=== 実行されたプログラム ===
%s`, executionResult, preprocessedProgram)
	
	return formattedResults, nil
}

// 5段階生成システムの実装（新しいプロセス）

// GenerateProblemFiveStage 全体の5段階生成プロセスを実行（会話形式）
func (s *problemService) GenerateProblemFiveStage(ctx context.Context, req models.FiveStageGenerationRequest, userSchoolCode string) (*models.FiveStageGenerationResponse, error) {
	return s.GenerateProblemFiveStageWithProgress(ctx, req, userSchoolCode, nil)
}

// GenerateProblemFiveStageWithProgress 全体の5段階生成プロセスを実行（進捗コールバック付き）
func (s *problemService) GenerateProblemFiveStageWithProgress(ctx context.Context, req models.FiveStageGenerationRequest, userSchoolCode string, progressCallback func(stage int, message string)) (*models.FiveStageGenerationResponse, error) {
	fmt.Printf("🚀 [FiveStage] Starting NEW five-stage problem generation (CONVERSATION MODE) for user: %s\n", userSchoolCode)
	fmt.Printf("🔍 [FiveStage] Request details: Prompt length=%d, Subject=%s\n", len(req.Prompt), req.Subject)
	
	// 会話履歴を初期化
	conversationHistory := &models.ConversationHistory{
		Messages: make([]models.ConversationMessage, 0),
	}
	fmt.Printf("💬 [FiveStage] Initialized conversation history\n")
	
	// ユーザー情報を取得して生成制限をチェック
	fmt.Printf("📋 [FiveStage] Fetching user info for: %s\n", userSchoolCode)
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		fmt.Printf("❌ [FiveStage] Failed to get user info: %v\n", err)
		return &models.FiveStageGenerationResponse{
			Success: false,
			Error:   fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err),
		}, nil
	}
	
	fmt.Printf("👤 [FiveStage] User found: ID=%d, SchoolCode=%s, Email=%s\n", user.ID, user.SchoolCode, user.Email)
	fmt.Printf("🔢 [FiveStage] Current generation count: %d (limit: %d)\n", user.ProblemGenerationCount, user.ProblemGenerationLimit)
	
	// 生成制限チェック（-1は制限なし）
	if user.ProblemGenerationLimit >= 0 && user.ProblemGenerationCount >= user.ProblemGenerationLimit {
		fmt.Printf("🚫 [FiveStage] Generation limit reached: %d/%d\n", user.ProblemGenerationCount, user.ProblemGenerationLimit)
		return &models.FiveStageGenerationResponse{
			Success: false,
			Error:   fmt.Sprintf("問題生成回数の上限（%d回）に達しました", user.ProblemGenerationLimit),
		}, nil
	}
	
	fmt.Printf("🔢 [FiveStage] BEFORE UPDATE: User %s has %d/%d problems generated\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit)
	
	// 問題生成成功時にユーザーの生成回数を更新（処理開始前に更新）
	oldCount := user.ProblemGenerationCount
	user.ProblemGenerationCount++
	user.UpdatedAt = time.Now()
	
	fmt.Printf("📝 [FiveStage] Attempting to update user generation count from %d to %d\n", oldCount, user.ProblemGenerationCount)
	fmt.Printf("🕒 [FiveStage] Update timestamp: %s\n", user.UpdatedAt.Format("2006-01-02 15:04:05"))
	
	if err := s.userRepo.Update(ctx, user); err != nil {
		fmt.Printf("❌ [FiveStage] Failed to update user generation count: %v\n", err)
		fmt.Printf("🔍 [FiveStage] User data at failure: ID=%d, Count=%d, Limit=%d\n", user.ID, user.ProblemGenerationCount, user.ProblemGenerationLimit)
		return &models.FiveStageGenerationResponse{
			Success: false,
			Error:   fmt.Sprintf("問題生成カウントの更新に失敗しました: %w", err),
		}, nil
	} else {
		fmt.Printf("✅ [FiveStage] Successfully updated generation count: %s = %d/%d (was %d)\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit, oldCount)
		
		// 更新後に再度ユーザー情報を取得して確認
		verifyUser, verifyErr := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
		if verifyErr != nil {
			fmt.Printf("⚠️ [FiveStage] Failed to verify user update: %v\n", verifyErr)
		} else {
			fmt.Printf("🔍 [FiveStage] VERIFICATION: User %s now has %d/%d problems generated (DB check)\n", userSchoolCode, verifyUser.ProblemGenerationCount, verifyUser.ProblemGenerationLimit)
		}
	}
	
	// 新しいプロセス：1段階目：小問構成と解答プロセス生成（会話形式）
	fmt.Printf("💬 [FiveStage] Stage 1: Starting with conversation history (messages: %d)\n", len(conversationHistory.Messages))
	stage1Req := models.Stage1Request{
		Prompt:         req.Prompt,
		Subject:        req.Subject,
		OpinionProfile: req.OpinionProfile,
		SkipCount:      true, // FiveStage全体呼び出し時は既にカウント済みなのでスキップ
	}
	stage1Resp, err := s.GenerateStage1WithHistory(ctx, stage1Req, userSchoolCode, conversationHistory)
	if err != nil || !stage1Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:   false,
			Error:     fmt.Sprintf("1段階目（小問構成と解答プロセス生成）に失敗しました: %v", err),
			Stage1Log: stage1Resp.Log,
		}, nil
	}
	
	// Stage 1完了を通知
	if progressCallback != nil {
		progressCallback(1, "Stage 1完了：小問構成と解答プロセスを生成しました")
	}
	
	// 新しいプロセス：2段階目：パラメータ設定と動的検証（数値計算）（会話形式）
	fmt.Printf("💬 [FiveStage] Stage 2: Continuing conversation (messages: %d)\n", len(conversationHistory.Messages))
	stage2Req := models.Stage2Request{
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
	}
	stage2Resp, err := s.GenerateStage2WithHistory(ctx, stage2Req, userSchoolCode, conversationHistory)
	if err != nil || !stage2Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("2段階目（パラメータ設定と動的検証）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
		}, nil
	}
	
	// Stage 2完了を通知
	if progressCallback != nil {
		progressCallback(2, "Stage 2完了：パラメータ設定と動的検証を実行しました")
	}
	
	// 新しいプロセス：3段階目：問題文用の図形描画（会話形式）
	fmt.Printf("💬 [FiveStage] Stage 3: Continuing conversation (messages: %d)\n", len(conversationHistory.Messages))
	fmt.Printf("🔍 [FiveStage] Stage 3: About to call GenerateStage3WithHistory\n")
	stage3Req := models.Stage3Request{
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
		CompleteProblem:       stage2Resp.CompleteProblem, // Stage2の検証済みパラメータ
	}
	fmt.Printf("🔍 [FiveStage] Stage 3: Request prepared, calling GenerateStage3WithHistory...\n")
	stage3Resp, err := s.GenerateStage3WithHistory(ctx, stage3Req, userSchoolCode, conversationHistory)
	fmt.Printf("🔍 [FiveStage] Stage 3: GenerateStage3WithHistory returned, success=%t, err=%v\n", stage3Resp != nil && stage3Resp.Success, err)
	if err != nil || !stage3Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("3段階目（図形描画）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			CalculationProgram:    stage2Resp.CompleteProblem,
			CalculationResults:    stage2Resp.CompleteProblem,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
			Stage3Log:             stage3Resp.Log,
		}, nil
	}
	
	// Stage 3完了を通知
	if progressCallback != nil {
		progressCallback(3, "Stage 3完了：問題文用の図形を描画しました")
	}
	
	// 新しいプロセス：4段階目：完全な問題文の生成（会話形式）
	fmt.Printf("💬 [FiveStage] Stage 4: Continuing conversation (messages: %d)\n", len(conversationHistory.Messages))
	stage4Req := models.Stage4Request{
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
		CompleteProblem:       stage2Resp.CompleteProblem, // Stage2の検証済みパラメータ
		CalculationResults:    stage2Resp.CompleteProblem, // Stage2の検証済みパラメータ
	}
	stage4Resp, err := s.GenerateStage4WithHistory(ctx, stage4Req, userSchoolCode, conversationHistory)
	if err != nil || !stage4Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("4段階目（完全な問題文の生成）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			CalculationProgram:    stage2Resp.CompleteProblem,
			CalculationResults:    stage2Resp.CompleteProblem,
			GeometryCode:          stage3Resp.CalculationProgram,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
			Stage3Log:             stage3Resp.Log,
			Stage4Log:             stage4Resp.Log,
		}, nil
	}
	
	// Stage 4完了を通知
	if progressCallback != nil {
		progressCallback(4, "Stage 4完了：完全な問題文を生成しました")
	}
	
	// 新しいプロセス：5段階目：完全な解答・解説の生成（会話形式）
	fmt.Printf("💬 [FiveStage] Stage 5: Continuing conversation (messages: %d)\n", len(conversationHistory.Messages))
	stage5Req := models.Stage5Request{
		CompleteProblem: stage4Resp.FinalExplanation, // Stage4の完全な問題文
	}
	stage5Resp, err := s.GenerateStage5WithHistory(ctx, stage5Req, userSchoolCode, conversationHistory)
	if err != nil || !stage5Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("5段階目（完全な解答・解説の生成）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			CalculationProgram:    stage2Resp.CompleteProblem,
			CalculationResults:    stage2Resp.CompleteProblem,
			GeometryCode:          stage3Resp.CalculationProgram,
			CompleteProblem:       stage4Resp.FinalExplanation,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
			Stage3Log:             stage3Resp.Log,
			Stage4Log:             stage4Resp.Log,
			Stage5Log:             stage5Resp.Log,
		}, nil
	}
	
	// Stage 5完了を通知
	if progressCallback != nil {
		progressCallback(5, "Stage 5完了：完全な解答・解説を生成しました")
	}
	
	// 5段階生成完了後、問題をproblemsテーブルに保存（会話履歴も含む）
	fmt.Printf("💾 [FiveStage] Saving generated problem to database\n")
	
	// 会話履歴のnullチェック
	if conversationHistory == nil {
		fmt.Printf("⚠️ [FiveStage] Conversation history is nil, initializing empty history\n")
		conversationHistory = &models.ConversationHistory{
			Messages: make([]models.ConversationMessage, 0),
		}
	}
	
	fmt.Printf("💬 [FiveStage] Conversation history messages: %d\n", len(conversationHistory.Messages))
	
	// 正しいマッピング:
	// - Content: Stage 4の完全な問題文
	// - Solution: Stage 5の完全な解答・解説
	// - ImageBase64: Stage 3の図形画像
	problem := &models.Problem{
		UserID:              user.ID,
		Subject:             req.Subject,
		Prompt:              req.Prompt,
		Content:             stage4Resp.FinalExplanation,     // Stage 4: 完全な問題文
		Solution:            stage5Resp.GeometryCode,         // Stage 5: 完全な解答・解説
		ImageBase64:         stage3Resp.CalculationResults,   // Stage 3: 図形画像（Base64）
		OpinionProfile:      req.OpinionProfile,              // レガシー（後方互換性）
		OpinionProfileV2:    req.OpinionProfileV2,            // 新基準（Ver.2）
		ConversationHistory: conversationHistory,             // 5段階生成プロセスの会話履歴
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// リポジトリが実装されている場合のみ保存
	if s.problemRepo != nil {
		if err := s.problemRepo.Create(ctx, problem); err != nil {
			fmt.Printf("⚠️ [FiveStage] Failed to save problem to database: %v\n", err)
			// データベース保存に失敗してもレスポンスは成功として返す（問題生成自体は成功）
		} else {
			fmt.Printf("✅ [FiveStage] Problem saved to database with ID: %d\n", problem.ID)
		}
	} else {
		fmt.Printf("⚠️ [FiveStage] Problem repository is not initialized, skipping database save\n")
	}
	
	fmt.Printf("✅ [FiveStage] NEW Five-stage problem generation (CONVERSATION MODE) completed successfully\n")
	fmt.Printf("💬 [FiveStage] Total conversation messages: %d\n", len(conversationHistory.Messages))
	
	// レスポンスの正しいマッピング:
	// - complete_problem: Stage 4の完全な問題文
	// - final_explanation: Stage 5の完全な解答・解説
	// - image_base64: Stage 3の図形画像
	// - conversation_history: 5段階生成プロセスの会話履歴
	return &models.FiveStageGenerationResponse{
		Success:               true,
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,        // Stage 1: 小問構成と解答プロセス
		CalculationProgram:    stage2Resp.CompleteProblem,              // Stage 2: 検証済みパラメータ（数値計算結果）
		CalculationResults:    stage2Resp.CompleteProblem,              // Stage 2: 検証済みパラメータ（数値計算結果）
		GeometryCode:          stage3Resp.CalculationProgram,           // Stage 3: 図形コード
		CompleteProblem:       stage4Resp.FinalExplanation,             // Stage 4: 完全な問題文
		FinalExplanation:      stage5Resp.GeometryCode,                 // Stage 5: 完全な解答・解説
		ImageBase64:           stage3Resp.CalculationResults,           // Stage 3: 図形画像（Base64）
		ConversationHistory:   conversationHistory,                     // 会話履歴
		Stage1Log:             stage1Resp.Log,
		Stage2Log:             stage2Resp.Log,
		Stage3Log:             stage3Resp.Log,
		Stage4Log:             stage4Resp.Log,
		Stage5Log:             stage5Resp.Log,
	}, nil
}

// GenerateStage1 1段階目：小問構成と解答プロセス生成（新しいプロセス）
func (s *problemService) GenerateStage1(ctx context.Context, req models.Stage1Request, userSchoolCode string) (*models.Stage1Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage1] 1段階目を開始：小問構成と解答プロセス生成 (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	// 個別呼び出し時のみカウント処理を実行（FiveStage全体呼び出し時は既にカウント済み）
	// リクエストにSkipCountフラグがない場合のみカウント
	if !req.SkipCount {
		logBuilder.WriteString(fmt.Sprintf("🔢 現在の生成回数: %d/%d\n", user.ProblemGenerationCount, user.ProblemGenerationLimit))
		
		// 生成制限チェック（-1は制限なし）
		if user.ProblemGenerationLimit >= 0 && user.ProblemGenerationCount >= user.ProblemGenerationLimit {
			errorMsg := fmt.Sprintf("問題生成回数の上限（%d回）に達しました", user.ProblemGenerationLimit)
			logBuilder.WriteString(fmt.Sprintf("🚫 %s\n", errorMsg))
			return &models.Stage1Response{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, fmt.Errorf(errorMsg)
		}
		
		// 問題生成回数を更新
		oldCount := user.ProblemGenerationCount
		user.ProblemGenerationCount++
		user.UpdatedAt = time.Now()
		
		logBuilder.WriteString(fmt.Sprintf("📝 生成回数を更新: %d → %d\n", oldCount, user.ProblemGenerationCount))
		
		if err := s.userRepo.Update(ctx, user); err != nil {
			errorMsg := fmt.Sprintf("問題生成カウントの更新に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.Stage1Response{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, fmt.Errorf(errorMsg)
		}
		
		logBuilder.WriteString(fmt.Sprintf("✅ 問題生成カウントを更新: %s = %d/%d\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit))
	} else {
		logBuilder.WriteString("ℹ️ カウント処理をスキップ（FiveStage全体呼び出し時）\n")
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 1段階目は非推奨（会話形式を使用してください）
	logBuilder.WriteString("⚠️ 個別ステージの呼び出しは非推奨です\n")
	
	return &models.Stage1Response{
		Success: false,
		Error:   "個別ステージの呼び出しは非推奨です。GenerateProblemFiveStageを使用してください。",
		Log:     logBuilder.String(),
	}, fmt.Errorf("deprecated: use GenerateProblemFiveStage instead")
}

// GenerateStage1WithHistory 1段階目：小問構成と解答プロセス生成（会話履歴付き）
func (s *problemService) GenerateStage1WithHistory(ctx context.Context, req models.Stage1Request, userSchoolCode string, history *models.ConversationHistory) (*models.Stage1Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage1-Chat] 1段階目を開始：小問構成と解答プロセス生成（会話形式） (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	// 個別呼び出し時のみカウント処理を実行
	if !req.SkipCount {
		logBuilder.WriteString(fmt.Sprintf("🔢 現在の生成回数: %d/%d\n", user.ProblemGenerationCount, user.ProblemGenerationLimit))
		
		if user.ProblemGenerationLimit >= 0 && user.ProblemGenerationCount >= user.ProblemGenerationLimit {
			errorMsg := fmt.Sprintf("問題生成回数の上限（%d回）に達しました", user.ProblemGenerationLimit)
			logBuilder.WriteString(fmt.Sprintf("🚫 %s\n", errorMsg))
			return &models.Stage1Response{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, fmt.Errorf(errorMsg)
		}
		
		oldCount := user.ProblemGenerationCount
		user.ProblemGenerationCount++
		user.UpdatedAt = time.Now()
		
		logBuilder.WriteString(fmt.Sprintf("📝 生成回数を更新: %d → %d\n", oldCount, user.ProblemGenerationCount))
		
		if err := s.userRepo.Update(ctx, user); err != nil {
			errorMsg := fmt.Sprintf("問題生成カウントの更新に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.Stage1Response{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, fmt.Errorf(errorMsg)
		}
		
		logBuilder.WriteString(fmt.Sprintf("✅ 問題生成カウントを更新: %s = %d/%d\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit))
	} else {
		logBuilder.WriteString("ℹ️ カウント処理をスキップ（FiveStage全体呼び出し時）\n")
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// OpinionProfileをJSON文字列に変換
	var opinionProfileStr string
	if req.OpinionProfile != nil {
		opinionProfileStr = fmt.Sprintf("%+v", req.OpinionProfile)
	}
	
	// Stage 1では詳細な初期プロンプトを作成（全ステージの指示を含む）
	prompt := s.createFiveStageInitialPrompt(req.Prompt, req.Subject, opinionProfileStr)
	logBuilder.WriteString("📝 5段階生成の初期プロンプト（全ステージの指示を含む）を作成しました\n")
	
	// 会話履歴にユーザーメッセージを追加
	s.buildConversationHistory(history, prompt, "", 1)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴を更新（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	// 会話履歴を使用してAPI呼び出し
	var content string
	clientMessages := s.convertToClientMessages(history)
	
	// 会話履歴の内容をログ出力（デバッグ用）
	fmt.Printf("📋 [Stage1-Chat] Conversation history details:\n")
	for i, msg := range clientMessages {
		contentPreview := msg.Content
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		fmt.Printf("  Message %d: role=%s, content_length=%d, preview=%s\n", i+1, msg.Role, len(msg.Content), contentPreview)
	}
	
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		fmt.Printf("🔄 [Stage1-Chat] Calling OpenAI API...\n")
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
		fmt.Printf("🔄 [Stage1-Chat] OpenAI API call completed, err=%v\n", err)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		fmt.Printf("🔄 [Stage1-Chat] Calling Google API...\n")
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
		fmt.Printf("🔄 [Stage1-Chat] Google API call completed, err=%v\n", err)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		fmt.Printf("🔄 [Stage1-Chat] Calling Claude API with %d messages...\n", len(clientMessages))
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
		fmt.Printf("🔄 [Stage1-Chat] Claude API call completed, err=%v, content_length=%d\n", err, len(content))
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		fmt.Printf("❌ [Stage1-Chat] %s\n", errorMsg)
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	if err != nil {
		errorMsg := fmt.Sprintf("%s APIでの小問構成と解答プロセス生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		fmt.Printf("❌ [Stage1-Chat] API Error: %v\n", err)
		fmt.Printf("❌ [Stage1-Chat] Error type: %T\n", err)
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 会話履歴にアシスタントメッセージを追加
	s.buildConversationHistory(history, "", content, 1)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴にアシスタントの応答を追加（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	// 小問構成と解答プロセスを抽出
	subProblemsAndProcess := s.extractSubProblemsAndProcess(content)
	if subProblemsAndProcess == "" {
		subProblemsAndProcess = strings.TrimSpace(content)
	}
	
	if subProblemsAndProcess == "" {
		errorMsg := "小問構成と解答プロセスの抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📝 小問構成と解答プロセスを抽出しました (長さ: %d文字)\n", len(subProblemsAndProcess)))
	logBuilder.WriteString("✅ [Stage1-Chat] 1段階目（会話形式）が完了しました\n")
	
	return &models.Stage1Response{
		Success:               true,
		SubProblemsAndProcess: subProblemsAndProcess,
		Log:                   logBuilder.String(),
	}, nil
}

// GenerateStage2WithHistory 2段階目：パラメータ設定と動的検証（数値計算）（会話履歴付き）
func (s *problemService) GenerateStage2WithHistory(ctx context.Context, req models.Stage2Request, userSchoolCode string, history *models.ConversationHistory) (*models.Stage2Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage2-Chat] 2段階目を開始：パラメータ設定と動的検証（会話形式） (ユーザー: %s)\n", userSchoolCode))
	
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage2Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// Stage 2以降はトリガーのみ送信
	prompt := s.loadStageTrigger()
	logBuilder.WriteString("📝 Stage 2トリガーを送信（次のステージに進む）\n")
	
	s.buildConversationHistory(history, prompt, "", 2)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴を更新（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	var content string
	clientMessages := s.convertToClientMessages(history)
	
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage2Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	if err != nil {
		errorMsg := fmt.Sprintf("%s APIでのパラメータ設定と動的検証に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage2Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	s.buildConversationHistory(history, "", content, 2)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴にアシスタントの応答を追加（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	// Stage 2では数値計算プログラムと結果を抽出
	calculationProgram := s.extractCalculationProgram(content)
	logBuilder.WriteString(fmt.Sprintf("🧮 計算プログラムの抽出: %t (長さ: %d文字)\n", calculationProgram != "", len(calculationProgram)))
	
	var calculationResults string
	if calculationProgram != "" {
		logBuilder.WriteString("🧮 数値計算プログラムを実行中...\n")
		calculationResults, err = s.executeCalculationProgram(ctx, calculationProgram)
		if err != nil {
			logBuilder.WriteString(fmt.Sprintf("⚠️ 数値計算の実行に失敗: %v\n", err))
			calculationResults = fmt.Sprintf("計算実行エラー: %v", err)
		} else {
			logBuilder.WriteString("✅ 数値計算を実行しました\n")
		}
	}
	
	// CompleteProblemフィールドに計算結果を格納（後方互換性のため）
	completeProblem := calculationResults
	if completeProblem == "" {
		completeProblem = strings.TrimSpace(content)
	}
	logBuilder.WriteString("✅ [Stage2-Chat] 2段階目（会話形式）が完了しました\n")
	
	return &models.Stage2Response{
		Success:         true,
		CompleteProblem: completeProblem,
		Log:             logBuilder.String(),
	}, nil
}

// GenerateStage4WithHistory 4段階目：完全な問題文の生成（会話履歴付き）
func (s *problemService) GenerateStage4WithHistory(ctx context.Context, req models.Stage4Request, userSchoolCode string, history *models.ConversationHistory) (*models.Stage4Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage4-Chat] 4段階目を開始：完全な問題文の生成（会話形式） (ユーザー: %s)\n", userSchoolCode))
	
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// Stage 4もトリガーのみ送信
	prompt := s.loadStageTrigger()
	logBuilder.WriteString("📝 Stage 4トリガーを送信（次のステージに進む）\n")
	
	s.buildConversationHistory(history, prompt, "", 4)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴を更新（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	var content string
	clientMessages := s.convertToClientMessages(history)
	
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	if err != nil {
		errorMsg := fmt.Sprintf("%s APIでの完全な問題文の生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	s.buildConversationHistory(history, "", content, 4)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴にアシスタントの応答を追加（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	completeProblem := s.extractCompleteProblem(content)
	if completeProblem == "" {
		completeProblem = strings.TrimSpace(content)
	}
	
	if completeProblem == "" {
		errorMsg := "完全な問題文の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📝 完全な問題文を抽出しました (長さ: %d文字)\n", len(completeProblem)))
	logBuilder.WriteString("✅ [Stage4-Chat] 4段階目（会話形式）が完了しました\n")
	
	// FinalExplanationフィールドに完全な問題文を格納（フィールド名は後方互換性のため変更しない）
	return &models.Stage4Response{
		Success:          true,
		FinalExplanation: completeProblem, // 実際は「完全な問題文」だが、フィールド名は変更しない
		Log:              logBuilder.String(),
	}, nil
}

// GenerateStage5WithHistory 5段階目：完全な解答・解説の生成（会話履歴付き）
func (s *problemService) GenerateStage5WithHistory(ctx context.Context, req models.Stage5Request, userSchoolCode string, history *models.ConversationHistory) (*models.Stage5Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage5-Chat] 5段階目を開始：完全な解答・解説の生成（会話形式） (ユーザー: %s)\n", userSchoolCode))
	
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage5Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// Stage 5もトリガーのみ送信
	prompt := s.loadStageTrigger()
	logBuilder.WriteString("📝 Stage 5トリガーを送信（次のステージに進む）\n")
	
	s.buildConversationHistory(history, prompt, "", 5)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴を更新（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	var content string
	clientMessages := s.convertToClientMessages(history)
	
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage5Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	if err != nil {
		errorMsg := fmt.Sprintf("%s APIでの完全な解答・解説生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage5Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	s.buildConversationHistory(history, "", content, 5)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴にアシスタントの応答を追加（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	finalExplanation := s.extractFinalSolution(content)
	if finalExplanation == "" {
		finalExplanation = strings.TrimSpace(content)
	}
	
	if finalExplanation == "" {
		errorMsg := "完全な解答・解説の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage5Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📚 完全な解答・解説を抽出しました (長さ: %d文字)\n", len(finalExplanation)))
	logBuilder.WriteString("✅ [Stage5-Chat] 5段階目（会話形式）が完了しました\n")
	
	// GeometryCodeフィールドに解答・解説を格納（フィールド名は後方互換性のため変更しない）
	return &models.Stage5Response{
		Success:      true,
		GeometryCode: finalExplanation, // 実際は「完全な解答・解説」だが、フィールド名は変更しない
		ImageBase64:  "",
		Log:          logBuilder.String(),
	}, nil
}

// GenerateStage3WithHistory 3段階目：問題文用の図形描画（会話履歴付き）
func (s *problemService) GenerateStage3WithHistory(ctx context.Context, req models.Stage3Request, userSchoolCode string, history *models.ConversationHistory) (*models.Stage3Response, error) {
	fmt.Printf("🎨 [Stage3-Chat] ===== STAGE 3 STARTED ===== (user: %s)\n", userSchoolCode)
	fmt.Printf("🎨 [Stage3-Chat] History messages count: %d\n", len(history.Messages))
	
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage3-Chat] 3段階目を開始：問題文用の図形描画（会話形式） (ユーザー: %s)\n", userSchoolCode))
	
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage3Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// Stage 3もトリガーのみ送信
	prompt := s.loadStageTrigger()
	logBuilder.WriteString("📝 Stage 3トリガーを送信（次のステージに進む）\n")
	
	s.buildConversationHistory(history, prompt, "", 3)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴を更新（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	var content string
	clientMessages := s.convertToClientMessages(history)
	
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage3Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	if err != nil {
		logBuilder.WriteString(fmt.Sprintf("⚠️ AIによる図形コード生成に失敗: %v\n", err))
		logBuilder.WriteString("ℹ️ この問題は図形なしで続行します\n")
		logBuilder.WriteString("✅ [Stage3-Chat] 3段階目（会話形式）が完了しました（図形なし）\n")
		
		return &models.Stage3Response{
			Success:            true,
			CalculationProgram: "",
			CalculationResults: "",
			Log:                logBuilder.String(),
		}, nil
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	s.buildConversationHistory(history, "", content, 3)
	logBuilder.WriteString(fmt.Sprintf("💬 会話履歴にアシスタントの応答を追加（現在のメッセージ数: %d）\n", len(history.Messages)))
	
	geometryCode := s.extractPythonCode(content)
	fmt.Printf("🐍 [Stage3-Chat] Geometry code extracted: %t (length: %d)\n", geometryCode != "", len(geometryCode))
	logBuilder.WriteString(fmt.Sprintf("🐍 図形コードの抽出: %t (長さ: %d文字)\n", geometryCode != "", len(geometryCode)))
	
	var imageBase64 string
	if geometryCode != "" {
		fmt.Printf("🎨 [Stage3-Chat] Generating geometry with code...\n")
		logBuilder.WriteString("🎨 図形を生成中...\n")
		imageBase64, err = s.coreClient.GenerateCustomGeometry(ctx, geometryCode, req.CompleteProblem)
		if err != nil {
			fmt.Printf("❌ [Stage3-Chat] Geometry generation failed: %v\n", err)
			logBuilder.WriteString(fmt.Sprintf("⚠️ 図形生成に失敗: %v\n", err))
		} else {
			fmt.Printf("✅ [Stage3-Chat] Geometry generated successfully (length: %d)\n", len(imageBase64))
			logBuilder.WriteString("✅ 図形を生成しました\n")
		}
	} else {
		fmt.Printf("ℹ️ [Stage3-Chat] No geometry code found, skipping geometry generation\n")
		logBuilder.WriteString("ℹ️ この問題には図形は必要ありません\n")
	}
	
	fmt.Printf("🖼️ [Stage3-Chat] Final image base64 length: %d\n", len(imageBase64))
	logBuilder.WriteString(fmt.Sprintf("🖼️ 最終的な図形データの長さ: %d\n", len(imageBase64)))
	logBuilder.WriteString("✅ [Stage3-Chat] 3段階目（会話形式）が完了しました\n")
	
	// CalculationProgramフィールドに図形コードを、CalculationResultsに画像を格納（後方互換性のため）
	return &models.Stage3Response{
		Success:            true,
		CalculationProgram: geometryCode,
		CalculationResults: imageBase64,
		Log:                logBuilder.String(),
	}, nil
}


// GenerateStage2 2段階目：完全な問題生成（新しいプロセス）
func (s *problemService) GenerateStage2(ctx context.Context, req models.Stage2Request, userSchoolCode string) (*models.Stage2Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage2] 2段階目を開始：完全な問題生成 (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage2Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 2段階目は非推奨（会話形式を使用してください）
	logBuilder.WriteString("⚠️ 個別ステージの呼び出しは非推奨です\n")
	
	return &models.Stage2Response{
		Success: false,
		Error:   "個別ステージの呼び出しは非推奨です。GenerateProblemFiveStageを使用してください。",
		Log:     logBuilder.String(),
	}, fmt.Errorf("deprecated: use GenerateProblemFiveStage instead")
}

// createStage2Prompt 2段階目用のプロンプト（図形生成専用）
func (s *problemService) createStage2Prompt(problemText string) string {
	return s.createGeometryPromptWithSamples(problemText)
}

// GenerateStage3 3段階目：数値計算プログラム生成・実行（新しいプロセス）
func (s *problemService) GenerateStage3(ctx context.Context, req models.Stage3Request, userSchoolCode string) (*models.Stage3Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage3] 3段階目を開始：数値計算プログラム生成・実行 (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage3Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 3段階目は非推奨（会話形式を使用してください）
	logBuilder.WriteString("⚠️ 個別ステージの呼び出しは非推奨です\n")
	
	return &models.Stage3Response{
		Success: false,
		Error:   "個別ステージの呼び出しは非推奨です。GenerateProblemFiveStageを使用してください。",
		Log:     logBuilder.String(),
	}, fmt.Errorf("deprecated: use GenerateProblemFiveStage instead")
}

// buildConversationHistory 会話履歴を構築
func (s *problemService) buildConversationHistory(history *models.ConversationHistory, userMessage, assistantMessage string, stage int) {
	if history == nil {
		return
	}
	
	// ユーザーメッセージを追加
	if userMessage != "" {
		history.Messages = append(history.Messages, models.ConversationMessage{
			Role:    "user",
			Content: userMessage,
			Stage:   stage,
		})
	}
	
	// アシスタントメッセージを追加
	if assistantMessage != "" {
		history.Messages = append(history.Messages, models.ConversationMessage{
			Role:    "assistant",
			Content: assistantMessage,
			Stage:   stage,
		})
	}
}

// convertToClientMessages 会話履歴をクライアント用メッセージに変換
func (s *problemService) convertToClientMessages(history *models.ConversationHistory) []clients.ChatMessage {
	if history == nil || len(history.Messages) == 0 {
		return nil
	}
	
	messages := make([]clients.ChatMessage, 0, len(history.Messages))
	for _, msg := range history.Messages {
		messages = append(messages, clients.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return messages
}

// 3問生成システムの実装（15段階プロセス）

// GenerateThreeProblems 3問生成プロセス全体を実行
func (s *problemService) GenerateThreeProblems(ctx context.Context, req models.ThreeProblemGenerationRequest, userSchoolCode string) (*models.ThreeProblemGenerationResponse, error) {
	return s.GenerateThreeProblemsWithProgress(ctx, req, userSchoolCode, nil)
}

// GenerateThreeProblemsWithProgress 3問生成プロセス全体を実行（進捗コールバック付き）
func (s *problemService) GenerateThreeProblemsWithProgress(ctx context.Context, req models.ThreeProblemGenerationRequest, userSchoolCode string, progressCallback func(stage int, message string)) (*models.ThreeProblemGenerationResponse, error) {
	fmt.Printf("🚀 [ThreeProblems] Starting three-problem generation (15-stage process) for user: %s\n", userSchoolCode)
	fmt.Printf("🔍 [ThreeProblems] Uploaded problem content length: %d, PDF data length: %d\n", len(req.UploadedProblemContent), len(req.UploadedProblemPDF))
	
	// 1. ユーザー情報を取得して生成制限をチェック
	fmt.Printf("📋 [ThreeProblems] Fetching user info for: %s\n", userSchoolCode)
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		fmt.Printf("❌ [ThreeProblems] Failed to get user info: %v\n", err)
		return &models.ThreeProblemGenerationResponse{
			Success: false,
			Error:   fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err),
		}, nil
	}
	
	// 1.5. PDFデータがある場合は、Google Files APIを使用してテキストを抽出
	var uploadedProblemContent string
	if len(req.UploadedProblemPDF) > 0 {
		fmt.Printf("📄 [ThreeProblems] PDF data detected, using Google Files API for extraction\n")
		
		// Google Clientを使用してPDFを処理
		if user.PreferredAPI != "google" && user.PreferredAPI != "gemini" {
			fmt.Printf("⚠️ [ThreeProblems] PDF upload requires Google API, but user prefers %s\n", user.PreferredAPI)
			return &models.ThreeProblemGenerationResponse{
				Success: false,
				Error:   "PDFファイルのアップロードにはGoogle API（Gemini）が必要です。設定ページでAPIをGoogleに変更してください。",
			}, nil
		}
		
		// PDFからテキストを抽出するための簡単なプロンプト
		extractPrompt := "このPDFファイルに含まれる問題文を正確に抽出してください。数式、図形の説明、問題番号などすべての情報を含めてください。"
		
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		extractedText, err := dynamicClient.GenerateContentWithPDF(ctx, extractPrompt, req.UploadedProblemPDF)
		if err != nil {
			fmt.Printf("❌ [ThreeProblems] Failed to extract text from PDF: %v\n", err)
			return &models.ThreeProblemGenerationResponse{
				Success: false,
				Error:   fmt.Sprintf("PDFからのテキスト抽出に失敗しました: %v", err),
			}, nil
		}
		
		uploadedProblemContent = extractedText
		fmt.Printf("✅ [ThreeProblems] Successfully extracted text from PDF (length: %d)\n", len(uploadedProblemContent))
	} else {
		// テキストコンテンツを使用
		uploadedProblemContent = req.UploadedProblemContent
		fmt.Printf("📝 [ThreeProblems] Using text content (length: %d)\n", len(uploadedProblemContent))
	}
	
	fmt.Printf("👤 [ThreeProblems] User found: ID=%d, SchoolCode=%s\n", user.ID, user.SchoolCode)
	fmt.Printf("🔢 [ThreeProblems] Current generation count: %d (limit: %d)\n", user.ProblemGenerationCount, user.ProblemGenerationLimit)
	
	// 2. 生成制限チェック（-1は制限なし）
	if user.ProblemGenerationLimit >= 0 && user.ProblemGenerationCount >= user.ProblemGenerationLimit {
		fmt.Printf("🚫 [ThreeProblems] Generation limit reached: %d/%d\n", user.ProblemGenerationCount, user.ProblemGenerationLimit)
		return &models.ThreeProblemGenerationResponse{
			Success: false,
			Error:   fmt.Sprintf("問題生成回数の上限（%d回）に達しました", user.ProblemGenerationLimit),
		}, nil
	}
	
	// 3. 問題生成回数を更新（3問で1回とカウント）
	oldCount := user.ProblemGenerationCount
	user.ProblemGenerationCount++
	user.UpdatedAt = time.Now()
	
	fmt.Printf("📝 [ThreeProblems] Updating generation count from %d to %d\n", oldCount, user.ProblemGenerationCount)
	
	if err := s.userRepo.Update(ctx, user); err != nil {
		fmt.Printf("❌ [ThreeProblems] Failed to update user generation count: %v\n", err)
		return &models.ThreeProblemGenerationResponse{
			Success: false,
			Error:   fmt.Sprintf("問題生成カウントの更新に失敗しました: %v", err),
		}, nil
	}
	
	fmt.Printf("✅ [ThreeProblems] Successfully updated generation count: %s = %d/%d\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit)
	
	// 4. 大ステップ会話履歴を初期化
	mainHistory := &models.ConversationHistory{
		Messages: make([]models.ConversationMessage, 0),
	}
	fmt.Printf("💬 [ThreeProblems] Initialized main conversation history\n")
	
	// 5. パターンA生成（Stage 1-5）
	fmt.Printf("🎯 [ThreeProblems] Starting Pattern A generation (stages 1-5)\n")
	patternA, err := s.generateSinglePattern(ctx, "パターンA: 数値だけ違う", uploadedProblemContent, userSchoolCode, mainHistory, progressCallback, 1)
	if err != nil || !patternA.Success {
		errorMsg := fmt.Sprintf("パターンA生成に失敗しました: %v", err)
		if patternA != nil && patternA.Error != "" {
			errorMsg = patternA.Error
		}
		return &models.ThreeProblemGenerationResponse{
			Success:                 false,
			Error:                   errorMsg,
			PatternA:                *patternA,
			MainConversationHistory: mainHistory,
		}, nil
	}
	fmt.Printf("✅ [ThreeProblems] Pattern A generation completed\n")
	
	// 6. パターンB生成（Stage 6-10）
	fmt.Printf("🎯 [ThreeProblems] Starting Pattern B generation (stages 6-10)\n")
	patternB, err := s.generateSinglePattern(ctx, "パターンB: 必要な公式は同じだが、問題自体は違う", uploadedProblemContent, userSchoolCode, mainHistory, progressCallback, 6)
	if err != nil || !patternB.Success {
		errorMsg := fmt.Sprintf("パターンB生成に失敗しました: %v", err)
		if patternB != nil && patternB.Error != "" {
			errorMsg = patternB.Error
		}
		return &models.ThreeProblemGenerationResponse{
			Success:                 false,
			Error:                   errorMsg,
			PatternA:                *patternA,
			PatternB:                *patternB,
			MainConversationHistory: mainHistory,
		}, nil
	}
	fmt.Printf("✅ [ThreeProblems] Pattern B generation completed\n")
	
	// 7. パターンC生成（Stage 11-15）
	fmt.Printf("🎯 [ThreeProblems] Starting Pattern C generation (stages 11-15)\n")
	patternC, err := s.generateSinglePattern(ctx, "パターンC: 全体的な構成は同じだが、問われている部分が違う", uploadedProblemContent, userSchoolCode, mainHistory, progressCallback, 11)
	if err != nil || !patternC.Success {
		errorMsg := fmt.Sprintf("パターンC生成に失敗しました: %v", err)
		if patternC != nil && patternC.Error != "" {
			errorMsg = patternC.Error
		}
		return &models.ThreeProblemGenerationResponse{
			Success:                 false,
			Error:                   errorMsg,
			PatternA:                *patternA,
			PatternB:                *patternB,
			PatternC:                *patternC,
			MainConversationHistory: mainHistory,
		}, nil
	}
	fmt.Printf("✅ [ThreeProblems] Pattern C generation completed\n")
	
	// 8. データベースに3問を保存
	fmt.Printf("💾 [ThreeProblems] Saving 3 problems to database\n")
	
	// パターンA保存
	problemA := &models.Problem{
		UserID:              user.ID,
		Subject:             req.Subject,
		Prompt:              "パターンA: 数値だけ違う（元問題参照）",
		Content:             patternA.Content,
		Solution:            patternA.Solution,
		ImageBase64:         patternA.ImageBase64,
		ConversationHistory: patternA.ConversationHistory,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	
	if s.problemRepo != nil {
		if err := s.problemRepo.Create(ctx, problemA); err != nil {
			fmt.Printf("⚠️ [ThreeProblems] Failed to save Pattern A: %v\n", err)
		} else {
			fmt.Printf("✅ [ThreeProblems] Pattern A saved with ID: %d\n", problemA.ID)
		}
	}
	
	// パターンB保存
	problemB := &models.Problem{
		UserID:              user.ID,
		Subject:             req.Subject,
		Prompt:              "パターンB: 必要な公式は同じだが、問題自体は違う（元問題参照）",
		Content:             patternB.Content,
		Solution:            patternB.Solution,
		ImageBase64:         patternB.ImageBase64,
		ConversationHistory: patternB.ConversationHistory,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	
	if s.problemRepo != nil {
		if err := s.problemRepo.Create(ctx, problemB); err != nil {
			fmt.Printf("⚠️ [ThreeProblems] Failed to save Pattern B: %v\n", err)
		} else {
			fmt.Printf("✅ [ThreeProblems] Pattern B saved with ID: %d\n", problemB.ID)
		}
	}
	
	// パターンC保存
	problemC := &models.Problem{
		UserID:              user.ID,
		Subject:             req.Subject,
		Prompt:              "パターンC: 全体的な構成は同じだが、問われている部分が違う（元問題参照）",
		Content:             patternC.Content,
		Solution:            patternC.Solution,
		ImageBase64:         patternC.ImageBase64,
		ConversationHistory: patternC.ConversationHistory,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	
	if s.problemRepo != nil {
		if err := s.problemRepo.Create(ctx, problemC); err != nil {
			fmt.Printf("⚠️ [ThreeProblems] Failed to save Pattern C: %v\n", err)
		} else {
			fmt.Printf("✅ [ThreeProblems] Pattern C saved with ID: %d\n", problemC.ID)
		}
	}
	
	fmt.Printf("✅ [ThreeProblems] Three-problem generation (15-stage process) completed successfully\n")
	
	return &models.ThreeProblemGenerationResponse{
		Success:                 true,
		PatternA:                *patternA,
		PatternB:                *patternB,
		PatternC:                *patternC,
		MainConversationHistory: mainHistory,
		Log:                     "3問生成が完了しました",
	}, nil
}

// generateSinglePattern 1つのパターンを5段階プロセスで生成
func (s *problemService) generateSinglePattern(
	ctx context.Context,
	patternName string,
	uploadedProblemContent string,
	userSchoolCode string,
	mainHistory *models.ConversationHistory,
	progressCallback func(stage int, message string),
	baseStage int, // 1, 6, 11
) (*models.PatternResult, error) {
	fmt.Printf("🎨 [Pattern] Starting pattern generation: %s (base stage: %d)\n", patternName, baseStage)
	
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Pattern] %s の生成を開始\n", patternName))
	
	// 1. 小ステップ会話履歴を初期化
	patternHistory := &models.ConversationHistory{
		Messages: make([]models.ConversationMessage, 0),
	}
	fmt.Printf("💬 [Pattern] Initialized pattern conversation history\n")
	
	// 2. 初期プロンプトを作成
	initialPrompt, err := s.promptLoader.LoadThreeProblemGenerationPrompt(uploadedProblemContent, patternName)
	if err != nil {
		errorMsg := fmt.Sprintf("プロンプトの読み込みに失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.PatternResult{
			Success: false,
			Error:   errorMsg,
		}, err
	}
	
	// 初期プロンプトを会話履歴に追加
	s.buildConversationHistory(patternHistory, initialPrompt, "", 1)
	logBuilder.WriteString("📝 初期プロンプトを作成しました\n")
	
	result := &models.PatternResult{
		Success:             true,
		ConversationHistory: patternHistory,
	}
	
	// 3. Stage 1-5を順次実行
	for stage := 1; stage <= 5; stage++ {
		globalStage := baseStage + stage - 1
		fmt.Printf("🔄 [Pattern] Executing stage %d (global stage %d)\n", stage, globalStage)
		
		if progressCallback != nil {
			progressCallback(globalStage, fmt.Sprintf("%s - Stage %d を実行中", patternName, stage))
		}
		
		// ステージを実行
		stageResult, stageLog, err := s.executePatternStage(ctx, stage, patternHistory, userSchoolCode)
		if err != nil {
			errorMsg := fmt.Sprintf("Stage %d の実行に失敗しました: %v", stage, err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			result.Success = false
			result.Error = errorMsg
			return result, err
		}
		
		logBuilder.WriteString(stageLog)
		
		// 結果を格納
		switch stage {
		case 1:
			result.Stage1Result = stageResult
			result.Stage1Log = stageLog
		case 2:
			result.Stage2Result = stageResult
			result.Stage2Log = stageLog
		case 3:
			result.Stage3Result = stageResult
			result.Stage3Log = stageLog
			// Stage 3では図形を生成
			pythonCode := s.extractPythonCode(stageResult)
			if pythonCode != "" {
				fmt.Printf("🎨 [Pattern] Generating geometry from Python code\n")
				imageBase64, err := s.coreClient.GenerateCustomGeometry(ctx, pythonCode, uploadedProblemContent)
				if err != nil {
					fmt.Printf("⚠️ [Pattern] Geometry generation failed: %v\n", err)
					logBuilder.WriteString(fmt.Sprintf("⚠️ 図形生成に失敗: %v\n", err))
				} else {
					result.ImageBase64 = imageBase64
					fmt.Printf("✅ [Pattern] Geometry generated successfully\n")
					logBuilder.WriteString("✅ 図形を生成しました\n")
				}
			}
		case 4:
			result.Stage4Result = stageResult
			result.Stage4Log = stageLog
			result.Content = s.extractCompleteProblem(stageResult)
			if result.Content == "" {
				result.Content = strings.TrimSpace(stageResult)
			}
		case 5:
			result.Stage5Result = stageResult
			result.Stage5Log = stageLog
			result.Solution = s.extractFinalSolution(stageResult)
			if result.Solution == "" {
				result.Solution = strings.TrimSpace(stageResult)
			}
		}
		
		fmt.Printf("✅ [Pattern] Stage %d completed\n", stage)
	}
	
	fmt.Printf("✅ [Pattern] Pattern generation completed: %s\n", patternName)
	
	return result, nil
}

// executePatternStage パターンの1つのステージを実行
func (s *problemService) executePatternStage(
	ctx context.Context,
	stage int,
	history *models.ConversationHistory,
	userSchoolCode string,
) (string, string, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("🔄 [Stage%d] ステージ%dを開始\n", stage, stage))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return "", logBuilder.String(), err
	}
	
	// Stage 2以降はトリガーを送信
	if stage > 1 {
		trigger := s.loadStageTrigger()
		s.buildConversationHistory(history, trigger, "", stage)
		logBuilder.WriteString("📝 次のステージへのトリガーを送信しました\n")
	}
	
	// AI呼び出し
	clientMessages := s.convertToClientMessages(history)
	var content string
	
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateWithHistory(ctx, clientMessages)
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return "", logBuilder.String(), fmt.Errorf(errorMsg)
	}
	
	if err != nil {
		errorMsg := fmt.Sprintf("%s APIでの生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return "", logBuilder.String(), err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 会話履歴にアシスタントメッセージを追加
	s.buildConversationHistory(history, "", content, stage)
	logBuilder.WriteString("💬 会話履歴にアシスタントの応答を追加しました\n")
	
	logBuilder.WriteString(fmt.Sprintf("✅ [Stage%d] ステージ%dが完了しました\n", stage, stage))
	
	return content, logBuilder.String(), nil
}
