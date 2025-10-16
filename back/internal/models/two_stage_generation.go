package models

import "time"

// TwoStageGenerationRequest 2段階生成リクエスト
type TwoStageGenerationRequest struct {
	Prompt  string                 `json:"prompt" validate:"required"`
	Subject string                 `json:"subject" validate:"required"`
	Filters map[string]interface{} `json:"filters"`
}

// FirstStageResponse 1回目API呼び出しのレスポンス
type FirstStageResponse struct {
	Success      bool   `json:"success"`
	ProblemText  string `json:"problem_text"`
	GeometryCode string `json:"geometry_code"`
	ImageBase64  string `json:"image_base64,omitempty"`
	Error        string `json:"error,omitempty"`
	Log          string `json:"log"`
}

// SecondStageRequest 2回目API呼び出しのリクエスト
type SecondStageRequest struct {
	ProblemText  string `json:"problem_text" validate:"required"`
	GeometryCode string `json:"geometry_code,omitempty"`
}

// SecondStageResponse 2回目API呼び出しのレスポンス
type SecondStageResponse struct {
	Success             bool   `json:"success"`
	SolutionSteps       string `json:"solution_steps"`
	CalculationProgram  string `json:"calculation_program"`
	FinalSolution       string `json:"final_solution"`
	CalculationResults  string `json:"calculation_results"`
	Error               string `json:"error,omitempty"`
	Log                 string `json:"log"`
}

// TwoStageGenerationResponse 全体のレスポンス
type TwoStageGenerationResponse struct {
	Success             bool   `json:"success"`
	ProblemText         string `json:"problem_text"`
	ImageBase64         string `json:"image_base64,omitempty"`
	SolutionSteps       string `json:"solution_steps"`
	FinalSolution       string `json:"final_solution"`
	CalculationResults  string `json:"calculation_results"`
	Error               string `json:"error,omitempty"`
	FirstStageLog       string `json:"first_stage_log"`
	SecondStageLog      string `json:"second_stage_log"`
	GeometryCode        string `json:"geometry_code,omitempty"`
	CalculationProgram  string `json:"calculation_program,omitempty"`
}

// GenerationLog 生成プロセスのログ
type GenerationLog struct {
	Timestamp   time.Time `json:"timestamp"`
	Stage       string    `json:"stage"` // "first" or "second"
	Message     string    `json:"message"`
	APIProvider string    `json:"api_provider,omitempty"`
	Model       string    `json:"model,omitempty"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
}
