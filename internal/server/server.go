package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	srv    *http.Server
	logger *slog.Logger
}

func New(logger *slog.Logger, addr string, opts ...Option) *Server {
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	for _, opt := range opts {
		opt(&Server{srv: srv})
	}

	return &Server{
		srv:    srv,
		logger: logger,
	}
}

type Option func(*Server)

func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.srv.WriteTimeout = timeout
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.srv.ReadTimeout = timeout
	}
}

func WithHandler(handler http.Handler) Option {
	return func(s *Server) {
		s.srv.Handler = handler
	}
}

func (s *Server) StartAndWait() {
	s.Start()
	s.GracefulShutdown()
}

func (s *Server) Start() {
	go func() {
		s.logger.InfoContext(context.Background(), "starting server", "addr", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(context.Background(), "failed to start server", "error", err)
		}
	}()
}

func (s *Server) GracefulShutdown() {
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = s.srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	s.logger.InfoContext(context.Background(), "shutting down")
}
