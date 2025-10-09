package utils

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// WriteErrorResponse writes an error response to the client
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
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
	// 環境に応じてオリジンを設定
	allowedOrigins := []string{
		"http://localhost:3000",           // 開発環境
		"https://mon-gene.wakatsuki.app",  // 本番環境
	}
	
	// すべての許可されたオリジンを設定（簡単な実装）
	// 本来はリクエストのOriginヘッダーをチェックして適切なものを返すべき
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
