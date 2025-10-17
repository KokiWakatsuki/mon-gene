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
	// opinion.md基準の評価データ（従来のfiltersを削除し、opinion_profileのみ使用）
	OpinionProfile *OpinionProfile `json:"opinion_profile,omitempty" db:"opinion_profile"` // opinion.md基準のプロファイル
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// OpinionProfile は opinion.md の評価基準に基づく問題プロファイル
type OpinionProfile struct {
	Domain             int    `json:"domain"`               // 出題分野コード (1-6)
	SkillLevel         int    `json:"skill_level"`          // コアスキル評価 (1-10)
	StructureComplexity [2]int `json:"structure_complexity"` // 問題構造評価 [A, B] (各1-10)
	DifficultyScore    int    `json:"difficulty_score"`     // 総合難易度スコア (1-20)
}

type GenerateProblemRequest struct {
	Prompt         string          `json:"prompt" validate:"required"`
	Subject        string          `json:"subject" validate:"required"`
	OpinionProfile *OpinionProfile `json:"opinion_profile,omitempty"` // opinion.md基準での問題生成（従来のfiltersを削除）
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
