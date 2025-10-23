package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"tx-processor/handlers"
	"tx-processor/logger"
)

type Server struct {
	httpServer *http.Server
	logger     logger.Logger
}

type Config struct {
	Port   string
	Logger logger.Logger
}

func New(cfg Config, handler *handlers.Handler) *Server {
	mux := http.NewServeMux()

	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	return &Server{
		httpServer: srv,
		logger:     cfg.Logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server failed to listen and serve", "error", err)
		}
	}()

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	s.logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server shutdown failed", "error", err)
		return err
	}
	return nil
}
