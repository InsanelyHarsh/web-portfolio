package comment

import (
	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/comment/models"
	"github.com/insanelyharsh/web-portfolio/internal/constants"
	"github.com/insanelyharsh/web-portfolio/internal/types"
)

func toCommentDTO(c *models.Comment) *dtos.Comment {
	return &dtos.Comment{
		Id:          c.Id,
		ParentId:    c.ParentId,
		AuthorName:  displayName(c),
		IsAnonymous: c.IsAnonymous,
		Content:     c.Content,
		CreatedAt:   c.CreatedAt,
		Replies:     []*dtos.Comment{}, // non-nil so it marshals to [] not null
	}
}

func toCommentTree(flat []*models.Comment) []*dtos.Comment {
	topLevel := []*dtos.Comment{}
	byId := make(map[types.CommentId]*dtos.Comment, len(flat))

	for _, c := range flat {
		dto := toCommentDTO(c)
		byId[c.Id] = dto
		if c.ParentId == nil {
			topLevel = append(topLevel, dto)
		}
	}

	for _, c := range flat {
		if c.ParentId == nil {
			continue
		}
		if parent, ok := byId[*c.ParentId]; ok {
			parent.Replies = append(parent.Replies, byId[c.Id])
		}
	}

	return topLevel
}

func displayName(c *models.Comment) string {
	if c.IsAnonymous || c.AuthorName == nil || *c.AuthorName == "" {
		return constants.AnonymousDisplayName
	}
	return *c.AuthorName
}
