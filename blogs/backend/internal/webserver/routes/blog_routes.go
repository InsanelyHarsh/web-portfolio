package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/blog"
	"github.com/insanelyharsh/web-portfolio/internal/types"
)

func RegisterBlogRoutes(mux *http.ServeMux, manager *blog.BlogManager) {
	mux.HandleFunc("GET /blogs", getBlogListHandler(manager))
	mux.HandleFunc("GET /blogs/id/{id}", getBlogByIdHandler(manager))
	mux.HandleFunc("GET /blogs/{slug}", getBlogBySlugHandler(manager))
}

func getBlogListHandler(manager *blog.BlogManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := manager.GetBlogList(r.Context())
		if err != nil {
			writeError(r.Context(), w, err, "blog not found")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

func getBlogByIdHandler(manager *blog.BlogManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawId := r.PathValue("id")
		id, err := strconv.Atoi(rawId)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		content, err := manager.GetBlogContentById(r.Context(), types.BlogId(id))
		writeBlogHTMLResult(w, r, content, err)
	}
}

func getBlogBySlugHandler(manager *blog.BlogManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		content, err := manager.GetBlogContentBySlug(r.Context(), types.BlogSlug(slug))
		writeBlogHTMLResult(w, r, content, err)
	}
}

func writeBlogHTMLResult(w http.ResponseWriter, r *http.Request, content *dtos.Blog, err error) {
	if err != nil {
		writeError(r.Context(), w, err, "blog not found")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(content.Content))
}
