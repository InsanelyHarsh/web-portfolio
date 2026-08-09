package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
		writeBlogResult(w, list, err)
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
		writeBlogResult(w, content, err)
	}
}

func getBlogBySlugHandler(manager *blog.BlogManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		content, err := manager.GetBlogContentBySlug(r.Context(), types.BlogSlug(slug))
		writeBlogResult(w, content, err)
	}
}

func writeBlogResult(w http.ResponseWriter, content any, err error) {
	if err != nil {
		if errors.Is(err, blog.ErrNotFound) {
			http.Error(w, "blog not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(content)
}
