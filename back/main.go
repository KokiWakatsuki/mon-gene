package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// リクエスト構造体
type GenerateProblemRequest struct {
	Prompt  string                 `json:"prompt"`
	Subject string                 `json:"subject"`
	Filters map[string]interface{} `json:"filters"`
}

// レスポンス構造体
type GenerateProblemResponse struct {
	Content string `json:"content"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Claude APIリクエスト構造体
type ClaudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Claude APIレスポンス構造体
type ClaudeResponse struct {
	Content []ContentItem `json:"content"`
	Error   *ClaudeError  `json:"error,omitempty"`
}

type ContentItem struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// CORS設定
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// 問題生成エンドポイント
func generateProblemHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	// OPTIONSリクエストの処理
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// リクエストボディの読み取り
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	var req GenerateProblemRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received request - Subject: %s, Prompt: %s", req.Subject, req.Prompt)

	// Claude APIキーの取得
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		// 環境変数が設定されていない場合はハードコードされた値を使用（開発用）
		apiKey = "<REDACTED>"
	}

	// Claude APIの呼び出し
	content, err := callClaudeAPI(req.Prompt, apiKey)
	if err != nil {
		log.Printf("Error calling Claude API: %v", err)
		response := GenerateProblemResponse{
			Content: "",
			Success: false,
			Error:   fmt.Sprintf("Claude API呼び出しエラー: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 成功レスポンス
	response := GenerateProblemResponse{
		Content: content,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Claude APIを呼び出す関数
func callClaudeAPI(prompt, apiKey string) (string, error) {
	claudeReq := ClaudeRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1000,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return "", fmt.Errorf("JSON marshaling error: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("request creation error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response reading error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", fmt.Errorf("JSON unmarshaling error: %v", err)
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("Claude API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude API")
	}

	return claudeResp.Content[0].Text, nil
}

// ヘルスチェックエンドポイント
func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"message": "Mongene Backend Server is running",
	})
}

func main() {
	// ルートの設定
	http.HandleFunc("/", healthHandler)
	http.HandleFunc("/api/generate-problem", generateProblemHandler)

	log.Println("Starting Mongene Backend Server on :8080")
	log.Println("Endpoints:")
	log.Println("  GET  / - Health check")
	log.Println("  POST /api/generate-problem - Generate problem using Claude API")

	server := http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
