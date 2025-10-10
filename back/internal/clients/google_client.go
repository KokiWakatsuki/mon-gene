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
			MaxOutputTokens: 8000,
		},
	}

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
			case 404:
				return "", fmt.Errorf("指定されたGoogleモデル「%s」が見つかりません。利用可能なモデルを確認してください。エラー: %s", c.model, errorResponse.Error.Message)
			case 403:
				return "", fmt.Errorf("Google APIキーが無効または権限がありません。設定を確認してください。エラー: %s", errorResponse.Error.Message)
			case 429:
				return "", fmt.Errorf("Google APIのレート制限に達しました。しばらく待ってから再試行してください。エラー: %s", errorResponse.Error.Message)
			default:
				return "", fmt.Errorf("Google API error (code %d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
			}
		}
		return "", fmt.Errorf("Google API error (status %d): %s", resp.StatusCode, string(body))
	}

	// デバッグ用：レスポンス全体を記録
	fmt.Printf("🔍 Google API raw response: %s\n", string(body))

	var response GoogleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("Google API error: %s", response.Error.Message)
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from Google API")
	}

	candidate := response.Candidates[0]
	fmt.Printf("🔍 Candidate info: FinishReason=%s, Parts count=%d\n", candidate.FinishReason, len(candidate.Content.Parts))
	
	// finishReasonをチェック
	if candidate.FinishReason == "MAX_TOKENS" {
		fmt.Printf("⚠️ Google API response truncated due to MAX_TOKENS\n")
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
