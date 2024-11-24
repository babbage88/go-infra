package swaggerui

import (
	"embed"
	"io/fs"
	"net/http"
)

// Embed the static Swagger UI assets into the binary.
//
//go:embed embed
var swaggerFS embed.FS

// ServeSwaggerUI creates a new HTTP handler for the Swagger UI.
func ServeSwaggerUI(spec []byte) http.Handler {
	// Create a sub-filesystem for the embedded assets
	staticFiles, err := fs.Sub(swaggerFS, "embed")
	if err != nil {
		panic("Failed to create sub-filesystem: " + err.Error())
	}

	// Serve the spec dynamically at /swagger_spec
	specHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(spec)
	}

	// Setup a multiplexer to serve both the static files and the spec
	mux := http.NewServeMux()
	mux.Handle("/swagger_spec", http.HandlerFunc(specHandler))
	mux.Handle("/", http.FileServer(http.FS(staticFiles)))

	return mux
}
