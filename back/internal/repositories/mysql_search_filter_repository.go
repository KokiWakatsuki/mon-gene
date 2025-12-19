package repositories

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mon-gene/back/internal/models"
)

type MySQLSearchFilterRepository struct {
	db *sqlx.DB
}

func NewMySQLSearchFilterRepository(db *sqlx.DB) SearchFilterRepository {
	repo := &MySQLSearchFilterRepository{db: db}
	
	// seedデータを読み込む
	if err := repo.loadSeedData(); err != nil {
		log.Printf("Warning: Failed to load search_filter seed data: %v", err)
	}
	
	return repo
}

// loadSeedData はsearch_filter.csvからseedデータを読み込む
func (r *MySQLSearchFilterRepository) loadSeedData() error {
	ctx := context.Background()
	
	// すでにデータが存在するかチェック
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM search_filters")
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}
	
	// データが既に存在する場合はスキップ
	if count > 0 {
		log.Printf("Search filters already exist (%d records), skipping seed data load", count)
		return nil
	}
	
	// CSVファイルを開く
	file, err := os.Open("data/search_filter.csv")
	if err != nil {
		return fmt.Errorf("failed to open search_filter.csv: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	reader.LazyQuotes = true // JSON配列内のクォートを許容
	reader.FieldsPerRecord = -1 // フィールド数の不一致を許容
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) < 2 {
		log.Println("No search filter data to load")
		return nil
	}
	
	// ヘッダー行をスキップして各レコードを処理
	loadedCount := 0
	for i, record := range records[1:] {
		// デバッグ: レコードの内容を表示
		log.Printf("📋 Line %d: record length=%d, columns: %v", i+2, len(record), record)
		
		if len(record) < 11 {
			log.Printf("Skipping invalid record at line %d: insufficient columns (got %d, expected 11)", i+2, len(record))
			continue
		}
		
		// IDをパース
		id, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			log.Printf("Skipping record at line %d: invalid id: %v", i+2, err)
			continue
		}
		
		// UserIDをパース
		userID, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			log.Printf("Skipping record at line %d: invalid user_id: %v", i+2, err)
			continue
		}
		
		// UnitsをJSON文字列として処理
		// LazyQuotesにより、"["...""]"が複数のフィールドに分割される可能性がある
		var units sql.NullString
		unitsStartCol := 5
		
		// unitsフィールドを探す（["で始まり]"で終わる）
		if len(record) > unitsStartCol && record[unitsStartCol] != "" {
			// ["で始まるか確認
			if strings.HasPrefix(record[unitsStartCol], `["`) {
				// ]"で終わるカラムを探す
				unitsEndCol := unitsStartCol
				for j := unitsStartCol; j < len(record); j++ {
					if strings.HasSuffix(record[j], `]"`) {
						unitsEndCol = j
						break
					}
				}
				
				// problem_repository.goと同じアプローチ：カンマで結合してから修正
				unitsParts := record[unitsStartCol : unitsEndCol+1]
				unitsStr := strings.Join(unitsParts, ",")
				
				// CSVパーサーによる分割で欠けた引用符を修正
				// Step 1: ["三平方の定理, -> ["三平方の定理",
				// 最初の要素の閉じクォートを追加（最初のカンマの前に"を挿入）
				firstCommaIdx := strings.Index(unitsStr, ",")
				if firstCommaIdx > 0 && !strings.Contains(unitsStr[:firstCommaIdx+1], `",`) {
					unitsStr = unitsStr[:firstCommaIdx] + `"` + unitsStr[firstCommaIdx:]
				}
				
				// Step 2: 先頭の["を削除して[に
				unitsStr = strings.TrimPrefix(unitsStr, `["`)
				unitsStr = `["` + unitsStr
				
				// Step 3: 末尾の]"を削除して]に
				unitsStr = strings.TrimSuffix(unitsStr, `]"`)
				unitsStr = unitsStr + `]`
				
				// JSON配列として検証
				var testArray []string
				if err := json.Unmarshal([]byte(unitsStr), &testArray); err == nil {
					units = sql.NullString{String: unitsStr, Valid: true}
					log.Printf("✅ Line %d: Successfully parsed units: %v", i+2, testArray)
				} else {
					log.Printf("⚠️ Line %d: Invalid JSON for units: %v (data: %s)", i+2, err, unitsStr)
				}
			}
		}
		
		// Year, ExamSession, IsChecked, CreatedAt, UpdatedAtは最後の5フィールド
		yearCol := len(record) - 5
		examSessionCol := len(record) - 4
		isCheckedCol := len(record) - 3
		createdAtCol := len(record) - 2
		updatedAtCol := len(record) - 1
		
		// Year
		year := ""
		if yearCol >= 0 && yearCol < len(record) {
			year = strings.TrimSpace(record[yearCol])
		}
		
		// ExamSession
		examSession := ""
		if examSessionCol >= 0 && examSessionCol < len(record) {
			examSession = strings.TrimSpace(record[examSessionCol])
		}
		
		// IsCheckedをパース
		var isChecked sql.NullBool
		if isCheckedCol >= 0 && isCheckedCol < len(record) && record[isCheckedCol] != "" {
			val, err := strconv.ParseBool(record[isCheckedCol])
			if err == nil {
				isChecked = sql.NullBool{Bool: val, Valid: true}
			}
		}
		
		// CreatedAtをパース
		var createdAt time.Time
		if createdAtCol >= 0 && createdAtCol < len(record) && record[createdAtCol] != "" && record[createdAtCol] != "NULL" {
			var err error
			createdAt, err = time.Parse("2006-01-02 15:04:05", record[createdAtCol])
			if err != nil {
				log.Printf("Warning: invalid created_at at line %d, using current time: %v", i+2, err)
				createdAt = time.Now()
			}
		} else {
			createdAt = time.Now()
		}
		
		// UpdatedAtをパース
		var updatedAt time.Time
		if updatedAtCol >= 0 && updatedAtCol < len(record) && record[updatedAtCol] != "" && record[updatedAtCol] != "NULL" {
			var err error
			updatedAt, err = time.Parse("2006-01-02 15:04:05", record[updatedAtCol])
			if err != nil {
				log.Printf("Warning: invalid updated_at at line %d, using current time: %v", i+2, err)
				updatedAt = time.Now()
			}
		} else {
			updatedAt = time.Now()
		}
		
		// データベースに挿入
		query := `
			INSERT INTO search_filters (id, user_id, name, keyword, subject, units, year, exam_session, is_checked, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		
		_, err = r.db.ExecContext(ctx, query,
			id,
			userID,
			record[2], // name
			record[3], // keyword
			record[4], // subject
			units,
			year,
			examSession,
			isChecked,
			createdAt,
			updatedAt,
		)
		
		if err != nil {
			log.Printf("Failed to insert search filter at line %d: %v", i+2, err)
			continue
		}
		
		loadedCount++
	}
	
	log.Printf("Successfully loaded %d search filter records from CSV", loadedCount)
	return nil
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