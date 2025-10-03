package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
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

// 認証関連の構造体
type LoginRequest struct {
	SchoolCode string `json:"schoolCode"`
	Password   string `json:"password"`
	Remember   bool   `json:"remember"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ForgotPasswordRequest struct {
	SchoolCode string `json:"schoolCode"`
}

type ForgotPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type School struct {
	Code     string `json:"code"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// セッション管理用
type Session struct {
	Token      string    `json:"token"`
	SchoolCode string    `json:"schoolCode"`
	ExpiresAt  time.Time `json:"expiresAt"`
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

// インメモリデータストア（本来はデータベースを使用）
var schools = map[string]School{
	"00000": {
		Code:     "00000",
		Password: "password",
		Email:    "nutfes.script@gmail.com",
	},
}

var sessions = make(map[string]Session)

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
		log.Printf("Error: CLAUDE_API_KEY environment variable is not set")
		response := GenerateProblemResponse{
			Content: "",
			Success: false,
			Error:   "APIキーが設定されていません",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
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

// ランダムトークン生成
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// メール送信機能
func sendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// 設定値の詳細チェック
	if from == "" {
		log.Printf("SMTP設定エラー: SMTP_FROM が設定されていません")
	}
	if password == "" || password == "your-gmail-app-password" {
		log.Printf("SMTP設定エラー: SMTP_PASSWORD が設定されていないか、デフォルト値のままです")
	}
	if smtpHost == "" {
		log.Printf("SMTP設定エラー: SMTP_HOST が設定されていません")
	}
	if smtpPort == "" {
		log.Printf("SMTP設定エラー: SMTP_PORT が設定されていません")
	}

	if from == "" || password == "" || password == "your-gmail-app-password" || smtpHost == "" || smtpPort == "" {
		log.Printf("SMTP設定が不完全です。デモ用にログ出力します。")
		log.Printf("To: %s", to)
		log.Printf("Subject: %s", subject)
		log.Printf("Body: %s", body)
		log.Printf("")
		log.Printf("実際のメール送信を有効にするには:")
		log.Printf("1. Gmailでアプリパスワードを生成してください")
		log.Printf("2. .envファイルのSMTP_PASSWORDを実際のアプリパスワードに変更してください")
		return nil
	}

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	
	if err != nil {
		log.Printf("メール送信エラー: %v", err)
		return err
	}
	
	log.Printf("メールを正常に送信しました: %s", to)
	return nil
}

// ログインエンドポイント
func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	var req LoginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 認証チェック
	school, exists := schools[req.SchoolCode]
	if !exists || school.Password != req.Password {
		response := LoginResponse{
			Success: false,
			Error:   "塾コードまたはパスワードが正しくありません",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// トークン生成
	token, err := generateToken()
	if err != nil {
		log.Printf("Error generating token: %v", err)
		response := LoginResponse{
			Success: false,
			Error:   "認証トークンの生成に失敗しました",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// セッション保存（24時間有効）
	expiresAt := time.Now().Add(24 * time.Hour)
	if req.Remember {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30日間
	}

	sessions[token] = Session{
		Token:      token,
		SchoolCode: req.SchoolCode,
		ExpiresAt:  expiresAt,
	}

	response := LoginResponse{
		Success: true,
		Token:   token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// パスワードリセットエンドポイント
func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	var req ForgotPasswordRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 塾コードの存在確認
	school, exists := schools[req.SchoolCode]
	if !exists {
		response := ForgotPasswordResponse{
			Success: false,
			Error:   "指定された塾コードが見つかりません",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// メール送信
	subject := "【Mongene】パスワードのお知らせ"
	emailBody := fmt.Sprintf(`
こんにちは、

お忘れになったパスワードをお知らせいたします。

塾コード: %s
パスワード: %s

今後ともMongeneをよろしくお願いいたします。

Mongeneサポートチーム
`, school.Code, school.Password)

	err = sendEmail(school.Email, subject, emailBody)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		response := ForgotPasswordResponse{
			Success: false,
			Error:   "メールの送信に失敗しました",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ForgotPasswordResponse{
		Success: true,
		Message: "パスワードを記載したメールを送信しました",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 認証チェック関数
func isAuthenticated(token string) bool {
	session, exists := sessions[token]
	if !exists {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(sessions, token)
		return false
	}

	return true
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
	// .envファイルを読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// ルートの設定
	http.HandleFunc("/", healthHandler)
	http.HandleFunc("/api/generate-problem", generateProblemHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/forgot-password", forgotPasswordHandler)

	log.Println("Starting Mongene Backend Server on :8080")
	log.Println("Endpoints:")
	log.Println("  GET  / - Health check")
	log.Println("  POST /api/generate-problem - Generate problem using Claude API")
	log.Println("  POST /api/login - User authentication")
	log.Println("  POST /api/forgot-password - Password reset")

	server := http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
