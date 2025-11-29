package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mon-gene/back/internal/api/handlers"
	"github.com/mon-gene/back/internal/api/routes"
	"github.com/mon-gene/back/internal/clients"
	"github.com/mon-gene/back/internal/config"
	"github.com/mon-gene/back/internal/repositories"
	"github.com/mon-gene/back/internal/services"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// データベース接続の初期化（リトライ機能付き）
	dbConfig := config.LoadDatabaseConfig()
	db, err := config.NewDatabaseWithRetry(dbConfig)
	if err != nil {
		log.Printf("❌ データベース接続に失敗しました: %v", err)
		log.Printf("⚠️ メモリベースのリポジトリを使用します")
	} else {
		defer db.Close()
	}

	// サービスの初期化
	emailService := services.NewEmailService()
	
	// 実際のクライアントを初期化（空のモデル名で初期化、ユーザー設定に基づいて動的に作成）
	claudeClient := clients.NewClaudeClient("")  // ユーザー設定に基づいて動的に作成
	openaiClient := clients.NewOpenAIClient("")  // ユーザー設定に基づいて動的に作成
	googleClient := clients.NewGoogleClient("")  // ユーザー設定に基づいて動的に作成
	coreClient := clients.NewCoreClient()
	
	// リポジトリを初期化（データベース接続が成功した場合はMySQL、失敗した場合はメモリベース）
	var userRepo repositories.UserRepository
	var sessionRepo repositories.SessionRepository
	var problemRepo repositories.ProblemRepository
	
	if db != nil {
		// MySQLベースのリポジトリを使用
		userRepo = repositories.NewMySQLUserRepository(db)
		sessionRepo = repositories.NewMemorySessionRepository() // Sessionは引き続きメモリベース
		problemRepo = repositories.NewMySQLProblemRepository(db)
		log.Printf("✅ MySQLベースのリポジトリを初期化しました")
	} else {
		// メモリベースのリポジトリを使用
		userRepo = repositories.NewMemoryUserRepository()
		sessionRepo = repositories.NewMemorySessionRepository()
		problemRepo = nil
		log.Printf("✅ メモリベースのリポジトリを初期化しました")
	}
	
	log.Printf("📧 seedデータ: 塾コード=00000, メール=nutfes.script@gmail.com")
	log.Printf("🤖 AIクライアントを初期化しました（Claude, OpenAI, Google）")
	
	// サービスを初期化
	authService := services.NewAuthService(userRepo, sessionRepo, emailService)
	problemService := services.NewProblemService(claudeClient, openaiClient, googleClient, coreClient, problemRepo, userRepo)

	// ハンドラーの初期化
	authHandler := handlers.NewAuthHandler(authService)
	problemHandler := handlers.NewProblemHandler(problemService, authService)
	healthHandler := handlers.NewHealthHandler()
	chatHandler := handlers.NewChatHandler(authService)
	sseHandler := handlers.NewSSEHandler(problemService, authService)

	// ルーターの設定
	router := routes.NewRouter(authHandler, problemHandler, healthHandler, chatHandler, sseHandler)

	// サーバーの起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Mongene Backend Server starting on port %s", port)
	log.Printf("📋 Available endpoints:")
	log.Printf("  - GET  /health")
	log.Printf("  - POST /api/login")
	log.Printf("  - POST /api/forgot-password")
	log.Printf("  - POST /api/logout")
	log.Printf("  - POST /api/generate-problem")
	log.Printf("  - POST /api/generate-problem-five-stage")
	log.Printf("  - POST /api/generate-problem-five-stage-sse (SSE)")
	log.Printf("  - POST /api/generate-three-problems-sse (SSE)")
	log.Printf("  - POST /api/generate-pdf")
	log.Printf("  - GET  /api/problems/search?keyword=<keyword>")
	log.Printf("  - GET  /api/problems/history")
	log.Printf("  - POST /api/chat")
	
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
