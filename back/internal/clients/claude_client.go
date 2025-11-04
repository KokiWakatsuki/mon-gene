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

type claudeClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []Message       `json:"messages"`
	Thinking  *ThinkingConfig `json:"thinking,omitempty"`
}

type ThinkingConfig struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ImageContent struct {
	Type   string      `json:"type"`
	Source ImageSource `json:"source"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type ClaudeResponse struct {
	Content []ContentBlock `json:"content"`
	Usage   Usage          `json:"usage"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func NewClaudeClient(model string) ClaudeClient {
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		fmt.Printf("⚠️ CLAUDE_API_KEY not found in environment variables\n")
	}
	
	// モデル名が空の場合はデフォルトを使用しない
	if model == "" {
		fmt.Printf("⚠️ Claude model not specified\n")
	}
	
	return &claudeClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1/messages",
		client:  &http.Client{},
	}
}

func (c *claudeClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Claude API key not configured")
	}

	if c.model == "" {
		return "", fmt.Errorf("Claude model not specified. Please configure your AI settings in the settings page")
	}

	fmt.Printf("🤖 Using Claude API with model: %s\n", c.model)

	// 推論トークンの設定
	// 用途: 1段階生成（旧方式）- 単一プロンプトから問題を一度に生成
	thinkingBudget := 10000
	maxTokens := 10000 + thinkingBudget // 推論トークン + 10000

	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Thinking: &ThinkingConfig{
			Type:         "enabled",
			BudgetTokens: thinkingBudget,
		},
	}

	fmt.Printf("🧠 Using thinking tokens: %d, max tokens: %d\n", thinkingBudget, maxTokens)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// より詳細なエラー情報を提供
		var errorData map[string]interface{}
		if err := json.Unmarshal(body, &errorData); err == nil {
			if errorObj, exists := errorData["error"]; exists {
				if errorMap, ok := errorObj.(map[string]interface{}); ok {
					errorType := ""
					errorMessage := ""
					if t, exists := errorMap["type"]; exists {
						if typeStr, ok := t.(string); ok {
							errorType = typeStr
						}
					}
					if m, exists := errorMap["message"]; exists {
						if msgStr, ok := m.(string); ok {
							errorMessage = msgStr
						}
					}

					switch errorType {
					case "invalid_request_error":
						if strings.Contains(errorMessage, "maximum context length") || strings.Contains(errorMessage, "too many tokens") {
							return "", NewTokenLimitError(fmt.Sprintf("入力テキストが長すぎます。テキストを短くして再度お試しください。詳細: %s", errorMessage))
						}
						return "", NewGeneralError(fmt.Sprintf("Claude API リクエストエラー: %s", errorMessage))
					case "authentication_error":
						return "", NewInvalidAPIKeyError(fmt.Sprintf("設定を確認してください。詳細: %s", errorMessage))
					case "permission_error":
						return "", NewInvalidAPIKeyError(fmt.Sprintf("APIキーの権限を確認してください。詳細: %s", errorMessage))
					case "rate_limit_error":
						return "", NewRateLimitError(fmt.Sprintf("しばらく待ってから再試行してください。詳細: %s", errorMessage))
					case "api_error", "overloaded_error":
						return "", NewGeneralError(fmt.Sprintf("Claude APIサーバーエラー: %s", errorMessage))
					default:
						return "", NewGeneralError(fmt.Sprintf("Claude API error (%s): %s", errorType, errorMessage))
					}
				}
			}
		}
		return "", NewGeneralError(fmt.Sprintf("Claude API error (status %d): %s", resp.StatusCode, string(body)))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content returned from Claude API")
	}

	content := claudeResp.Content[0].Text
	fmt.Printf("✅ Claude API response received (length: %d)\n", len(content))

	return content, nil
}

func (c *claudeClient) GenerateMultimodalContent(ctx context.Context, prompt string, files []FileContent) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Claude API key not configured")
	}

	if c.model == "" {
		return "", fmt.Errorf("Claude model not specified. Please configure your AI settings in the settings page")
	}

	fmt.Printf("🤖 Using Claude API with model: %s (multimodal with %d files)\n", c.model, len(files))

	// コンテンツの配列を構築
	var contentArray []interface{}
	
	// テキストコンテンツを追加
	contentArray = append(contentArray, TextContent{
		Type: "text",
		Text: prompt,
	})

	// ファイルコンテンツを追加
	for _, file := range files {
		if file.Type == "image" && strings.HasPrefix(file.MimeType, "image/") {
			// 画像の場合
			contentArray = append(contentArray, ImageContent{
				Type: "image",
				Source: ImageSource{
					Type:      "base64",
					MediaType: file.MimeType,
					Data:      file.Data,
				},
			})
		} else {
			// その他のファイルの場合、テキストとして説明を追加
			contentArray = append(contentArray, TextContent{
				Type: "text",
				Text: fmt.Sprintf("\n\n[添付ファイル: %s (%s)]\nファイルの内容について分析してください。", file.Name, file.MimeType),
			})
		}
	}

	// 推論トークンの設定
	// 用途: マルチモーダル生成 - 画像やファイルを含む入力からの問題生成
	// 画像解析には追加の推論が必要なため、十分なトークン数を確保
	thinkingBudget := 10000
	baseOutputTokens := 10000
	maxTokens := baseOutputTokens + thinkingBudget // 基本出力トークン + 推論トークン

	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: contentArray,
			},
		},
		Thinking: &ThinkingConfig{
			Type:         "enabled",
			BudgetTokens: thinkingBudget,
		},
	}

	fmt.Printf("🧠 [Claude] Thinking budget: %d tokens, Base output: %d tokens, Total max: %d tokens (multimodal)\n", thinkingBudget, baseOutputTokens, maxTokens)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal multimodal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create multimodal request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send multimodal request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read multimodal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// より詳細なエラー情報を提供
		var errorData map[string]interface{}
		if err := json.Unmarshal(body, &errorData); err == nil {
			if errorObj, exists := errorData["error"]; exists {
				if errorMap, ok := errorObj.(map[string]interface{}); ok {
					errorType := ""
					errorMessage := ""
					if t, exists := errorMap["type"]; exists {
						if typeStr, ok := t.(string); ok {
							errorType = typeStr
						}
					}
					if m, exists := errorMap["message"]; exists {
						if msgStr, ok := m.(string); ok {
							errorMessage = msgStr
						}
					}

					switch errorType {
					case "invalid_request_error":
						if strings.Contains(errorMessage, "maximum context length") || strings.Contains(errorMessage, "too many tokens") {
							return "", NewTokenLimitError(fmt.Sprintf("入力テキストまたは画像が大きすぎます。サイズを小さくして再度お試しください。詳細: %s", errorMessage))
						}
						return "", NewGeneralError(fmt.Sprintf("Claude API リクエストエラー: %s", errorMessage))
					case "authentication_error":
						return "", NewInvalidAPIKeyError(fmt.Sprintf("設定を確認してください。詳細: %s", errorMessage))
					case "permission_error":
						return "", NewInvalidAPIKeyError(fmt.Sprintf("APIキーの権限を確認してください。詳細: %s", errorMessage))
					case "rate_limit_error":
						return "", NewRateLimitError(fmt.Sprintf("しばらく待ってから再試行してください。詳細: %s", errorMessage))
					case "api_error", "overloaded_error":
						return "", NewGeneralError(fmt.Sprintf("Claude APIサーバーエラー: %s", errorMessage))
					default:
						return "", NewGeneralError(fmt.Sprintf("Claude API error (%s): %s", errorType, errorMessage))
					}
				}
			}
		}
		return "", NewGeneralError(fmt.Sprintf("Claude API multimodal error (status %d): %s", resp.StatusCode, string(body)))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal multimodal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content returned from Claude API multimodal")
	}

	content := claudeResp.Content[0].Text
	fmt.Printf("✅ Claude API multimodal response received (length: %d)\n", len(content))

	return content, nil
}

func (c *claudeClient) GenerateWithHistory(ctx context.Context, messages []ChatMessage) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Claude API key not configured")
	}

	if c.model == "" {
		return "", fmt.Errorf("Claude model not specified. Please configure your AI settings in the settings page")
	}

	fmt.Printf("🤖 Using Claude API with model: %s (with %d messages in history)\n", c.model, len(messages))

	// ChatMessageをClaudeのMessage形式に変換
	claudeMessages := make([]Message, 0, len(messages))
	for _, msg := range messages {
		claudeMessages = append(claudeMessages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 推論トークンの設定
	// 用途: 5段階生成（新方式）- 会話履歴を使った段階的な問題生成
	// Stage 1-5の各段階で会話の文脈を理解し、適切な応答を生成
	thinkingBudget := 10000
	baseOutputTokens := 10000
	maxTokens := baseOutputTokens + thinkingBudget // 基本出力トークン + 推論トークン

	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		Messages:  claudeMessages,
		Thinking: &ThinkingConfig{
			Type:         "enabled",
			BudgetTokens: thinkingBudget,
		},
	}

	fmt.Printf("🧠 [Claude] Thinking budget: %d tokens, Base output: %d tokens, Total max: %d tokens (with history)\n", thinkingBudget, baseOutputTokens, maxTokens)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorData map[string]interface{}
		if err := json.Unmarshal(body, &errorData); err == nil {
			if errorObj, exists := errorData["error"]; exists {
				if errorMap, ok := errorObj.(map[string]interface{}); ok {
					errorType := ""
					errorMessage := ""
					if t, exists := errorMap["type"]; exists {
						if typeStr, ok := t.(string); ok {
							errorType = typeStr
						}
					}
					if m, exists := errorMap["message"]; exists {
						if msgStr, ok := m.(string); ok {
							errorMessage = msgStr
						}
					}

					switch errorType {
					case "invalid_request_error":
						if strings.Contains(errorMessage, "maximum context length") || strings.Contains(errorMessage, "too many tokens") {
							return "", NewTokenLimitError(fmt.Sprintf("会話履歴が長すぎます。履歴を短くして再度お試しください。詳細: %s", errorMessage))
						}
						return "", NewGeneralError(fmt.Sprintf("Claude API リクエストエラー: %s", errorMessage))
					case "authentication_error":
						return "", NewInvalidAPIKeyError(fmt.Sprintf("設定を確認してください。詳細: %s", errorMessage))
					case "permission_error":
						return "", NewInvalidAPIKeyError(fmt.Sprintf("APIキーの権限を確認してください。詳細: %s", errorMessage))
					case "rate_limit_error":
						return "", NewRateLimitError(fmt.Sprintf("しばらく待ってから再試行してください。詳細: %s", errorMessage))
					case "api_error", "overloaded_error":
						return "", NewGeneralError(fmt.Sprintf("Claude APIサーバーエラー: %s", errorMessage))
					default:
						return "", NewGeneralError(fmt.Sprintf("Claude API error (%s): %s", errorType, errorMessage))
					}
				}
			}
		}
		return "", NewGeneralError(fmt.Sprintf("Claude API error (status %d): %s", resp.StatusCode, string(body)))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content returned from Claude API")
	}

	content := claudeResp.Content[0].Text
	fmt.Printf("✅ Claude API response with history received (length: %d)\n", len(content))

	return content, nil
}
