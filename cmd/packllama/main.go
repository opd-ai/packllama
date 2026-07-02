package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/opd-ai/packllama/internal/api"
	"github.com/opd-ai/packllama/internal/config"
	"github.com/opd-ai/packllama/internal/modelstore"
	"github.com/opd-ai/packllama/internal/service"
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

	svc := buildService(cfg, logger)

	return api.NewServer(api.Config{
		Host:            cfg.Host,
		Port:            cfg.Port,
		AllowedOrigins:  cfg.AllowedOrigins,
		ShutdownTimeout: cfg.ShutdownTimeout,
		Logger:          logger,
	}, svc).Start(ctx)
}

// buildService constructs an InferenceService from cfg. When ModelsDir is set
// the service discovers local GGUF files so the /v1/models endpoints work
// without a full inference backend.
func buildService(cfg config.Config, logger *slog.Logger) service.InferenceService {
	if cfg.ModelsDir == "" {
		return nil
	}
	registry := modelstore.New()
	if err := registry.Scan(cfg.ModelsDir, false); err != nil {
		logger.Warn("model discovery failed", "dir", cfg.ModelsDir, "error", err)
	}
	if cfg.DefaultModel != "" {
		_ = registry.AddAlias("default", cfg.DefaultModel)
	}
	return service.NewRegistryService(registry)
}

// loadConfig builds a Config by merging defaults, optional file, env vars, and flags.
// Precedence order (highest wins): flags > env > file > defaults.
func loadConfig() (config.Config, error) {
	cfg := config.Default()

	// Declare flag destinations separately so flag.Visit can detect which flags
	// were explicitly provided and apply them last, after file and env.
	var (
		configFile      string
		host            string
		port            int
		logLevel        string
		logFormat       string
		modelsDir       string
		defaultModel    string
		disableUI       bool
		shutdownTimeout time.Duration
	)

	flag.StringVar(&configFile, "config", "", "path to JSON configuration file")
	flag.StringVar(&host, "host", cfg.Host, "bind address")
	flag.IntVar(&port, "port", cfg.Port, "listen port")
	flag.StringVar(&logLevel, "log-level", cfg.LogLevel, "log level: debug, info, warn, error")
	flag.StringVar(&logFormat, "log-format", cfg.LogFormat, "log format: text, json")
	flag.StringVar(&modelsDir, "models-dir", cfg.ModelsDir, "directory containing .gguf model files")
	flag.StringVar(&defaultModel, "default-model", cfg.DefaultModel, "default model to load on startup")
	flag.BoolVar(&disableUI, "no-ui", cfg.DisableUI, "run in API-only mode (no desktop UI)")
	flag.DurationVar(&shutdownTimeout, "shutdown-timeout", cfg.ShutdownTimeout, "graceful shutdown timeout")
	flag.Parse()

	// Apply lower-precedence sources first: file, then env.
	if configFile != "" {
		if err := cfg.LoadFile(configFile); err != nil {
			return config.Config{}, err
		}
	}
	cfg.ApplyEnv()

	// Apply only flags that were explicitly provided so they win over file and env.
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "host":
			cfg.Host = host
		case "port":
			cfg.Port = port
		case "log-level":
			cfg.LogLevel = logLevel
		case "log-format":
			cfg.LogFormat = logFormat
		case "models-dir":
			cfg.ModelsDir = modelsDir
		case "default-model":
			cfg.DefaultModel = defaultModel
		case "no-ui":
			cfg.DisableUI = disableUI
		case "shutdown-timeout":
			cfg.ShutdownTimeout = shutdownTimeout
		}
	})

	return cfg, cfg.Validate()
}

func newLogger(cfg config.Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
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
	var handler slog.Handler
	if strings.ToLower(cfg.LogFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
