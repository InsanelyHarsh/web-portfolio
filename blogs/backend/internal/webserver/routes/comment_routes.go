package routes

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"unicode/utf8"

	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/comment"
	"github.com/insanelyharsh/web-portfolio/internal/constants"
	"github.com/insanelyharsh/web-portfolio/internal/types"
)

// Comment routes live under /comments/{slug} rather than nested as
// /blogs/{slug}/comments: Go's net/http ServeMux rejects that nested form
// as ambiguous against the existing GET /blogs/id/{id} pattern (a request
// like "/blogs/id/comments" would match both, and neither pattern is more
// specific than the other), so a distinct top-level prefix is used instead.
func RegisterCommentRoutes(mux *http.ServeMux, manager *comment.CommentManager) {
	mux.HandleFunc("GET /comments/{slug}", getCommentsHandler(manager))
	mux.HandleFunc("POST /comments/{slug}", createCommentHandler(manager))
}

func getCommentsHandler(manager *comment.CommentManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		list, err := manager.GetCommentsBySlug(r.Context(), types.BlogSlug(slug))
		if err != nil {
			writeError(r.Context(), w, err, "blog not found")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

func createCommentHandler(manager *comment.CommentManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		// Read the raw body and check it's valid UTF-8 before decoding:
		// encoding/json silently substitutes U+FFFD for invalid UTF-8 byte
		// sequences while decoding into a string, so checking a decoded
		// dtos.CreateCommentRequest field would never catch malformed input.
		body, err := io.ReadAll(io.LimitReader(r.Body, int64(constants.CommentMaxLength)*4+1024))
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if !utf8.Valid(body) {
			http.Error(w, "request body must be valid UTF-8", http.StatusBadRequest)
			return
		}

		var req dtos.CreateCommentRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		created, err := manager.CreateComment(r.Context(), types.BlogSlug(slug), req, clientIP(r))
		if err != nil {
			writeError(r.Context(), w, err, "blog not found")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	}
}

// clientIP returns the requester's address without the port, for storage
// against a comment (abuse tracking only - never returned by the API).
// Falls back to the raw RemoteAddr if it isn't a host:port pair.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
