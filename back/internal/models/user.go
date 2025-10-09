package models

import "time"

type User struct {
	ID                    int64     `json:"id" db:"id"`
	SchoolCode           string    `json:"school_code" db:"school_code"`
	PasswordHash         string    `json:"-" db:"password_hash"`
	Email                string    `json:"email" db:"email"`
	ProblemGenerationLimit int     `json:"problem_generation_limit" db:"problem_generation_limit"` // -1 = 制限なし, 0以上 = 制限回数
	ProblemGenerationCount int     `json:"problem_generation_count" db:"problem_generation_count"` // 現在の生成回数
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type LoginRequest struct {
	SchoolCode string `json:"schoolCode" validate:"required"`
	Password   string `json:"password" validate:"required"`
	Remember   bool   `json:"remember"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ForgotPasswordRequest struct {
	SchoolCode string `json:"schoolCode" validate:"required"`
}

type ForgotPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
