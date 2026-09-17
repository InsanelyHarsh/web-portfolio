package repository

import (
	"context"
	"errors"

	"github.com/insanelyharsh/web-portfolio/internal/comment/models"
	"github.com/insanelyharsh/web-portfolio/internal/config"
	"github.com/insanelyharsh/web-portfolio/internal/types"
	"github.com/jackc/pgx/v5"
)

type CommentRepository interface {
	GetBlogComments(ctx context.Context, blogId types.BlogId) ([]*models.Comment, error)
	GetCommentById(ctx context.Context, id types.CommentId) (*models.Comment, error)
	CreateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error)
}

type CommentRepositoryImpl struct {
	db config.PgxIface
}

func NewCommentRepository(db config.PgxIface) CommentRepository {
	return &CommentRepositoryImpl{
		db: db,
	}
}

func (c *CommentRepositoryImpl) GetBlogComments(ctx context.Context, blogId types.BlogId) ([]*models.Comment, error) {
	query := `
	SELECT id, blog_id, parent_id, is_anonymous, author_name, author_email, author_ip, content, created_at, updated_at
	FROM comments WHERE blog_id = $1 ORDER BY created_at ASC`

	rows, err := c.db.Query(ctx, query, blogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*models.Comment{} // non-nil so an empty result marshals to [] not null
	for rows.Next() {
		var item models.Comment
		if err := rows.Scan(&item.Id, &item.BlogId, &item.ParentId, &item.IsAnonymous,
			&item.AuthorName, &item.AuthorEmail, &item.AuthorIP, &item.Content,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *CommentRepositoryImpl) GetCommentById(ctx context.Context, id types.CommentId) (*models.Comment, error) {
	query := `
	SELECT id, blog_id, parent_id, is_anonymous, author_name, author_email, author_ip, content, created_at, updated_at
	FROM comments WHERE id = $1`

	row := c.db.QueryRow(ctx, query, id)
	return scanComment(row)
}

func (c *CommentRepositoryImpl) CreateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	query := `
	INSERT INTO comments (blog_id, parent_id, is_anonymous, author_name, author_email, author_ip, content)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, blog_id, parent_id, is_anonymous, author_name, author_email, author_ip, content, created_at, updated_at`

	row := c.db.QueryRow(ctx, query,
		comment.BlogId, comment.ParentId, comment.IsAnonymous,
		comment.AuthorName, comment.AuthorEmail, comment.AuthorIP, comment.Content)

	created, err := scanComment(row)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func scanComment(row pgx.Row) (*models.Comment, error) {
	var comment models.Comment
	err := row.Scan(&comment.Id, &comment.BlogId, &comment.ParentId, &comment.IsAnonymous,
		&comment.AuthorName, &comment.AuthorEmail, &comment.AuthorIP, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return &comment, nil
}
