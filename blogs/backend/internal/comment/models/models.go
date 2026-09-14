package models

import (
	"time"

	"github.com/insanelyharsh/web-portfolio/internal/types"
)

type Comment struct {
	Id          types.CommentId
	BlogId      types.BlogId
	ParentId    *types.CommentId
	IsAnonymous bool
	AuthorName  *string
	AuthorEmail *string
	AuthorIP    *string
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
