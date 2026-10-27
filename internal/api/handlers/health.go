package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tanzhongyan-org/template-go/internal/core/config"
	"github.com/tanzhongyan-org/template-go/internal/domain/health"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	cfg *config.Config
}

// NewHealthHandler constructs a HealthHandler instance.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

// ServeHTTP writes the health check JSON response.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := health.GetStatus(h.cfg.AppName, h.cfg.AppEnv)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
