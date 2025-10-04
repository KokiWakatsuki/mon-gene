package middleware

import (
	"net/http"

	"github.com/mon-gene/back/internal/utils"
)

// CORSMiddleware handles CORS preflight requests and sets CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		utils.EnableCORS(w)

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}
