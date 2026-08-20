package repository

import (
	"context"
	"errors"

	"github.com/insanelyharsh/web-portfolio/internal/blog/models"
	"github.com/insanelyharsh/web-portfolio/internal/types"
	"github.com/jackc/pgx/v5"
)

type BlogRepository interface {
	GetBlogContentById(ctx context.Context, id types.BlogId) (*models.BlogContent, error)
	GetBlogContentBySlug(ctx context.Context, slug types.BlogSlug) (*models.BlogContent, error)
	GetBlogList(ctx context.Context) ([]*models.BlogListItem, error)
}

type BlogRepositoryImpl struct {
	db pgx.Conn
}

func NewBlogRepository(db pgx.Conn) BlogRepository {
	return &BlogRepositoryImpl{
		db: db,
	}
}

func (b *BlogRepositoryImpl) GetBlogContentById(ctx context.Context, id types.BlogId) (*models.BlogContent, error) {
	query := `
	SELECT id, slug, title, content, created_at, updated_at
	FROM blogs WHERE id = $1`

	row := b.db.QueryRow(ctx, query, id)

	return scanBlogContent(row)
}

func (b *BlogRepositoryImpl) GetBlogContentBySlug(ctx context.Context, slug types.BlogSlug) (*models.BlogContent, error) {
	query := `
	SELECT id, slug, title, content, created_at, updated_at
	FROM blogs WHERE slug = $1`

	row := b.db.QueryRow(ctx, query, slug)
	return scanBlogContent(row)
}

func (b *BlogRepositoryImpl) GetBlogList(ctx context.Context) ([]*models.BlogListItem, error) {
	query := `
	SELECT id, slug, title, LEFT(content, 500) AS content
	FROM blogs ORDER BY created_at DESC`

	rows, err := b.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*models.BlogListItem{} // non-nil so an empty result marshals to [] not null
	for rows.Next() {
		var item models.BlogListItem
		if err := rows.Scan(&item.Id, &item.Slug, &item.Title, &item.Content); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanBlogContent(row pgx.Row) (*models.BlogContent, error) {
	var blog models.BlogContent
	err := row.Scan(&blog.Id, &blog.Slug, &blog.Title, &blog.Content, &blog.CreatedAt, &blog.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &blog, nil
}
