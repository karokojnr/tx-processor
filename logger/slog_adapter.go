package logger

import "log/slog"

// SlogAdapter is an adapter that implements the Logger interface using slog
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new SlogAdapter
func NewSlogAdapter(logger *slog.Logger) Logger {
	return &SlogAdapter{logger: logger}
}

// Debug logs a debug message
func (s *SlogAdapter) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

// Info logs an info message
func (s *SlogAdapter) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

// Warn logs a warning message
func (s *SlogAdapter) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

// Error logs an error message
func (s *SlogAdapter) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

// With returns a new logger with the given key-value pairs
func (s *SlogAdapter) With(args ...any) Logger {
	return &SlogAdapter{logger: s.logger.With(args...)}
}
