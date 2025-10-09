package routes

import (
	"net/http"

	"github.com/mon-gene/back/internal/api/handlers"
	"github.com/mon-gene/back/internal/api/middleware"
)

// Router sets up all the routes for the application
func NewRouter(
	authHandler *handlers.AuthHandler,
	problemHandler *handlers.ProblemHandler,
	healthHandler *handlers.HealthHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/", healthHandler.Health)
	mux.HandleFunc("/health", healthHandler.Health)

	// Authentication endpoints
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("/api/logout", authHandler.Logout)
	mux.HandleFunc("/api/user-info", authHandler.GetUserInfo)

	// Problem generation endpoints
	mux.HandleFunc("/api/generate-problem", problemHandler.GenerateProblem)
	mux.HandleFunc("/api/generate-pdf", problemHandler.GeneratePDF)

	// Apply CORS middleware to all routes
	return middleware.CORSMiddleware(mux)
}
