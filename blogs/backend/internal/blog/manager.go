package blog

import (
	"context"
	"errors"

	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/blog/repository"
	"github.com/insanelyharsh/web-portfolio/internal/types"
)

// ErrNotFound is returned when a requested blog doesn't exist, so callers
// (e.g. the HTTP layer) can distinguish "not found" from other failures via
// errors.Is instead of matching on an error string.
var ErrNotFound = errors.New("not found")

type BlogManager struct {
	repo repository.BlogRepository
}

func NewBlogManager(repo repository.BlogRepository) *BlogManager {
	return &BlogManager{
		repo: repo,
	}
}

func (m *BlogManager) GetBlogContentById(ctx context.Context, id types.BlogId) (*dtos.Blog, error) {
	blog, err := m.repo.GetBlogContentById(ctx, id)
	if err != nil {
		return nil, err
	}

	if blog == nil {
		return nil, ErrNotFound
	}

	return toBlogDTO(blog), nil
}

func (m *BlogManager) GetBlogContentBySlug(ctx context.Context, slug types.BlogSlug) (*dtos.Blog, error) {
	blog, err := m.repo.GetBlogContentBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if blog == nil {
		return nil, ErrNotFound
	}

	return toBlogDTO(blog), nil
}
