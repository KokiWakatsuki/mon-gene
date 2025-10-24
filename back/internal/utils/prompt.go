package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptLoader プロンプトファイルを読み込むためのユーティリティ
type PromptLoader struct {
	baseDir string
}

// NewPromptLoader プロンプトローダーを初期化
func NewPromptLoader(baseDir string) *PromptLoader {
	return &PromptLoader{
		baseDir: baseDir,
	}
}

// LoadPrompt プロンプトファイルを読み込み、変数を置換して返す
func (p *PromptLoader) LoadPrompt(filename string, variables map[string]string) (string, error) {
	filePath := filepath.Join(p.baseDir, filename)
	
	// ファイルの存在確認
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("prompt file not found: %s", filePath)
	}
	
	// ファイル読み込み
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", filePath, err)
	}
	
	// 変数の置換
	prompt := string(content)
	for key, value := range variables {
		placeholder := "{" + key + "}"
		prompt = strings.ReplaceAll(prompt, placeholder, value)
	}
	
	return prompt, nil
}

// LoadGeometryRegenerationPrompt 図形再生成プロンプトを読み込み
func (p *PromptLoader) LoadGeometryRegenerationPrompt(problemText string) (string, error) {
	variables := map[string]string{
		"PROBLEM_TEXT": problemText,
	}
	return p.LoadPrompt("geometry_regeneration.txt", variables)
}

// LoadConversationFormatPrompt 会話形式プロンプトを読み込み
func (p *PromptLoader) LoadConversationFormatPrompt(userPrompt string) (string, error) {
	variables := map[string]string{
		"USER_PROMPT": userPrompt,
	}
	return p.LoadPrompt("conversation_format.txt", variables)
}

// LoadStandardFormatPrompt 標準形式プロンプトを読み込み
func (p *PromptLoader) LoadStandardFormatPrompt(userPrompt string) (string, error) {
	variables := map[string]string{
		"USER_PROMPT": userPrompt,
	}
	return p.LoadPrompt("standard_format.txt", variables)
}

// LoadStage1Prompt 1段階目プロンプトを読み込み
func (p *PromptLoader) LoadStage1Prompt(userPrompt, subject string) (string, error) {
	variables := map[string]string{
		"USER_PROMPT": userPrompt,
		"SUBJECT":     subject,
	}
	return p.LoadPrompt("stage1_problem_text.txt", variables)
}

// LoadStage3Prompt 3段階目プロンプトを読み込み
func (p *PromptLoader) LoadStage3Prompt(problemText, geometryCode string) (string, error) {
	variables := map[string]string{
		"PROBLEM_TEXT": problemText,
	}
	
	// 図形コードがある場合の追加セクション
	if geometryCode != "" {
		variables["GEOMETRY_CODE_SECTION"] = `
【図形描画コード】
` + geometryCode
	} else {
		variables["GEOMETRY_CODE_SECTION"] = ""
	}
	
	return p.LoadPrompt("stage3_solution_steps.txt", variables)
}

// LoadStage4Prompt 4段階目プロンプトを読み込み
func (p *PromptLoader) LoadStage4Prompt(problemText, solutionSteps string) (string, error) {
	variables := map[string]string{
		"PROBLEM_TEXT":    problemText,
		"SOLUTION_STEPS":  solutionSteps,
	}
	return p.LoadPrompt("stage4_calculation_program.txt", variables)
}

// LoadStage5Prompt 5段階目プロンプトを読み込み
func (p *PromptLoader) LoadStage5Prompt(problemText, solutionSteps, calculationResults string) (string, error) {
	variables := map[string]string{
		"PROBLEM_TEXT":        problemText,
		"SOLUTION_STEPS":      solutionSteps,
		"CALCULATION_RESULTS": calculationResults,
	}
	return p.LoadPrompt("stage5_final_explanation.txt", variables)
}
