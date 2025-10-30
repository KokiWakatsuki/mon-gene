package models

// TwoStageGenerationRequest 2段階生成のリクエスト
type TwoStageGenerationRequest struct {
	Prompt         string          `json:"prompt"`
	Subject        string          `json:"subject"`
	OpinionProfile *OpinionProfile `json:"opinion_profile,omitempty"`
}

// TwoStageGenerationResponse 2段階生成の最終レスポンス
type TwoStageGenerationResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	
	// 生成結果
	ProblemText         string `json:"problem_text"`
	ImageBase64         string `json:"image_base64"`
	SolutionSteps       string `json:"solution_steps"`
	FinalSolution       string `json:"final_solution"`
	CalculationResults  string `json:"calculation_results"`
	GeometryCode        string `json:"geometry_code"`
	CalculationProgram  string `json:"calculation_program"`
	
	// ログ
	FirstStageLog  string `json:"first_stage_log"`
	SecondStageLog string `json:"second_stage_log"`
}

// FirstStageResponse 1回目API呼び出しのレスポンス
type FirstStageResponse struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	ProblemText  string `json:"problem_text"`
	GeometryCode string `json:"geometry_code"`
	ImageBase64  string `json:"image_base64"`
	Log          string `json:"log"`
}

// SecondStageRequest 2回目API呼び出しのリクエスト
type SecondStageRequest struct {
	ProblemText  string `json:"problem_text"`
	GeometryCode string `json:"geometry_code,omitempty"`
}

// SecondStageResponse 2回目API呼び出しのレスポンス
type SecondStageResponse struct {
	Success             bool   `json:"success"`
	Error               string `json:"error,omitempty"`
	SolutionSteps       string `json:"solution_steps"`
	CalculationProgram  string `json:"calculation_program"`
	FinalSolution       string `json:"final_solution"`
	CalculationResults  string `json:"calculation_results"`
	Log                 string `json:"log"`
}

// FiveStageGenerationRequest 5段階生成のリクエスト
type FiveStageGenerationRequest struct {
	Prompt         string          `json:"prompt"`
	Subject        string          `json:"subject"`
	OpinionProfile *OpinionProfile `json:"opinion_profile,omitempty"`
}

// FiveStageGenerationResponse 5段階生成の最終レスポンス（修正版）
type FiveStageGenerationResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	
	// 各段階の結果（修正後の順序に対応）
	SubProblemsAndProcess string `json:"sub_problems_and_process"` // Stage1: 小問構成と解答プロセス
	CompleteProblem       string `json:"complete_problem"`        // Stage2: 完全な問題（大問と小問）
	CalculationProgram    string `json:"calculation_program"`     // Stage3: 数値計算プログラム
	CalculationResults    string `json:"calculation_results"`     // Stage3: 数値計算結果
	FinalExplanation      string `json:"final_explanation"`       // Stage4: 完全な解答・解説
	GeometryCode          string `json:"geometry_code"`           // Stage5: 図形描画プログラム
	ImageBase64           string `json:"image_base64"`            // Stage5: 図形画像
	
	// 各段階のログ
	Stage1Log string `json:"stage1_log"`
	Stage2Log string `json:"stage2_log"`
	Stage3Log string `json:"stage3_log"`
	Stage4Log string `json:"stage4_log"`
	Stage5Log string `json:"stage5_log"`
}

// Stage1Request 1段階目のリクエスト（小問構成と解答プロセス生成）
type Stage1Request struct {
	Prompt  string `json:"prompt"`
	Subject string `json:"subject"`
}

// Stage1Response 1段階目のレスポンス（小問構成と解答プロセス生成）
type Stage1Response struct {
	Success               bool   `json:"success"`
	Error                 string `json:"error,omitempty"`
	SubProblemsAndProcess string `json:"sub_problems_and_process"` // 小問構成と解答プロセス
	Log                   string `json:"log"`
}

// Stage2Request 2段階目のリクエスト（完全な問題生成）
type Stage2Request struct {
	SubProblemsAndProcess string `json:"sub_problems_and_process"` // Stage1で生成された小問構成と解答プロセス
}

// Stage2Response 2段階目のレスポンス（完全な問題生成）
type Stage2Response struct {
	Success         bool   `json:"success"`
	Error           string `json:"error,omitempty"`
	CompleteProblem string `json:"complete_problem"` // 完全な問題（大問と小問）
	Log             string `json:"log"`
}

// Stage3Request 3段階目のリクエスト（数値計算プログラム生成・実行）
type Stage3Request struct {
	CompleteProblem       string `json:"complete_problem"`        // Stage2の完全な問題
	SubProblemsAndProcess string `json:"sub_problems_and_process"` // Stage1の解答プロセス
}

// Stage3Response 3段階目のレスポンス（数値計算プログラム生成・実行）
type Stage3Response struct {
	Success            bool   `json:"success"`
	Error              string `json:"error,omitempty"`
	CalculationProgram string `json:"calculation_program"` // 数値計算プログラム
	CalculationResults string `json:"calculation_results"` // 計算結果
	Log                string `json:"log"`
}

// Stage4Request 4段階目のリクエスト（完全な解答・解説生成）
type Stage4Request struct {
	CompleteProblem       string `json:"complete_problem"`        // Stage2の完全な問題
	SubProblemsAndProcess string `json:"sub_problems_and_process"` // Stage1の解答プロセス
	CalculationResults    string `json:"calculation_results"`     // Stage3の計算結果
}

// Stage4Response 4段階目のレスポンス（完全な解答・解説生成）
type Stage4Response struct {
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
	FinalExplanation string `json:"final_explanation"` // 完全な解答・解説
	Log              string `json:"log"`
}

// Stage5Request 5段階目のリクエスト（図形描画プログラム生成・実行）
type Stage5Request struct {
	CompleteProblem       string `json:"complete_problem"`        // Stage2の完全な問題
	SubProblemsAndProcess string `json:"sub_problems_and_process"` // Stage1の解答プロセス
	CalculationResults    string `json:"calculation_results"`     // Stage3の計算結果
	FinalExplanation      string `json:"final_explanation"`       // Stage4の完全な解答・解説
	
	// 5段階生成完了後のDB保存用（オプション）
	FiveStageData *FiveStageDataForSave `json:"five_stage_data,omitempty"`
}

// FiveStageDataForSave 5段階生成完了後のDB保存用データ
type FiveStageDataForSave struct {
	Prompt         string          `json:"prompt"`
	Subject        string          `json:"subject"`
	OpinionProfile *OpinionProfile `json:"opinion_profile,omitempty"`
	ImageBase64    string          `json:"image_base64,omitempty"`
}

// Stage5Response 5段階目のレスポンス（図形描画プログラム生成・実行）
type Stage5Response struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	GeometryCode string `json:"geometry_code"` // 図形描画プログラム
	ImageBase64  string `json:"image_base64"`  // 生成された図形画像
	Log          string `json:"log"`
}

// ProgressUpdate 進捗更新用の構造体
type ProgressUpdate struct {
	Stage       int     `json:"stage"`
	MaxStages   int     `json:"max_stages"`
	Progress    float64 `json:"progress"`
	Message     string  `json:"message"`
	IsCompleted bool    `json:"is_completed"`
	Error       string  `json:"error,omitempty"`
}
