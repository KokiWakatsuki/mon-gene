package repositories

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/utils"
)

type MySQLUserRepository struct {
	db *sqlx.DB
}

func NewMySQLUserRepository(db *sqlx.DB) UserRepository {
	repo := &MySQLUserRepository{db: db}
	
	// CSVからseedデータを読み込み
	if err := repo.loadSeedData(); err != nil {
		log.Printf("⚠️ seedデータの読み込みに失敗: %v", err)
	}
	
	return repo
}

func (r *MySQLUserRepository) GetBySchoolCode(ctx context.Context, schoolCode string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, school_code, email, password_hash, problem_generation_limit,
			   problem_generation_count, figure_regeneration_limit, figure_regeneration_count,
			   preview_limit, preview_count,
			   role, preferred_api, preferred_model, created_at, updated_at
		FROM users WHERE school_code = ?
	`
	
	err := r.db.Get(user, query, schoolCode)
	if err != nil {
		return nil, fmt.Errorf("ユーザーが見つかりません: %w", err)
	}
	
	return user, nil
}

func (r *MySQLUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, school_code, email, password_hash, problem_generation_limit,
			   problem_generation_count, figure_regeneration_limit, figure_regeneration_count,
			   preview_limit, preview_count,
			   role, preferred_api, preferred_model, created_at, updated_at
		FROM users WHERE id = ?
	`
	
	err := r.db.Get(user, query, id)
	if err != nil {
		return nil, fmt.Errorf("ユーザーが見つかりません: %w", err)
	}
	
	return user, nil
}

func (r *MySQLUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (school_code, email, password_hash, problem_generation_limit, 
						  problem_generation_count, figure_regeneration_limit, figure_regeneration_count,
						  role, preferred_api, preferred_model)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, 
		user.SchoolCode, user.Email, user.PasswordHash, user.ProblemGenerationLimit,
		user.ProblemGenerationCount, user.FigureRegenerationLimit, user.FigureRegenerationCount,
		user.Role, user.PreferredAPI, user.PreferredModel)
	if err != nil {
		return fmt.Errorf("ユーザーの作成に失敗: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("IDの取得に失敗: %w", err)
	}
	
	user.ID = id
	return nil
}

func (r *MySQLUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = ?, password_hash = ?, problem_generation_limit = ?, 
			problem_generation_count = ?, figure_regeneration_limit = ?, figure_regeneration_count = ?,
			role = ?, preferred_api = ?, preferred_model = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, 
		user.Email, user.PasswordHash, user.ProblemGenerationLimit, user.ProblemGenerationCount,
		user.FigureRegenerationLimit, user.FigureRegenerationCount,
		user.Role, user.PreferredAPI, user.PreferredModel, user.ID)
	if err != nil {
		return fmt.Errorf("ユーザーの更新に失敗: %w", err)
	}
	
	return nil
}

func (r *MySQLUserRepository) UpdateGenerationCount(userID int64, count int) error {
	query := `UPDATE users SET problem_generation_count = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, count, userID)
	if err != nil {
		return fmt.Errorf("生成回数の更新に失敗: %w", err)
	}
	
	return nil
}

func (r *MySQLUserRepository) UpdateFigureRegenerationCount(userID int64, count int) error {
	query := `UPDATE users SET figure_regeneration_count = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, count, userID)
	if err != nil {
		return fmt.Errorf("図形再生成回数の更新に失敗: %w", err)
	}
	
	return nil
}

// IncrementPreviewCount プレビュー回数をインクリメント
func (r *MySQLUserRepository) IncrementPreviewCount(ctx context.Context, userID int64) error {
	query := `UPDATE users SET preview_count = preview_count + 1 WHERE id = ?`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("プレビュー回数の更新に失敗: %w", err)
	}
	
	return nil
}

// loadSeedData はCSVファイルからseedデータを読み込んでデータベースに挿入します
func (r *MySQLUserRepository) loadSeedData() error {
	// 既存のユーザー数をチェック
	var count int
	if err := r.db.Get(&count, "SELECT COUNT(*) FROM users"); err != nil {
		return fmt.Errorf("ユーザー数の取得に失敗: %w", err)
	}
	
	// 既にユーザーが存在する場合はseedデータの読み込みをスキップ
	if count > 0 {
		log.Printf("✅ 既存のユーザーが%d件存在するため、seedデータの読み込みをスキップします", count)
		return nil
	}

	file, err := os.Open("data/users.csv")
	if err != nil {
		return fmt.Errorf("CSVファイルの読み込みに失敗: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("CSV解析に失敗: %w", err)
	}

	// ヘッダーをスキップ
	for i, record := range records[1:] {
		if len(record) < 11 {
			log.Printf("⚠️ 行%d: データが不完全です: %v", i+2, record)
			continue
		}

		limit, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("⚠️ 行%d: 問題生成制限数の解析に失敗: %v", i+2, err)
			continue
		}

		generationCount, err := strconv.Atoi(record[5])
		if err != nil {
			log.Printf("⚠️ 行%d: 問題生成カウントの解析に失敗: %v", i+2, err)
			continue
		}

		figureLimit, err := strconv.Atoi(record[6])
		if err != nil {
			log.Printf("⚠️ 行%d: 図形再生成制限数の解析に失敗: %v", i+2, err)
			continue
		}

		figureCount, err := strconv.Atoi(record[7])
		if err != nil {
			log.Printf("⚠️ 行%d: 図形再生成カウントの解析に失敗: %v", i+2, err)
			continue
		}

		// パスワードをハッシュ化
		hashedPassword, err := utils.HashPassword(record[3])
		if err != nil {
			log.Printf("⚠️ 行%d: パスワードハッシュ化に失敗: %v", i+2, err)
			continue
		}

		user := &models.User{
			SchoolCode:              record[1],
			Email:                   record[2],
			PasswordHash:           hashedPassword,
			ProblemGenerationLimit: limit,
			ProblemGenerationCount: generationCount,
			FigureRegenerationLimit: figureLimit,
			FigureRegenerationCount: figureCount,
			Role:                   record[8],
			PreferredAPI:           record[9],
			PreferredModel:         record[10],
		}

		if err := r.Create(context.Background(), user); err != nil {
			log.Printf("⚠️ 行%d: ユーザー作成に失敗: %v", i+2, err)
			continue
		}

		log.Printf("📝 ユーザー追加: SchoolCode=%s, Email=%s, Role=%s, API=%s, Model=%s", 
			user.SchoolCode, user.Email, user.Role, user.PreferredAPI, user.PreferredModel)
	}

	log.Printf("✅ CSVファイルから %d 人のユーザーを読み込みました", len(records)-1)
	return nil
}
