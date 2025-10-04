package handlers

import (
	"net/http"

	"github.com/mon-gene/back/internal/utils"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "ok",
		"message": "Mongene Backend Server is running",
		"service": "mongene-backend",
		"version": "1.0.0",
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}
