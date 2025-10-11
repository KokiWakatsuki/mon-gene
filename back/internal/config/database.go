package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func LoadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "user"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "develop"),
	}
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func NewDatabase(config *DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 接続プールの設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// NewDatabaseWithRetry はリトライ機能付きでデータベースに接続し、マイグレーションを実行します
func NewDatabaseWithRetry(config *DatabaseConfig) (*sqlx.DB, error) {
	maxRetries := 30
	retryInterval := 2 * time.Second

	log.Printf("📦 データベース接続を開始します: %s@%s:%s/%s", config.User, config.Host, config.Port, config.DBName)

	for i := 0; i < maxRetries; i++ {
		db, err := sqlx.Connect("mysql", config.DSN())
		if err == nil {
			// 接続プールの設定
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(5)
			
			// 接続テスト
			if pingErr := db.Ping(); pingErr == nil {
				log.Printf("✅ データベース接続成功: %s@%s:%s/%s", config.User, config.Host, config.Port, config.DBName)
				
				// マイグレーションを実行
				if migErr := runMigrations(db); migErr != nil {
					log.Printf("⚠️ マイグレーション警告: %v", migErr)
				}
				
				return db, nil
			} else {
				db.Close()
				err = pingErr
			}
		}

		if i == 0 {
			log.Printf("⏳ データベースが起動するまで待機中... (最大%d回試行)", maxRetries)
		}
		
		if i < maxRetries-1 {
			log.Printf("⏳ リトライ %d/%d: %v", i+1, maxRetries, err)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("データベース接続に失敗しました (最大%d回試行): %w", maxRetries, fmt.Errorf("connection timeout"))
}

// runMigrations はデータベースマイグレーションを実行します
func runMigrations(db *sqlx.DB) error {
	log.Printf("🔧 データベースマイグレーションを開始します...")
	
	// usersテーブルを作成
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		school_code VARCHAR(10) NOT NULL UNIQUE COMMENT '塾コード',
		email VARCHAR(255) NOT NULL COMMENT 'メールアドレス',
		password_hash VARCHAR(255) NOT NULL COMMENT 'パスワード（ハッシュ化）',
		problem_generation_limit INT NOT NULL DEFAULT 10 COMMENT '問題生成制限数（-1は無制限）',
		problem_generation_count INT NOT NULL DEFAULT 0 COMMENT '現在の生成回数',
		role ENUM('teacher', 'admin') NOT NULL DEFAULT 'teacher' COMMENT 'ユーザーロール',
		preferred_api VARCHAR(50) NOT NULL DEFAULT 'claude' COMMENT '優先API',
		preferred_model VARCHAR(100) NOT NULL DEFAULT 'claude-3-5-sonnet-20241022' COMMENT '優先モデル',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_school_code (school_code),
		INDEX idx_email (email)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ユーザーテーブル'`
	
	if _, err := db.Exec(createUsersTableSQL); err != nil {
		return fmt.Errorf("usersテーブルの作成に失敗: %w", err)
	}
	log.Printf("✅ usersテーブルを作成/確認しました")

	// problemsテーブルを作成
	createProblemsTableSQL := `
	CREATE TABLE IF NOT EXISTS problems (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		subject VARCHAR(100) NOT NULL COMMENT '科目（数学、物理など）',
		prompt TEXT NOT NULL COMMENT '生成時のプロンプト',
		content TEXT NOT NULL COMMENT '問題文',
		solution TEXT COMMENT '解答',
		image_base64 LONGTEXT COMMENT '図（Base64エンコード）',
		filters JSON COMMENT '生成パラメータ（フィルタ条件）',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_subject (subject),
		INDEX idx_created_at (created_at),
		FULLTEXT INDEX idx_fulltext_search (content, solution, prompt, subject),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='問題テーブル'`
	
	if _, err := db.Exec(createProblemsTableSQL); err != nil {
		return fmt.Errorf("problemsテーブルの作成に失敗: %w", err)
	}
	log.Printf("✅ problemsテーブルを作成/確認しました")
	
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
