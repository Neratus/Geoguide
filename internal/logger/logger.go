package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	config "github.com/Neratus/geoguide/internal/repository/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

func Init(cfg config.LoggingConfig) (*lumberjack.Logger, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
		return nil, err
	}

	logFile := &lumberjack.Logger{
		Filename:   cfg.File,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   true,
	}

	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	fileHandler := slog.NewJSONHandler(logFile, opts)
	consoleHandler := slog.NewTextHandler(os.Stdout, opts)

	multi := &MultiHandler{handlers: []slog.Handler{fileHandler, consoleHandler}}
	logger := slog.New(multi)
	slog.SetDefault(logger)

	return logFile, nil
}
