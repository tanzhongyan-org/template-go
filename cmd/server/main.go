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
	dbschema "github.com/tanzhongyan-org/template-go/sql"
)

// maxBodyBytes caps request bodies at the HTTP boundary.
const maxBodyBytes = 1 << 20

// SetupRouter initializes the Chi router with standard middleware and routes.
func SetupRouter(cfg *config.Config, queries *db.Queries) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handler.LimitBody(maxBodyBytes))

	healthFn := handler.HealthHandler(cfg)

	r.Get("/", handler.RootHandler(cfg))
	r.Get("/health", healthFn)

	var notesH *handler.NotesHandler
	if queries != nil {
		notesH = handler.NewNotesHandler(queries, cfg)
		r.Mount("/notes", notesH.Routes())
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthFn)
		r.Get("/", handler.APIRootHandler(cfg))
		if notesH != nil {
			r.Mount("/notes", notesH.Routes())
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

	if _, err := pool.Exec(ctx, dbschema.Schema); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	queries := db.New(pool)
	router := SetupRouter(cfg, queries)

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
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
