package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
)

type Server struct {
	cfg        Config
	httpServer *http.Server
	logger     *slog.Logger

	mu       sync.RWMutex
	listener net.Listener
}

func NewServer(cfg Config) *Server {
	cfg = cfg.withDefaults()
	return &Server{
		cfg:    cfg,
		logger: cfg.Logger,
		httpServer: &http.Server{
			Addr:    cfg.addr(),
			Handler: NewHandler(cfg.Logger, cfg.AllowedOrigins),
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	listener, err := net.Listen("tcp", s.cfg.addr())
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.setListener(listener)

	go s.shutdownOnCancel(ctx)
	s.logger.Info("server listening", "addr", listener.Addr().String())

	err = s.httpServer.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) ListenAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener == nil {
		return s.cfg.addr()
	}
	return s.listener.Addr().String()
}

func (s *Server) setListener(listener net.Listener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listener = listener
}

func (s *Server) shutdownOnCancel(ctx context.Context) {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("server shutdown failed", "error", err)
	}
}
