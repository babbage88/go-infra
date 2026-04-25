package cors

import (
	"log"
	"log/slog"
	"net/http"
)

func VerifyRequestPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Error("Invalid request method", slog.String("Method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
func EnableCors(w *http.ResponseWriter) {
	applyCORSHeaders(*w, &http.Request{Header: http.Header{}}, "GET, POST, PUT, DELETE, OPTIONS")
}

func HandlerCorsAndOptions(w http.ResponseWriter, r *http.Request) {
	applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")
	if r.Method == "OPTIONS" {
		slog.Info("Received OPTIONS request")
		applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")
	}
}

// CORSMiddleware adds CORS headers and handles OPTIONS requests.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")

		// Handle OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// handleOPTIONS handles CORS preflight OPTIONS requests.
func handleOPTIONS(w http.ResponseWriter, r *http.Request) {
	applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")
	w.WriteHeader(http.StatusOK)
}

// CORSWithPOST is a middleware for handling CORS with POST requests.
func CORSWithPOST(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Printf("Recovered from panic: %v\n", rvr)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		// Handle OPTIONS requests using the helper function
		if r.Method == http.MethodOptions {
			handleOPTIONS(w, r)
			return
		}

		if r.Method != http.MethodPost {
			slog.Error("Invalid request method", slog.String("Method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// CORSWithPOST is a middleware for handling CORS with POST requests.
func CORSWithGET(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Printf("Recovered from panic: %v\n", rvr)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		// Handle OPTIONS requests using the helper function
		if r.Method == http.MethodOptions {
			handleOPTIONS(w, r)
			return
		}

		if r.Method != http.MethodGet {
			slog.Error("Invalid request method", slog.String("Method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// CORSWithPOST is a middleware for handling CORS with POST requests.
func CORSWithDELETE(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS requests using the helper function
		if r.Method == http.MethodOptions {
			handleOPTIONS(w, r)
			return
		}

		if r.Method != http.MethodDelete {
			slog.Error("Invalid request method", slog.String("Method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// CORSWithPOST is a middleware for handling CORS with POST requests.
func CORSWithPUT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS requests using the helper function
		if r.Method == http.MethodOptions {
			handleOPTIONS(w, r)
			return
		}

		if r.Method != http.MethodPut {
			slog.Error("Invalid request method", slog.String("Method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		applyCORSHeaders(w, r, "GET, POST, PUT, DELETE, OPTIONS")

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
