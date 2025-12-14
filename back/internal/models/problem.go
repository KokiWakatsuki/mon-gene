package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Problem struct {
	ID          int64                  `json:"id" db:"id"`
	UserID      int64                  `json:"user_id" db:"user_id"`
	Subject     string                 `json:"subject" db:"subject"`
	Prompt      string                 `json:"prompt" db:"prompt"`                           // 生成時のプロンプト
	Content     string                 `json:"content" db:"content"`                         // 問題文
	Solution    string                 `json:"solution,omitempty" db:"solution"`             // 解答
	ImageBase64 string                 `json:"image_base64,omitempty" db:"image_base64"`     // 図
	// opinion.md基準の評価データ
	OpinionProfile      *OpinionProfile      `json:"opinion_profile,omitempty" db:"opinion_profile"`           // opinion_ver1.md基準のプロファイル（レガシー）
	OpinionProfileV2    *OpinionProfileV2    `json:"opinion_profile_v2,omitempty" db:"opinion_profile_v2"`     // opinion_ver2.md基準のプロファイル
	ConversationHistory *ConversationHistory `json:"conversation_history,omitempty" db:"conversation_history"` // 5段階生成プロセスの会話履歴
	CheckInfo           *CheckInfo           `json:"check_info,omitempty" db:"check_info"`                     // チェック情報
	CreatedAt           time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at" db:"updated_at"`
}

// CheckInfo はチェック情報を格納する構造体
type CheckInfo struct {
	ProblemTextOK  bool     `json:"problem_text_ok"`  // 問題文が適切か
	SolutionOK     bool     `json:"solution_ok"`      // 解答・解説が適切か
	FigureOK       bool     `json:"figure_ok"`        // 図が適切か
	Units          []string `json:"units"`            // 使用されている単元・公式・定理（複数選択）
	Year           string   `json:"year"`             // 年度（2020~2025）
	ExamSession    string   `json:"exam_session"`     // 回数（第1回、第2回、第3回、プレ、追試）
}

// OpinionProfile は opinion_ver1.md の評価基準に基づく問題プロファイル（レガシー）
type OpinionProfile struct {
	Domain             int    `json:"domain"`               // 出題分野コード (1-6)
	SkillLevel         int    `json:"skill_level"`          // コアスキル評価 (1-10)
	StructureComplexity [2]int `json:"structure_complexity"` // 問題構造評価 [A, B] (各1-10)
	DifficultyScore    int    `json:"difficulty_score"`     // 総合難易度スコア (1-20)
}

// OpinionProfileV2 は opinion_ver2.md の評価基準に基づく問題プロファイル
type OpinionProfileV2 struct {
	// 1. 文章量・構成に関する指標
	ProblemTextLength    int      `json:"problem_text_length"`     // 大問の問題文文字数
	SubProblemTextLength int      `json:"sub_problem_text_length"` // 小問の総文字数
	GivenValuesCount     int      `json:"given_values_count"`      // 与えられる数値の個数
	SubProblemCount      int      `json:"sub_problem_count"`       // 小問の数（解答要求数）
	SubProblemTypes      []string `json:"sub_problem_types"`       // 小問ごとの要求種別（長さ、面積、体積、最短距離、角度、比）
	SolidComposition     string   `json:"solid_composition"`       // 立体の構成（単一/複数の組み合わせ）

	// 2. 解答形式に関する指標
	AnswerFormats       []string `json:"answer_formats"`        // 解答の形式（整数、既約分数、無理数）
	AnswerUnits         []string `json:"answer_units"`          // 小問ごとの要求単位（cm、cm²、cm³、度、単位なし）
	UsesAuxiliaryPoints bool     `json:"uses_auxiliary_points"` // 正答例での補助点の使用

	// 3. 使用単元に関する指標
	SetupUnits    []string `json:"setup_units"`    // 単元（設定）：問題文や図の構成要素
	SolutionUnits []string `json:"solution_units"` // 単元（解法）：正答例で使用される定理・手法

	// 4. 図形に関する指標
	TotalVertices     int  `json:"total_vertices"`      // 総頂点・点の数
	HasMovingPoint    bool `json:"has_moving_point"`    // 動点の有無
	FigureValuesCount int  `json:"figure_values_count"` // 図中に明記された数値の個数

	// 5. 解法プロセスと認知負荷に関する指標
	SolutionSteps                int  `json:"solution_steps"`                  // 解法のステップ数
	HasLogicalBranching          bool `json:"has_logical_branching"`           // 論理的分岐の有無
	TheoremCount                 int  `json:"theorem_count"`                   // 使用定理・公式の数
	RequiresMultiUnitIntegration bool `json:"requires_multi_unit_integration"` // 複数単元の知識統合の要否
	HasIrrelevantInfo            bool `json:"has_irrelevant_info"`             // 無関係な情報の有無
}

type GenerateProblemRequest struct {
	Prompt           string            `json:"prompt" validate:"required"`
	Subject          string            `json:"subject" validate:"required"`
	OpinionProfile   *OpinionProfile   `json:"opinion_profile,omitempty"`   // opinion_ver1.md基準での問題生成（レガシー）
	OpinionProfileV2 *OpinionProfileV2 `json:"opinion_profile_v2,omitempty"` // opinion_ver2.md基準での問題生成
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

type UpdateCheckInfoRequest struct {
	ID        int64      `json:"id" validate:"required"`
	CheckInfo *CheckInfo `json:"check_info" validate:"required"`
}

type UpdateCheckInfoResponse struct {
	Success bool     `json:"success"`
	Problem *Problem `json:"problem,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// SearchFilter は保存された検索条件を表す構造体
type SearchFilter struct {
	ID          int64      `json:"id" db:"id"`
	UserID      int64      `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	Keyword     string     `json:"keyword,omitempty" db:"keyword"`
	Subject     string     `json:"subject,omitempty" db:"subject"`
	Units       StringList `json:"units,omitempty" db:"units"`
	Year        string     `json:"year,omitempty" db:"year"`
	ExamSession string     `json:"exam_session,omitempty" db:"exam_session"`
	IsChecked   *bool      `json:"is_checked,omitempty" db:"is_checked"` // null許容のためポインタ
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// StringList はJSON配列として保存される文字列リスト
type StringList []string

// Scan はデータベースからの読み込み時に呼ばれる
func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value はデータベースへの書き込み時に呼ばれる
func (s StringList) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// CreateSearchFilterRequest は検索条件作成リクエスト
type CreateSearchFilterRequest struct {
	Name        string   `json:"name" validate:"required"`
	Keyword     string   `json:"keyword,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	Units       []string `json:"units,omitempty"`
	Year        string   `json:"year,omitempty"`
	ExamSession string   `json:"exam_session,omitempty"`
	IsChecked   *bool    `json:"is_checked,omitempty"`
}

// UpdateSearchFilterRequest は検索条件更新リクエスト
type UpdateSearchFilterRequest struct {
	ID          int64    `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Keyword     string   `json:"keyword,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	Units       []string `json:"units,omitempty"`
	Year        string   `json:"year,omitempty"`
	ExamSession string   `json:"exam_session,omitempty"`
	IsChecked   *bool    `json:"is_checked,omitempty"`
}

// SearchFilterResponse は検索条件のレスポンス
type SearchFilterResponse struct {
	Success      bool            `json:"success"`
	SearchFilter *SearchFilter   `json:"search_filter,omitempty"`
	Filters      []*SearchFilter `json:"filters,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// SourceListItem は出典リストの項目を表す
type SourceListItem struct {
	ID          int64     `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Year        string    `db:"year" json:"year"`
	ExamSession string    `db:"exam_session" json:"exam_session"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// CreateSourceListItemRequest は出典リスト項目作成リクエスト
type CreateSourceListItemRequest struct {
	Year        string `json:"year" binding:"required"`
	ExamSession string `json:"exam_session" binding:"required"`
}

// SourceListResponse は出典リスト取得レスポンス
type SourceListResponse struct {
	Success bool              `json:"success"`
	Items   []*SourceListItem `json:"items,omitempty"`
	Error   string            `json:"error,omitempty"`
}
