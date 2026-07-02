package api

import (
	"log/slog"
	"net"
	"strconv"
	"time"
)

const defaultShutdownTimeout = 5 * time.Second

type Config struct {
	Host            string
	Port            int
	AllowedOrigins  []string
	ShutdownTimeout time.Duration
	Logger          *slog.Logger

	// LogRequests enables debug-level logging of request bodies.
	// This is expensive and should be used for troubleshooting only.
	LogRequests bool
	// LogResponses enables debug-level logging of response bodies.
	// This is expensive and should be used for troubleshooting only.
	LogResponses bool

	// EnableMetrics exposes a Prometheus /metrics endpoint when true.
	EnableMetrics bool
}

func (c Config) withDefaults() Config {
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = defaultShutdownTimeout
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
	return c
}

func (c Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
