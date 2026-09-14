package dtos

import (
	"time"

	"github.com/insanelyharsh/web-portfolio/internal/types"
)

// Comment is the public shape of a comment or reply. AuthorEmail and any
// IP address are deliberately absent - they're never returned by the API.
type Comment struct {
	Id          types.CommentId  `json:"id"`
	ParentId    *types.CommentId `json:"parentId,omitempty"`
	AuthorName  string           `json:"authorName"`
	IsAnonymous bool             `json:"isAnonymous"`
	Content     string           `json:"content"`
	CreatedAt   time.Time        `json:"createdAt"`
	Replies     []*Comment       `json:"replies"`
}

// CreateCommentRequest is the expected request body for posting a comment
// or reply. AuthorName/AuthorEmail are ignored when IsAnonymous is true.
type CreateCommentRequest struct {
	ParentId    *types.CommentId `json:"parentId,omitempty"`
	AuthorName  string           `json:"authorName,omitempty"`
	AuthorEmail string           `json:"authorEmail,omitempty"`
	IsAnonymous bool             `json:"isAnonymous"`
	Content     string           `json:"content"`
}
