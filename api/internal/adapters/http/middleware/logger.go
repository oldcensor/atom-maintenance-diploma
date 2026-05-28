package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"atom-maintenance/platform/logger"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func Logger(base *slog.Logger) func(http.Handler) http.Handler {
	base = base.With("module", "http")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()

			log := logger.WithReqID(ctx, base)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			log.Info("response",
				"status", wrapped.status,
				"bytes", wrapped.size,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
