package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/packllama/internal/api"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := api.NewServer(api.Config{
		Host:   "127.0.0.1",
		Port:   8080,
		Logger: logger,
	})
	if err := srv.Start(ctx); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
