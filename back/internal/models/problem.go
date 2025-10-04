package models

import "time"

type Problem struct {
	ID          int64                  `json:"id" db:"id"`
	UserID      int64                  `json:"user_id" db:"user_id"`
	Subject     string                 `json:"subject" db:"subject"`
	Content     string                 `json:"content" db:"content"`
	Solution    string                 `json:"solution,omitempty" db:"solution"`
	ImageBase64 string                 `json:"image_base64,omitempty" db:"image_base64"`
	Filters     map[string]interface{} `json:"filters" db:"filters"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
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
}

type PDFGenerateRequest struct {
	ProblemText string `json:"problem_text" validate:"required"`
	ImageBase64 string `json:"image_base64,omitempty"`
}

type PDFGenerateResponse struct {
	Success   bool   `json:"success"`
	PDFBase64 string `json:"pdf_base64,omitempty"`
	Error     string `json:"error,omitempty"`
}
