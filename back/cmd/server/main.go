package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mon-gene/back/internal/api/handlers"
	"github.com/mon-gene/back/internal/api/routes"
	"github.com/mon-gene/back/internal/clients"
	"github.com/mon-gene/back/internal/repositories"
	"github.com/mon-gene/back/internal/services"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// サービスの初期化
	emailService := services.NewEmailService()
	
	// 実際のクライアントを初期化（空のモデル名で初期化、ユーザー設定に基づいて動的に作成）
	claudeClient := clients.NewClaudeClient("")  // ユーザー設定に基づいて動的に作成
	openaiClient := clients.NewOpenAIClient("")  // ユーザー設定に基づいて動的に作成
	googleClient := clients.NewGoogleClient("")  // ユーザー設定に基づいて動的に作成
	coreClient := clients.NewCoreClient()
	
	// メモリベースのリポジトリを初期化
	userRepo := repositories.NewMemoryUserRepository()
	sessionRepo := repositories.NewMemorySessionRepository()
	
	log.Printf("✅ リポジトリを初期化しました（メモリベース）")
	log.Printf("📧 seedデータ: 塾コード=00000, メール=nutfes.script@gmail.com")
	log.Printf("🤖 AIクライアントを初期化しました（Claude, OpenAI, Google）")
	
	// サービスを初期化（実際のリポジトリを使用）
	authService := services.NewAuthService(userRepo, sessionRepo, emailService)
	problemService := services.NewProblemService(claudeClient, openaiClient, googleClient, coreClient, nil, userRepo)

	// ハンドラーの初期化
	authHandler := handlers.NewAuthHandler(authService)
	problemHandler := handlers.NewProblemHandler(problemService, authService)
	healthHandler := handlers.NewHealthHandler()

	// ルーターの設定
	router := routes.NewRouter(authHandler, problemHandler, healthHandler)

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
	log.Printf("  - POST /api/generate-pdf")
	
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
