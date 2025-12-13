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
		
		// 3. 解答手順（解答プロセスとしても使用）
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


// LoadGeometryPromptWithSamples 図形描画プロンプトにサンプルを追加して読み込み（新Stage5用）
func (p *PromptLoader) LoadGeometryPromptWithSamples(problemText string) (string, error) {
	samples, err := p.LoadSampleProblems()
	if err != nil {
		// サンプルが読み込めない場合は通常のプロンプトを返す
		return p.LoadGeometryRegenerationPrompt(problemText)
	}
	
	// few-shotサンプルを構築（図形描画コードを使用）
	var fewShotExamples strings.Builder
	fewShotExamples.WriteString("\n<few_shot_examples>\n")
	fewShotExamples.WriteString("以下は参考となる図形描画プログラムの例です：\n\n")
	
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

// LoadFiveStageInitialPrompt 5段階生成の初期プロンプトを読み込み（全ステージの指示を含む）
func (p *PromptLoader) LoadFiveStageInitialPrompt(userPrompt, subject, opinionProfile string) (string, error) {
	samples, err := p.LoadSampleProblems()
	
	// few-shotサンプルを構築
	var fewShotExamples strings.Builder
	if err == nil && len(samples) > 0 {
		fewShotExamples.WriteString("\n<few_shot_examples>\n")
		fewShotExamples.WriteString("以下は参考となる問題生成の例です：\n\n")
		
		for i, sample := range samples {
			fewShotExamples.WriteString(fmt.Sprintf("【例%d】\n", i+1))
			fewShotExamples.WriteString("問題文：\n")
			fewShotExamples.WriteString(sample.ProblemText)
			fewShotExamples.WriteString("\n\n")
			
			if sample.GeometryCode != "" {
				fewShotExamples.WriteString("図形コード：\n```python\n")
				fewShotExamples.WriteString(sample.GeometryCode)
				fewShotExamples.WriteString("\n```\n\n")
			}
			
			if sample.FinalExplanation != "" {
				fewShotExamples.WriteString("解答・解説：\n")
				fewShotExamples.WriteString(sample.FinalExplanation)
				fewShotExamples.WriteString("\n\n")
			}
		}
		fewShotExamples.WriteString("</few_shot_examples>\n")
	} else {
		fewShotExamples.WriteString("")
	}
	
	variables := map[string]string{
		"USER_PROMPT":      userPrompt,
		"SUBJECT":          subject,
		"OPINION_PROFILE":  opinionProfile,
		"FEW_SHOT_SAMPLES": fewShotExamples.String(),
	}
	
	return p.LoadPrompt("five_stage_initial.txt", variables)
}

// LoadStageTrigger ステージトリガーを読み込む
func (p *PromptLoader) LoadStageTrigger() (string, error) {
	content, err := os.ReadFile(filepath.Join(p.baseDir, "stage_trigger.txt"))
	if err != nil {
		return "", fmt.Errorf("failed to read stage trigger file: %w", err)
	}
	return string(content), nil
}

// LoadThreeProblemGenerationPrompt 3問生成プロンプトを読み込み
func (p *PromptLoader) LoadThreeProblemGenerationPrompt(uploadedProblemContent, currentPattern string, excludedUnits []string) (string, error) {
	// デバッグログ：除外単元を確認
	fmt.Printf("🔍 [PromptLoader] LoadThreeProblemGenerationPrompt called\n")
	fmt.Printf("🔍 [PromptLoader] excludedUnits count: %d\n", len(excludedUnits))
	for i, unit := range excludedUnits {
		fmt.Printf("🔍 [PromptLoader] excludedUnits[%d]: %s\n", i, unit)
	}
	
	// 除外単元の指示を構築
	excludedUnitsInstruction := ""
	if len(excludedUnits) > 0 {
		excludedUnitsInstruction = "以下の単元はまだ習っていないため、問題に含めないでください:\n"
		for _, unit := range excludedUnits {
			excludedUnitsInstruction += fmt.Sprintf("- %s\n", unit)
		}
		fmt.Printf("✅ [PromptLoader] excludedUnitsInstruction generated (length: %d)\n", len(excludedUnitsInstruction))
		fmt.Printf("📝 [PromptLoader] excludedUnitsInstruction content:\n%s\n", excludedUnitsInstruction)
	} else {
		fmt.Printf("ℹ️ [PromptLoader] No excluded units specified\n")
	}
	
	variables := map[string]string{
		"UPLOADED_PROBLEM_CONTENT":    uploadedProblemContent,
		"UPLOADED_SOLUTION_CONTENT":   "(解答ファイルはアップロードされていません)",
		"CURRENT_PATTERN":             currentPattern,
		"EXCLUDED_UNITS_INSTRUCTION":  excludedUnitsInstruction,
	}
	
	fmt.Printf("🔍 [PromptLoader] Variables prepared, loading prompt file...\n")
	return p.LoadPrompt("three_problem_generation.txt", variables)
}

// LoadThreeProblemGenerationPromptWithSolution 3問生成プロンプトを読み込み（解答付き）
func (p *PromptLoader) LoadThreeProblemGenerationPromptWithSolution(uploadedProblemContent, uploadedSolutionContent, currentPattern string) (string, error) {
	solutionText := uploadedSolutionContent
	if solutionText == "" {
		solutionText = "(解答ファイルはアップロードされていません)"
	}
	
	variables := map[string]string{
		"UPLOADED_PROBLEM_CONTENT":  uploadedProblemContent,
		"UPLOADED_SOLUTION_CONTENT": solutionText,
		"CURRENT_PATTERN":           currentPattern,
	}
	return p.LoadPrompt("three_problem_generation.txt", variables)
}
