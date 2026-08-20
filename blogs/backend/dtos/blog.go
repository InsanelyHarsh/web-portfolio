package dtos

import (
	"time"

	"github.com/insanelyharsh/web-portfolio/internal/types"
)

type Blog struct {
	Id        types.BlogId   `json:"id"`
	Slug      types.BlogSlug `json:"slug"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type BlogListItem struct {
	Id      types.BlogId   `json:"id"`
	Slug    types.BlogSlug `json:"slug"`
	Title   string         `json:"title"`
	Excerpt string         `json:"excerpt"`
}
