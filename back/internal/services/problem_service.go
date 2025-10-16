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
)

type ProblemService interface {
	GenerateProblem(ctx context.Context, req models.GenerateProblemRequest, userSchoolCode string) (*models.Problem, error)
	GeneratePDF(ctx context.Context, req models.PDFGenerateRequest) (string, error)
	UpdateProblem(ctx context.Context, req models.UpdateProblemRequest, userID int64) (*models.Problem, error)
	RegenerateGeometry(ctx context.Context, req models.RegenerateGeometryRequest, userID int64) (string, error)
	SearchProblemsByParameters(ctx context.Context, userID int64, subject string, prompt string, filters map[string]interface{}) ([]*models.Problem, error)
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
	claudeClient clients.ClaudeClient
	openaiClient clients.OpenAIClient
	googleClient clients.GoogleClient
	coreClient   clients.CoreClient
	problemRepo  repositories.ProblemRepository
	userRepo     repositories.UserRepository
}

func NewProblemService(
	claudeClient clients.ClaudeClient,
	openaiClient clients.OpenAIClient,
	googleClient clients.GoogleClient,
	coreClient clients.CoreClient,
	problemRepo repositories.ProblemRepository,
	userRepo repositories.UserRepository,
) ProblemService {
	return &problemService{
		claudeClient: claudeClient,
		openaiClient: openaiClient,
		googleClient: googleClient,
		coreClient:   coreClient,
		problemRepo:  problemRepo,
		userRepo:     userRepo,
	}
}

func (s *problemService) GenerateProblem(ctx context.Context, req models.GenerateProblemRequest, userSchoolCode string) (*models.Problem, error) {
	// 1. ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// 2. 同じパラメータで既に生成された問題があるか検索
	if s.problemRepo != nil {
		existingProblems, err := s.problemRepo.SearchByParameters(ctx, user.ID, req.Subject, req.Prompt, req.Filters)
		if err == nil && len(existingProblems) > 0 {
			fmt.Printf("♻️ Found existing problem with same parameters. Reusing problem ID: %d\n", existingProblems[0].ID)
			return existingProblems[0], nil
		}
	}
	
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
		analysis, err := s.coreClient.AnalyzeProblem(ctx, problemText, req.Filters)
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
		UserID:      user.ID,
		Subject:     req.Subject,
		Prompt:      req.Prompt,
		Content:     problemText,
		Solution:    solutionText,
		ImageBase64: imageBase64,
		Filters:     req.Filters,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
	return `あなたは日本の中学校の数学教師です。以下の問題文から、図形描画用のPythonコードを生成してください。

【既存の問題文】
` + problemText + `

**出力形式**：
図形が必要な場合は、以下の形式で図形描画用のPythonコードを出力してください：

---GEOMETRY_CODE_START---
# 図形描画コード（問題に特化した図形を描画）
# 重要: import文は絶対に記述しないでください（事前にインポート済み）
# 利用可能な変数: plt, patches, np, numpy, Axes3D, Poly3DCollection

# 2D図形の場合
fig, ax = plt.subplots(1, 1, figsize=(8, 6))

# 3D図形の場合は以下を使用
# fig = plt.figure(figsize=(8, 8))
# ax = fig.add_subplot(111, projection='3d')

# ここに問題文に応じた具体的な図形描画コードを記述
# 例：正方形ABCD、点P、Q、Rの位置、線分、座標軸など

ax.set_aspect('equal')
ax.grid(True, alpha=0.3)
plt.tight_layout()
---GEOMETRY_CODE_END---

重要な指示：
1. 問題文に含まれる具体的な数値や条件を図形に正確に反映してください
2. 点の位置、線分の長さ、比率などを問題文通りに描画してください
3. **座標軸の表示判定**：
   - 問題文のキーワードで判定
   - 「座標」「グラフ」「関数」「x軸」「y軸」があれば、ax.grid(True, alpha=0.3) で座標軸を表示
   - 「体積」「面積」「角度」「長さ」「直方体」「円錐」「球」があれば、ax.axis('off') で座標軸を非表示
4. 図形のラベルは必ずアルファベット（A、B、C、P、Q、R等）を使用してください
5. ax.text()で日本語を使用しないでください
6. タイトルやラベルは英語またはアルファベットのみを使用してください
7. import文は記述しないでください（plt, np, patches, Axes3D, Poly3DCollectionは既に利用可能です）
8. numpy関数はnp.array(), np.linspace(), np.meshgrid()等で使用してください
9. 3D図形が必要な場合は以下を使用してください：
   - fig = plt.figure(figsize=(8, 8))
   - ax = fig.add_subplot(111, projection='3d')
   - ax.plot_surface(), ax.add_collection3d(Poly3DCollection())等
   - ax.view_init(elev=20, azim=-75)で視点を調整
10. 切断図形や断面図が必要な場合は、切断面をPoly3DCollectionで描画してください
11. **頂点ラベル（必須）**: 
   - 全ての頂点にアルファベット（A、B、C、D、E、F、G、H等）を表示
   - ax.text(x, y, z, 'A', size=16, color='black', weight='bold')
   - 立方体: A,B,C,D（下面）、E,F,G,H（上面）
   - 円錐: O（頂点）、A,B,C...（底面）

**注意**: 問題文に図形が不要な場合は、コードブロックを出力しないでください。`
}

// enhancePromptForGeometry enhances the prompt to include geometry generation instructions
func (s *problemService) enhancePromptForGeometry(prompt string) string {
	return `あなたは日本の中学校の数学教師です。以下の条件に従って、日本語で数学の問題を作成してください。

` + prompt + `

**出力形式**：
1. 問題文
2. 図形描画コード（必要な場合）
3. 解答・解説（別ページ用）

以下の形式で出力してください：

---PROBLEM_START---
【問題】
（ここに問題文を記述）
---PROBLEM_END---

もし問題に図形が必要な場合は、以下の形式で図形描画用のPythonコードを追加してください：

---GEOMETRY_CODE_START---
# 図形描画コード（問題に特化した図形を描画）
# 重要: import文は絶対に記述しないでください（事前にインポート済み）
# 利用可能な変数: plt, patches, np, numpy, Axes3D, Poly3DCollection

# 2D図形の場合
fig, ax = plt.subplots(1, 1, figsize=(8, 6))

# 3D図形の場合は以下を使用
# fig = plt.figure(figsize=(8, 8))
# ax = fig.add_subplot(111, projection='3d')

# ここに問題に応じた具体的な図形描画コードを記述
# 例：正方形ABCD、点P、Q、Rの位置、線分、座標軸など

ax.set_aspect('equal')
ax.grid(True, alpha=0.3)
plt.tight_layout()
---GEOMETRY_CODE_END---

---SOLUTION_START---
【解答・解説】
（ここに詳しい解答と解説を記述）

【解答】
（最終的な答え）

【解説】
（解法の手順と考え方を詳しく説明）
---SOLUTION_END---

重要：
1. 問題文に含まれる具体的な数値や条件を図形に正確に反映してください
2. 点の位置、線分の長さ、比率などを問題文通りに描画してください
3. **座標軸の表示判定**：
   - 問題文のキーワードで判定
   - 「座標」「グラフ」「関数」「x軸」「y軸」があれば、ax.grid(True, alpha=0.3) で座標軸を表示
   - 「体積」「面積」「角度」「長さ」「直方体」「円錐」「球」があれば、ax.axis('off') で座標軸を非表示
4. 図形のラベルは必ずアルファベット（A、B、C、P、Q、R等）を使用してください
5. ax.text()で日本語を使用しないでください
6. タイトルやラベルは英語またはアルファベットのみを使用してください
7. import文は記述しないでください（plt, np, patches, Axes3D, Poly3DCollectionは既に利用可能です）
8. numpy関数はnp.array(), np.linspace(), np.meshgrid()等で使用してください
9. 3D図形が必要な場合は以下を使用してください：
   - fig = plt.figure(figsize=(8, 8))
   - ax = fig.add_subplot(111, projection='3d')
   - ax.plot_surface(), ax.add_collection3d(Poly3DCollection())等
   - ax.view_init(elev=20, azim=-75)で視点を調整
10. 切断図形や断面図が必要な場合は、切断面をPoly3DCollectionで描画してください
11. **頂点ラベル（必須）**: 
   - 全ての頂点にアルファベット（A、B、C、D、E、F、G、H等）を表示
   - ax.text(x, y, z, 'A', size=16, color='black', weight='bold')
   - 立方体: A,B,C,D（下面）、E,F,G,H（上面）
   - 円錐: O（頂点）、A,B,C...（底面）`
}

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

// SearchProblemsByParameters パラメータで問題を検索
func (s *problemService) SearchProblemsByParameters(ctx context.Context, userID int64, subject string, prompt string, filters map[string]interface{}) ([]*models.Problem, error) {
	if s.problemRepo == nil {
		return nil, fmt.Errorf("problem repository is not initialized")
	}
	
	problems, err := s.problemRepo.SearchByParameters(ctx, userID, subject, prompt, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by parameters: %w", err)
	}
	
	return problems, nil
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
	geometryPrompt := s.createGeometryRegenerationPrompt(contentToAnalyze)
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
		
		analysis, err := s.coreClient.AnalyzeProblem(ctx, contentToAnalyze, problem.Filters)
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

// 2段階生成システムの実装

// GenerateProblemTwoStage 全体の2段階生成プロセスを実行
func (s *problemService) GenerateProblemTwoStage(ctx context.Context, req models.TwoStageGenerationRequest, userSchoolCode string) (*models.TwoStageGenerationResponse, error) {
	fmt.Printf("🚀 [TwoStage] Starting two-stage problem generation for user: %s\n", userSchoolCode)
	
	// 1回目のAPI呼び出し
	firstStageResp, err := s.GenerateFirstStage(ctx, req, userSchoolCode)
	if err != nil {
		return &models.TwoStageGenerationResponse{
			Success:       false,
			Error:         fmt.Sprintf("1回目のAPI呼び出しに失敗しました: %v", err),
			FirstStageLog: firstStageResp.Log,
		}, nil
	}
	
	if !firstStageResp.Success {
		return &models.TwoStageGenerationResponse{
			Success:       false,
			Error:         "1回目のAPI呼び出しが失敗しました",
			FirstStageLog: firstStageResp.Log,
		}, nil
	}
	
	// 2回目のAPI呼び出し
	secondStageReq := models.SecondStageRequest{
		ProblemText:  firstStageResp.ProblemText,
		GeometryCode: firstStageResp.GeometryCode,
	}
	
	secondStageResp, err := s.GenerateSecondStage(ctx, secondStageReq, userSchoolCode)
	if err != nil {
		return &models.TwoStageGenerationResponse{
			Success:        false,
			Error:          fmt.Sprintf("2回目のAPI呼び出しに失敗しました: %v", err),
			ProblemText:    firstStageResp.ProblemText,
			ImageBase64:    firstStageResp.ImageBase64,
			GeometryCode:   firstStageResp.GeometryCode,
			FirstStageLog:  firstStageResp.Log,
			SecondStageLog: secondStageResp.Log,
		}, nil
	}
	
	fmt.Printf("✅ [TwoStage] Two-stage problem generation completed successfully\n")
	
	return &models.TwoStageGenerationResponse{
		Success:             true,
		ProblemText:         firstStageResp.ProblemText,
		ImageBase64:         firstStageResp.ImageBase64,
		SolutionSteps:       secondStageResp.SolutionSteps,
		FinalSolution:       secondStageResp.FinalSolution,
		CalculationResults:  secondStageResp.CalculationResults,
		FirstStageLog:       firstStageResp.Log,
		SecondStageLog:      secondStageResp.Log,
		GeometryCode:        firstStageResp.GeometryCode,
		CalculationProgram:  secondStageResp.CalculationProgram,
	}, nil
}

// GenerateFirstStage 1回目のAPI呼び出し（問題文・図形生成）
func (s *problemService) GenerateFirstStage(ctx context.Context, req models.TwoStageGenerationRequest, userSchoolCode string) (*models.FirstStageResponse, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [FirstStage] 1回目のAPI呼び出しを開始 (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.FirstStageResponse{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	// API設定の確認
	if user.PreferredAPI == "" || user.PreferredModel == "" {
		errorMsg := fmt.Sprintf("AI設定が不完全です。設定ページでAPIとモデルを選択してください。現在の設定: API=%s, モデル=%s", user.PreferredAPI, user.PreferredModel)
		logBuilder.WriteString(fmt.Sprintf("⚠️ %s\n", errorMsg))
		return &models.FirstStageResponse{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 1回目用のプロンプトを作成
	prompt := s.createFirstStagePrompt(req.Prompt)
	logBuilder.WriteString("📝 1回目用プロンプトを作成しました\n")
	
	// AIクライアントを選択してAPI呼び出し
	var content string
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
		if err != nil {
			errorMsg := fmt.Sprintf("OpenAI APIでの問題生成に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.FirstStageResponse{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, err
		}
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
		if err != nil {
			errorMsg := fmt.Sprintf("Google APIでの問題生成に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.FirstStageResponse{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, err
		}
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
		if err != nil {
			errorMsg := fmt.Sprintf("Claude APIでの問題生成に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.FirstStageResponse{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, err
		}
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.FirstStageResponse{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 問題文とPythonコードを抽出
	problemText := s.extractProblemText(content)
	pythonCode := s.extractPythonCode(content)
	
	if problemText == "" {
		errorMsg := "問題文の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.FirstStageResponse{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📝 問題文を抽出しました (長さ: %d文字)\n", len(problemText)))
	logBuilder.WriteString(fmt.Sprintf("🐍 図形コードの抽出: %t\n", pythonCode != ""))
	
	// 図形生成
	var imageBase64 string
	if pythonCode != "" {
		logBuilder.WriteString("🎨 カスタム図形を生成中...\n")
		imageBase64, err = s.coreClient.GenerateCustomGeometry(ctx, pythonCode, problemText)
		if err != nil {
			logBuilder.WriteString(fmt.Sprintf("⚠️ カスタム図形生成に失敗: %v\n", err))
		} else {
			logBuilder.WriteString("✅ カスタム図形を生成しました\n")
		}
	} else {
		logBuilder.WriteString("🔍 従来の図形分析を実行中...\n")
		analysis, err := s.coreClient.AnalyzeProblem(ctx, problemText, req.Filters)
		if err != nil {
			logBuilder.WriteString(fmt.Sprintf("⚠️ 図形分析に失敗: %v\n", err))
		} else if analysis.NeedsGeometry && len(analysis.DetectedShapes) > 0 {
			shapeType := analysis.DetectedShapes[0]
			logBuilder.WriteString(fmt.Sprintf("🎨 %s図形を生成中...\n", shapeType))
			if params, exists := analysis.SuggestedParameters[shapeType]; exists {
				imageBase64, err = s.coreClient.GenerateGeometry(ctx, shapeType, params)
				if err != nil {
					logBuilder.WriteString(fmt.Sprintf("⚠️ 図形生成に失敗: %v\n", err))
				} else {
					logBuilder.WriteString(fmt.Sprintf("✅ %s図形を生成しました\n", shapeType))
				}
			}
		} else {
			logBuilder.WriteString("ℹ️ この問題には図形は必要ありません\n")
		}
	}
	
	logBuilder.WriteString(fmt.Sprintf("🖼️ 最終的な図形データの長さ: %d\n", len(imageBase64)))
	logBuilder.WriteString("✅ [FirstStage] 1回目のAPI呼び出しが完了しました\n")
	
	return &models.FirstStageResponse{
		Success:      true,
		ProblemText:  problemText,
		GeometryCode: pythonCode,
		ImageBase64:  imageBase64,
		Log:          logBuilder.String(),
	}, nil
}

// GenerateSecondStage 2回目のAPI呼び出し（解答手順・数値計算プログラム生成・実行）
func (s *problemService) GenerateSecondStage(ctx context.Context, req models.SecondStageRequest, userSchoolCode string) (*models.SecondStageResponse, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [SecondStage] 2回目のAPI呼び出しを開始 (ユーザー: %s)\n", userSchoolCode))
	
	// ユーザー情報を取得
	user, err := s.userRepo.GetBySchoolCode(ctx, userSchoolCode)
	if err != nil {
		errorMsg := fmt.Sprintf("ユーザー情報の取得に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.SecondStageResponse{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 2回目用のプロンプトを作成（解答手順・数値計算プログラム生成のみ）
	prompt := s.createSecondStagePrompt(req.ProblemText, req.GeometryCode)
	logBuilder.WriteString("📝 2回目用プロンプトを作成しました\n")
	
	// AIクライアントを選択してAPI呼び出し
	var content string
	switch user.PreferredAPI {
	case "openai", "chatgpt":
		dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
		if err != nil {
			errorMsg := fmt.Sprintf("OpenAI APIでの解答生成に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.SecondStageResponse{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, err
		}
	case "google", "gemini":
		dynamicClient := clients.NewGoogleClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
		if err != nil {
			errorMsg := fmt.Sprintf("Google APIでの解答生成に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.SecondStageResponse{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, err
		}
	case "claude", "laboratory":
		dynamicClient := clients.NewClaudeClient(user.PreferredModel)
		content, err = dynamicClient.GenerateContent(ctx, prompt)
		if err != nil {
			errorMsg := fmt.Sprintf("Claude APIでの解答生成に失敗しました: %v", err)
			logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
			return &models.SecondStageResponse{
				Success: false,
				Error:   errorMsg,
				Log:     logBuilder.String(),
			}, err
		}
	default:
		errorMsg := fmt.Sprintf("サポートされていないAPI「%s」が指定されています", user.PreferredAPI)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.SecondStageResponse{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 解答手順と数値計算プログラムを抽出
	solutionSteps := s.extractSolutionSteps(content)
	calculationProgram := s.extractCalculationProgram(content)
	
	logBuilder.WriteString(fmt.Sprintf("📚 解答手順の抽出: %t (長さ: %d文字)\n", solutionSteps != "", len(solutionSteps)))
	logBuilder.WriteString(fmt.Sprintf("🧮 計算プログラムの抽出: %t (長さ: %d文字)\n", calculationProgram != "", len(calculationProgram)))
	
	// 計算プログラムの内容をログに表示
	if calculationProgram != "" {
		logBuilder.WriteString(strings.Repeat("=", 50) + "\n")
		logBuilder.WriteString("🧮 [生成された数値計算プログラム]\n")
		logBuilder.WriteString(strings.Repeat("=", 50) + "\n")
		logBuilder.WriteString(calculationProgram + "\n")
		logBuilder.WriteString(strings.Repeat("=", 50) + "\n")
	}
	
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
	
	// 3回目のAPI呼び出し：解答手順と計算結果を統合して最終解説文を生成
	var finalSolution string
	if solutionSteps != "" && calculationResults != "" {
		logBuilder.WriteString("⭐ [ThirdStage] 3回目のAPI呼び出しを開始：解答手順と計算結果の統合\n")
		
		finalPrompt := s.createThirdStagePrompt(req.ProblemText, solutionSteps, calculationResults)
		logBuilder.WriteString("📝 3回目用プロンプトを作成しました\n")
		
		// 同じAIクライアントで最終解説文を生成
		switch user.PreferredAPI {
		case "openai", "chatgpt":
			dynamicClient := clients.NewOpenAIClient(user.PreferredModel)
			finalContent, err := dynamicClient.GenerateContent(ctx, finalPrompt)
			if err != nil {
				logBuilder.WriteString(fmt.Sprintf("⚠️ 3回目のAPI呼び出しに失敗: %v\n", err))
				finalSolution = solutionSteps // フォールバック：解答手順をそのまま使用
			} else {
				finalSolution = s.extractFinalSolution(finalContent)
				if finalSolution == "" {
					finalSolution = finalContent // マーカーがない場合は全体を使用
				}
				logBuilder.WriteString("✅ 3回目のAPI呼び出しで最終解説文を生成しました\n")
			}
		case "google", "gemini":
			dynamicClient := clients.NewGoogleClient(user.PreferredModel)
			finalContent, err := dynamicClient.GenerateContent(ctx, finalPrompt)
			if err != nil {
				logBuilder.WriteString(fmt.Sprintf("⚠️ 3回目のAPI呼び出しに失敗: %v\n", err))
				finalSolution = solutionSteps
			} else {
				finalSolution = s.extractFinalSolution(finalContent)
				if finalSolution == "" {
					finalSolution = finalContent
				}
				logBuilder.WriteString("✅ 3回目のAPI呼び出しで最終解説文を生成しました\n")
			}
		case "claude", "laboratory":
			dynamicClient := clients.NewClaudeClient(user.PreferredModel)
			finalContent, err := dynamicClient.GenerateContent(ctx, finalPrompt)
			if err != nil {
				logBuilder.WriteString(fmt.Sprintf("⚠️ 3回目のAPI呼び出しに失敗: %v\n", err))
				finalSolution = solutionSteps
			} else {
				finalSolution = s.extractFinalSolution(finalContent)
				if finalSolution == "" {
					finalSolution = finalContent
				}
				logBuilder.WriteString("✅ 3回目のAPI呼び出しで最終解説文を生成しました\n")
			}
		}
	} else {
		finalSolution = solutionSteps // 計算結果がない場合は解答手順をそのまま使用
	}
	
	logBuilder.WriteString("✅ [SecondStage] 2回目のAPI呼び出しが完了しました（3回目含む）\n")
	
	return &models.SecondStageResponse{
		Success:             true,
		SolutionSteps:       solutionSteps,
		CalculationProgram:  calculationProgram,
		FinalSolution:       finalSolution,
		CalculationResults:  calculationResults,
		Log:                 logBuilder.String(),
	}, nil
}

// createFirstStagePrompt 1回目API呼び出し用のプロンプトを作成
func (s *problemService) createFirstStagePrompt(userPrompt string) string {
	return `あなたは日本の中学校の数学教師です。以下の条件に従って、日本語で数学の問題を作成してください。

` + userPrompt + `

**重要：この段階では問題文と図形のみを生成し、解答・解説は生成しないでください。**

**出力形式**：

---PROBLEM_START---
【問題】
（ここに問題文を記述）
---PROBLEM_END---

もし問題に図形が必要な場合は、以下の形式で図形描画用のPythonコードを追加してください：

---GEOMETRY_CODE_START---
# 図形描画コード（問題に特化した図形を描画）
# 重要: import文は絶対に記述しないでください（事前にインポート済み）
# 利用可能な変数: plt, patches, np, numpy, Axes3D, Poly3DCollection

# 2D図形の場合
fig, ax = plt.subplots(1, 1, figsize=(8, 6))

# 3D図形の場合は以下を使用
# fig = plt.figure(figsize=(8, 8))
# ax = fig.add_subplot(111, projection='3d')

# ここに問題に応じた具体的な図形描画コードを記述
ax.set_aspect('equal')
ax.grid(True, alpha=0.3)
plt.tight_layout()
---GEOMETRY_CODE_END---

**注意事項**：
1. 解答・解説は絶対に含めないでください
2. 問題文は完全で自己完結的にしてください
3. 図形が必要な場合は、問題文の内容に正確に対応した図形コードを作成してください
4. import文は記述しないでください
5. 図形のラベルはアルファベット（A、B、C等）を使用してください`
}

// createSecondStagePrompt 2回目API呼び出し用のプロンプトを作成
func (s *problemService) createSecondStagePrompt(problemText, geometryCode string) string {
	prompt := `以下の問題について、詳細な解答の手順と数値計算を行うPythonプログラムを必ず作成してください。

【問題文】
` + problemText

	if geometryCode != "" {
		prompt += `

【図形描画コード】
` + geometryCode
	}

	prompt += `

**必須出力形式**：以下の3つのセクションを必ず全て含めて出力してください。

---SOLUTION_STEPS_START---
【解答の手順】
1. （手順1の詳細説明）
2. （手順2の詳細説明） 
3. （手順3の詳細説明）
...
（問題で問われている各小問について、段階的に解法を説明）
---SOLUTION_STEPS_END---

---CALCULATION_PROGRAM_START---
# 数値計算プログラム（Python）
# この問題の解答に必要な全ての数値計算を実行するプログラムです
# import文は使用しないでください（numpy は np として、math は math として利用可能）

print("=== 数値計算結果 ===")

# **必須**: 問題文の各小問に対応する具体的な数値計算を以下に記述してください

# 【重要】以下は計算例です。実際の問題に合わせて具体的な計算を実装してください：

# ===== 計算例1: 連立方程式を解く =====
# 例：3点を通る2次関数 y = ax² + bx + c を求める
# 点A(1,8), B(3,2), C(-1,18) を通る場合
# A = np.array([[1, 1, 1], [9, 3, 1], [1, -1, 1]])
# B = np.array([8, 2, 18])
# solution = np.linalg.solve(A, B)
# a, b, c = solution
# print(f"係数: a={a}, b={b}, c={c}")
# print(f"2次関数: y = {a}x² + ({b})x + {c}")

# ===== 計算例2: 2次関数の最大値・最小値 =====
# 例：f(x) = -2x² + 80x + 1000 の最大値
# a, b, c = -2, 80, 1000
# vertex_x = -b / (2*a)
# vertex_y = a * vertex_x**2 + b * vertex_x + c
# print(f"最大値: x={vertex_x}で y={vertex_y}")

# ===== 計算例3: 関数値の範囲計算 =====
# 例：区間[1,30]での関数値を計算
# for x_val in range(1, 31):
#     y_val = -2 * x_val**2 + 80 * x_val + 1000
#     print(f"x={x_val}: y={y_val}")

# ===== 計算例4: 方程式を解く =====
# 例：y = 3000 - 25x = 0 を解く
# x_solution = 3000 / 25
# print(f"販売終了日: {x_solution}日後")

# ===== 計算例5: 利益計算（複雑な場合） =====
# 例：価格変動による利益計算
# price_per_day = []
# for day in range(1, 121):
#     if 1 <= day <= 30:
#         price = -2 * day**2 + 80 * day + 1000
#     elif 31 <= day <= 60:
#         price = -day**2 + 20 * day + 2400
#     else:
#         price = 3000 - 25 * day
#     
#     if price <= 0:
#         break
#     
#     cost = 800
#     quantity = (price - 500) / 50
#     daily_profit = (price - cost) * quantity - 15000
#     
#     # 広告費を考慮
#     if day == 20 or day == 40:
#         daily_profit -= 50000
#     
#     print(f"日{day}: 価格{price}円, 利益{daily_profit}円")

# ===== 計算例6: 積分計算（総利益など） =====
# 例：総利益を数値的に計算
# total_profit = 0
# for day in range(1, 121):
#     # 日毎の利益を計算して累積
#     daily_profit = calculate_daily_profit(day)  # 関数定義が必要
#     total_profit += daily_profit
# print(f"総利益: {total_profit}円")

# ===== 計算例7: 最適化問題 =====
# 例：利益が最大となる日を探索
# max_profit = -float('inf')
# best_day = 0
# for day in range(1, 121):
#     profit = calculate_profit(day)
#     if profit > max_profit:
#         max_profit = profit
#         best_day = day
# print(f"最大利益: {best_day}日目で {max_profit}円")

# **実装必須**: 上記の例を参考に、実際の問題の各小問に対応する計算を以下に記述：

# 小問(1)の計算:
# （ここに1番目の小問に対する具体的な計算を記述）

# 小問(2)の計算:
# （ここに2番目の小問に対する具体的な計算を記述）

# 小問(3)の計算:
# （ここに3番目の小問に対する具体的な計算を記述）

# 小問(4)の計算:
# （ここに4番目の小問に対する具体的な計算を記述）

# さらに小問がある場合は継続して記述してください
---CALCULATION_PROGRAM_END---

---FINAL_SOLUTION_START---
【最終解答】
（問題の各小問に対する具体的な数値を含む最終的な答え）

例：
(1) y = 2x² - 5x + 11
(2) x = 3のとき、y = 14
(3) ...
---FINAL_SOLUTION_END---

**重要な指示**：
1. 解答の手順では、数学的な解法を段階的に説明してください
2. **数値計算プログラムは必須です**。問題に含まれる全ての数値計算を実行可能なPythonコードで記述してください
3. 連立方程式、2次方程式、関数の値計算など、問題に応じた適切な計算方法を使用してください
4. numpy（np）、math ライブラリは利用可能ですが、import文は記述しないでください
5. print文で計算結果を必ず出力してください
6. 最終解答には問題の各小問に対する具体的な数値を含めてください`

	return prompt
}

// GenerateStage4 4段階目：数値計算プログラム生成・実行
func (s *problemService) GenerateStage4(ctx context.Context, req models.Stage4Request, userSchoolCode string) (*models.Stage4Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage4] 4段階目を開始：数値計算プログラム生成・実行 (ユーザー: %s)\n", userSchoolCode))
	
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
	
	// 4段階目用のプロンプトを作成（数値計算プログラム生成）
	prompt := s.createStage4Prompt(req.ProblemText, req.SolutionSteps)
	logBuilder.WriteString("📝 4段階目用プロンプトを作成しました\n")
	
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
		errorMsg := fmt.Sprintf("%s APIでの数値計算プログラム生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage4Response{
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
	
	// 計算プログラムの内容をログに表示
	if calculationProgram != "" {
		logBuilder.WriteString(strings.Repeat("=", 50) + "\n")
		logBuilder.WriteString("🧮 [生成された数値計算プログラム]\n")
		logBuilder.WriteString(strings.Repeat("=", 50) + "\n")
		logBuilder.WriteString(calculationProgram + "\n")
		logBuilder.WriteString(strings.Repeat("=", 50) + "\n")
	}
	
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
	
	logBuilder.WriteString("✅ [Stage4] 4段階目が完了しました\n")
	
	return &models.Stage4Response{
		Success:            true,
		CalculationProgram: calculationProgram,
		CalculationResults: calculationResults,
		Log:                logBuilder.String(),
	}, nil
}

// createStage4Prompt 4段階目用のプロンプト（数値計算プログラム生成）
func (s *problemService) createStage4Prompt(problemText, solutionSteps string) string {
	return `以下の問題と解答手順について、全ての計算をPythonで実行する数値計算プログラムを作成してください。

【問題文】
` + problemText + `

【解答の手順】
` + solutionSteps + `

**重要：全ての数値計算はPythonで実行し、推測や手計算の結果を直接書き込まないでください。**

**必須出力形式**：

---CALCULATION_PROGRAM_START---
# 数値計算プログラム（Python）
# この問題の解答に必要な全ての数値計算を実行するプログラムです
# import文は使用しないでください（numpy は np として、math は math として利用可能）

print("=== 数値計算結果 ===")

# **絶対に守るべきルール**：
# 1. print文で計算結果の数値を直接書かないでください
# 2. 全ての計算はPythonの変数と演算で行ってください  
# 3. math.sqrt(), np.sqrt(), **, +, -, *, / を使って正確に計算してください

# **悪い例（絶対にやってはいけません）**：
# print(f"= √(144 + 144 + 81)")  # 数値を直接書いている
# print(f"= √369")              # 計算結果を推測している  
# print(f"= 19.2 cm")           # 最終結果を推測している

# **良い例（必ずこの方法で書いてください）**：
# a = 6 - (-6)
# b = 6 - (-6) 
# c = 9 - 0
# result = math.sqrt(a**2 + b**2 + c**2)
# print(f"= √({a}² + {b}² + {c}²)")
# print(f"= √({a**2} + {b**2} + {c**2})")
# print(f"= √{a**2 + b**2 + c**2}")
# print(f"= {result:.1f} cm")

# 以下に問題に応じた具体的な計算を記述してください：

# 座標系の設定（問題文に応じて調整）
print("1. 座標系の設定")

# 小問ごとの計算を実装してください
# 小問(1)の計算:
print("\n2. 小問(1)の計算")

# 小問(2)の計算:  
print("\n3. 小問(2)の計算")

# 小問(3)の計算:
print("\n4. 小問(3)の計算")

# 小問(4)の計算:
print("\n5. 小問(4)の計算")

# さらに小問がある場合は継続

print("\n=== 計算完了 ===")
---CALCULATION_PROGRAM_END---

**厳格な指示**：
1. **計算結果を推測しないでください** - 全ての数値はPythonで計算してください
2. **print文で数値を直接書かないでください** - 変数の値を表示してください
3. **math.sqrt(), **, +, -, *, / を使用してください** - 電卓的な推測は禁止です
4. **変数を使って段階的に計算してください** - 一度に複雑な式を書かないでください
5. **各小問について具体的な計算コードを記述してください**
6. **座標、距離、面積、体積など、問題に応じた適切な計算を実装してください**
7. **計算過程も含めて、全てをPythonの演算で実行してください**`
}

// GenerateStage5 5段階目：最終解説生成
func (s *problemService) GenerateStage5(ctx context.Context, req models.Stage5Request, userSchoolCode string) (*models.Stage5Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage5] 5段階目を開始：最終解説生成 (ユーザー: %s)\n", userSchoolCode))
	
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
	
	// 5段階目用のプロンプトを作成（最終解説生成）
	prompt := s.createStage5Prompt(req.ProblemText, req.SolutionSteps, req.CalculationResults)
	logBuilder.WriteString("📝 5段階目用プロンプトを作成しました\n")
	
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
		errorMsg := fmt.Sprintf("%s APIでの最終解説生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage5Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 最終解説を抽出
	finalExplanation := s.extractFinalSolution(content)
	if finalExplanation == "" {
		finalExplanation = strings.TrimSpace(content) // フォールバック：全体を解説として使用
	}
	
	if finalExplanation == "" {
		errorMsg := "最終解説の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage5Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📝 最終解説を抽出しました (長さ: %d文字)\n", len(finalExplanation)))
	logBuilder.WriteString("✅ [Stage5] 5段階目が完了しました\n")
	
	return &models.Stage5Response{
		Success:          true,
		FinalExplanation: finalExplanation,
		Log:              logBuilder.String(),
	}, nil
}

// createStage5Prompt 5段階目用のプロンプト（最終解説生成）
func (s *problemService) createStage5Prompt(problemText, solutionSteps, calculationResults string) string {
	return s.createThirdStagePrompt(problemText, solutionSteps, calculationResults) // 既存の統合ロジックを再利用
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
	return `以下の問題について、解答手順と数値計算結果を統合して、完全で理解しやすい解説文を作成してください。

【問題文】
` + problemText + `

【生成された解答手順】
` + solutionSteps + `

【数値計算の実行結果】
` + calculationResults + `

**出力形式**：

---FINAL_SOLUTION_START---
【完全な解答・解説】

（解答手順と計算結果を統合し、以下の構成で記述してください）

【解法】
（数学的な解法手順を、計算結果の具体的な数値を交えながら詳しく説明）

【計算過程】
（重要な計算過程を、実際の数値を使って示す）

【解答】
（問題の各小問に対する最終的な答えを具体的な数値で記述）

---FINAL_SOLUTION_END---

**重要な指示**：
1. 解答手順で述べた数学的な方法と、実際の計算結果を自然に統合してください
2. 抽象的な説明ではなく、具体的な数値を使った説明を心がけてください
3. 読み手が理解しやすいよう、計算過程と結果を明確に示してください
4. 問題の各小問について、明確で具体的な答えを提示してください
5. 数値の間違いがないよう、計算結果をそのまま活用してください`
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

// 5段階生成システムの実装

// GenerateProblemFiveStage 全体の5段階生成プロセスを実行
func (s *problemService) GenerateProblemFiveStage(ctx context.Context, req models.FiveStageGenerationRequest, userSchoolCode string) (*models.FiveStageGenerationResponse, error) {
	fmt.Printf("🚀 [FiveStage] Starting five-stage problem generation for user: %s\n", userSchoolCode)
	fmt.Printf("🔍 [FiveStage] Request details: Prompt length=%d, Subject=%s, Filters=%v\n", len(req.Prompt), req.Subject, req.Filters)
	
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
	
	// 1段階目：問題文生成
	stage1Req := models.Stage1Request{
		Prompt:  req.Prompt,
		Subject: req.Subject,
		Filters: req.Filters,
	}
	stage1Resp, err := s.GenerateStage1(ctx, stage1Req, userSchoolCode)
	if err != nil || !stage1Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:   false,
			Error:     fmt.Sprintf("1段階目に失敗しました: %v", err),
			Stage1Log: stage1Resp.Log,
		}, nil
	}
	
	// 2段階目：図形生成
	stage2Req := models.Stage2Request{
		ProblemText: stage1Resp.ProblemText,
	}
	stage2Resp, err := s.GenerateStage2(ctx, stage2Req, userSchoolCode)
	if err != nil || !stage2Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:     false,
			Error:       fmt.Sprintf("2段階目に失敗しました: %v", err),
			ProblemText: stage1Resp.ProblemText,
			Stage1Log:   stage1Resp.Log,
			Stage2Log:   stage2Resp.Log,
		}, nil
	}
	
	// 3段階目：解答手順生成
	stage3Req := models.Stage3Request{
		ProblemText:  stage1Resp.ProblemText,
		GeometryCode: stage2Resp.GeometryCode,
		ImageBase64:  stage2Resp.ImageBase64,
	}
	stage3Resp, err := s.GenerateStage3(ctx, stage3Req, userSchoolCode)
	if err != nil || !stage3Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:      false,
			Error:        fmt.Sprintf("3段階目に失敗しました: %v", err),
			ProblemText:  stage1Resp.ProblemText,
			GeometryCode: stage2Resp.GeometryCode,
			ImageBase64:  stage2Resp.ImageBase64,
			Stage1Log:    stage1Resp.Log,
			Stage2Log:    stage2Resp.Log,
			Stage3Log:    stage3Resp.Log,
		}, nil
	}
	
	// 4段階目：数値計算プログラム生成・実行
	stage4Req := models.Stage4Request{
		ProblemText:   stage1Resp.ProblemText,
		SolutionSteps: stage3Resp.SolutionSteps,
	}
	stage4Resp, err := s.GenerateStage4(ctx, stage4Req, userSchoolCode)
	if err != nil || !stage4Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:        false,
			Error:          fmt.Sprintf("4段階目に失敗しました: %v", err),
			ProblemText:    stage1Resp.ProblemText,
			GeometryCode:   stage2Resp.GeometryCode,
			ImageBase64:    stage2Resp.ImageBase64,
			SolutionSteps:  stage3Resp.SolutionSteps,
			Stage1Log:      stage1Resp.Log,
			Stage2Log:      stage2Resp.Log,
			Stage3Log:      stage3Resp.Log,
			Stage4Log:      stage4Resp.Log,
		}, nil
	}
	
	// 5段階目：最終解説生成
	stage5Req := models.Stage5Request{
		ProblemText:        stage1Resp.ProblemText,
		SolutionSteps:      stage3Resp.SolutionSteps,
		CalculationResults: stage4Resp.CalculationResults,
	}
	stage5Resp, err := s.GenerateStage5(ctx, stage5Req, userSchoolCode)
	if err != nil || !stage5Resp.Success {
		return &models.FiveStageGenerationResponse{
			Success:            false,
			Error:              fmt.Sprintf("5段階目に失敗しました: %v", err),
			ProblemText:        stage1Resp.ProblemText,
			GeometryCode:       stage2Resp.GeometryCode,
			ImageBase64:        stage2Resp.ImageBase64,
			SolutionSteps:      stage3Resp.SolutionSteps,
			CalculationProgram: stage4Resp.CalculationProgram,
			CalculationResults: stage4Resp.CalculationResults,
			Stage1Log:          stage1Resp.Log,
			Stage2Log:          stage2Resp.Log,
			Stage3Log:          stage3Resp.Log,
			Stage4Log:          stage4Resp.Log,
			Stage5Log:          stage5Resp.Log,
		}, nil
	}
	
	// 5段階生成完了後、問題をproblemsテーブルに保存
	fmt.Printf("💾 [FiveStage] Saving generated problem to database\n")
	
	problem := &models.Problem{
		UserID:      user.ID,
		Subject:     req.Subject,
		Prompt:      req.Prompt,
		Content:     stage1Resp.ProblemText,
		Solution:    stage5Resp.FinalExplanation,
		ImageBase64: stage2Resp.ImageBase64,
		Filters:     req.Filters,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
	
	fmt.Printf("✅ [FiveStage] Five-stage problem generation completed successfully\n")
	
	return &models.FiveStageGenerationResponse{
		Success:            true,
		ProblemText:        stage1Resp.ProblemText,
		GeometryCode:       stage2Resp.GeometryCode,
		ImageBase64:        stage2Resp.ImageBase64,
		SolutionSteps:      stage3Resp.SolutionSteps,
		CalculationProgram: stage4Resp.CalculationProgram,
		CalculationResults: stage4Resp.CalculationResults,
		FinalExplanation:   stage5Resp.FinalExplanation,
		Stage1Log:          stage1Resp.Log,
		Stage2Log:          stage2Resp.Log,
		Stage3Log:          stage3Resp.Log,
		Stage4Log:          stage4Resp.Log,
		Stage5Log:          stage5Resp.Log,
	}, nil
}

// GenerateStage1 1段階目：問題文のみ生成
func (s *problemService) GenerateStage1(ctx context.Context, req models.Stage1Request, userSchoolCode string) (*models.Stage1Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage1] 1段階目を開始：問題文生成 (ユーザー: %s)\n", userSchoolCode))
	
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
	
	// 生成制限チェック（-1は制限なし）
	if user.ProblemGenerationLimit >= 0 && user.ProblemGenerationCount >= user.ProblemGenerationLimit {
		errorMsg := fmt.Sprintf("問題生成回数の上限（%d回）に達しました", user.ProblemGenerationLimit)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("🔢 [Stage1] BEFORE UPDATE: User %s has %d/%d problems generated\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit))
	
	// 問題生成成功時にユーザーの生成回数を更新（Stage1で1回のみ更新）
	oldCount := user.ProblemGenerationCount
	user.ProblemGenerationCount++
	user.UpdatedAt = time.Now()
	
	logBuilder.WriteString(fmt.Sprintf("📝 [Stage1] Attempting to update user generation count from %d to %d\n", oldCount, user.ProblemGenerationCount))
	
	if err := s.userRepo.Update(ctx, user); err != nil {
		errorMsg := fmt.Sprintf("問題生成カウントの更新に失敗しました: %v", err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	} else {
		logBuilder.WriteString(fmt.Sprintf("✅ [Stage1] Successfully updated generation count: %s = %d/%d (was %d)\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit, oldCount))
	}
	
	logBuilder.WriteString(fmt.Sprintf("🔢 User %s: %d/%d problems generated\n", userSchoolCode, user.ProblemGenerationCount, user.ProblemGenerationLimit))
	
	// API設定の確認
	if user.PreferredAPI == "" || user.PreferredModel == "" {
		errorMsg := fmt.Sprintf("AI設定が不完全です。現在の設定: API=%s, モデル=%s", user.PreferredAPI, user.PreferredModel)
		logBuilder.WriteString(fmt.Sprintf("⚠️ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("🤖 使用するAPI: %s, モデル: %s\n", user.PreferredAPI, user.PreferredModel))
	
	// 1段階目用のプロンプトを作成（問題文のみ生成）
	prompt := s.createStage1Prompt(req.Prompt, req.Subject, req.Filters)
	logBuilder.WriteString("📝 1段階目用プロンプトを作成しました\n")
	
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
		errorMsg := fmt.Sprintf("%s APIでの問題生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 問題文を抽出
	problemText := s.extractProblemText(content)
	if problemText == "" {
		problemText = strings.TrimSpace(content) // フォールバック：全体を問題文として使用
	}
	
	if problemText == "" {
		errorMsg := "問題文の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage1Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📝 問題文を抽出しました (長さ: %d文字)\n", len(problemText)))
	logBuilder.WriteString("✅ [Stage1] 1段階目が完了しました\n")
	
	return &models.Stage1Response{
		Success:     true,
		ProblemText: problemText,
		Log:         logBuilder.String(),
	}, nil
}

// createStage1Prompt 1段階目用のプロンプトを作成（問題文のみ）
func (s *problemService) createStage1Prompt(userPrompt, subject string, filters map[string]interface{}) string {
	return `あなたは日本の中学校の数学教師です。以下の条件に従って、日本語で数学の問題文のみを作成してください。

` + userPrompt + `

**重要：この段階では問題文のみを生成し、図形・解答・解説は一切含めないでください。**

**出力形式**：

---PROBLEM_START---
【問題】
（ここに完全で自己完結的な問題文を記述）
---PROBLEM_END---

**注意事項**：
1. 図形描画コード、解答、解説は絶対に含めないでください
2. 問題文は完全で自己完結的にしてください
3. 具体的な数値や条件を含む詳細な問題文を作成してください
4. 図形が必要な問題でも、この段階では図形は生成しません
5. 問題文だけで読者が何を求められているかが明確に分かるようにしてください`
}

// GenerateStage2 2段階目：問題文から図形生成
func (s *problemService) GenerateStage2(ctx context.Context, req models.Stage2Request, userSchoolCode string) (*models.Stage2Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage2] 2段階目を開始：図形生成 (ユーザー: %s)\n", userSchoolCode))
	
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
	
	// 2段階目用のプロンプトを作成（図形生成専用）
	prompt := s.createStage2Prompt(req.ProblemText)
	logBuilder.WriteString("📝 2段階目用プロンプトを作成しました\n")
	
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
		logBuilder.WriteString(fmt.Sprintf("⚠️ AIによる図形コード生成に失敗: %v\n", err))
		// フォールバックとして図形なしで続行
		logBuilder.WriteString("ℹ️ この問題は図形なしで続行します\n")
		logBuilder.WriteString("✅ [Stage2] 2段階目が完了しました（図形なし）\n")
		
		return &models.Stage2Response{
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
		imageBase64, err = s.coreClient.GenerateCustomGeometry(ctx, geometryCode, req.ProblemText)
		if err != nil {
			logBuilder.WriteString(fmt.Sprintf("⚠️ 図形生成に失敗: %v\n", err))
		} else {
			logBuilder.WriteString("✅ 図形を生成しました\n")
		}
	} else {
		logBuilder.WriteString("ℹ️ この問題には図形は必要ありません\n")
	}
	
	logBuilder.WriteString(fmt.Sprintf("🖼️ 最終的な図形データの長さ: %d\n", len(imageBase64)))
	logBuilder.WriteString("✅ [Stage2] 2段階目が完了しました\n")
	
	return &models.Stage2Response{
		Success:      true,
		GeometryCode: geometryCode,
		ImageBase64:  imageBase64,
		Log:          logBuilder.String(),
	}, nil
}

// createStage2Prompt 2段階目用のプロンプト（図形生成専用）
func (s *problemService) createStage2Prompt(problemText string) string {
	return s.createGeometryRegenerationPrompt(problemText)
}

// GenerateStage3 3段階目：解答手順生成
func (s *problemService) GenerateStage3(ctx context.Context, req models.Stage3Request, userSchoolCode string) (*models.Stage3Response, error) {
	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("⭐ [Stage3] 3段階目を開始：解答手順生成 (ユーザー: %s)\n", userSchoolCode))
	
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
	
	// 3段階目用のプロンプトを作成（解答手順のみ）
	prompt := s.createStage3Prompt(req.ProblemText, req.GeometryCode)
	logBuilder.WriteString("📝 3段階目用プロンプトを作成しました\n")
	
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
		errorMsg := fmt.Sprintf("%s APIでの解答手順生成に失敗しました: %v", user.PreferredAPI, err)
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage3Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, err
	}
	
	logBuilder.WriteString(fmt.Sprintf("✅ AIからのレスポンスを受信しました (長さ: %d文字)\n", len(content)))
	
	// 解答手順を抽出
	solutionSteps := s.extractSolutionSteps(content)
	if solutionSteps == "" {
		solutionSteps = strings.TrimSpace(content) // フォールバック：全体を解答手順として使用
	}
	
	if solutionSteps == "" {
		errorMsg := "解答手順の抽出に失敗しました"
		logBuilder.WriteString(fmt.Sprintf("❌ %s\n", errorMsg))
		return &models.Stage3Response{
			Success: false,
			Error:   errorMsg,
			Log:     logBuilder.String(),
		}, fmt.Errorf(errorMsg)
	}
	
	logBuilder.WriteString(fmt.Sprintf("📚 解答手順を抽出しました (長さ: %d文字)\n", len(solutionSteps)))
	logBuilder.WriteString("✅ [Stage3] 3段階目が完了しました\n")
	
	return &models.Stage3Response{
		Success:       true,
		SolutionSteps: solutionSteps,
		Log:           logBuilder.String(),
	}, nil
}

// createStage3Prompt 3段階目用のプロンプト（解答手順のみ）
func (s *problemService) createStage3Prompt(problemText, geometryCode string) string {
	prompt := `以下の問題について、詳細な解答の手順のみを作成してください。数値計算は行わず、解法の流れのみを説明してください。

【問題文】
` + problemText

	if geometryCode != "" {
		prompt += `

【図形描画コード】
` + geometryCode
	}

	prompt += `

**重要：この段階では解答の手順のみを生成し、具体的な数値計算は行わないでください。**

**出力形式**：

---SOLUTION_STEPS_START---
【解答の手順】
1. （手順1：どのような考え方で解くか）
2. （手順2：どのような公式や定理を使うか）
3. （手順3：計算の流れはどうなるか）
4. （手順4：最終的に何を求めるか）
...
（問題で問われている各小問について、段階的に解法の手順を説明）
---SOLUTION_STEPS_END---

**注意事項**：
1. 具体的な数値での計算は行わず、解法の手順のみを説明してください
2. 使用する公式や定理を明記してください
3. 各小問について、どのような流れで解答するかを詳しく説明してください
4. 数値計算プログラムや最終的な答えは含めないでください
5. 読み手が解法の流れを理解できるような詳細な手順を記述してください`

	return prompt
}
