package comment

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/apperrors"
	blogmodels "github.com/insanelyharsh/web-portfolio/internal/blog/models"
	"github.com/insanelyharsh/web-portfolio/internal/comment/models"
	"github.com/insanelyharsh/web-portfolio/internal/constants"
	"github.com/insanelyharsh/web-portfolio/internal/types"
)

// errBoom is a generic sentinel used to prove a repo error is propagated
// unwrapped (via errors.Is), rather than swallowed or re-wrapped.
var errBoom = errors.New("boom")

// fakeBlogRepo and fakeCommentRepo are hand-written function-field fakes
// for blogrepository.BlogRepository and repository.CommentRepository. Each
// method delegates to its func field when set, else returns a nil/zero
// value - letting most subtests supply just the one method they need while
// a few capture their call arguments via closures.

type fakeBlogRepo struct {
	getBlogContentBySlugFunc func(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error)
	getBlogContentByIdFunc   func(ctx context.Context, id types.BlogId) (*blogmodels.BlogContent, error)
	getBlogListFunc          func(ctx context.Context) ([]*blogmodels.BlogListItem, error)
}

func (f *fakeBlogRepo) GetBlogContentBySlug(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error) {
	if f.getBlogContentBySlugFunc != nil {
		return f.getBlogContentBySlugFunc(ctx, slug)
	}
	return nil, nil
}

func (f *fakeBlogRepo) GetBlogContentById(ctx context.Context, id types.BlogId) (*blogmodels.BlogContent, error) {
	if f.getBlogContentByIdFunc != nil {
		return f.getBlogContentByIdFunc(ctx, id)
	}
	return nil, nil
}

func (f *fakeBlogRepo) GetBlogList(ctx context.Context) ([]*blogmodels.BlogListItem, error) {
	if f.getBlogListFunc != nil {
		return f.getBlogListFunc(ctx)
	}
	return nil, nil
}

type fakeCommentRepo struct {
	getBlogCommentsFunc func(ctx context.Context, blogId types.BlogId) ([]*models.Comment, error)
	getCommentByIdFunc  func(ctx context.Context, id types.CommentId) (*models.Comment, error)
	createCommentFunc   func(ctx context.Context, comment *models.Comment) (*models.Comment, error)
}

func (f *fakeCommentRepo) GetBlogComments(ctx context.Context, blogId types.BlogId) ([]*models.Comment, error) {
	if f.getBlogCommentsFunc != nil {
		return f.getBlogCommentsFunc(ctx, blogId)
	}
	return nil, nil
}

func (f *fakeCommentRepo) GetCommentById(ctx context.Context, id types.CommentId) (*models.Comment, error) {
	if f.getCommentByIdFunc != nil {
		return f.getCommentByIdFunc(ctx, id)
	}
	return nil, nil
}

func (f *fakeCommentRepo) CreateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	if f.createCommentFunc != nil {
		return f.createCommentFunc(ctx, comment)
	}
	return nil, nil
}

var fixedTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

const (
	testSlug   = types.BlogSlug("test-slug")
	testBlogId = types.BlogId(1)
)

func testBlog() *blogmodels.BlogContent {
	return &blogmodels.BlogContent{
		Id:        testBlogId,
		Slug:      testSlug,
		Title:     "Test Blog",
		Content:   "blog content",
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}
}

// foundBlogRepo returns a fakeBlogRepo resolving testSlug to testBlog() -
// the "blog exists" precondition shared by nearly every test below.
func foundBlogRepo() *fakeBlogRepo {
	return &fakeBlogRepo{
		getBlogContentBySlugFunc: func(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error) {
			return testBlog(), nil
		},
	}
}

// echoCreateCommentRepo builds a fakeCommentRepo whose CreateComment
// captures the *models.Comment it was asked to persist into *capture, and
// echoes it back with Id/CreatedAt/UpdatedAt filled in as a real DB insert
// would.
func echoCreateCommentRepo(capture **models.Comment) *fakeCommentRepo {
	return &fakeCommentRepo{
		createCommentFunc: func(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
			*capture = comment
			created := *comment
			created.Id = types.CommentId(42)
			created.CreatedAt = fixedTime
			created.UpdatedAt = fixedTime
			return &created, nil
		},
	}
}

func wantValidationErr(t *testing.T, err error, wantSubstr string) {
	t.Helper()
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected ErrValidation, got: %v", err)
	}
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("expected error to contain %q, got: %v", wantSubstr, err)
	}
}

func TestCreateComment_ValidationErrors(t *testing.T) {
	otherBlogId := types.BlogId(999)
	someGrandparentId := types.CommentId(50)
	missingParentId := types.CommentId(404)
	otherBlogParentId := types.CommentId(5)
	replyParentId := types.CommentId(6)

	cases := []struct {
		name        string
		req         dtos.CreateCommentRequest
		commentRepo *fakeCommentRepo // defaults to an empty fake if nil
		wantSubstr  string
	}{
		{
			name:       "content empty",
			req:        dtos.CreateCommentRequest{Content: "", IsAnonymous: true},
			wantSubstr: "content is required",
		},
		{
			name:       "content whitespace only",
			req:        dtos.CreateCommentRequest{Content: "   \t\n", IsAnonymous: true},
			wantSubstr: "content is required",
		},
		{
			name:       "content invalid utf8",
			req:        dtos.CreateCommentRequest{Content: "hello \xff\xfe world", IsAnonymous: true},
			wantSubstr: "content must be valid UTF-8",
		},
		{
			name:       "content exceeds max length",
			req:        dtos.CreateCommentRequest{Content: strings.Repeat("a", constants.CommentMaxLength+1), IsAnonymous: true},
			wantSubstr: fmt.Sprintf("content exceeds %d characters", constants.CommentMaxLength),
		},
		{
			// 1251 runes but 5004 bytes: proves the length check is on
			// byte length (len(content)), not rune count.
			name:       "content multi-byte utf8 exceeds byte length",
			req:        dtos.CreateCommentRequest{Content: strings.Repeat("😀", 1251), IsAnonymous: true},
			wantSubstr: fmt.Sprintf("content exceeds %d characters", constants.CommentMaxLength),
		},
		{
			name:       "author name empty when not anonymous",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: ""},
			wantSubstr: "author name is required unless posting anonymously",
		},
		{
			name:       "author name whitespace only",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "   "},
			wantSubstr: "author name is required unless posting anonymously",
		},
		{
			name:       "author name invalid utf8",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Al\xffice"},
			wantSubstr: "author name must be valid UTF-8",
		},
		{
			name:       "author name exceeds max length",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: strings.Repeat("a", constants.AuthorNameMaxLength+1)},
			wantSubstr: fmt.Sprintf("author name exceeds %d characters", constants.AuthorNameMaxLength),
		},
		{
			name:       "author email invalid utf8",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "bob@ex\xffample.com"},
			wantSubstr: "author email must be valid UTF-8",
		},
		{
			name:       "author email invalid format - no at sign",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "not-an-email"},
			wantSubstr: "author email is not a valid email address",
		},
		{
			name:       "author email invalid format - empty local part",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "@example.com"},
			wantSubstr: "author email is not a valid email address",
		},
		{
			name:       "author email invalid format - empty domain",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "bob@"},
			wantSubstr: "author email is not a valid email address",
		},
		{
			name:       "author email invalid format - domain starts with dot",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "bob@.com"},
			wantSubstr: "author email is not a valid email address",
		},
		{
			name:       "author email invalid format - consecutive dots",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "bob@example..com"},
			wantSubstr: "author email is not a valid email address",
		},
		{
			name:       "author email invalid format - space in local part",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "bob smith@example.com"},
			wantSubstr: "author email is not a valid email address",
		},
		{
			name:       "author email invalid format - domain ends with hyphen",
			req:        dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "bob@example.com-"},
			wantSubstr: "author email is not a valid email address",
		},
		{
			name: "parent not found",
			req:  dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: true, ParentId: &missingParentId},
			commentRepo: &fakeCommentRepo{
				getCommentByIdFunc: func(ctx context.Context, id types.CommentId) (*models.Comment, error) {
					return nil, nil
				},
			},
			wantSubstr: "parent comment not found",
		},
		{
			name: "parent belongs to a different blog",
			req:  dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: true, ParentId: &otherBlogParentId},
			commentRepo: &fakeCommentRepo{
				getCommentByIdFunc: func(ctx context.Context, id types.CommentId) (*models.Comment, error) {
					return &models.Comment{Id: otherBlogParentId, BlogId: otherBlogId}, nil
				},
			},
			wantSubstr: "parent comment not found",
		},
		{
			name: "parent is itself a reply",
			req:  dtos.CreateCommentRequest{Content: "valid content", IsAnonymous: true, ParentId: &replyParentId},
			commentRepo: &fakeCommentRepo{
				getCommentByIdFunc: func(ctx context.Context, id types.CommentId) (*models.Comment, error) {
					return &models.Comment{Id: replyParentId, BlogId: testBlogId, ParentId: &someGrandparentId}, nil
				},
			},
			wantSubstr: "cannot reply to a reply",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			commentRepo := tc.commentRepo
			if commentRepo == nil {
				commentRepo = &fakeCommentRepo{}
			}
			mgr := NewCommentManager(commentRepo, foundBlogRepo())

			_, err := mgr.CreateComment(context.Background(), testSlug, tc.req, "")

			wantValidationErr(t, err, tc.wantSubstr)
		})
	}
}

func TestCreateComment_BlogNotFound(t *testing.T) {
	blogRepo := &fakeBlogRepo{
		getBlogContentBySlugFunc: func(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error) {
			return nil, nil
		},
	}
	mgr := NewCommentManager(&fakeCommentRepo{}, blogRepo)

	got, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: true}, "")

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
	if got != nil {
		t.Fatalf("expected a nil result, got: %#v", got)
	}
}

func TestCreateComment_BlogRepoError(t *testing.T) {
	blogRepo := &fakeBlogRepo{
		getBlogContentBySlugFunc: func(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error) {
			return nil, errBoom
		},
	}
	mgr := NewCommentManager(&fakeCommentRepo{}, blogRepo)

	_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: true}, "")

	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom to be propagated, got: %v", err)
	}
	if errors.Is(err, apperrors.ErrValidation) || errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected the raw repo error, not re-wrapped as ErrValidation/ErrNotFound, got: %v", err)
	}
}

func TestCreateComment_ParentLookupRepoError(t *testing.T) {
	parentId := types.CommentId(1)
	commentRepo := &fakeCommentRepo{
		getCommentByIdFunc: func(ctx context.Context, id types.CommentId) (*models.Comment, error) {
			return nil, errBoom
		},
	}
	mgr := NewCommentManager(commentRepo, foundBlogRepo())

	_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: true, ParentId: &parentId}, "")

	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom to be propagated, got: %v", err)
	}
}

func TestCreateComment_RepoCreateCommentError(t *testing.T) {
	commentRepo := &fakeCommentRepo{
		createCommentFunc: func(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
			return nil, errBoom
		},
	}
	mgr := NewCommentManager(commentRepo, foundBlogRepo())

	_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: true}, "")

	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom to be propagated, got: %v", err)
	}
}

func TestCreateComment_ContentAtMaxLength_Succeeds(t *testing.T) {
	var captured *models.Comment
	mgr := NewCommentManager(echoCreateCommentRepo(&captured), foundBlogRepo())

	content := strings.Repeat("a", constants.CommentMaxLength)
	_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: content, IsAnonymous: true}, "")

	if err != nil {
		t.Fatalf("expected no error at exactly the max length, got: %v", err)
	}
	if captured == nil || len(captured.Content) != constants.CommentMaxLength {
		t.Fatalf("expected captured content of length %d, got %#v", constants.CommentMaxLength, captured)
	}
}

func TestCreateComment_AuthorNameAtMaxLength_Succeeds(t *testing.T) {
	var captured *models.Comment
	mgr := NewCommentManager(echoCreateCommentRepo(&captured), foundBlogRepo())

	authorName := strings.Repeat("a", constants.AuthorNameMaxLength)
	_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: false, AuthorName: authorName}, "")

	if err != nil {
		t.Fatalf("expected no error at exactly the max length, got: %v", err)
	}
	if captured == nil || captured.AuthorName == nil || len(*captured.AuthorName) != constants.AuthorNameMaxLength {
		t.Fatalf("expected captured author name of length %d, got %#v", constants.AuthorNameMaxLength, captured)
	}
}

func TestCreateComment_ValidReply_Success(t *testing.T) {
	parentId := types.CommentId(7)
	var captured *models.Comment
	commentRepo := echoCreateCommentRepo(&captured)
	commentRepo.getCommentByIdFunc = func(ctx context.Context, id types.CommentId) (*models.Comment, error) {
		return &models.Comment{Id: parentId, BlogId: testBlogId, ParentId: nil}, nil
	}
	mgr := NewCommentManager(commentRepo, foundBlogRepo())

	dto, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: true, ParentId: &parentId}, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if captured == nil || captured.ParentId == nil || *captured.ParentId != parentId {
		t.Fatalf("expected captured comment's ParentId to be %v, got %#v", parentId, captured)
	}
	if dto.ParentId == nil || *dto.ParentId != parentId {
		t.Fatalf("expected returned dto's ParentId to be %v, got %v", parentId, dto.ParentId)
	}
}

// TestCreateComment_Anonymous_AuthorFieldsIgnoredEvenIfPresent is a
// regression test for the rule that AuthorName/AuthorEmail on the request
// are silently discarded when IsAnonymous is true, even if the caller set
// them.
func TestCreateComment_Anonymous_AuthorFieldsIgnoredEvenIfPresent(t *testing.T) {
	var captured *models.Comment
	mgr := NewCommentManager(echoCreateCommentRepo(&captured), foundBlogRepo())

	req := dtos.CreateCommentRequest{
		Content:     "hi",
		IsAnonymous: true,
		AuthorName:  "Alice",
		AuthorEmail: "alice@example.com",
	}
	dto, err := mgr.CreateComment(context.Background(), testSlug, req, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if captured == nil {
		t.Fatalf("expected repo.CreateComment to be invoked")
	}
	if captured.AuthorName != nil {
		t.Fatalf("expected AuthorName to be discarded for an anonymous comment, got: %v", *captured.AuthorName)
	}
	if captured.AuthorEmail != nil {
		t.Fatalf("expected AuthorEmail to be discarded for an anonymous comment, got: %v", *captured.AuthorEmail)
	}
	if dto.AuthorName != constants.AnonymousDisplayName {
		t.Fatalf("expected returned AuthorName %q, got %q", constants.AnonymousDisplayName, dto.AuthorName)
	}
}

func TestCreateComment_NonAnonymous_EmailOptional_EmptyAllowed(t *testing.T) {
	var captured *models.Comment
	mgr := NewCommentManager(echoCreateCommentRepo(&captured), foundBlogRepo())

	_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: ""}, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if captured == nil {
		t.Fatalf("expected repo.CreateComment to be invoked")
	}
	if captured.AuthorEmail != nil {
		t.Fatalf("expected AuthorEmail to stay nil when not supplied, got: %v", *captured.AuthorEmail)
	}
}

func TestCreateComment_NonAnonymous_EmailWhitespaceTrimmedBeforeValidation(t *testing.T) {
	var captured *models.Comment
	mgr := NewCommentManager(echoCreateCommentRepo(&captured), foundBlogRepo())

	_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: "  bob@example.com  "}, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if captured == nil || captured.AuthorEmail == nil || *captured.AuthorEmail != "bob@example.com" {
		t.Fatalf("expected trimmed email %q, got %#v", "bob@example.com", captured)
	}
}

func TestCreateComment_NonAnonymous_ValidEmailFormats(t *testing.T) {
	emails := []string{
		"bob@example.com",
		"bob.smith+tag@sub.example.co.uk",
		"bob@localhost",
		"u@x",
	}

	for _, email := range emails {
		t.Run(email, func(t *testing.T) {
			var captured *models.Comment
			mgr := NewCommentManager(echoCreateCommentRepo(&captured), foundBlogRepo())

			_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: false, AuthorName: "Bob", AuthorEmail: email}, "")

			if err != nil {
				t.Fatalf("expected %q to be accepted as valid, got: %v", email, err)
			}
			if captured == nil || captured.AuthorEmail == nil || *captured.AuthorEmail != email {
				t.Fatalf("expected captured email %q, got %#v", email, captured)
			}
		})
	}
}

func TestCreateComment_ClientIPPropagation(t *testing.T) {
	cases := []struct {
		name     string
		clientIP string
	}{
		{"non-empty client ip is stored", "203.0.113.5"},
		{"empty client ip stays nil", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var captured *models.Comment
			mgr := NewCommentManager(echoCreateCommentRepo(&captured), foundBlogRepo())

			_, err := mgr.CreateComment(context.Background(), testSlug, dtos.CreateCommentRequest{Content: "hi", IsAnonymous: true}, tc.clientIP)

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if captured == nil {
				t.Fatalf("expected repo.CreateComment to be invoked")
			}
			if tc.clientIP == "" {
				if captured.AuthorIP != nil {
					t.Fatalf("expected AuthorIP to stay nil, got: %v", *captured.AuthorIP)
				}
				return
			}
			if captured.AuthorIP == nil || *captured.AuthorIP != tc.clientIP {
				t.Fatalf("expected AuthorIP %q, got %#v", tc.clientIP, captured.AuthorIP)
			}
		})
	}
}

func TestCreateComment_Success_ReturnsMappedDTO(t *testing.T) {
	commentRepo := &fakeCommentRepo{
		createCommentFunc: func(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
			created := *comment
			created.Id = types.CommentId(42)
			created.CreatedAt = fixedTime
			created.UpdatedAt = fixedTime
			return &created, nil
		},
	}
	mgr := NewCommentManager(commentRepo, foundBlogRepo())

	req := dtos.CreateCommentRequest{Content: "hello", IsAnonymous: false, AuthorName: "Bob"}
	dto, err := mgr.CreateComment(context.Background(), testSlug, req, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if dto.Id != types.CommentId(42) {
		t.Fatalf("expected Id 42, got %v", dto.Id)
	}
	if dto.AuthorName != "Bob" {
		t.Fatalf("expected AuthorName %q, got %q", "Bob", dto.AuthorName)
	}
	if dto.IsAnonymous {
		t.Fatalf("expected IsAnonymous false, got true")
	}
	if dto.Content != "hello" {
		t.Fatalf("expected Content %q, got %q", "hello", dto.Content)
	}
	if !dto.CreatedAt.Equal(fixedTime) {
		t.Fatalf("expected CreatedAt %v, got %v", fixedTime, dto.CreatedAt)
	}
	if dto.ParentId != nil {
		t.Fatalf("expected a nil ParentId, got %v", *dto.ParentId)
	}
	if dto.Replies == nil || len(dto.Replies) != 0 {
		t.Fatalf("expected a non-nil empty Replies, got %#v", dto.Replies)
	}
}

func TestGetCommentsBySlug_BlogNotFound(t *testing.T) {
	blogRepo := &fakeBlogRepo{
		getBlogContentBySlugFunc: func(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error) {
			return nil, nil
		},
	}
	mgr := NewCommentManager(&fakeCommentRepo{}, blogRepo)

	got, err := mgr.GetCommentsBySlug(context.Background(), testSlug)

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
	if got != nil {
		t.Fatalf("expected a nil result, got: %#v", got)
	}
}

func TestGetCommentsBySlug_BlogRepoError(t *testing.T) {
	blogRepo := &fakeBlogRepo{
		getBlogContentBySlugFunc: func(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error) {
			return nil, errBoom
		},
	}
	mgr := NewCommentManager(&fakeCommentRepo{}, blogRepo)

	_, err := mgr.GetCommentsBySlug(context.Background(), testSlug)

	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom to be propagated, got: %v", err)
	}
}

func TestGetCommentsBySlug_CommentRepoError(t *testing.T) {
	commentRepo := &fakeCommentRepo{
		getBlogCommentsFunc: func(ctx context.Context, blogId types.BlogId) ([]*models.Comment, error) {
			return nil, errBoom
		},
	}
	mgr := NewCommentManager(commentRepo, foundBlogRepo())

	_, err := mgr.GetCommentsBySlug(context.Background(), testSlug)

	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom to be propagated, got: %v", err)
	}
}

func TestGetCommentsBySlug_EmptyList(t *testing.T) {
	cases := []struct {
		name string
		flat []*models.Comment
	}{
		{"nil from repo", nil},
		{"empty slice from repo", []*models.Comment{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			commentRepo := &fakeCommentRepo{
				getBlogCommentsFunc: func(ctx context.Context, blogId types.BlogId) ([]*models.Comment, error) {
					return tc.flat, nil
				},
			}
			mgr := NewCommentManager(commentRepo, foundBlogRepo())

			got, err := mgr.GetCommentsBySlug(context.Background(), testSlug)

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got == nil || len(got) != 0 {
				t.Fatalf("expected a non-nil empty result, got: %#v", got)
			}
		})
	}
}

func TestGetCommentsBySlug_PassesResolvedBlogIdToRepo(t *testing.T) {
	blogId := types.BlogId(77)
	blogRepo := &fakeBlogRepo{
		getBlogContentBySlugFunc: func(ctx context.Context, slug types.BlogSlug) (*blogmodels.BlogContent, error) {
			return &blogmodels.BlogContent{Id: blogId, Slug: testSlug}, nil
		},
	}
	var capturedBlogId types.BlogId
	commentRepo := &fakeCommentRepo{
		getBlogCommentsFunc: func(ctx context.Context, id types.BlogId) ([]*models.Comment, error) {
			capturedBlogId = id
			return nil, nil
		},
	}
	mgr := NewCommentManager(commentRepo, blogRepo)

	if _, err := mgr.GetCommentsBySlug(context.Background(), testSlug); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if capturedBlogId != blogId {
		t.Fatalf("expected blogId %v passed to repo, got %v", blogId, capturedBlogId)
	}
}

func TestGetCommentsBySlug_BuildsTreeFromFlatList(t *testing.T) {
	parent1 := types.CommentId(1)
	parent2 := types.CommentId(2)
	flat := []*models.Comment{
		{Id: parent1, Content: "parent1", CreatedAt: fixedTime},
		{Id: 11, ParentId: &parent1, Content: "reply1a", CreatedAt: fixedTime},
		{Id: parent2, Content: "parent2", CreatedAt: fixedTime},
		{Id: 12, ParentId: &parent1, Content: "reply1b", CreatedAt: fixedTime},
	}
	commentRepo := &fakeCommentRepo{
		getBlogCommentsFunc: func(ctx context.Context, blogId types.BlogId) ([]*models.Comment, error) {
			return flat, nil
		},
	}
	mgr := NewCommentManager(commentRepo, foundBlogRepo())

	got, err := mgr.GetCommentsBySlug(context.Background(), testSlug)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	want := []*dtos.Comment{
		{
			Id:          parent1,
			AuthorName:  constants.AnonymousDisplayName,
			IsAnonymous: false,
			Content:     "parent1",
			CreatedAt:   fixedTime,
			Replies: []*dtos.Comment{
				{Id: 11, ParentId: &parent1, AuthorName: constants.AnonymousDisplayName, Content: "reply1a", CreatedAt: fixedTime, Replies: []*dtos.Comment{}},
				{Id: 12, ParentId: &parent1, AuthorName: constants.AnonymousDisplayName, Content: "reply1b", CreatedAt: fixedTime, Replies: []*dtos.Comment{}},
			},
		},
		{
			Id:          parent2,
			AuthorName:  constants.AnonymousDisplayName,
			IsAnonymous: false,
			Content:     "parent2",
			CreatedAt:   fixedTime,
			Replies:     []*dtos.Comment{},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}
