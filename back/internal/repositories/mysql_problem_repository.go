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
