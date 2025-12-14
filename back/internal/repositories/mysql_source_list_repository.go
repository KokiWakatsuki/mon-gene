package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/mon-gene/back/internal/models"
)

type MySQLSourceListRepository struct {
	db *sqlx.DB
}

func NewMySQLSourceListRepository(db *sqlx.DB) *MySQLSourceListRepository {
	return &MySQLSourceListRepository{db: db}
}

// Create は新しい出典リスト項目を作成
func (r *MySQLSourceListRepository) Create(item *models.SourceListItem) error {
	query := `
		INSERT INTO source_list (user_id, year, exam_session)
		VALUES (:user_id, :year, :exam_session)
		ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)
	`
	result, err := r.db.NamedExec(query, item)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = id
	return nil
}

// GetByUserID はユーザーIDで出典リストを取得
func (r *MySQLSourceListRepository) GetByUserID(userID int64) ([]*models.SourceListItem, error) {
	var items []*models.SourceListItem
	query := `SELECT * FROM source_list WHERE user_id = ? ORDER BY created_at DESC`
	err := r.db.Select(&items, query, userID)
	return items, err
}

// Delete は出典リスト項目を削除
func (r *MySQLSourceListRepository) Delete(id int64, userID int64) error {
	query := `DELETE FROM source_list WHERE id = ? AND user_id = ?`
	_, err := r.db.Exec(query, id, userID)
	return err
}

// DeleteByYearAndExam は年度と回数で出典リスト項目を削除
func (r *MySQLSourceListRepository) DeleteByYearAndExam(userID int64, year string, examSession string) error {
	query := `DELETE FROM source_list WHERE user_id = ? AND year = ? AND exam_session = ?`
	_, err := r.db.Exec(query, userID, year, examSession)
	return err
}