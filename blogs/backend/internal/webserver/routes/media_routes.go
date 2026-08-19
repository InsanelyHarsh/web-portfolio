package routes

import (
	"errors"
	"io"
	"net/http"

	"github.com/insanelyharsh/web-portfolio/internal/media"
)

func RegisterMediaRoutes(mux *http.ServeMux, manager *media.MediaManager) {
	mux.HandleFunc("GET /media/{key...}", getMediaHandler(manager))
}

func getMediaHandler(manager *media.MediaManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if key == "" {
			http.Error(w, "invalid key", http.StatusBadRequest)
			return
		}

		res, err := manager.GetObject(r.Context(), key)
		if err != nil {
			if errors.Is(err, media.ErrNotFound) {
				http.Error(w, "media not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		// Forward the object's metadata headers as-is so clients get correct
		// MIME sniffing, caching, and conditional-request behavior.
		if ct := res.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		if etag := res.Header.Get("ETag"); etag != "" {
			w.Header().Set("ETag", etag)
		}
		if cl := res.Header.Get("Content-Length"); cl != "" {
			w.Header().Set("Content-Length", cl)
		}

		_, _ = io.Copy(w, res.Body)
	}
}
