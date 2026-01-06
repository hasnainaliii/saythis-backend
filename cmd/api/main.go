package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"saythis-backend/internal/server"
	"time"
)

func main() {

	router := server.NewRouter()

	server := &http.Server{
		Addr:         ":3000",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		log.Println("🚀 Server running on http://localhost:3000")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server Failed: %v", err)
		}
	}()

	<-quit
	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
