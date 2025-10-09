package routes

import (
	"net/http"

	"github.com/mon-gene/back/internal/api/handlers"
	"github.com/mon-gene/back/internal/api/middleware"
	"github.com/mon-gene/back/internal/utils"
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
	
	// User info endpoint (supports GET and OPTIONS)
	mux.HandleFunc("/api/user-info", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET", "OPTIONS":
			authHandler.GetUserInfo(w, r)
		default:
			utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Problem generation endpoints
	mux.HandleFunc("/api/generate-problem", problemHandler.GenerateProblem)
	mux.HandleFunc("/api/generate-pdf", problemHandler.GeneratePDF)

	// Apply CORS middleware to all routes
	return middleware.CORSMiddleware(mux)
}
