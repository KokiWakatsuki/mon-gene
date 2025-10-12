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

// runMigrations はデータベースマイグレーションファイルを実行します
func runMigrations(db *sqlx.DB) error {
	log.Printf("🔧 データベースマイグレーションを開始します...")
	
	return runMigrationFiles(db, "migrations")
}

// runMigrationFiles は指定されたディレクトリのマイグレーションファイルを順番に実行します
func runMigrationFiles(db *sqlx.DB, migrationDir string) error {
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("⚠️ マイグレーションディレクトリが存在しません: %s", migrationDir)
			return nil
		}
		return fmt.Errorf("マイグレーションディレクトリの読み込みに失敗: %w", err)
	}

	// .sqlファイルのみをフィルタリングしてソート
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 4 && file.Name()[len(file.Name())-4:] == ".sql" {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	if len(sqlFiles) == 0 {
		log.Printf("⚠️ マイグレーションファイルが見つかりません")
		return nil
	}

	// ファイルを順番に実行
	for _, filename := range sqlFiles {
		filepath := fmt.Sprintf("%s/%s", migrationDir, filename)
		log.Printf("📄 マイグレーション実行: %s", filename)
		
		content, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("マイグレーションファイルの読み込みに失敗 %s: %w", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("マイグレーションの実行に失敗 %s: %w", filename, err)
		}
		
		log.Printf("✅ マイグレーション完了: %s", filename)
	}
	
	log.Printf("🎉 全マイグレーションが完了しました")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
