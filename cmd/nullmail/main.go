package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"nullmail/internal/smtp"
)

func main() {
	initLogger()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startHealthServer()
	}()

	port := ":2525"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	server := smtp.NewSMTPServer(port)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		slog.Info("Shutting down servers...")
		os.Exit(0)
	}()

	slog.Info("Starting SMTP server", "port", port)
	if err := server.Start(port); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func startHealthServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"nullmail-smtp"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"nullmail-smtp","endpoints":["/health"]}`))
	})

	healthPort := os.Getenv("PORT")
	if healthPort == "" {
		healthPort = "8080"
	}

	server := &http.Server{
		Addr:         ":" + healthPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	slog.Info("Starting health check server", "port", healthPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Health server error", "error", err)
	}
}

func initLogger() {
	var handler slog.Handler
	var level slog.Level

	if os.Getenv("DEBUG") == "true" {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	if os.Getenv("ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	slog.SetDefault(slog.New(handler))
}
