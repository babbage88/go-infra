package api_server

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// requestLoggingMiddleware logs request paths and warns on 404s
func requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)

		if recorder.status == http.StatusNotFound {
			slog.Warn("no route matched",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.Int("status", recorder.status),
				slog.Any("duration", duration),
			)
		} else {
			slog.Info("handled request",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.Int("status", recorder.status),
				slog.Any("duration", duration),
			)
		}
	})
}
