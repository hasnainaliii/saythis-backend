package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"saythis-backend/internal/config"
	"saythis-backend/internal/database"
	"saythis-backend/internal/server"
)

func main() {

	logger := log.New(os.Stdout, "[SAYTHIS] ", log.LstdFlags|log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("❌ Failed to load configuration: %v", err)
	}

	pool, err := database.Connect(cfg.DataBaseUrl)
	if err != nil {
		logger.Fatalf("❌ Failed to connect to database: %v", err)
	}

	defer func() {
		logger.Println("🔌 Closing database connection...")
		pool.Close()
	}()

	router := server.NewRouter(pool, cfg, logger)

	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Printf("🚀 Server starting on http://localhost%s", cfg.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Fatalf("❌ Server failed: %v", err)
		}

	case sig := <-shutdown:
		logger.Printf("🛑 Shutdown signal received: %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Printf("⚠️  Graceful shutdown failed, forcing close: %v", err)
			if closeErr := srv.Close(); closeErr != nil {
				logger.Fatalf("❌ Server close failed: %v", closeErr)
			}
		}

		logger.Println("✅ Server stopped gracefully")
	}
}
