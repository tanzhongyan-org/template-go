package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tanzhongyan-org/template-go/internal/api/handlers"
	"github.com/tanzhongyan-org/template-go/internal/core/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.Handle("/health", handlers.NewHealthHandler(cfg))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprintf(w, "Welcome to %s", cfg.AppName)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting %s on %s...", cfg.AppName, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
