package models

import "time"

type Problem struct {
	ID          int64                  `json:"id" db:"id"`
	UserID      int64                  `json:"user_id" db:"user_id"`
	Subject     string                 `json:"subject" db:"subject"`
	Prompt      string                 `json:"prompt" db:"prompt"`                           // 生成時のプロンプト
	Content     string                 `json:"content" db:"content"`                         // 問題文
	Solution    string                 `json:"solution,omitempty" db:"solution"`             // 解答
	ImageBase64 string                 `json:"image_base64,omitempty" db:"image_base64"`     // 図
	Filters     map[string]interface{} `json:"filters" db:"filters"`                         // 生成パラメータ
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type GenerateProblemRequest struct {
	Prompt  string                 `json:"prompt" validate:"required"`
	Subject string                 `json:"subject" validate:"required"`
	Filters map[string]interface{} `json:"filters"`
}

type GenerateProblemResponse struct {
	Content     string `json:"content"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
	Solution    string `json:"solution,omitempty"`
}

type PDFGenerateRequest struct {
	ProblemText  string `json:"problem_text" validate:"required"`
	ImageBase64  string `json:"image_base64,omitempty"`
	SolutionText string `json:"solution_text,omitempty"`
}

type PDFGenerateResponse struct {
	Success   bool   `json:"success"`
	PDFBase64 string `json:"pdf_base64,omitempty"`
	Error     string `json:"error,omitempty"`
}

type UpdateProblemRequest struct {
	ID       int64  `json:"id" validate:"required"`
	Content  string `json:"content" validate:"required"`
	Solution string `json:"solution,omitempty"`
}

type UpdateProblemResponse struct {
	Success bool     `json:"success"`
	Problem *Problem `json:"problem,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type RegenerateGeometryRequest struct {
	ID      int64  `json:"id" validate:"required"`
	Content string `json:"content,omitempty"`
}

type RegenerateGeometryResponse struct {
	Success     bool   `json:"success"`
	ImageBase64 string `json:"image_base64,omitempty"`
	Error       string `json:"error,omitempty"`
}
