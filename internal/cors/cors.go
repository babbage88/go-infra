package cors

import (
	"net/http"
	"os"
	"strings"
)

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedMethods string) {
	origin := resolveAllowedOrigin(r)
	if origin == "" {
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func resolveAllowedOrigin(r *http.Request) string {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return ""
	}

	if isAllowedOrigin(origin) {
		return origin
	}

	return ""
}

func isAllowedOrigin(origin string) bool {
	allowedOrigins := []string{}
	if frontendURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/"); frontendURL != "" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}

	if corsOrigins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); corsOrigins != "" {
		for _, candidate := range strings.Split(corsOrigins, ",") {
			candidate = strings.TrimRight(strings.TrimSpace(candidate), "/")
			if candidate != "" {
				allowedOrigins = append(allowedOrigins, candidate)
			}
		}
	}

	normalizedOrigin := strings.TrimRight(origin, "/")
	for _, allowedOrigin := range allowedOrigins {
		if strings.EqualFold(normalizedOrigin, allowedOrigin) {
			return true
		}
	}

	return strings.HasPrefix(normalizedOrigin, "http://localhost:") ||
		strings.HasPrefix(normalizedOrigin, "https://localhost:") ||
		strings.HasPrefix(normalizedOrigin, "http://127.0.0.1:") ||
		strings.HasPrefix(normalizedOrigin, "https://127.0.0.1:")
}

func HandleCORSPreflightMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass CORS logic for WebSocket upgrade requests
		if strings.ToLower(r.Header.Get("Connection")) == "upgrade" && strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
			next.ServeHTTP(w, r)
			return
		}
		applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CORSWithMethods wraps a handler and sets allowed methods for CORS preflight
func CORSWithMethods(handler http.Handler, methods ...string) http.Handler {
	allowed := append(methods, "OPTIONS")
	allowedMethods := strings.Join(allowed, ", ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applyCORSHeaders(w, r, allowedMethods)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
