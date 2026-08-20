package models

import (
	"time"

	"github.com/insanelyharsh/web-portfolio/internal/types"
)

type BlogContent struct {
	Id        types.BlogId
	Slug      types.BlogSlug
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BlogListItem is the lightweight shape used for blog listings: Content
// here holds only a server-side-truncated slice of the raw markdown (see
// BlogRepository.GetBlogList), which the mapper reduces further into a
// plain-text excerpt.
type BlogListItem struct {
	Id      types.BlogId
	Slug    types.BlogSlug
	Title   string
	Content string
}
