package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type googleClient struct {
	apiKey string
	model  string
}

type GoogleRequest struct {
	Contents []GoogleContent `json:"contents"`
	GenerationConfig GoogleGenerationConfig `json:"generationConfig"`
}

type GoogleContent struct {
	Parts []GooglePart `json:"parts"`
}

type GooglePart struct {
	Text string `json:"text"`
}

type GoogleGenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
	ThinkingConfig *GoogleThinkingConfig `json:"thinkingConfig,omitempty"`
}

type GoogleThinkingConfig struct {
	ThinkingBudget *int32 `json:"thinkingBudget,omitempty"`
}

type GoogleResponse struct {
	Candidates []GoogleCandidate `json:"candidates"`
	Error      *GoogleError      `json:"error,omitempty"`
}

type GoogleCandidate struct {
	Content      GoogleContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type GoogleError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func NewGoogleClient(model string) GoogleClient {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Printf("⚠️ GOOGLE_API_KEY not found in environment variables\n")
	}
	
	// モデル名が空の場合はデフォルトを使用しない
	if model == "" {
		fmt.Printf("⚠️ Google model not specified\n")
	}
	
	// 古いモデル名を新しいものに自動変換
	if model == "gemini-pro" {
		model = "gemini-1.5-flash"
		fmt.Printf("🔄 Converting deprecated model 'gemini-pro' to 'gemini-1.5-flash'\n")
	}
	
	// models/プレフィックスがない場合は自動的に追加
	if model != "" && !strings.HasPrefix(model, "models/") {
		model = "models/" + model
		fmt.Printf("🔄 Adding 'models/' prefix to Google model: %s\n", model)
	}
	
	return &googleClient{
		apiKey: apiKey,
		model:  model,
	}
}

func (c *googleClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Google API key not configured")
	}

	if c.model == "" {
		return "", fmt.Errorf("Google model not specified. Please configure your AI settings in the settings page")
	}

	fmt.Printf("🤖 Using Google API with model: %s\n", c.model)

	// 推論トークンの設定
	// 用途: 1段階生成（旧方式）- 単一プロンプトから問題を一度に生成
	thinkingBudget := int32(10000)
	maxOutputTokens := 10000 + int(thinkingBudget) // 推論トークン + 10000

	request := GoogleRequest{
		Contents: []GoogleContent{
			{
				Parts: []GooglePart{
					{
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: GoogleGenerationConfig{
			MaxOutputTokens: maxOutputTokens,
			ThinkingConfig: &GoogleThinkingConfig{
				ThinkingBudget: &thinkingBudget,
			},
		},
	}

	fmt.Printf("🧠 Using thinking tokens: %d, max output tokens: %d\n", thinkingBudget, maxOutputTokens)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/%s:generateContent?key=%s", c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Google API error - Status: %d, Body: %s\n", resp.StatusCode, string(body))
		// より詳細なエラー情報を提供
		var errorResponse GoogleResponse
		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error != nil {
			switch errorResponse.Error.Code {
			case 400:
				if strings.Contains(errorResponse.Error.Message, "too many tokens") || strings.Contains(errorResponse.Error.Message, "maximum context length") {
					return "", NewTokenLimitError(fmt.Sprintf("入力テキストが長すぎます。テキストを短くして再度お試しください。詳細: %s", errorResponse.Error.Message))
				}
				return "", NewGeneralError(fmt.Sprintf("Google API リクエストエラー: %s", errorResponse.Error.Message))
			case 403:
				return "", NewInvalidAPIKeyError(fmt.Sprintf("設定を確認してください。詳細: %s", errorResponse.Error.Message))
			case 404:
				return "", NewModelNotFoundError(fmt.Sprintf("モデル「%s」が利用できません。詳細: %s", c.model, errorResponse.Error.Message))
			case 429:
				return "", NewRateLimitError(fmt.Sprintf("しばらく待ってから再試行してください。詳細: %s", errorResponse.Error.Message))
			default:
				return "", NewGeneralError(fmt.Sprintf("Google API error (code %d): %s", errorResponse.Error.Code, errorResponse.Error.Message))
			}
		}
		return "", NewGeneralError(fmt.Sprintf("Google API error (status %d): %s", resp.StatusCode, string(body)))
	}

	// デバッグ用：レスポンス全体を記録
	fmt.Printf("🔍 Google API raw response: %s\n", string(body))

	var response GoogleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		switch response.Error.Code {
		case 400:
			if strings.Contains(response.Error.Message, "too many tokens") || strings.Contains(response.Error.Message, "maximum context length") {
				return "", NewTokenLimitError(fmt.Sprintf("入力テキストが長すぎます。テキストを短くして再度お試しください。詳細: %s", response.Error.Message))
			}
			return "", NewGeneralError(fmt.Sprintf("Google API リクエストエラー: %s", response.Error.Message))
		case 403:
			return "", NewInvalidAPIKeyError(fmt.Sprintf("設定を確認してください。詳細: %s", response.Error.Message))
		case 404:
			return "", NewModelNotFoundError(fmt.Sprintf("モデル「%s」が利用できません。詳細: %s", c.model, response.Error.Message))
		case 429:
			return "", NewRateLimitError(fmt.Sprintf("しばらく待ってから再試行してください。詳細: %s", response.Error.Message))
		default:
			return "", NewGeneralError(fmt.Sprintf("Google API error: %s", response.Error.Message))
		}
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from Google API")
	}

	candidate := response.Candidates[0]
	fmt.Printf("🔍 Candidate info: FinishReason=%s, Parts count=%d\n", candidate.FinishReason, len(candidate.Content.Parts))
	
	// finishReasonをチェック
	if candidate.FinishReason == "MAX_TOKENS" {
		fmt.Printf("⚠️ Google API response truncated due to MAX_TOKENS\n")
		return "", NewTokenLimitError("生成されるレスポンスが長すぎます。より短いプロンプトを使用するか、MaxOutputTokensを増やしてください。")
	}

	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned from Google API. FinishReason: %s", candidate.FinishReason)
	}

	content := candidate.Content.Parts[0].Text
	fmt.Printf("🔍 Content extracted: '%s' (length: %d)\n", content, len(content))
	
	// 空のコンテンツの場合
	if content == "" {
		return "", fmt.Errorf("empty content returned from Google API. FinishReason: %s, Parts count: %d", candidate.FinishReason, len(candidate.Content.Parts))
	}

	fmt.Printf("✅ Google API response received (length: %d, finishReason: %s)\n", len(content), candidate.FinishReason)

	return content, nil
}

func (c *googleClient) GenerateMultimodalContent(ctx context.Context, prompt string, files []FileContent) (string, error) {
	// 現在は基本的な実装として、ファイルの説明をテキストに追加してGenerateContentを呼び出し
	enhancedPrompt := prompt
	
	if len(files) > 0 {
		enhancedPrompt += "\n\n添付ファイル:\n"
		for _, file := range files {
			enhancedPrompt += fmt.Sprintf("- %s (%s, タイプ: %s)\n", file.Name, file.MimeType, file.Type)
		}
		enhancedPrompt += "\n上記のファイルについて分析・処理してください。"
	}
	
	return c.GenerateContent(ctx, enhancedPrompt)
}

func (c *googleClient) GenerateWithHistory(ctx context.Context, messages []ChatMessage) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Google API key not configured")
	}

	if c.model == "" {
		return "", fmt.Errorf("Google model not specified. Please configure your AI settings in the settings page")
	}

	fmt.Printf("🤖 Using Google API with model: %s (with %d messages in history)\n", c.model, len(messages))

	// ChatMessageをGoogleのContent形式に変換
	// Google APIは会話履歴を単一のテキストとして結合する必要がある
	var conversationText strings.Builder
	for i, msg := range messages {
		if i > 0 {
			conversationText.WriteString("\n\n")
		}
		if msg.Role == "user" {
			conversationText.WriteString("【ユーザー】\n")
		} else {
			conversationText.WriteString("【アシスタント】\n")
		}
		conversationText.WriteString(msg.Content)
	}

	// 推論トークンの設定
	// 用途: 5段階生成（新方式）- 会話履歴を使った段階的な問題生成
	// Stage 1-5の各段階で会話の文脈を理解し、適切な応答を生成
	thinkingBudget := int32(10000)
	baseOutputTokens := 10000
	maxOutputTokens := baseOutputTokens + int(thinkingBudget) // 基本出力トークン + 推論トークン

	request := GoogleRequest{
		Contents: []GoogleContent{
			{
				Parts: []GooglePart{
					{
						Text: conversationText.String(),
					},
				},
			},
		},
		GenerationConfig: GoogleGenerationConfig{
			MaxOutputTokens: maxOutputTokens,
			ThinkingConfig: &GoogleThinkingConfig{
				ThinkingBudget: &thinkingBudget,
			},
		},
	}

	fmt.Printf("🧠 [Google] Thinking budget: %d tokens, Base output: %d tokens, Total max: %d tokens (with history)\n", thinkingBudget, baseOutputTokens, maxOutputTokens)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/%s:generateContent?key=%s", c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Google API error - Status: %d, Body: %s\n", resp.StatusCode, string(body))
		var errorResponse GoogleResponse
		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error != nil {
			switch errorResponse.Error.Code {
			case 400:
				if strings.Contains(errorResponse.Error.Message, "too many tokens") || strings.Contains(errorResponse.Error.Message, "maximum context length") {
					return "", NewTokenLimitError(fmt.Sprintf("会話履歴が長すぎます。履歴を短くして再度お試しください。詳細: %s", errorResponse.Error.Message))
				}
				return "", NewGeneralError(fmt.Sprintf("Google API リクエストエラー: %s", errorResponse.Error.Message))
			case 403:
				return "", NewInvalidAPIKeyError(fmt.Sprintf("設定を確認してください。詳細: %s", errorResponse.Error.Message))
			case 404:
				return "", NewModelNotFoundError(fmt.Sprintf("モデル「%s」が利用できません。詳細: %s", c.model, errorResponse.Error.Message))
			case 429:
				return "", NewRateLimitError(fmt.Sprintf("しばらく待ってから再試行してください。詳細: %s", errorResponse.Error.Message))
			default:
				return "", NewGeneralError(fmt.Sprintf("Google API error (code %d): %s", errorResponse.Error.Code, errorResponse.Error.Message))
			}
		}
		return "", NewGeneralError(fmt.Sprintf("Google API error (status %d): %s", resp.StatusCode, string(body)))
	}

	fmt.Printf("🔍 Google API raw response: %s\n", string(body))

	var response GoogleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		switch response.Error.Code {
		case 400:
			if strings.Contains(response.Error.Message, "too many tokens") || strings.Contains(response.Error.Message, "maximum context length") {
				return "", NewTokenLimitError(fmt.Sprintf("会話履歴が長すぎます。履歴を短くして再度お試しください。詳細: %s", response.Error.Message))
			}
			return "", NewGeneralError(fmt.Sprintf("Google API リクエストエラー: %s", response.Error.Message))
		case 403:
			return "", NewInvalidAPIKeyError(fmt.Sprintf("設定を確認してください。詳細: %s", response.Error.Message))
		case 404:
			return "", NewModelNotFoundError(fmt.Sprintf("モデル「%s」が利用できません。詳細: %s", c.model, response.Error.Message))
		case 429:
			return "", NewRateLimitError(fmt.Sprintf("しばらく待ってから再試行してください。詳細: %s", response.Error.Message))
		default:
			return "", NewGeneralError(fmt.Sprintf("Google API error: %s", response.Error.Message))
		}
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from Google API")
	}

	candidate := response.Candidates[0]
	fmt.Printf("🔍 Candidate info: FinishReason=%s, Parts count=%d\n", candidate.FinishReason, len(candidate.Content.Parts))
	
	if candidate.FinishReason == "MAX_TOKENS" {
		fmt.Printf("⚠️ Google API response truncated due to MAX_TOKENS\n")
		return "", NewTokenLimitError("生成されるレスポンスが長すぎます。より短いプロンプトを使用するか、MaxOutputTokensを増やしてください。")
	}

	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned from Google API. FinishReason: %s", candidate.FinishReason)
	}

	content := candidate.Content.Parts[0].Text
	fmt.Printf("🔍 Content extracted: '%s' (length: %d)\n", content, len(content))
	
	if content == "" {
		return "", fmt.Errorf("empty content returned from Google API. FinishReason: %s, Parts count: %d", candidate.FinishReason, len(candidate.Content.Parts))
	}

	fmt.Printf("✅ Google API response with history received (length: %d, finishReason: %s)\n", len(content), candidate.FinishReason)

	return content, nil
}
