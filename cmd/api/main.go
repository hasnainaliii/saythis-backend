package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"saythis-backend/internal/config"
	"saythis-backend/internal/database"
	"saythis-backend/internal/server"

	"go.uber.org/zap"
)

func main() {

	// *******************
	// Logger initialization
	// *******************
	if err := config.InitLogger(); err != nil {
		panic("Logger is not initialized")
	}
	defer config.Sync()

	// *******************
	// Env initialization
	// *******************
	cfg, err := config.Load()
	if err != nil {
		zap.S().Fatalf("❌ Failed to load configuration: %v", err)
	}

	// *******************
	// Database initialization
	// *******************
	pool, err := database.Connect(cfg.DataBaseURL)
	if err != nil {
		zap.S().Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		zap.S().Info("🔌 Closing database connection...")
		pool.Close()
	}()

	// *******************
	// Router initialization
	// *******************
	router := server.NewRouter(pool, cfg)
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		zap.S().Info("🚀 Server is starting http://localhost", cfg.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	// *******************
	// Graceful Shutdown
	// *******************

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("❌ Server failed: %v", err)
		}
	case sig := <-shutDown:
		zap.S().Info("🛑 Shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			zap.S().Error("⚠️  Graceful shutdown failed, forcing close: ", err)
			if closeErr := srv.Close(); closeErr != nil {
				zap.S().Error("❌ Server close failed: ", closeErr)
			}
		}

		zap.S().Info("✅ Server stopped gracefully")

	}
}
