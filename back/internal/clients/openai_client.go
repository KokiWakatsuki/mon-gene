package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type openAIClient struct {
	apiKey string
	model  string
}

type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	MaxTokens int            `json:"max_tokens"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message OpenAIMessage `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func NewOpenAIClient(model string) OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Printf("⚠️ OPENAI_API_KEY not found in environment variables\n")
	}
	
	// モデル名が空の場合はデフォルトを使用しない
	if model == "" {
		fmt.Printf("⚠️ OpenAI model not specified\n")
	}
	
	// フロントエンド設定のモデル名を実際のAPIモデル名にマッピング
	modelMapping := map[string]string{
		"gpt-5":           "gpt-4o",
		"gpt-4.1":         "gpt-4o",
		"gpt-4.5":         "gpt-4o",
		"o3-pro":          "gpt-4o",
		"o4-mini-high":    "gpt-3.5-turbo",
	}
	
	if mappedModel, exists := modelMapping[model]; exists {
		fmt.Printf("🔄 Mapping OpenAI model '%s' to '%s'\n", model, mappedModel)
		model = mappedModel
	}
	
	return &openAIClient{
		apiKey: apiKey,
		model:  model,
	}
}

func (c *openAIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	if c.model == "" {
		return "", fmt.Errorf("OpenAI model not specified. Please configure your AI settings in the settings page")
	}

	fmt.Printf("🤖 Using OpenAI API with model: %s\n", c.model)

	request := OpenAIRequest{
		Model: c.model,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 2000,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		// より詳細なエラー情報を提供
		var errorResponse OpenAIResponse
		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error != nil {
			switch errorResponse.Error.Code {
			case "insufficient_quota":
				return "", fmt.Errorf("OpenAI APIのクォータが不足しています。プランと請求詳細を確認してください。エラー: %s", errorResponse.Error.Message)
			case "invalid_api_key":
				return "", fmt.Errorf("OpenAI APIキーが無効です。設定を確認してください。エラー: %s", errorResponse.Error.Message)
			default:
				return "", fmt.Errorf("OpenAI API error (%s): %s", errorResponse.Error.Code, errorResponse.Error.Message)
			}
		}
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI API")
	}

	content := response.Choices[0].Message.Content
	fmt.Printf("✅ OpenAI API response received (length: %d)\n", len(content))

	return content, nil
}
