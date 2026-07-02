package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/packllama/internal/api"
	"github.com/opd-ai/packllama/internal/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// run is the real entry point, returning an error so main stays simple.
func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	logger := newLogger(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return api.NewServer(api.Config{
		Host:            cfg.Host,
		Port:            cfg.Port,
		AllowedOrigins:  cfg.AllowedOrigins,
		ShutdownTimeout: cfg.ShutdownTimeout,
		Logger:          logger,
	}, nil).Start(ctx)
}

// loadConfig builds a Config by merging defaults, optional file, env vars, and flags.
func loadConfig() (config.Config, error) {
	cfg := config.Default()

	configFile := flag.String("config", "", "path to JSON configuration file")
	flag.StringVar(&cfg.Host, "host", cfg.Host, "bind address")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "listen port")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log level: debug, info, warn, error")
	flag.StringVar(&cfg.LogFormat, "log-format", cfg.LogFormat, "log format: text, json")
	flag.StringVar(&cfg.ModelsDir, "models-dir", cfg.ModelsDir, "directory containing .gguf model files")
	flag.StringVar(&cfg.DefaultModel, "default-model", cfg.DefaultModel, "default model to load on startup")
	flag.BoolVar(&cfg.DisableUI, "no-ui", cfg.DisableUI, "run in API-only mode (no desktop UI)")
	flag.DurationVar(&cfg.ShutdownTimeout, "shutdown-timeout", cfg.ShutdownTimeout, "graceful shutdown timeout")
	flag.Parse()

	if *configFile != "" {
		if err := cfg.LoadFile(*configFile); err != nil {
			return config.Config{}, err
		}
	}
	cfg.ApplyEnv()
	return cfg, cfg.Validate()
}

func newLogger(cfg config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}
	_ = level // used below
	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
