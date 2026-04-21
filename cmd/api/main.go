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
	// Logger intilization
	// *******************

	_, err := config.InitLogger(config.LoggerConfig{
		Env:     "development",
		Service: "Saythis",
		Level:   slog.LevelDebug,
		Output:  os.Stdout,
	})

	if err != nil {
		panic(err)
	}

	// *******************
	// Env intilization
	// *******************

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("❌ Failed to load configuration: ", "error", err)
	}

	// *******************
	// Database intilization
	// *******************

	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("❌ Failed to Connect to database: ", "error", err)
	}
	defer func() {
		slog.Error("🔌 Closing database connection...")
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
		slog.Info("🚀 server is started at http://localhost", "PORT", cfg.Port)
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
			slog.Error("❌ Server failed: ", "Error", err)
		}
	case sig := <-shutDown:
		slog.Info("🛑 Shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("⚠️  Graceful shutdown failed, forcing close: ", "Error", err)
			if closeErr := srv.Close(); closeErr != nil {
				slog.Error("❌ Server close failed: ", "error", closeErr)
			}

		}

		slog.Info("✅ Server stopped gracefully")
	}
}
