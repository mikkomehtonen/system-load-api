package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sysload/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/v1/stats", handlers.Stats)
	mux.HandleFunc("/api/v1/cpu", handlers.CPU)
	mux.HandleFunc("/api/v1/memory", handlers.Memory)
	mux.HandleFunc("/api/v1/disk", handlers.Disk)
	mux.HandleFunc("/api/v1/gpu", handlers.GPU)
	mux.HandleFunc("/api/v1/network", handlers.Network)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handlers.TimeoutMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      12 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// Shutdown handling.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("System Load API listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped")
}
