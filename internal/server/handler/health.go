package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
	Version   string            `json:"version"`
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
		Version:   "1.0.0", // This should come from build info
	}

	// Check database connectivity
	if db, ok := h.getDB(); ok {
		if err := db.Ping(); err != nil {
			status.Status = "unhealthy"
			status.Services["database"] = "unhealthy"
		} else {
			status.Services["database"] = "healthy"
		}
	} else {
		status.Status = "unhealthy"
		status.Services["database"] = "unknown"
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Set status code based on health
	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Encode response
	json.NewEncoder(w).Encode(status)
}

func (h *Handler) getDB() (*sqlx.DB, bool) {
	return h.DB, true
}
