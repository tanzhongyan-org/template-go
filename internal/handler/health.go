package handler

import (
	"net/http"
	"time"

	"github.com/tanzhongyan-org/template-go/internal/config"
)

// HealthStatus represents the health check response payload.
type HealthStatus struct {
	Status      string    `json:"status"`
	AppName     string    `json:"appName"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
}

// HealthHandler returns an http.HandlerFunc that writes the health status JSON.
func HealthHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := HealthStatus{
			Status:      "healthy",
			AppName:     cfg.AppName,
			Environment: cfg.AppEnv,
			Timestamp:   time.Now().UTC(),
		}

		writeJSON(w, http.StatusOK, payload)
	}
}
