package httputils

import (
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
