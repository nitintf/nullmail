package main

import (
	"log/slog"
	"os"

	"nullmail/internal/smtp"
)

func main() {
	initLogger()

	port := ":2525" // Using non-privileged port for development
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	server := smtp.NewSMTPServer(port)

	slog.Info("Starting SMTP server", "port", port)
	if err := server.Start(port); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func initLogger() {
	// Configure slog globally
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
