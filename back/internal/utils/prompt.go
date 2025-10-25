package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"regexp"
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

// LoadGeometryRegenerationPromptWithSamples 図形再生成プロンプトにサンプルを追加して読み込み
func (p *PromptLoader) LoadGeometryRegenerationPromptWithSamples(problemText string) (string, error) {
	samples, err := p.LoadSampleProblems()
	if err != nil {
		// サンプルが読み込めない場合は通常のプロンプトを返す
		return p.LoadGeometryRegenerationPrompt(problemText)
	}
	
	// few-shotサンプルを構築
	var fewShotExamples strings.Builder
	fewShotExamples.WriteString("\n<few_shot_examples>\n")
	fewShotExamples.WriteString("以下は参考となる図形描画コードの例です：\n\n")
	
	for i, sample := range samples {
		if sample.GeometryCode != "" {
			fewShotExamples.WriteString(fmt.Sprintf("【例%d】\n", i+1))
			fewShotExamples.WriteString("```python\n")
			fewShotExamples.WriteString(sample.GeometryCode)
			fewShotExamples.WriteString("\n```\n\n")
		}
	}
	fewShotExamples.WriteString("</few_shot_examples>\n")
	
	variables := map[string]string{
		"PROBLEM_TEXT":     problemText,
		"FEW_SHOT_SAMPLES": fewShotExamples.String(),
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

// SampleData サンプル問題のデータ構造
type SampleData struct {
	ProblemText        string
	GeometryCode       string
	SolutionSteps      string
	CalculationProgram string
	FinalExplanation   string
}

// LoadSampleProblems サンプル問題を読み込む
func (p *PromptLoader) LoadSampleProblems() ([]SampleData, error) {
	sampleDir := filepath.Join(p.baseDir, "../sample")
	
	// サンプルファイル一覧を取得
	files, err := os.ReadDir(sampleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sample directory: %w", err)
	}
	
	var samples []SampleData
	
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		
		filePath := filepath.Join(sampleDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		
		sample, err := parseSampleContent(string(content))
		if err != nil {
			continue
		}
		
		samples = append(samples, sample)
	}
	
	return samples, nil
}

// parseSampleContent サンプルコンテンツを解析して各セクションに分離
func parseSampleContent(content string) (SampleData, error) {
	var sample SampleData
	
	// セクション分割（### で始まる見出しで分割）
	sections := strings.Split(content, "### ")
	
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		
		// 1. 問題文
		if strings.HasPrefix(section, "1. 問題文") || strings.Contains(section, "問題文（小問同士のつながりや") {
			sample.ProblemText = extractProblemText(section)
		}
		
		// 2. 図形描画のPythonコード
		if strings.HasPrefix(section, "2. 問題文から，図形描画のPythonコード") {
			sample.GeometryCode = extractCodeSection(section, "python")
		}
		
		// 3. 解答手順
		if strings.HasPrefix(section, "3. 問題文と図形から，解答手順") {
			sample.SolutionSteps = extractSolutionSteps(section)
		}
		
		// 4. 数値計算プログラム
		if strings.HasPrefix(section, "4. 解答手順から，数値計算を行うPythonプログラム") {
			sample.CalculationProgram = extractCodeSection(section, "python")
		}
		
		// 5. 完全な解答・解説
		if strings.HasPrefix(section, "5. 解答手順と数値計算の結果から，完全な解答・解説") {
			sample.FinalExplanation = extractFinalExplanation(section)
		}
	}
	
	return sample, nil
}

// extractProblemText 問題文を抽出
func extractProblemText(section string) string {
	// 見出しを取り除き、問題文本体のみを抽出
	lines := strings.Split(section, "\n")
	var problemLines []string
	inProblem := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 見出し部分をスキップ
		if strings.Contains(line, "1. 問題文") {
			continue
		}
		
		// 空行で開始を判断
		if !inProblem && line != "" {
			inProblem = true
		}
		
		if inProblem {
			problemLines = append(problemLines, line)
		}
	}
	
	return strings.Join(problemLines, "\n")
}

// extractCodeSection コードセクションを抽出
func extractCodeSection(section string, codeType string) string {
	// ```python と ```の間を抽出
	re := regexp.MustCompile("```" + codeType + "([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(section, -1)
	
	var codes []string
	for _, match := range matches {
		if len(match) > 1 {
			codes = append(codes, strings.TrimSpace(match[1]))
		}
	}
	
	return strings.Join(codes, "\n\n")
}

// extractSolutionSteps 解答手順を抽出
func extractSolutionSteps(section string) string {
	// 見出しの後のコンテンツを抽出（#### で始まる小見出しを含む）
	lines := strings.Split(section, "\n")
	var stepLines []string
	
	skipHeader := true
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// ヘッダー行をスキップ
		if skipHeader && strings.Contains(line, "3. 問題文と図形から") {
			skipHeader = false
			continue
		}
		
		if !skipHeader {
			stepLines = append(stepLines, line)
		}
	}
	
	return strings.Join(stepLines, "\n")
}

// extractFinalExplanation 最終解説を抽出
func extractFinalExplanation(section string) string {
	// "#### 解答" と "#### 解説" の部分を抽出
	lines := strings.Split(section, "\n")
	var explanationLines []string
	
	skipHeader := true
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// ヘッダー行をスキップ
		if skipHeader && strings.Contains(line, "5. 解答手順と数値計算") {
			skipHeader = false
			continue
		}
		
		if !skipHeader {
			explanationLines = append(explanationLines, line)
		}
	}
	
	return strings.Join(explanationLines, "\n")
}

// LoadStage1PromptWithSamples stage1プロンプトにサンプルを追加して読み込み
func (p *PromptLoader) LoadStage1PromptWithSamples(userPrompt, subject string) (string, error) {
	samples, err := p.LoadSampleProblems()
	if err != nil {
		// サンプルが読み込めない場合は通常のプロンプトを返す
		return p.LoadStage1Prompt(userPrompt, subject)
	}
	
	// few-shotサンプルを構築
	var fewShotExamples strings.Builder
	fewShotExamples.WriteString("\n<few_shot_examples>\n")
	fewShotExamples.WriteString("以下は参考となる問題文の例です：\n\n")
	
	for i, sample := range samples {
		if sample.ProblemText != "" {
			fewShotExamples.WriteString(fmt.Sprintf("【例%d】\n", i+1))
			fewShotExamples.WriteString(sample.ProblemText)
			fewShotExamples.WriteString("\n\n")
		}
	}
	fewShotExamples.WriteString("</few_shot_examples>\n")
	
	variables := map[string]string{
		"USER_PROMPT":      userPrompt,
		"SUBJECT":          subject,
		"FEW_SHOT_SAMPLES": fewShotExamples.String(),
	}
	
	return p.LoadPrompt("stage1_problem_text.txt", variables)
}

// LoadStage3PromptWithSamples stage3プロンプトにサンプルを追加して読み込み
func (p *PromptLoader) LoadStage3PromptWithSamples(problemText, geometryCode string) (string, error) {
	samples, err := p.LoadSampleProblems()
	if err != nil {
		return p.LoadStage3Prompt(problemText, geometryCode)
	}
	
	// few-shotサンプルを構築
	var fewShotExamples strings.Builder
	fewShotExamples.WriteString("\n<few_shot_examples>\n")
	fewShotExamples.WriteString("以下は参考となる解答手順の例です：\n\n")
	
	for i, sample := range samples {
		if sample.SolutionSteps != "" {
			fewShotExamples.WriteString(fmt.Sprintf("【例%d】\n", i+1))
			fewShotExamples.WriteString(sample.SolutionSteps)
			fewShotExamples.WriteString("\n\n")
		}
	}
	fewShotExamples.WriteString("</few_shot_examples>\n")
	
	variables := map[string]string{
		"PROBLEM_TEXT":     problemText,
		"FEW_SHOT_SAMPLES": fewShotExamples.String(),
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

// LoadStage4PromptWithSamples stage4プロンプトにサンプルを追加して読み込み
func (p *PromptLoader) LoadStage4PromptWithSamples(problemText, solutionSteps string) (string, error) {
	samples, err := p.LoadSampleProblems()
	if err != nil {
		return p.LoadStage4Prompt(problemText, solutionSteps)
	}
	
	// few-shotサンプルを構築
	var fewShotExamples strings.Builder
	fewShotExamples.WriteString("\n<few_shot_examples>\n")
	fewShotExamples.WriteString("以下は参考となる数値計算プログラムの例です：\n\n")
	
	for i, sample := range samples {
		if sample.CalculationProgram != "" {
			fewShotExamples.WriteString(fmt.Sprintf("【例%d】\n", i+1))
			fewShotExamples.WriteString("```python\n")
			fewShotExamples.WriteString(sample.CalculationProgram)
			fewShotExamples.WriteString("\n```\n\n")
		}
	}
	fewShotExamples.WriteString("</few_shot_examples>\n")
	
	variables := map[string]string{
		"PROBLEM_TEXT":      problemText,
		"SOLUTION_STEPS":    solutionSteps,
		"FEW_SHOT_SAMPLES":  fewShotExamples.String(),
	}
	
	return p.LoadPrompt("stage4_calculation_program.txt", variables)
}

// LoadStage5PromptWithSamples stage5プロンプトにサンプルを追加して読み込み
func (p *PromptLoader) LoadStage5PromptWithSamples(problemText, solutionSteps, calculationResults string) (string, error) {
	samples, err := p.LoadSampleProblems()
	if err != nil {
		return p.LoadStage5Prompt(problemText, solutionSteps, calculationResults)
	}
	
	// few-shotサンプルを構築
	var fewShotExamples strings.Builder
	fewShotExamples.WriteString("\n<few_shot_examples>\n")
	fewShotExamples.WriteString("以下は参考となる完全な解答・解説の例です：\n\n")
	
	for i, sample := range samples {
		if sample.FinalExplanation != "" {
			fewShotExamples.WriteString(fmt.Sprintf("【例%d】\n", i+1))
			fewShotExamples.WriteString(sample.FinalExplanation)
			fewShotExamples.WriteString("\n\n")
		}
	}
	fewShotExamples.WriteString("</few_shot_examples>\n")
	
	variables := map[string]string{
		"PROBLEM_TEXT":        problemText,
		"SOLUTION_STEPS":      solutionSteps,
		"CALCULATION_RESULTS": calculationResults,
		"FEW_SHOT_SAMPLES":    fewShotExamples.String(),
	}
	
	return p.LoadPrompt("stage5_final_explanation.txt", variables)
}
