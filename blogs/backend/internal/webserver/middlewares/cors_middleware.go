package middlewares

import (
	"net/http"
	"strings"
)

// CORSMiddleware returns a middleware allowing cross-origin requests from
// any of the given allowed origins (e.g. a local frontend dev server plus a
// deployed frontend URL), permitting the given request headers. Since a
// response can only ever name one origin, a matching request's exact Origin
// is reflected back rather than using a wildcard, and Vary: Origin is set so
// shared caches don't serve one origin's CORS headers to a different one.
//
// Preflight OPTIONS requests are short-circuited with a 204, since those
// never carry a method the underlying mux would otherwise match (an OPTIONS
// request left to fall through would hit a 404/405 with no CORS headers,
// failing the preflight).
func CORSMiddleware(allowedOrigins []string, allowedHeaders []string) func(http.Handler) http.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}

	allowedHeadersValue := strings.Join(allowedHeaders, ", ")

	return func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", allowedHeadersValue)
			w.Header().Set("Access-Control-Max-Age", "600")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			h.ServeHTTP(w, r)
		}
	}
}
