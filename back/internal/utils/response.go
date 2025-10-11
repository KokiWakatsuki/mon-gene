package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// WriteErrorResponse writes an error response to the client
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	EnableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Success: false,
		Error:   message,
	}
	
	json.NewEncoder(w).Encode(response)
}

// WriteJSONResponse writes a JSON response to the client
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	EnableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, write a simple error response
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": "Internal server error"}`))
	}
}

// EnableCORS enables CORS for the response
func EnableCORS(w http.ResponseWriter) {
	// 環境変数からALLOWED_ORIGINSを取得
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// デフォルトは本番環境のURL
		allowedOrigins = "https://mon-gene.wakatsuki.app"
	}
	
	// 開発環境の場合は複数のOriginを許可
	environment := os.Getenv("ENVIRONMENT")
	if environment == "development" {
		// 開発環境では複数のOriginをカンマ区切りで許可
		origins := strings.Split(allowedOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		// 最初のOriginを使用（通常はlocalhost）、または全て許可
		if len(origins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", origins[0])
			// 実際には、より安全に全てのOriginをチェックするべきですが、開発環境なので簡略化
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
	} else {
		// 本番環境では最初のOrigin（通常は本番URL）を使用
		origins := strings.Split(allowedOrigins, ",")
		if len(origins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", strings.TrimSpace(origins[len(origins)-1])) // 最後のOrigin（本番URL）
		}
	}
	
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "3600")
}
