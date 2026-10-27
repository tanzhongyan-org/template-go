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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanzhongyan-org/template-go/internal/config"
	"github.com/tanzhongyan-org/template-go/internal/db"
	"github.com/tanzhongyan-org/template-go/internal/handler"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS notes (
    id VARCHAR(36) PRIMARY KEY,
    owner_id VARCHAR(255) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notes_owner_id ON notes (owner_id);
`

// SetupRouter initializes the Chi router with standard middleware and routes.
func SetupRouter(cfg *config.Config, queries *db.Queries) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	healthFn := handler.HealthHandler(cfg)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "Welcome to %s\n", cfg.AppName)
	})
	r.Get("/health", healthFn)

	if queries != nil {
		notesH := handler.NewNotesHandler(queries)
		r.Route("/notes", func(r chi.Router) {
			r.Get("/", notesH.List)
			r.Post("/", notesH.Create)
			r.Get("/{id}", notesH.Get)
			r.Put("/{id}", notesH.Update)
			r.Delete("/{id}", notesH.Delete)
		})
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthFn)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"message":"Welcome to %s API v1"}`+"\n", cfg.AppName)
		})
		if queries != nil {
			notesH := handler.NewNotesHandler(queries)
			r.Route("/notes", func(r chi.Router) {
				r.Get("/", notesH.List)
				r.Post("/", notesH.Create)
				r.Get("/{id}", notesH.Get)
				r.Put("/{id}", notesH.Update)
				r.Delete("/{id}", notesH.Delete)
			})
		}
	})

	return r
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	queries := db.New(pool)
	router := SetupRouter(cfg, queries)

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting %s on %s...", cfg.AppName, addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, initiating graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			log.Fatalf("Error closing server: %v", closeErr)
		}
	}
	log.Println("Server stopped gracefully")
}