package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"saythis-backend/internal/config"
	"saythis-backend/internal/database"
	"saythis-backend/internal/server"
	"syscall"
	"time"
)

func main() {

	// *******************
	// Env intilization
	// *******************

	cfg, err := config.LoadConfig()
	if err != nil {

		slog.Error(" ❌  Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// *******************
	// Logger intilization
	// *******************

	_, err = config.InitLogger(config.LoggerConfig{
		Env:     cfg.AppEnv,
		Service: "saythis",
		Level:   slog.LevelDebug,
		Output:  os.Stdout,
	})
	if err != nil {
		slog.Error("❌ Failed to initialise logger", "error", err)
		os.Exit(1)
	}

	// *******************
	// Database intilization
	// *******************

	pool, err := database.ConnectWithRetry(cfg.DatabaseURL)
	if err != nil {
		slog.Error("❌ Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("🔌 Closing database connection pool...")
		pool.Close()
	}()

	// *******************
	// Router intilization
	// *******************

	router := server.NewRouter(pool, cfg)

	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		slog.Info("🚀 Server listening", "addr", "http://localhost"+cfg.Port)
		serverError <- srv.ListenAndServe()
	}()

	// *******************
	// Graceful Shutdown
	// *******************

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverError:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("❌ Server error", "error", err)
			os.Exit(1)
		}

	case sig := <-shutDown:
		slog.Info("🛑 Shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("⚠️  Graceful shutdown failed, forcing close", "error", err)
			if closeErr := srv.Close(); closeErr != nil {
				slog.Error("❌ Forced close failed", "error", closeErr)
			}
		}

		slog.Info("✅ Server stopped gracefully")
	}
}
