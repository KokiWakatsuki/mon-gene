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
	"regexp"
	"strings"
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
	Content     string `json:"content"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
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

// Core API関連の構造体
type CoreAnalysisRequest struct {
	ProblemText     string                 `json:"problem_text"`
	UnitParameters  map[string]interface{} `json:"unit_parameters"`
	Subject         string                 `json:"subject"`
}

type CoreAnalysisResponse struct {
	Success             bool                              `json:"success"`
	NeedsGeometry       bool                              `json:"needs_geometry"`
	DetectedShapes      []string                          `json:"detected_shapes"`
	SuggestedParameters map[string]map[string]interface{} `json:"suggested_parameters"`
}

type CoreGeometryRequest struct {
	ShapeType  string                 `json:"shape_type"`
	Parameters map[string]interface{} `json:"parameters"`
	Labels     map[string]string      `json:"labels,omitempty"`
}

type CoreGeometryResponse struct {
	Success     bool   `json:"success"`
	ImageBase64 string `json:"image_base64"`
	ShapeType   string `json:"shape_type"`
}

type CorePDFRequest struct {
	ProblemText string `json:"problem_text"`
	ImageBase64 string `json:"image_base64,omitempty"`
}

type CorePDFResponse struct {
	Success    bool   `json:"success"`
	PDFBase64  string `json:"pdf_base64"`
}

type CoreCustomGeometryRequest struct {
	PythonCode  string `json:"python_code"`
	ProblemText string `json:"problem_text"`
}

type CoreCustomGeometryResponse struct {
	Success     bool   `json:"success"`
	ImageBase64 string `json:"image_base64"`
	ProblemText string `json:"problem_text"`
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

	// Pythonコードが含まれているかチェック
	pythonCode := extractPythonCode(content)
	cleanContent := removePythonCode(content)
	
	log.Printf("Original content length: %d", len(content))
	log.Printf("Python code length: %d", len(pythonCode))
	log.Printf("Clean content length: %d", len(cleanContent))
	
	if pythonCode != "" {
		preview := pythonCode
		if len(pythonCode) > 100 {
			preview = pythonCode[:100] + "..."
		}
		log.Printf("Python code found: %s", preview)
	}
	
	var imageBase64 string
	
	if pythonCode != "" {
		// カスタムPythonコードで図形を生成
		log.Printf("Custom Python code detected, generating custom geometry")
		imageBase64, err = generateCustomGeometryWithCore(pythonCode, cleanContent)
		if err != nil {
			log.Printf("Error generating custom geometry: %v", err)
		}
	} else {
		// 従来の方法で図形が必要かどうかを分析
		analysisResp, err := analyzeProblemWithCore(cleanContent, req.Filters)
		if err != nil {
			log.Printf("Error analyzing problem: %v", err)
		} else if analysisResp.NeedsGeometry && len(analysisResp.DetectedShapes) > 0 {
			// 最初に検出された図形を描画
			shapeType := analysisResp.DetectedShapes[0]
			imageBase64, err = generateGeometryWithCore(shapeType, analysisResp.SuggestedParameters[shapeType])
			if err != nil {
				log.Printf("Error generating geometry: %v", err)
			}
		}
	}

	// 成功レスポンス
	response := GenerateProblemResponse{
		Content:     cleanContent,
		Success:     true,
		ImageBase64: imageBase64,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Claude APIを呼び出す関数
func callClaudeAPI(prompt, apiKey string) (string, error) {
	// 図形描画コード生成を含むプロンプトに拡張
	enhancedPrompt := prompt + `

もし問題に図形が必要な場合は、問題文の後に以下の形式で図形描画用のPythonコードを追加してください：

---GEOMETRY_CODE_START---
# 図形描画コード（問題に特化した図形を描画）
# import文は不要です（事前にインポート済み）
fig, ax = plt.subplots(1, 1, figsize=(8, 6))

# ここに問題に応じた具体的な図形描画コードを記述
# 例：正方形ABCD、点P、Q、Rの位置、線分、座標軸など
# 利用可能な変数: plt, patches, np, numpy

ax.set_aspect('equal')
ax.grid(True, alpha=0.3)
plt.tight_layout()
---GEOMETRY_CODE_END---

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
10. 切断図形や断面図が必要な場合は、切断面をPoly3DCollectionで描画してください`

	claudeReq := ClaudeRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 2000, // トークン数を増やす
		Messages: []Message{
			{
				Role:    "user",
				Content: enhancedPrompt,
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

// PDF生成エンドポイント
func generatePDFHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CorePDFRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received PDF generation request for problem text length: %d", len(req.ProblemText))

	// Core APIでPDF生成
	pdfBase64, err := generatePDFWithCore(req.ProblemText, req.ImageBase64)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		response := CorePDFResponse{
			Success: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 成功レスポンス
	response := CorePDFResponse{
		Success:   true,
		PDFBase64: pdfBase64,
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

// Core APIとの連携関数
func analyzeProblemWithCore(problemText string, filters map[string]interface{}) (*CoreAnalysisResponse, error) {
	coreURL := os.Getenv("CORE_API_URL")
	if coreURL == "" {
		coreURL = "http://core:1234" // Dockerコンテナ内ではサービス名を使用
	}

	reqData := CoreAnalysisRequest{
		ProblemText:    problemText,
		UnitParameters: filters,
		Subject:        "math",
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("JSON marshaling error: %v", err)
	}

	resp, err := http.Post(coreURL+"/analyze-problem", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response reading error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Core API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var analysisResp CoreAnalysisResponse
	if err := json.Unmarshal(respBody, &analysisResp); err != nil {
		return nil, fmt.Errorf("JSON unmarshaling error: %v", err)
	}

	return &analysisResp, nil
}

func generateGeometryWithCore(shapeType string, parameters map[string]interface{}) (string, error) {
	coreURL := os.Getenv("CORE_API_URL")
	if coreURL == "" {
		coreURL = "http://core:1234" // Dockerコンテナ内ではサービス名を使用
	}

	reqData := CoreGeometryRequest{
		ShapeType:  shapeType,
		Parameters: parameters,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("JSON marshaling error: %v", err)
	}

	resp, err := http.Post(coreURL+"/draw-geometry", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response reading error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Core API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var geometryResp CoreGeometryResponse
	if err := json.Unmarshal(respBody, &geometryResp); err != nil {
		return "", fmt.Errorf("JSON unmarshaling error: %v", err)
	}

	return geometryResp.ImageBase64, nil
}

func generatePDFWithCore(problemText, imageBase64 string) (string, error) {
	coreURL := os.Getenv("CORE_API_URL")
	if coreURL == "" {
		coreURL = "http://core:1234" // Dockerコンテナ内ではサービス名を使用
	}

	reqData := CorePDFRequest{
		ProblemText: problemText,
		ImageBase64: imageBase64,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("JSON marshaling error: %v", err)
	}

	resp, err := http.Post(coreURL+"/generate-pdf", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response reading error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Core API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var pdfResp CorePDFResponse
	if err := json.Unmarshal(respBody, &pdfResp); err != nil {
		return "", fmt.Errorf("JSON unmarshaling error: %v", err)
	}

	return pdfResp.PDFBase64, nil
}

// Pythonコードを抽出する関数
func extractPythonCode(content string) string {
	re := regexp.MustCompile(`(?s)---GEOMETRY_CODE_START---(.*?)---GEOMETRY_CODE_END---`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// 問題文からPythonコードを除去する関数
func removePythonCode(content string) string {
	re := regexp.MustCompile(`(?s)---GEOMETRY_CODE_START---.*?---GEOMETRY_CODE_END---`)
	return strings.TrimSpace(re.ReplaceAllString(content, ""))
}

// カスタム図形生成関数
func generateCustomGeometryWithCore(pythonCode, problemText string) (string, error) {
	coreURL := os.Getenv("CORE_API_URL")
	if coreURL == "" {
		coreURL = "http://core:1234" // Dockerコンテナ内ではサービス名を使用
	}

	reqData := CoreCustomGeometryRequest{
		PythonCode:  pythonCode,
		ProblemText: problemText,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("JSON marshaling error: %v", err)
	}

	resp, err := http.Post(coreURL+"/draw-custom-geometry", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response reading error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Core API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var customGeometryResp CoreCustomGeometryResponse
	if err := json.Unmarshal(respBody, &customGeometryResp); err != nil {
		return "", fmt.Errorf("JSON unmarshaling error: %v", err)
	}

	return customGeometryResp.ImageBase64, nil
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
	http.HandleFunc("/api/generate-pdf", generatePDFHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/forgot-password", forgotPasswordHandler)

	log.Println("Starting Mongene Backend Server on :8080")
	log.Println("Endpoints:")
	log.Println("  GET  / - Health check")
	log.Println("  POST /api/generate-problem - Generate problem using Claude API")
	log.Println("  POST /api/generate-pdf - Generate PDF with problem and image")
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
