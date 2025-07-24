package middleware

import (
	"log/slog"
	"net/http"
)

// RecoverMiddleware is a middleware that recovers from panics in HTTP handlers,
// logs the error, and returns a 500 Internal Server Error response.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered in handler", slog.Any("error", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
