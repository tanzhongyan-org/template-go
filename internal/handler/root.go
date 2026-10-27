package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tanzhongyan-org/template-go/internal/config"
)

// RootHandler answers GET / with a plain-text greeting.
func RootHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := fmt.Fprintf(w, "Welcome to %s\n", cfg.AppName); err != nil {
			log.Printf("write root response: %v", err)
		}
	}
}

type apiWelcome struct {
	Message string `json:"message"`
}

// APIRootHandler answers GET /api/v1 with a JSON greeting.
func APIRootHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, apiWelcome{Message: fmt.Sprintf("Welcome to %s API v1", cfg.AppName)})
	}
}
