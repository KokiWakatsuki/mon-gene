package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mon-gene/back/internal/models"
)

type MySQLProblemRepository struct {
	db *sqlx.DB
}

func NewMySQLProblemRepository(db *sqlx.DB) ProblemRepository {
	return &MySQLProblemRepository{db: db}
}

func (r *MySQLProblemRepository) Create(ctx context.Context, problem *models.Problem) error {
	filtersJSON, err := json.Marshal(problem.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	query := `
		INSERT INTO problems (user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := r.db.ExecContext(ctx, query,
		problem.UserID,
		problem.Subject,
		problem.Prompt,
		problem.Content,
		problem.Solution,
		problem.ImageBase64,
		filtersJSON,
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
	var problem models.Problem
	var filtersJSON []byte

	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at
		FROM problems
		WHERE id = ?
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&problem.ID,
		&problem.UserID,
		&problem.Subject,
		&problem.Prompt,
		&problem.Content,
		&problem.Solution,
		&problem.ImageBase64,
		&filtersJSON,
		&problem.CreatedAt,
		&problem.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("problem not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	if err := json.Unmarshal(filtersJSON, &problem.Filters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
	}

	return &problem, nil
}

func (r *MySQLProblemRepository) GetByIDAndUserID(ctx context.Context, id, userID int64) (*models.Problem, error) {
	var problem models.Problem
	var filtersJSON []byte

	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at
		FROM problems
		WHERE id = ? AND user_id = ?
	`

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&problem.ID,
		&problem.UserID,
		&problem.Subject,
		&problem.Prompt,
		&problem.Content,
		&problem.Solution,
		&problem.ImageBase64,
		&filtersJSON,
		&problem.CreatedAt,
		&problem.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("problem not found or access denied")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	if err := json.Unmarshal(filtersJSON, &problem.Filters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
	}

	return &problem, nil
}

func (r *MySQLProblemRepository) Update(ctx context.Context, problem *models.Problem) error {
	filtersJSON, err := json.Marshal(problem.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	query := `
		UPDATE problems 
		SET subject = ?, prompt = ?, content = ?, solution = ?, image_base64 = ?, filters = ?, updated_at = NOW()
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		problem.Subject,
		problem.Prompt,
		problem.Content,
		problem.Solution,
		problem.ImageBase64,
		filtersJSON,
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
		SELECT id, user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at
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
		var problem models.Problem
		var filtersJSON []byte

		err := rows.Scan(
			&problem.ID,
			&problem.UserID,
			&problem.Subject,
			&problem.Prompt,
			&problem.Content,
			&problem.Solution,
			&problem.ImageBase64,
			&filtersJSON,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}

		if err := json.Unmarshal(filtersJSON, &problem.Filters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}

		problems = append(problems, &problem)
	}

	return problems, nil
}

func (r *MySQLProblemRepository) SearchCombined(ctx context.Context, userID int64, keyword string, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error) {
	// 基本クエリの構築
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at
		FROM problems
		WHERE user_id = ?`

	queryArgs := []interface{}{userID}

	// キーワード検索条件
	if keyword != "" {
		query += " AND (content LIKE ? OR solution LIKE ? OR prompt LIKE ? OR subject LIKE ?)"
		searchPattern := "%" + keyword + "%"
		queryArgs = append(queryArgs, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// 科目での絞り込み
	if subject != "" {
		query += " AND subject = ?"
		queryArgs = append(queryArgs, subject)
	}

	// フィルター条件での絞り込み（部分一致・完全一致の切り替え）
	if filters != nil && len(filters) > 0 {
		if matchType == "exact" {
			// 完全一致検索：すべての条件が一致する必要がある
			for _, value := range filters {
				if valueArray, ok := value.([]interface{}); ok && len(valueArray) > 0 {
					// 配列の場合、すべての値が一致する必要がある
					for _, v := range valueArray {
						valueStr := fmt.Sprintf("%v", v)
						query += " AND filters LIKE ?"
						queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
					}
				} else if valueStr, ok := value.(string); ok && valueStr != "" {
					// 単一の文字列値での検索
					query += " AND filters LIKE ?"
					queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
				} else if value != nil {
					// その他の型での検索
					valueStr := fmt.Sprintf("%v", value)
					query += " AND filters LIKE ?"
					queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
				}
			}
		} else {
			// 部分一致検索（デフォルト）：いずれかの条件が一致すればよい
			allConditions := []string{}
			for _, value := range filters {
				if valueArray, ok := value.([]interface{}); ok && len(valueArray) > 0 {
					// 配列の場合、いずれかの値が一致すればよい
					for _, v := range valueArray {
						valueStr := fmt.Sprintf("%v", v)
						allConditions = append(allConditions, "filters LIKE ?")
						queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
					}
				} else if valueStr, ok := value.(string); ok && valueStr != "" {
					// 単一の文字列値での検索
					allConditions = append(allConditions, "filters LIKE ?")
					queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
				} else if value != nil {
					// その他の型での検索
					valueStr := fmt.Sprintf("%v", value)
					allConditions = append(allConditions, "filters LIKE ?")
					queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
				}
			}
			if len(allConditions) > 0 {
				query += fmt.Sprintf(" AND (%s)", strings.Join(allConditions, " OR "))
			}
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
		var problem models.Problem
		var filtersJSON []byte

		err := rows.Scan(
			&problem.ID,
			&problem.UserID,
			&problem.Subject,
			&problem.Prompt,
			&problem.Content,
			&problem.Solution,
			&problem.ImageBase64,
			&filtersJSON,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}

		if err := json.Unmarshal(filtersJSON, &problem.Filters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}

		problems = append(problems, &problem)
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
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filters: %w", err)
	}

	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at
		FROM problems
		WHERE user_id = ? AND subject = ? AND prompt = ? AND filters = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, subject, prompt, filtersJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by parameters: %w", err)
	}
	defer rows.Close()

	var problems []*models.Problem
	for rows.Next() {
		var problem models.Problem
		var filtersJSON []byte

		err := rows.Scan(
			&problem.ID,
			&problem.UserID,
			&problem.Subject,
			&problem.Prompt,
			&problem.Content,
			&problem.Solution,
			&problem.ImageBase64,
			&filtersJSON,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}

		if err := json.Unmarshal(filtersJSON, &problem.Filters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}

		problems = append(problems, &problem)
	}

	return problems, nil
}

func (r *MySQLProblemRepository) SearchByFilters(ctx context.Context, userID int64, subject string, filters map[string]interface{}, matchType string, limit, offset int) ([]*models.Problem, error) {
	// 基本クエリの構築
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at
		FROM problems
		WHERE user_id = ?`

	queryArgs := []interface{}{userID}

	// 科目での絞り込み
	if subject != "" {
		query += " AND subject = ?"
		queryArgs = append(queryArgs, subject)
	}

	// フィルター条件での絞り込み（部分一致・完全一致の切り替え）
	if matchType == "exact" {
		// 完全一致検索：すべての条件が一致する必要がある
		for _, value := range filters {
			if valueArray, ok := value.([]interface{}); ok && len(valueArray) > 0 {
				// 配列の場合、すべての値が一致する必要がある
				for _, v := range valueArray {
					valueStr := fmt.Sprintf("%v", v)
					query += " AND filters LIKE ?"
					queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
				}
			} else if valueStr, ok := value.(string); ok && valueStr != "" {
				// 単一の文字列値での検索
				query += " AND filters LIKE ?"
				queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
			} else if value != nil {
				// その他の型での検索
				valueStr := fmt.Sprintf("%v", value)
				query += " AND filters LIKE ?"
				queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
			}
		}
	} else {
		// 部分一致検索（デフォルト）：いずれかの条件が一致すればよい
		allConditions := []string{}
		for _, value := range filters {
			if valueArray, ok := value.([]interface{}); ok && len(valueArray) > 0 {
				// 配列の場合、いずれかの値が一致すればよい
				for _, v := range valueArray {
					valueStr := fmt.Sprintf("%v", v)
					allConditions = append(allConditions, "filters LIKE ?")
					queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
				}
			} else if valueStr, ok := value.(string); ok && valueStr != "" {
				// 単一の文字列値での検索
				allConditions = append(allConditions, "filters LIKE ?")
				queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
			} else if value != nil {
				// その他の型での検索
				valueStr := fmt.Sprintf("%v", value)
				allConditions = append(allConditions, "filters LIKE ?")
				queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", valueStr))
			}
		}
		if len(allConditions) > 0 {
			query += fmt.Sprintf(" AND (%s)", strings.Join(allConditions, " OR "))
		}
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to search problems by filters: %w", err)
	}
	defer rows.Close()

	var problems []*models.Problem
	for rows.Next() {
		var problem models.Problem
		var filtersJSON []byte

		err := rows.Scan(
			&problem.ID,
			&problem.UserID,
			&problem.Subject,
			&problem.Prompt,
			&problem.Content,
			&problem.Solution,
			&problem.ImageBase64,
			&filtersJSON,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}

		if err := json.Unmarshal(filtersJSON, &problem.Filters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}

		problems = append(problems, &problem)
	}

	return problems, nil
}

func (r *MySQLProblemRepository) SearchByKeyword(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*models.Problem, error) {
	query := `
		SELECT id, user_id, subject, prompt, content, solution, image_base64, filters, created_at, updated_at
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
		var problem models.Problem
		var filtersJSON []byte

		err := rows.Scan(
			&problem.ID,
			&problem.UserID,
			&problem.Subject,
			&problem.Prompt,
			&problem.Content,
			&problem.Solution,
			&problem.ImageBase64,
			&filtersJSON,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problem: %w", err)
		}

		if err := json.Unmarshal(filtersJSON, &problem.Filters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}

		problems = append(problems, &problem)
	}

	return problems, nil
}
