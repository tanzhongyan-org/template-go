package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tanzhongyan-org/template-go/internal/api/handlers"
	"github.com/tanzhongyan-org/template-go/internal/core/config"
)

// SetupRouter initializes the Chi router with standard middleware and routes.
func SetupRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Production middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	healthHandler := handlers.NewHealthHandler(cfg)

	// Root routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "Welcome to %s\n", cfg.AppName)
	})
	r.Get("/health", healthHandler.ServeHTTP)

	// API v1 route group
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.ServeHTTP)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"message":"Welcome to %s API v1"}`+"\n", cfg.AppName)
		})
	})

	return r
}

func main() {
	cfg := config.Load()
	router := SetupRouter(cfg)

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Starting %s on %s...", cfg.AppName, addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("Shutdown signal %v received, initiating graceful shutdown...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if closeErr := server.Close(); closeErr != nil {
				log.Fatalf("Error closing server: %v", closeErr)
			}
		}
		log.Println("Server stopped gracefully")
	}
}
