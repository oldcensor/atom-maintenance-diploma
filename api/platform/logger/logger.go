package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"atom-maintenance/internal/config"
	"atom-maintenance/pkg"

	"github.com/lmittmann/tint"
)

func New(cfg config.LoggerConfig) *slog.Logger {
	lvl := parseLevel(cfg.Level)
	addSource := lvl == slog.LevelDebug

	var h slog.Handler

	if cfg.Env == "local" || cfg.Env == "dev" {
		h = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      lvl,
			AddSource:  addSource,
			TimeFormat: time.TimeOnly,
		})
	} else {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     lvl,
			AddSource: addSource,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					a.Value = slog.StringValue(time.Now().UTC().Format(time.RFC3339Nano))
				}
				return a
			},
		})
	}

	return slog.New(h).With(
		"service", cfg.Service,
		"env", cfg.Env,
		"version", cfg.Version,
	)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func WithReqID(ctx context.Context, base *slog.Logger) *slog.Logger {
	if rid := pkg.RequestIDFrom(ctx); rid != "" {
		return base.With("req_id", rid)
	}
	return base
}
