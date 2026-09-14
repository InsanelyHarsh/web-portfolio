package comment

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/apperrors"
	blogrepository "github.com/insanelyharsh/web-portfolio/internal/blog/repository"
	"github.com/insanelyharsh/web-portfolio/internal/comment/models"
	"github.com/insanelyharsh/web-portfolio/internal/comment/repository"
	"github.com/insanelyharsh/web-portfolio/internal/constants"
	"github.com/insanelyharsh/web-portfolio/internal/types"
)

// emailPattern is compiled once from constants.EmailPattern rather than on
// every request.
var emailPattern = regexp.MustCompile(constants.EmailPattern)

type CommentManager struct {
	repo     repository.CommentRepository
	blogRepo blogrepository.BlogRepository
}

func NewCommentManager(repo repository.CommentRepository, blogRepo blogrepository.BlogRepository) *CommentManager {
	return &CommentManager{
		repo:     repo,
		blogRepo: blogRepo,
	}
}

func (m *CommentManager) GetCommentsBySlug(ctx context.Context, slug types.BlogSlug) ([]*dtos.Comment, error) {
	blog, err := m.blogRepo.GetBlogContentBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if blog == nil {
		return nil, apperrors.ErrNotFound
	}

	flat, err := m.repo.GetBlogComments(ctx, blog.Id)
	if err != nil {
		return nil, err
	}

	return toCommentTree(flat), nil
}

func (m *CommentManager) CreateComment(ctx context.Context, slug types.BlogSlug, req dtos.CreateCommentRequest, clientIP string) (*dtos.Comment, error) {
	blog, err := m.blogRepo.GetBlogContentBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if blog == nil {
		return nil, apperrors.ErrNotFound
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", apperrors.ErrValidation)
	}
	if !utf8.ValidString(content) {
		return nil, fmt.Errorf("%w: content must be valid UTF-8", apperrors.ErrValidation)
	}
	if len(content) > constants.CommentMaxLength {
		return nil, fmt.Errorf("%w: content exceeds %d characters", apperrors.ErrValidation, constants.CommentMaxLength)
	}

	// Flat nesting: a reply's parent must exist, belong to this blog, and
	// itself be top-level - replying to a reply isn't allowed.
	var parentId *types.CommentId
	if req.ParentId != nil {
		parent, err := m.repo.GetCommentById(ctx, *req.ParentId)
		if err != nil {
			return nil, err
		}
		if parent == nil || parent.BlogId != blog.Id {
			return nil, fmt.Errorf("%w: parent comment not found", apperrors.ErrValidation)
		}
		if parent.ParentId != nil {
			return nil, fmt.Errorf("%w: cannot reply to a reply", apperrors.ErrValidation)
		}
		parentId = req.ParentId
	}

	comment := &models.Comment{
		BlogId:      blog.Id,
		ParentId:    parentId,
		IsAnonymous: req.IsAnonymous,
		Content:     content,
	}

	if clientIP != "" {
		ip := clientIP
		comment.AuthorIP = &ip
	}

	if !req.IsAnonymous {
		authorName := strings.TrimSpace(req.AuthorName)
		if authorName == "" {
			return nil, fmt.Errorf("%w: author name is required unless posting anonymously", apperrors.ErrValidation)
		}
		if !utf8.ValidString(authorName) {
			return nil, fmt.Errorf("%w: author name must be valid UTF-8", apperrors.ErrValidation)
		}
		if len(authorName) > constants.AuthorNameMaxLength {
			return nil, fmt.Errorf("%w: author name exceeds %d characters", apperrors.ErrValidation, constants.AuthorNameMaxLength)
		}
		comment.AuthorName = &authorName

		if email := strings.TrimSpace(req.AuthorEmail); email != "" {
			if !utf8.ValidString(email) {
				return nil, fmt.Errorf("%w: author email must be valid UTF-8", apperrors.ErrValidation)
			}
			if !emailPattern.MatchString(email) {
				return nil, fmt.Errorf("%w: author email is not a valid email address", apperrors.ErrValidation)
			}
			comment.AuthorEmail = &email
		}
	}

	created, err := m.repo.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	return toCommentDTO(created), nil
}
