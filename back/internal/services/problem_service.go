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
	SearchProblemsByParameters(ctx context.Context, userID int64, subject string, prompt string, filters map[string]interface{}) ([]*models.Problem, error)
	SearchProblemsByKeyword(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*models.Problem, error)
	GetUserProblems(ctx context.Context, userID int64, limit, offset int) ([]*models.Problem, error)
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

	// 4. 問題生成成功時にユーザーの生成回数を更新
	user.ProblemGenerationCount++
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		// ログに記録するが、問題生成は成功として扱う
		fmt.Printf("⚠️ Failed to update user generation count: %v\n", err)
	} else {
		fmt.Printf("✅ Updated user %s generation count to %d\n", userSchoolCode, user.ProblemGenerationCount)
	}

	return problem, nil
}

func (s *problemService) GeneratePDF(ctx context.Context, req models.PDFGenerateRequest) (string, error) {
	pdfBase64, err := s.coreClient.GeneratePDF(ctx, req.ProblemText, req.ImageBase64, req.SolutionText)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}
	return pdfBase64, nil
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
