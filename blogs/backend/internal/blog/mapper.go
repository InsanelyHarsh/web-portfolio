package blog

import (
	"strings"

	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/blog/models"
	"github.com/insanelyharsh/web-portfolio/internal/constants"
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

func toBlogListItemDTO(item *models.BlogListItem) *dtos.BlogListItem {
	return &dtos.BlogListItem{
		Id:      item.Id,
		Slug:    item.Slug,
		Title:   item.Title,
		Excerpt: buildExcerpt(item.Content),
	}
}

// buildExcerpt reduces markdown content to plain text and takes the first
// couple of non-empty lines, capped at a max character length so one very
// long line can't blow up the response.
func buildExcerpt(rawMarkdown string) string {
	plain := string(parser.MarkdownToPlainText([]byte(rawMarkdown)))

	var lines []string
	for line := range strings.SplitSeq(plain, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) == constants.ExcerptMaxLines {
			break
		}
	}

	excerpt := strings.Join(lines, " ")
	if len(excerpt) > constants.ExcerptMaxChars {
		excerpt = strings.TrimSpace(excerpt[:constants.ExcerptMaxChars]) + "…"
	}
	return excerpt
}
