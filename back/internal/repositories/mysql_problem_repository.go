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

// 共通のスキャン処理（opinion_profile対応、filters削除済み）
func (r *MySQLProblemRepository) scanProblem(rows *sql.Rows) (*models.Problem, error) {
	var problem models.Problem
	var opinionProfileJSON []byte

	err := rows.Scan(
		&problem.ID,
		&problem.UserID,
		&problem.Subject,
		&problem.Prompt,
		&problem.Content,
		&problem.Solution,
		&problem.ImageBase64,
		&opinionProfileJSON,
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

	return &problem, nil
}

// 共通のスキャン処理（単一行用、filters削除済み）
func (r *MySQLProblemRepository) scanProblemRow(row *sql.Row) (*models.Problem, error) {
	var problem models.Problem
	var opinionProfileJSON []byte

	err := row.Scan(
		&problem.ID,
		&problem.UserID,
		&problem.Subject,
		&problem.Prompt,
		&problem.Content,
		&problem.Solution,
		&problem.ImageBase64,
		&opinionProfileJSON,
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

	return &problem, nil
}

func (r *MySQLProblemRepository) Create(ctx context.Context, problem *models.Problem) error {
	var opinionProfileJSON []byte
	var err error
	if problem.OpinionProfile != nil {
		opinionProfileJSON, err = json.Marshal(problem.OpinionProfile)
		if err != nil {
			return fmt.Errorf("failed to marshal opinion_profile: %w", err)
		}
	}

	query := `
		INSERT INTO problems (user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := r.db.ExecContext(ctx, query,
		problem.UserID,
		problem.Subject,
		problem.Prompt,
		problem.Content,
		problem.Solution,
		problem.ImageBase64,
		opinionProfileJSON,
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
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at
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
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at
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
	var err error
	if problem.OpinionProfile != nil {
		opinionProfileJSON, err = json.Marshal(problem.OpinionProfile)
		if err != nil {
			return fmt.Errorf("failed to marshal opinion_profile: %w", err)
		}
	}

	query := `
		UPDATE problems 
		SET subject = ?, prompt = ?, content = ?, solution = ?, image_base64 = ?, opinion_profile = ?, updated_at = NOW()
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		problem.Subject,
		problem.Prompt,
		problem.Content,
		problem.Solution,
		problem.ImageBase64,
		opinionProfileJSON,
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
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at
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
	
	// 基本クエリの構築（opinion_profileに統一）
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at
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

	// OpinionProfileベースのフィルター検索を実装（matchType対応）
	if filters != nil && len(filters) > 0 {
		fmt.Printf("  📊 Processing filters (%d entries):\n", len(filters))
		var filterConditions []string
		var filterArgs []interface{}

		// 出題分野コードでの絞り込み
		if domainValues, exists := filters["出題分野コード"]; exists {
			if domains, ok := domainValues.([]string); ok && len(domains) > 0 {
				if len(domains) == 1 {
					if domain := domains[0]; domain != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.domain') = ?")
						filterArgs = append(filterArgs, domain)
					}
				}
			}
		}

		// コアスキルレベルでの絞り込み
		if skillValues, exists := filters["コアスキルレベル"]; exists {
			if skills, ok := skillValues.([]string); ok && len(skills) > 0 {
				if len(skills) == 1 {
					if skill := skills[0]; skill != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.skill_level') = ?")
						filterArgs = append(filterArgs, skill)
					}
				}
			}
		}

		// 読解・設定の複雑度での絞り込み
		if complexityValues, exists := filters["読解・設定の複雑度"]; exists {
			if complexities, ok := complexityValues.([]string); ok && len(complexities) > 0 {
				if len(complexities) == 1 {
					if complexity := complexities[0]; complexity != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.structure_complexity[0]') = ?")
						filterArgs = append(filterArgs, complexity)
					}
				}
			}
		}

		// 設問の誘導性での絞り込み
		if guidanceValues, exists := filters["設問の誘導性"]; exists {
			if guidances, ok := guidanceValues.([]string); ok && len(guidances) > 0 {
				if len(guidances) == 1 {
					if guidance := guidances[0]; guidance != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.structure_complexity[1]') = ?")
						filterArgs = append(filterArgs, guidance)
					}
				}
			}
		}

		// 総合難易度スコアでの絞り込み（具体的な数値との完全一致）
		if difficultyValues, exists := filters["総合難易度スコア"]; exists {
			if difficulties, ok := difficultyValues.([]string); ok && len(difficulties) > 0 {
				if len(difficulties) == 1 {
					if difficulty := difficulties[0]; difficulty != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.difficulty_score') = ?")
						filterArgs = append(filterArgs, difficulty)
					}
				}
			}
		}

		// matchTypeに基づいてフィルター条件を結合
		if len(filterConditions) > 0 {
			if matchType == "partial" {
				// 部分一致: いずれかの条件が一致すればOK
				query += " AND (" + filterConditions[0]
				for i := 1; i < len(filterConditions); i++ {
					query += " OR " + filterConditions[i]
				}
				query += ")"
			} else {
				// 完全一致 (デフォルト): すべての条件が一致する必要がある
				for _, condition := range filterConditions {
					query += " AND " + condition
				}
			}
			queryArgs = append(queryArgs, filterArgs...)
		}
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by combined conditions: %w", err)
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

func (r *MySQLProblemRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM problems WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}
	return nil
}

func (r *MySQLProblemRepository) SearchByParameters(ctx context.Context, userID int64, subject string, prompt string, filters map[string]interface{}) ([]*models.Problem, error) {
	// 従来のfiltersベース検索は削除、基本的な検索のみ実行
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at
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
	
	// opinion_profileベースの検索を実装
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at
		FROM problems
		WHERE user_id = ?`

	queryArgs := []interface{}{userID}

	// 科目での絞り込み
	if subject != "" {
		query += " AND subject = ?"
		queryArgs = append(queryArgs, subject)
		fmt.Printf("  ✅ Subject filter added: %q\n", subject)
	}

	// OpinionProfileベースのフィルター検索を実装（matchType対応）
	if filters != nil && len(filters) > 0 {
		fmt.Printf("  📊 Processing filters (%d entries):\n", len(filters))
		var filterConditions []string
		var filterArgs []interface{}

		// 出題分野コードでの絞り込み
		if domainValues, exists := filters["出題分野コード"]; exists {
			fmt.Printf("    🔍 出題分野コード: %+v (type: %T)\n", domainValues, domainValues)
			// []interface{} から []string への変換を処理
			var domains []string
			if domainSlice, ok := domainValues.([]interface{}); ok {
				for _, v := range domainSlice {
					if str, ok := v.(string); ok {
						domains = append(domains, str)
					}
				}
			} else if domainSlice, ok := domainValues.([]string); ok {
				domains = domainSlice
			}
			
			if len(domains) > 0 {
				if len(domains) == 1 {
					if domain := domains[0]; domain != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.domain') = CAST(? AS UNSIGNED)")
						filterArgs = append(filterArgs, domain)
						fmt.Printf("      ✅ Added domain filter: %q (as UNSIGNED)\n", domain)
					}
				}
			} else {
				fmt.Printf("      ❌ Failed to parse domains: %+v\n", domainValues)
			}
		}

		// コアスキルレベルでの絞り込み
		if skillValues, exists := filters["コアスキルレベル"]; exists {
			fmt.Printf("    🔍 コアスキルレベル: %+v (type: %T)\n", skillValues, skillValues)
			// []interface{} から []string への変換を処理
			var skills []string
			if skillSlice, ok := skillValues.([]interface{}); ok {
				for _, v := range skillSlice {
					if str, ok := v.(string); ok {
						skills = append(skills, str)
					}
				}
			} else if skillSlice, ok := skillValues.([]string); ok {
				skills = skillSlice
			}
			
			if len(skills) > 0 {
				if len(skills) == 1 {
					if skill := skills[0]; skill != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.skill_level') = CAST(? AS UNSIGNED)")
						filterArgs = append(filterArgs, skill)
						fmt.Printf("      ✅ Added skill_level filter: %q (as UNSIGNED)\n", skill)
					}
				}
			} else {
				fmt.Printf("      ❌ Failed to parse skills: %+v\n", skillValues)
			}
		}

		// 読解・設定の複雑度での絞り込み
		if complexityValues, exists := filters["読解・設定の複雑度"]; exists {
			fmt.Printf("    🔍 読解・設定の複雑度: %+v (type: %T)\n", complexityValues, complexityValues)
			// []interface{} から []string への変換を処理
			var complexities []string
			if complexitySlice, ok := complexityValues.([]interface{}); ok {
				for _, v := range complexitySlice {
					if str, ok := v.(string); ok {
						complexities = append(complexities, str)
					}
				}
			} else if complexitySlice, ok := complexityValues.([]string); ok {
				complexities = complexitySlice
			}
			
			if len(complexities) > 0 {
				if len(complexities) == 1 {
					if complexity := complexities[0]; complexity != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.structure_complexity[0]') = CAST(? AS UNSIGNED)")
						filterArgs = append(filterArgs, complexity)
						fmt.Printf("      ✅ Added structure_complexity[0] filter: %q (as UNSIGNED)\n", complexity)
					}
				}
			} else {
				fmt.Printf("      ❌ Failed to parse complexities: %+v\n", complexityValues)
			}
		}

		// 設問の誘導性での絞り込み
		if guidanceValues, exists := filters["設問の誘導性"]; exists {
			fmt.Printf("    🔍 設問の誘導性: %+v (type: %T)\n", guidanceValues, guidanceValues)
			// []interface{} から []string への変換を処理
			var guidances []string
			if guidanceSlice, ok := guidanceValues.([]interface{}); ok {
				for _, v := range guidanceSlice {
					if str, ok := v.(string); ok {
						guidances = append(guidances, str)
					}
				}
			} else if guidanceSlice, ok := guidanceValues.([]string); ok {
				guidances = guidanceSlice
			}
			
			if len(guidances) > 0 {
				if len(guidances) == 1 {
					if guidance := guidances[0]; guidance != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.structure_complexity[1]') = CAST(? AS UNSIGNED)")
						filterArgs = append(filterArgs, guidance)
						fmt.Printf("      ✅ Added structure_complexity[1] filter: %q (as UNSIGNED)\n", guidance)
					}
				}
			} else {
				fmt.Printf("      ❌ Failed to parse guidances: %+v\n", guidanceValues)
			}
		}

		// 総合難易度スコアでの絞り込み（具体的な数値との完全一致）
		if difficultyValues, exists := filters["総合難易度スコア"]; exists {
			fmt.Printf("    🔍 総合難易度スコア: %+v (type: %T)\n", difficultyValues, difficultyValues)
			// []interface{} から []string への変換を処理
			var difficulties []string
			if difficultySlice, ok := difficultyValues.([]interface{}); ok {
				for _, v := range difficultySlice {
					if str, ok := v.(string); ok {
						difficulties = append(difficulties, str)
					}
				}
			} else if difficultySlice, ok := difficultyValues.([]string); ok {
				difficulties = difficultySlice
			}
			
			if len(difficulties) > 0 {
				if len(difficulties) == 1 {
					if difficulty := difficulties[0]; difficulty != "" {
						filterConditions = append(filterConditions, "JSON_EXTRACT(opinion_profile, '$.difficulty_score') = CAST(? AS UNSIGNED)")
						filterArgs = append(filterArgs, difficulty)
						fmt.Printf("      ✅ Added difficulty_score filter: %q (as UNSIGNED)\n", difficulty)
					}
				}
			} else {
				fmt.Printf("      ❌ Failed to parse difficulties: %+v\n", difficultyValues)
			}
		}

		fmt.Printf("  📊 Generated filter conditions (%d): %v\n", len(filterConditions), filterConditions)
		fmt.Printf("  📊 Filter args (%d): %v\n", len(filterArgs), filterArgs)

		// matchTypeに基づいてフィルター条件を結合
		if len(filterConditions) > 0 {
			if matchType == "partial" {
				// 部分一致: いずれかの条件が一致すればOK
				query += " AND (" + filterConditions[0]
				for i := 1; i < len(filterConditions); i++ {
					query += " OR " + filterConditions[i]
				}
				query += ")"
				fmt.Printf("  ✅ Applied PARTIAL matching (OR logic)\n")
			} else {
				// 完全一致 (デフォルト): すべての条件が一致する必要がある
				for _, condition := range filterConditions {
					query += " AND " + condition
				}
				fmt.Printf("  ✅ Applied EXACT matching (AND logic)\n")
			}
			queryArgs = append(queryArgs, filterArgs...)
		} else {
			fmt.Printf("  ⚠️ No filter conditions generated!\n")
		}
	} else {
		fmt.Printf("  ℹ️ No filters provided\n")
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, limit, offset)

	fmt.Printf("\n🔎 [FINAL QUERY]\n")
	fmt.Printf("SQL: %s\n", query)
	fmt.Printf("Args (%d): %v\n\n", len(queryArgs), queryArgs)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		fmt.Printf("❌ [ERROR] Query execution failed: %v\n", err)
		return nil, fmt.Errorf("failed to search problems by filters: %w", err)
	}
	defer rows.Close()

	var problems []*models.Problem
	for rows.Next() {
		problem, err := r.scanProblem(rows)
		if err != nil {
			fmt.Printf("❌ [ERROR] Row scanning failed: %v\n", err)
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, problem)
	}

	fmt.Printf("📋 [RESULT] Found %d problems\n", len(problems))
	for i, p := range problems {
		fmt.Printf("  - Problem %d: ID=%d, Subject=%q, OpinionProfile=%+v\n", i+1, p.ID, p.Subject, p.OpinionProfile)
	}

	return problems, nil
}

func (r *MySQLProblemRepository) SearchByKeyword(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*models.Problem, error) {
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, opinion_profile, created_at, updated_at
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
