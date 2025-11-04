
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mon-gene/back/internal/models"
)

type MySQLProblemRepository struct {
	db *sqlx.DB
}

func NewMySQLProblemRepository(db *sqlx.DB) ProblemRepository {
	return &MySQLProblemRepository{db: db}
}

// 共通のスキャン処理（opinion_profile + opinion_profile_v2対応）
func (r *MySQLProblemRepository) scanProblem(rows *sql.Rows) (*models.Problem, error) {
	var problem models.Problem
	var opinionProfileJSON []byte
	var opinionProfileV2JSON []byte

	err := rows.Scan(
		&problem.ID,
		&problem.UserID,
		&problem.Subject,
		&problem.Prompt,
		&problem.Content,
		&problem.Solution,
		&problem.ImageBase64,
		&opinionProfileJSON,
		&opinionProfileV2JSON,
		&problem.CreatedAt,
		&problem.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(opinionProfileJSON) > 0 {
		if err := json.Unmarshal(opinionProfileJSON, &problem.OpinionProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal opinion_profile: %w", err)
		}
	}

	if len(opinionProfileV2JSON) > 0 {
		if err := json.Unmarshal(opinionProfileV2JSON, &problem.OpinionProfileV2); err != nil {
			return nil, fmt.Errorf("failed to unmarshal opinion_profile_v2: %w", err)
		}
	}

	return &problem, nil
}

// 共通のスキャン処理（単一行用、opinion_profile + opinion_profile_v2対応）
func (r *MySQLProblemRepository) scanProblemRow(row *sql.Row) (*models.Problem, error) {
	var problem models.Problem
	var opinionProfileJSON []byte
	var opinionProfileV2JSON []byte

	err := row.Scan(
		&problem.ID,
		&problem.UserID,
		&problem.Subject,
		&problem.Prompt,
		&problem.Content,
		&problem.Solution,
		&problem.ImageBase64,
		&opinionProfileJSON,
		&opinionProfileV2JSON,
		&problem.CreatedAt,
		&problem.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(opinionProfileJSON) > 0 {
		if err := json.Unmarshal(opinionProfileJSON, &problem.OpinionProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal opinion_profile: %w", err)
		}
	}

	if len(opinionProfileV2JSON) > 0 {
		if err := json.Unmarshal(opinionProfileV2JSON, &problem.OpinionProfileV2); err != nil {
			return nil, fmt.Errorf("failed to unmarshal opinion_profile_v2: %w", err)
		}
	}

	return &problem, nil
}

func (r *MySQLProblemRepository) Create(ctx context.Context, problem *models.Problem) error {
	var opinionProfileJSON []byte
	var opinionProfileV2JSON []byte
	var err error
	
	if problem.OpinionProfile != nil {
		opinionProfileJSON, err = json.Marshal(problem.OpinionProfile)
		if err != nil {
			return fmt.Errorf("failed to marshal opinion_profile: %w", err)
		}
	}

	if problem.OpinionProfileV2 != nil {
		opinionProfileV2JSON, err = json.Marshal(problem.OpinionProfileV2)
		if err != nil {
			return fmt.Errorf("failed to marshal opinion_profile_v2: %w", err)
		}
	}

	query := `
		INSERT INTO problems (user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := r.db.ExecContext(ctx, query,
		problem.UserID,
		problem.Subject,
		problem.Prompt,
		problem.Content,
		problem.Solution,
		problem.ImageBase64,
		opinionProfileJSON,
		opinionProfileV2JSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create problem: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	problem.ID = id
	return nil
}

func (r *MySQLProblemRepository) GetByID(ctx context.Context, id int64) (*models.Problem, error) {
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at
		FROM problems
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	problem, err := r.scanProblemRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("problem not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	return problem, nil
}

func (r *MySQLProblemRepository) GetByIDAndUserID(ctx context.Context, id, userID int64) (*models.Problem, error) {
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at
		FROM problems
		WHERE id = ? AND user_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id, userID)
	problem, err := r.scanProblemRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("problem not found or access denied")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	return problem, nil
}

func (r *MySQLProblemRepository) Update(ctx context.Context, problem *models.Problem) error {
	var opinionProfileJSON []byte
	var opinionProfileV2JSON []byte
	var err error
	
	if problem.OpinionProfile != nil {
		opinionProfileJSON, err = json.Marshal(problem.OpinionProfile)
		if err != nil {
			return fmt.Errorf("failed to marshal opinion_profile: %w", err)
		}
	}

	if problem.OpinionProfileV2 != nil {
		opinionProfileV2JSON, err = json.Marshal(problem.OpinionProfileV2)
		if err != nil {
			return fmt.Errorf("failed to marshal opinion_profile_v2: %w", err)
		}
	}

	query := `
		UPDATE problems
		SET subject = ?, prompt = ?, content = ?, solution = ?, image_base64 = ?, opinion_profile = ?, opinion_profile_v2 = ?, updated_at = NOW()
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		problem.Subject,
		problem.Prompt,
		problem.Content,
		problem.Solution,
		problem.ImageBase64,
		opinionProfileJSON,
		opinionProfileV2JSON,
		problem.ID,
		problem.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update problem: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("problem not found or access denied")
	}

	return nil
}

func (r *MySQLProblemRepository) UpdateGeometry(ctx context.Context, id int64, imageBase64 string) error {
	query := `
		UPDATE problems 
		SET image_base64 = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, imageBase64, id)
	if err != nil {
		return fmt.Errorf("failed to update geometry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("problem not found")
	}

	return nil
}

func (r *MySQLProblemRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Problem, error) {
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at
		FROM problems
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get problems: %w", err)
	}
	defer rows.Close()

	var problems []*models.Problem
	for rows.Next() {
		problem, err := r.scanProblem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, problem)
	}

	return problems, nil
}

func (r *MySQLProblemRepository) SearchCombined(ctx context.Context, userID int64, keyword string, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error) {
	fmt.Printf("\n🔍 [DEBUG] SearchCombined called with:\n")
	fmt.Printf("  - userID: %d\n", userID)
	fmt.Printf("  - keyword: %q\n", keyword)
	fmt.Printf("  - subject: %q\n", subject)
	fmt.Printf("  - matchType: %q\n", matchType)
	fmt.Printf("  - limit: %d, offset: %d\n", limit, offset)
	fmt.Printf("  - filters: %+v\n", filters)
	
	// 基本クエリの構築（opinion_profile_v2対応）
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at
		FROM problems
		WHERE user_id = ?`

	queryArgs := []interface{}{userID}

	// キーワード検索条件
	if keyword != "" {
		query += " AND (content LIKE ? OR solution LIKE ? OR prompt LIKE ? OR subject LIKE ?)"
		searchPattern := "%" + keyword + "%"
		queryArgs = append(queryArgs, searchPattern, searchPattern, searchPattern, searchPattern)
		fmt.Printf("  ✅ Keyword filter added: %q (pattern: %q)\n", keyword, searchPattern)
	}

	// 科目での絞り込み
	if subject != "" {
		query += " AND subject = ?"
		queryArgs = append(queryArgs, subject)
		fmt.Printf("  ✅ Subject filter added: %q\n", subject)
	}

	// opinion_profile_v2が存在することを確認（フィルターがある場合のみ）
	if filters != nil && len(filters) > 0 {
		query += " AND opinion_profile_v2 IS NOT NULL"
	}

	query += " ORDER BY created_at DESC"

	fmt.Printf("\n🔎 [QUERY]\n")
	fmt.Printf("SQL: %s\n", query)
	fmt.Printf("Args: %v\n\n", queryArgs)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		fmt.Printf("❌ [ERROR] Query execution failed: %v\n", err)
		return nil, fmt.Errorf("failed to search problems by combined conditions: %w", err)
	}
	defer rows.Close()

	var allProblems []*models.Problem
	for rows.Next() {
		problem, err := r.scanProblem(rows)
		if err != nil {
			fmt.Printf("❌ [ERROR] Row scanning failed: %v\n", err)
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		allProblems = append(allProblems, problem)
	}

	fmt.Printf("📋 [ALL PROBLEMS] Found %d problems\n", len(allProblems))

	// フィルターがない場合はページネーションして返す
	if filters == nil || len(filters) == 0 {
		start := offset
		end := offset + limit
		if start > len(allProblems) {
			return []*models.Problem{}, nil
		}
		if end > len(allProblems) {
			end = len(allProblems)
		}
		return allProblems[start:end], nil
	}

	// アプリケーション層でフィルタリング（SearchByFiltersと同じロジック）
	var matchedProblems []*models.Problem
	
	for _, problem := range allProblems {
		if problem.OpinionProfileV2 == nil {
			continue
		}
		
		profile := problem.OpinionProfileV2
		matchCount := 0
		totalConditions := 0
		
		// 各フィールドをチェック
		checkField := func(key string, dbValue interface{}, searchValue interface{}) bool {
			totalConditions++
			
			switch v := dbValue.(type) {
			case int:
				if searchVal, ok := searchValue.(float64); ok {
					if v == int(searchVal) {
						matchCount++
						return true
					}
				}
			case bool:
				if searchVal, ok := searchValue.(bool); ok {
					if v == searchVal {
						matchCount++
						return true
					}
				}
			case string:
				if searchVal, ok := searchValue.(string); ok {
					if v == searchVal {
						matchCount++
						return true
					}
				}
			case []string:
				if searchArr, ok := searchValue.([]interface{}); ok {
					if len(v) == len(searchArr) {
						allMatch := true
						for _, searchItem := range searchArr {
							if searchStr, ok := searchItem.(string); ok {
								found := false
								for _, dbItem := range v {
									if dbItem == searchStr {
										found = true
										break
									}
								}
								if !found {
									allMatch = false
									break
								}
							}
						}
						if allMatch {
							matchCount++
							return true
						}
					}
				}
			}
			return false
		}
		
		// 各フィールドをチェック
		if val, exists := filters["problem_text_length"]; exists {
			checkField("problem_text_length", profile.ProblemTextLength, val)
		}
		if val, exists := filters["sub_problem_text_length"]; exists {
			checkField("sub_problem_text_length", profile.SubProblemTextLength, val)
		}
		if val, exists := filters["given_values_count"]; exists {
			checkField("given_values_count", profile.GivenValuesCount, val)
		}
		if val, exists := filters["sub_problem_count"]; exists {
			checkField("sub_problem_count", profile.SubProblemCount, val)
		}
		if val, exists := filters["sub_problem_types"]; exists {
			checkField("sub_problem_types", profile.SubProblemTypes, val)
		}
		if val, exists := filters["solid_composition"]; exists {
			checkField("solid_composition", profile.SolidComposition, val)
		}
		if val, exists := filters["answer_formats"]; exists {
			checkField("answer_formats", profile.AnswerFormats, val)
		}
		if val, exists := filters["answer_units"]; exists {
			checkField("answer_units", profile.AnswerUnits, val)
		}
		if val, exists := filters["uses_auxiliary_points"]; exists {
			checkField("uses_auxiliary_points", profile.UsesAuxiliaryPoints, val)
		}
		if val, exists := filters["setup_units"]; exists {
			checkField("setup_units", profile.SetupUnits, val)
		}
		if val, exists := filters["solution_units"]; exists {
			checkField("solution_units", profile.SolutionUnits, val)
		}
		if val, exists := filters["total_vertices"]; exists {
			checkField("total_vertices", profile.TotalVertices, val)
		}
		if val, exists := filters["has_moving_point"]; exists {
			checkField("has_moving_point", profile.HasMovingPoint, val)
		}
		if val, exists := filters["figure_values_count"]; exists {
			checkField("figure_values_count", profile.FigureValuesCount, val)
		}
		if val, exists := filters["solution_steps"]; exists {
			checkField("solution_steps", profile.SolutionSteps, val)
		}
		if val, exists := filters["has_logical_branching"]; exists {
			checkField("has_logical_branching", profile.HasLogicalBranching, val)
		}
		if val, exists := filters["theorem_count"]; exists {
			checkField("theorem_count", profile.TheoremCount, val)
		}
		if val, exists := filters["requires_multi_unit_integration"]; exists {
			checkField("requires_multi_unit_integration", profile.RequiresMultiUnitIntegration, val)
		}
		if val, exists := filters["has_irrelevant_info"]; exists {
			checkField("has_irrelevant_info", profile.HasIrrelevantInfo, val)
		}
		
		// マッチング判定
		if matchType == "exact" {
			if matchCount == totalConditions && totalConditions > 0 {
				matchedProblems = append(matchedProblems, problem)
				fmt.Printf("  ✅ Problem %d: EXACT MATCH (%d/%d)\n", problem.ID, matchCount, totalConditions)
			}
		} else {
			if matchCount > 0 {
				matchedProblems = append(matchedProblems, problem)
				fmt.Printf("  ✅ Problem %d: PARTIAL MATCH (%d/%d)\n", problem.ID, matchCount, totalConditions)
			}
		}
	}

	fmt.Printf("📋 [FILTERED] %d problems matched\n", len(matchedProblems))
	
	// ページネーション
	start := offset
	end := offset + limit
	if start > len(matchedProblems) {
		return []*models.Problem{}, nil
	}
	if end > len(matchedProblems) {
		end = len(matchedProblems)
	}
	
	return matchedProblems[start:end], nil
}

func (r *MySQLProblemRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM problems WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}
	return nil
}

func (r *MySQLProblemRepository) SearchByParameters(ctx context.Context, userID int64, subject string, prompt string, filters map[string]interface{}) ([]*models.Problem, error) {
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at
		FROM problems
		WHERE user_id = ? AND subject = ? AND prompt = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, subject, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by parameters: %w", err)
	}
	defer rows.Close()

	var problems []*models.Problem
	for rows.Next() {
		problem, err := r.scanProblem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, problem)
	}

	return problems, nil
}

func (r *MySQLProblemRepository) SearchByFilters(ctx context.Context, userID int64, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error) {
	fmt.Printf("\n🔍 [DEBUG] SearchByFilters called with:\n")
	fmt.Printf("  - userID: %d\n", userID)
	fmt.Printf("  - subject: %q\n", subject)
	fmt.Printf("  - matchType: %q\n", matchType)
	fmt.Printf("  - limit: %d, offset: %d\n", limit, offset)
	fmt.Printf("  - filters: %+v\n", filters)
	
	// まずすべての問題を取得（opinion_profile_v2があるもののみ）
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at
		FROM problems
		WHERE user_id = ? AND opinion_profile_v2 IS NOT NULL`

	queryArgs := []interface{}{userID}

	// 科目での絞り込み
	if subject != "" {
		query += " AND subject = ?"
		queryArgs = append(queryArgs, subject)
		fmt.Printf("  ✅ Subject filter added: %q\n", subject)
	}
	
	query += " ORDER BY created_at DESC"
	
	fmt.Printf("\n🔎 [QUERY]\n")
	fmt.Printf("SQL: %s\n", query)
	fmt.Printf("Args: %v\n\n", queryArgs)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		fmt.Printf("❌ [ERROR] Query execution failed: %v\n", err)
		return nil, fmt.Errorf("failed to search problems by filters: %w", err)
	}
	defer rows.Close()

	var allProblems []*models.Problem
	for rows.Next() {
		problem, err := r.scanProblem(rows)
		if err != nil {
			fmt.Printf("❌ [ERROR] Row scanning failed: %v\n", err)
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		allProblems = append(allProblems, problem)
	}

	fmt.Printf("📋 [ALL PROBLEMS] Found %d problems from database\n", len(allProblems))

	// フィルターがない場合はすべて返す
	if filters == nil || len(filters) == 0 {
		fmt.Printf("  ℹ️ No filters provided, returning all problems\n")
		start := offset
		end := offset + limit
		if start > len(allProblems) {
			return []*models.Problem{}, nil
		}
		if end > len(allProblems) {
			end = len(allProblems)
		}
		return allProblems[start:end], nil
	}

	// アプリケーション層でフィルタリング
	var matchedProblems []*models.Problem
	
	for _, problem := range allProblems {
		if problem.OpinionProfileV2 == nil {
			continue
		}
		
		profile := problem.OpinionProfileV2
		matchCount := 0
		totalConditions := 0
		
		// 各フィールドをチェック
		checkField := func(key string, dbValue interface{}, searchValue interface{}) bool {
			totalConditions++
			
			switch v := dbValue.(type) {
			case int:
				if searchVal, ok := searchValue.(float64); ok {
					if v == int(searchVal) {
						matchCount++
						return true
					}
				}
			case bool:
				if searchVal, ok := searchValue.(bool); ok {
					if v == searchVal {
						matchCount++
						return true
					}
				}
			case string:
				if searchVal, ok := searchValue.(string); ok {
					if v == searchVal {
						matchCount++
						return true
					}
				}
			case []string:
				// 配列の場合：検索値の配列とDB値の配列を比較
				if searchArr, ok := searchValue.([]interface{}); ok {
					// すべての要素が一致するかチェック
					if len(v) == len(searchArr) {
						allMatch := true
						for _, searchItem := range searchArr {
							if searchStr, ok := searchItem.(string); ok {
								found := false
								for _, dbItem := range v {
									if dbItem == searchStr {
										found = true
										break
									}
								}
								if !found {
									allMatch = false
									break
								}
							}
						}
						if allMatch {
							matchCount++
							return true
						}
					}
				}
			}
			return false
		}
		
		// 各フィールドをチェック
		if val, exists := filters["problem_text_length"]; exists {
			checkField("problem_text_length", profile.ProblemTextLength, val)
		}
		if val, exists := filters["sub_problem_text_length"]; exists {
			checkField("sub_problem_text_length", profile.SubProblemTextLength, val)
		}
		if val, exists := filters["given_values_count"]; exists {
			checkField("given_values_count", profile.GivenValuesCount, val)
		}
		if val, exists := filters["sub_problem_count"]; exists {
			checkField("sub_problem_count", profile.SubProblemCount, val)
		}
		if val, exists := filters["sub_problem_types"]; exists {
			checkField("sub_problem_types", profile.SubProblemTypes, val)
		}
		if val, exists := filters["solid_composition"]; exists {
			checkField("solid_composition", profile.SolidComposition, val)
		}
		if val, exists := filters["answer_formats"]; exists {
			checkField("answer_formats", profile.AnswerFormats, val)
		}
		if val, exists := filters["answer_units"]; exists {
			checkField("answer_units", profile.AnswerUnits, val)
		}
		if val, exists := filters["uses_auxiliary_points"]; exists {
			checkField("uses_auxiliary_points", profile.UsesAuxiliaryPoints, val)
		}
		if val, exists := filters["setup_units"]; exists {
			checkField("setup_units", profile.SetupUnits, val)
		}
		if val, exists := filters["solution_units"]; exists {
			checkField("solution_units", profile.SolutionUnits, val)
		}
		if val, exists := filters["total_vertices"]; exists {
			checkField("total_vertices", profile.TotalVertices, val)
		}
		if val, exists := filters["has_moving_point"]; exists {
			checkField("has_moving_point", profile.HasMovingPoint, val)
		}
		if val, exists := filters["figure_values_count"]; exists {
			checkField("figure_values_count", profile.FigureValuesCount, val)
		}
		if val, exists := filters["solution_steps"]; exists {
			checkField("solution_steps", profile.SolutionSteps, val)
		}
		if val, exists := filters["has_logical_branching"]; exists {
			checkField("has_logical_branching", profile.HasLogicalBranching, val)
		}
		if val, exists := filters["theorem_count"]; exists {
			checkField("theorem_count", profile.TheoremCount, val)
		}
		if val, exists := filters["requires_multi_unit_integration"]; exists {
			checkField("requires_multi_unit_integration", profile.RequiresMultiUnitIntegration, val)
		}
		if val, exists := filters["has_irrelevant_info"]; exists {
			checkField("has_irrelevant_info", profile.HasIrrelevantInfo, val)
		}
		
		// マッチング判定
		if matchType == "exact" {
			// 完全一致：すべての条件が一致
			if matchCount == totalConditions && totalConditions > 0 {
				matchedProblems = append(matchedProblems, problem)
				fmt.Printf("  ✅ Problem %d: EXACT MATCH (%d/%d conditions)\n", problem.ID, matchCount, totalConditions)
			} else {
				fmt.Printf("  ❌ Problem %d: Not exact match (%d/%d conditions)\n", problem.ID, matchCount, totalConditions)
			}
		} else {
			// 部分一致：どれか1つでも一致
			if matchCount > 0 {
				matchedProblems = append(matchedProblems, problem)
				fmt.Printf("  ✅ Problem %d: PARTIAL MATCH (%d/%d conditions)\n", problem.ID, matchCount, totalConditions)
			} else {
				fmt.Printf("  ❌ Problem %d: No match (%d/%d conditions)\n", problem.ID, matchCount, totalConditions)
			}
		}
	}

	fmt.Printf("📋 [FILTERED RESULT] %d problems matched (matchType: %s)\n", len(matchedProblems), matchType)
	
	// ページネーション適用
	start := offset
	end := offset + limit
	if start > len(matchedProblems) {
		return []*models.Problem{}, nil
	}
	if end > len(matchedProblems) {
		end = len(matchedProblems)
	}
	
	return matchedProblems[start:end], nil
}

func (r *MySQLProblemRepository) SearchByKeyword(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*models.Problem, error) {
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, opinion_profile_v2, created_at, updated_at
		FROM problems
		WHERE user_id = ? AND (
			content LIKE ? OR
			solution LIKE ? OR
			prompt LIKE ? OR
			subject LIKE ?
		)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	searchPattern := "%" + keyword + "%"
	rows, err := r.db.QueryContext(ctx, query, userID, searchPattern, searchPattern, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by keyword: %w", err)
	}
	defer rows.Close()

	var problems []*models.Problem
	for rows.Next() {
		problem, err := r.scanProblem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, problem)
	}

	return problems, nil
}
