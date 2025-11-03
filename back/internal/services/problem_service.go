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
	GenerateStage1(ctx context.Context, req models.Stage1Request, userSchoolCode string) (*models.Stage1Response, error)
	GenerateStage2(ctx context.Context, req models.Stage2Request, userSchoolCode string) (*models.Stage2Response, error)
	GenerateStage3(ctx context.Context, req models.Stage3Request, userSchoolCode string) (*models.Stage3Response, error)
	GenerateStage4(ctx context.Context, req models.Stage4Request, userSchoolCode string) (*models.Stage4Response, error)
	GenerateStage5(ctx context.Context, req models.Stage5Request, userSchoolCode string) (*models.Stage5Response, error)
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
	enhancedPrompt := s.enhancePromptForGeometry(req.Prompt)
	fmt.Printf("🔍 Enhanced prompt: %s\n", enhancedPrompt)
	
	var content string
	switch preferredAPI {
	case "openai", "chatgpt":
		// ユーザーの設定に基づいて新しいクライアントを作成
		dynamicClient := clients.NewOpenAIClient(preferredModel)
		content, err = dynamicClient.GenerateContent(ctx, enhancedPrompt)
		if err != nil {
			return nil, fmt.Errorf("OpenAI APIでの問題生成に失敗しました: %w", err)
		}
	case "google", "gemini":
		// ユーザーの設定に基づいて新しいクライアントを作成
		dynamicClient := clients.NewGoogleClient(preferredModel)
		content, err = dynamicClient.GenerateContent(ctx, enhancedPrompt)
		if err != nil {
			return nil, fmt.Errorf("Google APIでの問題生成に失敗しました: %w", err)
		}
	case "claude", "laboratory":
		// ユーザーの設定に基づいて新しいクライアントを作成
		// laboratoryもClaudeとして扱う
		dynamicClient := clients.NewClaudeClient(preferredModel)
		content, err = dynamicClient.GenerateContent(ctx, enhancedPrompt)
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

// enhancePromptForGeometry enhances the prompt to include geometry generation instructions
func (s *problemService) enhancePromptForGeometry(prompt string) string {
	// 会話形式が要求されているかチェック
	isConversationRequested := s.isConversationFormatRequested(prompt)
	
	if isConversationRequested {
		fmt.Printf("💬 [ConversationFormat] Conversation format requested by user\n")
		return s.createConversationPrompt(prompt)
	} else {
		fmt.Printf("📝 [StandardFormat] Standard problem format will be used\n")
		return s.createStandardPrompt(prompt)
	}
}

// isConversationFormatRequested ユーザーのプロンプトに会話文形式の要求があるかチェック
func (s *problemService) isConversationFormatRequested(prompt string) bool {
	conversationKeywords := []string{
		"会話文", "会話形式", "登場人物", "やり取り", "対話", 
		"条件を抽出", "条件抽出", "会話から", "話し合い",
		"二人の", "2人の", "キャラクター", "人物",
	}
	
	promptLower := strings.ToLower(prompt)
	for _, keyword := range conversationKeywords {
		if strings.Contains(promptLower, keyword) {
			return true
		}
	}
	return false
}

// createConversationPrompt 会話文形式の問題生成プロンプトを作成
func (s *problemService) createConversationPrompt(prompt string) string {
	promptText, err := s.promptLoader.LoadConversationFormatPrompt(prompt)
	if err != nil {
		fmt.Printf("⚠️ Failed to load conversation format prompt: %v\n", err)
		// フォールバック：エラー時は基本プロンプトを返す
		return "会話形式プロンプトの読み込みに失敗しました: " + err.Error()
	}
	return promptText
}

// createStandardPrompt 通常の問題生成プロンプトを作成
func (s *problemService) createStandardPrompt(prompt string) string {
	promptText, err := s.promptLoader.LoadStandardFormatPrompt(prompt)
	if err != nil {
		fmt.Printf("⚠️ Failed to load standard format prompt: %v\n", err)
		// フォールバック：エラー時は基本プロンプトを返す
		return "標準形式プロンプトの読み込みに失敗しました: " + err.Error()
	}
	return promptText
}

// createStage1Prompt 1段階目用のプロンプトを作成（問題文のみ）
func (s *problemService) createStage1Prompt(userPrompt, subject string) string {
	promptText, err := s.promptLoader.LoadStage1PromptWithSamples(userPrompt, subject)
	if err != nil {
		fmt.Printf("⚠️ Failed to load stage1 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadStage1Prompt(userPrompt, subject)
		if err != nil {
			return "1段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createStage3Prompt 3段階目用のプロンプト（解答手順のみ）
func (s *problemService) createStage3Prompt(problemText, geometryCode string) string {
	promptText, err := s.promptLoader.LoadStage3PromptWithSamples(problemText, geometryCode)
	if err != nil {
		fmt.Printf("⚠️ Failed to load stage3 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadStage3Prompt(problemText, geometryCode)
		if err != nil {
			return "3段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createStage4Prompt 4段階目用のプロンプト（数値計算プログラム生成）
func (s *problemService) createStage4Prompt(problemText, solutionSteps string) string {
	promptText, err := s.promptLoader.LoadStage4PromptWithSamples(problemText, solutionSteps)
	if err != nil {
		fmt.Printf("⚠️ Failed to load stage4 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadStage4Prompt(problemText, solutionSteps)
		if err != nil {
			return "4段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createStage5Prompt 5段階目用のプロンプト（最終解説生成）
func (s *problemService) createStage5Prompt(problemText, solutionSteps, calculationResults string) string {
	promptText, err := s.promptLoader.LoadStage5PromptWithSamples(problemText, solutionSteps, calculationResults)
	if err != nil {
		fmt.Printf("⚠️ Failed to load stage5 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadStage5Prompt(problemText, solutionSteps, calculationResults)
		if err != nil {
			return "5段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createNewStage1Prompt 新しい1段階目用のプロンプト（解答プロセス生成）
func (s *problemService) createNewStage1Prompt(userPrompt, subject string) string {
	promptText, err := s.promptLoader.LoadNewStage1PromptWithSamples(userPrompt, subject)
	if err != nil {
		fmt.Printf("⚠️ Failed to load new stage1 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadNewStage1Prompt(userPrompt, subject)
		if err != nil {
			return "新しい1段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createNewStage2Prompt 新しい2段階目用のプロンプト（完全な問題生成）
func (s *problemService) createNewStage2Prompt(subProblemsAndProcess string) string {
	promptText, err := s.promptLoader.LoadNewStage2PromptWithSamples(subProblemsAndProcess)
	if err != nil {
		fmt.Printf("⚠️ Failed to load new stage2 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadNewStage2Prompt(subProblemsAndProcess)
		if err != nil {
			return "新しい2段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createNewStage3Prompt 新しい3段階目用のプロンプト（数値計算プログラム生成）
func (s *problemService) createNewStage3Prompt(solutionProcess string) string {
	promptText, err := s.promptLoader.LoadNewStage3PromptWithSamples(solutionProcess)
	if err != nil {
		fmt.Printf("⚠️ Failed to load new stage3 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadNewStage3Prompt(solutionProcess)
		if err != nil {
			return "新しい3段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createNewStage4Prompt 新しい4段階目用のプロンプト（問題文生成）
func (s *problemService) createNewStage4Prompt(solutionProcess string) string {
	promptText, err := s.promptLoader.LoadNewStage4PromptWithSamples(solutionProcess)
	if err != nil {
		fmt.Printf("⚠️ Failed to load new stage4 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadNewStage4Prompt(solutionProcess)
		if err != nil {
			return "新しい4段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
	}
	return promptText
}

// createNewStage5Prompt 新しい5段階目用のプロンプト（完全な解答・解説生成）
func (s *problemService) createNewStage5Prompt(solutionProcess, calculationResults string) string {
	promptText, err := s.promptLoader.LoadNewStage5PromptWithSamples(solutionProcess, calculationResults)
	if err != nil {
		fmt.Printf("⚠️ Failed to load new stage5 prompt with samples: %v\n", err)
		// フォールバック：サンプルなしでプロンプトを読み込み
		promptText, err = s.promptLoader.LoadNewStage5Prompt(solutionProcess, calculationResults)
		if err != nil {
			return "新しい5段階目プロンプトの読み込みに失敗しました: " + err.Error()
		}
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
	re := regexp.MustCompile(`(?s)---GEOMETRY_CODE_START---(.*?)---GEOMETRY_CODE_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		pythonCode := strings.TrimSpace(matches[1])
		// import文を除去
		pythonCode = s.removeImportStatements(pythonCode)
		return pythonCode
	}
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

// RegenerateGeometry 問題の図形を再生成
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

	// 問題生成時と同じフローを適用：AIで図形コード生成→実行
	fmt.Printf("🤖 [RegenerateGeometry] Generating matplotlib code with AI\n")
	
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
	
	// 4段階目用のプロンプトを作成（完全な解答・解説生成）
	prompt := s.createNewStage5Prompt(req.SubProblemsAndProcess, req.CalculationResults)
	logBuilder.WriteString("📝 4段階目用プロンプト（完全な解答・解説生成）を作成しました\n")
	
	// AIクライアントを選択してAPI呼び出し
	var content string
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
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
		errorMsg := fmt.Sprintf("%s APIでの完全な解答・解説生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 完全な解答を抽出
	completeAnswer := s.extractFinalSolution(content)
	if completeAnswer == "" {
		completeAnswer = strings.TrimSpace(content) // フォールバック：全体を完全な解答として使用
	}
	
	if completeAnswer == "" {
		errorMsg := "完全な解答・解説の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📚 完全な解答・解説を抽出しました (長さ: %d文字)\n", len(completeAnswer)))
	logBuilder.WriteString("✅ [Stage4] 4段階目（完全な解答・解説生成）が完了しました\n")
	
	return &models.Stage4Response{
		Success:        true,
		FinalExplanation: completeAnswer,
		Log:            logBuilder.String(),
	}, nil
}


// GenerateStage5 5段階目：図形描画プログラム生成（新しいプロセス）
func (s *problemService) GenerateStage5(ctx context.Context, req models.Stage5Request, userSchoolCode string) (*models.Stage5Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage5] 5段階目を開始：図形描画プログラム生成 (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
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
	
	// 5段階目用のプロンプトを作成（図形描画プログラム生成）
	prompt := s.createGeometryPromptWithSamples(req.CompleteProblem)
	logBuilder.WriteString("📝 5段階目用プロンプト（図形描画プログラム生成）を作成しました\n")
	
	// AIクライアントを選択してAPI呼び出し
	var content string
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
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
		logBuilder.WriteString(fmt.Sprintf("⚠️ AIによる図形コード生成に失敗: %v\n", err))
		// フォールバックとして図形なしで続行
		logBuilder.WriteString("ℹ️ この問題は図形なしで続行します\n")
		logBuilder.WriteString("✅ [Stage5] 5段階目が完了しました（図形なし）\n")
		
		return &models.Stage5Response{
			Success:      true,
			GeometryCode: "",
			ImageBase64:  "",
			Log:          logBuilder.String(),
		}, nil
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 図形コードを抽出
	geometryCode := s.extractPythonCode(content)
	logBuilder.WriteString(fmt.Sprintf("🐍 図形コードの抽出: %t (長さ: %d文字)\n", geometryCode != "", len(geometryCode)))
	
	// 図形を実際に生成
	var imageBase64 string
	if geometryCode != "" {
		logBuilder.WriteString("🎨 図形を生成中...\n")
		imageBase64, err = s.coreClient.GenerateCustomGeometry(ctx, geometryCode, req.CompleteProblem)
		if err != nil {
			logBuilder.WriteString(fmt.Sprintf("⚠️ 図形生成に失敗: %v\n", err))
		} else {
			logBuilder.WriteString("✅ 図形を生成しました\n")
		}
	} else {
		logBuilder.WriteString("ℹ️ この問題には図形は必要ありません\n")
	}
	
	logBuilder.WriteString(fmt.Sprintf("🖼️ 最終的な図形データの長さ: %d\n", len(imageBase64)))
	logBuilder.WriteString("✅ [Stage5] 5段階目（図形描画）が完了しました\n")
	
	return &models.Stage5Response{
		Success:      true,
		GeometryCode: geometryCode,
		ImageBase64:  imageBase64,
		Log:          logBuilder.String(),
	}, nil
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
	
	// フォールバック：【最終解答】を探す
	re = regexp.MustCompile(`(?s)【最終解答】(.*?)(?:---|\n\n|\z)`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return ""
}

// createThirdStagePrompt 3回目API呼び出し用のプロンプトを作成（解答手順と計算結果の統合）
func (s *problemService) createThirdStagePrompt(problemText, solutionSteps, calculationResults string) string {
	promptText, err := s.promptLoader.LoadStage5Prompt(problemText, solutionSteps, calculationResults)
	if err != nil {
		fmt.Printf("⚠️ Failed to load third stage prompt: %v\n", err)
		return "統合解説プロンプトの読み込みに失敗しました: " + err.Error()
	}
	return promptText
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

// GenerateProblemFiveStage 全体の5段階生成プロセスを実行（新しい順序）
func (s *problemService) GenerateProblemFiveStage(ctx context.Context, req models.FiveStageGenerationRequest, userSchoolCode string) (*models.FiveStageGenerationResponse, error) {
	fmt.Printf("🚀 [FiveStage] Starting NEW five-stage problem generation for user: %s\n", userSchoolCode)
	fmt.Printf("🔍 [FiveStage] Request details: Prompt length=%d, Subject=%s\n", len(req.Prompt), req.Subject)
	
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
	
	// 新しいプロセス：1段階目：小問構成と解答プロセス生成
	stage1Req := models.Stage1Request{
		Prompt:  req.Prompt,
		Subject: req.Subject,
	}
	stage1Resp, err := s.GenerateStage1(ctx, stage1Req, userSchoolCode)
	if err != nil || !stage1Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:   false,
			Error:     fmt.Sprintf("1段階目（小問構成と解答プロセス生成）に失敗しました: %v", err),
			Stage1Log: stage1Resp.Log,
		}, nil
	}
	
	// 新しいプロセス：2段階目：完全な問題生成
	stage2Req := models.Stage2Request{
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
	}
	stage2Resp, err := s.GenerateStage2(ctx, stage2Req, userSchoolCode)
	if err != nil || !stage2Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("2段階目（完全な問題生成）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
		}, nil
	}
	
	// 新しいプロセス：3段階目：数値計算プログラム生成・実行
	stage3Req := models.Stage3Request{
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
		CompleteProblem:       stage2Resp.CompleteProblem,
	}
	stage3Resp, err := s.GenerateStage3(ctx, stage3Req, userSchoolCode)
	if err != nil || !stage3Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("3段階目（数値計算プログラム生成・実行）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			CompleteProblem:       stage2Resp.CompleteProblem,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
			Stage3Log:             stage3Resp.Log,
		}, nil
	}
	
	// 新しいプロセス：4段階目：完全な解答・解説生成
	stage4Req := models.Stage4Request{
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
		CompleteProblem:       stage2Resp.CompleteProblem,
		CalculationResults:    stage3Resp.CalculationResults,
	}
	stage4Resp, err := s.GenerateStage4(ctx, stage4Req, userSchoolCode)
	if err != nil || !stage4Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("4段階目（完全な解答・解説生成）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			CompleteProblem:       stage2Resp.CompleteProblem,
			CalculationProgram:    stage3Resp.CalculationProgram,
			CalculationResults:    stage3Resp.CalculationResults,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
			Stage3Log:             stage3Resp.Log,
			Stage4Log:             stage4Resp.Log,
		}, nil
	}
	
	// 新しいプロセス：5段階目：図形描画プログラム生成
	stage5Req := models.Stage5Request{
		CompleteProblem: stage2Resp.CompleteProblem,
	}
	stage5Resp, err := s.GenerateStage5(ctx, stage5Req, userSchoolCode)
	if err != nil || !stage5Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:               false,
			Error:                 fmt.Sprintf("5段階目（図形描画）に失敗しました: %v", err),
			SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
			CompleteProblem:       stage2Resp.CompleteProblem,
			CalculationProgram:    stage3Resp.CalculationProgram,
			CalculationResults:    stage3Resp.CalculationResults,
			FinalExplanation: stage4Resp.FinalExplanation,
			Stage1Log:             stage1Resp.Log,
			Stage2Log:             stage2Resp.Log,
			Stage3Log:             stage3Resp.Log,
			Stage4Log:             stage4Resp.Log,
			Stage5Log:             stage5Resp.Log,
		}, nil
	}
	
	// 5段階生成完了後、問題をproblemsテーブルに保存
	fmt.Printf("💾 [FiveStage] Saving generated problem to database\n")
	
	problem := &models.Problem{
		UserID:           user.ID,
		Subject:          req.Subject,
		Prompt:           req.Prompt,
		Content:          stage2Resp.CompleteProblem,   // Stage2で生成された完全な問題
		Solution:         stage4Resp.FinalExplanation,   // Stage4で生成された完全な解答・解説
		ImageBase64:      stage5Resp.ImageBase64,      // Stage5で生成された図形
		OpinionProfile:   req.OpinionProfile,          // レガシー（後方互換性）
		OpinionProfileV2: req.OpinionProfileV2,        // 新基準（Ver.2）
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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
	
	fmt.Printf("✅ [FiveStage] NEW Five-stage problem generation completed successfully\n")
	
	return &models.FiveStageGenerationResponse{
		Success:               true,
		SubProblemsAndProcess: stage1Resp.SubProblemsAndProcess,
		CompleteProblem:       stage2Resp.CompleteProblem,
		CalculationProgram:    stage3Resp.CalculationProgram,
		CalculationResults:    stage3Resp.CalculationResults,
		FinalExplanation: stage4Resp.FinalExplanation,
		GeometryCode:          stage5Resp.GeometryCode,
		ImageBase64:           stage5Resp.ImageBase64,
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
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 1段階目用のプロンプトを作成（小問構成と解答プロセス生成）
	prompt := s.createNewStage1Prompt(req.Prompt, req.Subject)
	logBuilder.WriteString("📝 1段階目用プロンプト（小問構成と解答プロセス生成）を作成しました\n")
	
	// AIクライアントを選択してAPI呼び出し
	var content string
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	if err != nil {
		errorMsg := fmt.Sprintf("%s APIでの小問構成と解答プロセス生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 小問構成と解答プロセスを抽出
	subProblemsAndProcess := s.extractSubProblemsAndProcess(content)
	if subProblemsAndProcess == "" {
		subProblemsAndProcess = strings.TrimSpace(content) // フォールバック：全体を小問構成と解答プロセスとして使用
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
	logBuilder.WriteString("✅ [Stage1] 1段階目（小問構成と解答プロセス生成）が完了しました\n")
	
	return &models.Stage1Response{
		Success:               true,
		SubProblemsAndProcess: subProblemsAndProcess,
		Log:                   logBuilder.String(),
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
	
	// 2段階目用のプロンプトを作成（完全な問題生成）
	prompt := s.createNewStage2Prompt(req.SubProblemsAndProcess)
	logBuilder.WriteString("📝 2段階目用プロンプト（完全な問題生成）を作成しました\n")
	
	// AIクライアントを選択してAPI呼び出し
	var content string
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
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
		errorMsg := fmt.Sprintf("%s APIでの完全な問題生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage2Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 完全な問題を抽出
	completeProblem := s.extractCompleteProblem(content)
	if completeProblem == "" {
		completeProblem = strings.TrimSpace(content) // フォールバック：全体を完全な問題として使用
	}
	
	if completeProblem == "" {
		errorMsg := "完全な問題の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage2Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📝 完全な問題を抽出しました (長さ: %d文字)\n", len(completeProblem)))
	logBuilder.WriteString("✅ [Stage2] 2段階目（完全な問題生成）が完了しました\n")
	
	return &models.Stage2Response{
		Success:         true,
		CompleteProblem: completeProblem,
		Log:             logBuilder.String(),
	}, nil
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
	
	// 3段階目用のプロンプトを作成（数値計算プログラム生成）
	prompt := s.createNewStage3Prompt(req.SubProblemsAndProcess)
	logBuilder.WriteString("📝 3段階目用プロンプト（数値計算プログラム生成）を作成しました\n")
	
	// AIクライアントを選択してAPI呼び出し
	var content string
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
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
		errorMsg := fmt.Sprintf("%s APIでの数値計算プログラム生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage3Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 数値計算プログラムを抽出
	calculationProgram := s.extractCalculationProgram(content)
	if calculationProgram == "" {
		calculationProgram = strings.TrimSpace(content) // フォールバック：全体をプログラムとして使用
	}
	
	logBuilder.WriteString(fmt.Sprintf("🧮 計算プログラムの抽出: %t (長さ: %d文字)\n", calculationProgram != "", len(calculationProgram)))
	
	// 数値計算プログラムを実行
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
	
	logBuilder.WriteString("✅ [Stage3] 3段階目（数値計算プログラム生成・実行）が完了しました\n")
	
	return &models.Stage3Response{
		Success:            true,
		CalculationProgram: calculationProgram,
		CalculationResults: calculationResults,
		Log:                logBuilder.String(),
	}, nil
}
