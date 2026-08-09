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
