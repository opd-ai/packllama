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
