package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mon-gene/back/internal/models"
)

type MySQLSearchFilterRepository struct {
	db *sqlx.DB
}

func NewMySQLSearchFilterRepository(db *sqlx.DB) SearchFilterRepository {
	return &MySQLSearchFilterRepository{db: db}
}

func (r *MySQLSearchFilterRepository) Create(ctx context.Context, filter *models.SearchFilter) error {
	query := `
		INSERT INTO search_filters (user_id, name, keyword, subject, units, year, exam_session, is_checked, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	
	result, err := r.db.ExecContext(ctx, query,
		filter.UserID,
		filter.Name,
		filter.Keyword,
		filter.Subject,
		filter.Units,
		filter.Year,
		filter.ExamSession,
		filter.IsChecked,
	)
	if err != nil {
		return fmt.Errorf("failed to create search filter: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	filter.ID = id
	return nil
}

func (r *MySQLSearchFilterRepository) GetByID(ctx context.Context, id int64) (*models.SearchFilter, error) {
	query := `
		SELECT id, user_id, name, keyword, subject, units, year, exam_session, is_checked, created_at, updated_at
		FROM search_filters
		WHERE id = ?
	`

	var filter models.SearchFilter
	err := r.db.GetContext(ctx, &filter, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("search filter not found")
		}
		return nil, fmt.Errorf("failed to get search filter: %w", err)
	}

	return &filter, nil
}

func (r *MySQLSearchFilterRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.SearchFilter, error) {
	query := `
		SELECT id, user_id, name, keyword, subject, units, year, exam_session, is_checked, created_at, updated_at
		FROM search_filters
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	var filters []*models.SearchFilter
	err := r.db.SelectContext(ctx, &filters, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get search filters: %w", err)
	}

	return filters, nil
}

func (r *MySQLSearchFilterRepository) Update(ctx context.Context, filter *models.SearchFilter) error {
	query := `
		UPDATE search_filters
		SET name = ?, keyword = ?, subject = ?, units = ?, year = ?, exam_session = ?, is_checked = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		filter.Name,
		filter.Keyword,
		filter.Subject,
		filter.Units,
		filter.Year,
		filter.ExamSession,
		filter.IsChecked,
		filter.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update search filter: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("search filter not found")
	}

	return nil
}

func (r *MySQLSearchFilterRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM search_filters WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete search filter: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("search filter not found")
	}

	return nil
}