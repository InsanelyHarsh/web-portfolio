package blog

import (
	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/blog/models"
	"github.com/insanelyharsh/web-portfolio/internal/parser"
)

func toBlogDTO(blog *models.BlogContent) *dtos.Blog {
	return &dtos.Blog{
		Id:        blog.Id,
		Slug:      blog.Slug,
		Title:     blog.Title,
		Content:   string(parser.MarkdownToHTML([]byte(blog.Content))),
		CreatedAt: blog.CreatedAt,
		UpdatedAt: blog.UpdatedAt,
	}
}
