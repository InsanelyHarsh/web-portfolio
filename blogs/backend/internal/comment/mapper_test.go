package comment

import (
	"reflect"
	"testing"
	"time"

	"github.com/insanelyharsh/web-portfolio/dtos"
	"github.com/insanelyharsh/web-portfolio/internal/comment/models"
	"github.com/insanelyharsh/web-portfolio/internal/constants"
	"github.com/insanelyharsh/web-portfolio/internal/types"
)

var mapperFixedTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func strPtr(s string) *string { return &s }

func commentIdPtr(id types.CommentId) *types.CommentId { return &id }

func TestToCommentDTO_MapsAllFields(t *testing.T) {
	// AuthorEmail/AuthorIP are populated on the input model to make clear
	// they're excluded from the DTO on purpose (dtos.Comment simply has no
	// such fields - enforced by the compiler), not by omission here.
	c := &models.Comment{
		Id:          types.CommentId(1),
		BlogId:      types.BlogId(10),
		ParentId:    nil,
		IsAnonymous: false,
		AuthorName:  strPtr("Bob"),
		AuthorEmail: strPtr("bob@example.com"),
		AuthorIP:    strPtr("203.0.113.5"),
		Content:     "hello world",
		CreatedAt:   mapperFixedTime,
		UpdatedAt:   mapperFixedTime,
	}

	got := toCommentDTO(c)

	if got.Id != c.Id {
		t.Fatalf("expected Id %v, got %v", c.Id, got.Id)
	}
	if got.IsAnonymous != c.IsAnonymous {
		t.Fatalf("expected IsAnonymous %v, got %v", c.IsAnonymous, got.IsAnonymous)
	}
	if got.Content != c.Content {
		t.Fatalf("expected Content %q, got %q", c.Content, got.Content)
	}
	if got.AuthorName != "Bob" {
		t.Fatalf("expected AuthorName %q, got %q", "Bob", got.AuthorName)
	}
	// Use .Equal rather than == / reflect.DeepEqual: a time.Time derived
	// from time.Now() carries a monotonic reading that can make two
	// otherwise-identical values compare unequal. mapperFixedTime is built
	// via time.Date so this particular case is safe either way, but .Equal
	// is the correct tool for comparing time.Time in general.
	if !got.CreatedAt.Equal(c.CreatedAt) {
		t.Fatalf("expected CreatedAt %v, got %v", c.CreatedAt, got.CreatedAt)
	}
	if !reflect.DeepEqual(got.Replies, []*dtos.Comment{}) {
		t.Fatalf("expected Replies to be a non-nil empty slice, got %#v", got.Replies)
	}
}

func TestToCommentDTO_ParentId(t *testing.T) {
	cases := []struct {
		name     string
		parentId *types.CommentId
	}{
		{"nil parent", nil},
		{"set parent", commentIdPtr(types.CommentId(7))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &models.Comment{Id: types.CommentId(1), ParentId: tc.parentId, IsAnonymous: true, Content: "x"}
			got := toCommentDTO(c)

			if tc.parentId == nil {
				if got.ParentId != nil {
					t.Fatalf("expected nil ParentId, got %v", *got.ParentId)
				}
				return
			}
			if got.ParentId == nil || *got.ParentId != *tc.parentId {
				t.Fatalf("expected ParentId %v, got %v", *tc.parentId, got.ParentId)
			}
		})
	}
}

func TestToCommentTree_EmptyInput(t *testing.T) {
	cases := []struct {
		name string
		in   []*models.Comment
	}{
		{"nil slice", nil},
		{"empty slice", []*models.Comment{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := toCommentTree(tc.in)
			if got == nil {
				t.Fatalf("expected a non-nil empty slice, got nil")
			}
			if len(got) != 0 {
				t.Fatalf("expected an empty result, got %#v", got)
			}
		})
	}
}

func TestToCommentTree_AllTopLevel_OrderPreserved(t *testing.T) {
	flat := []*models.Comment{
		{Id: 3, Content: "c3"},
		{Id: 1, Content: "c1"},
		{Id: 2, Content: "c2"},
	}

	got := toCommentTree(flat)

	if len(got) != 3 {
		t.Fatalf("expected 3 top-level comments, got %d", len(got))
	}
	wantOrder := []types.CommentId{3, 1, 2}
	for i, wantId := range wantOrder {
		if got[i].Id != wantId {
			t.Fatalf("expected top-level order %v at index %d, got id %v", wantOrder, i, got[i].Id)
		}
		if len(got[i].Replies) != 0 {
			t.Fatalf("expected comment %v to have no replies, got %d", got[i].Id, len(got[i].Replies))
		}
	}
}

func TestToCommentTree_SingleParentMultipleReplies_OrderPreserved(t *testing.T) {
	parentId := types.CommentId(1)
	flat := []*models.Comment{
		{Id: 1, Content: "parent"},
		{Id: 4, ParentId: &parentId, Content: "reply4"},
		{Id: 2, ParentId: &parentId, Content: "reply2"},
		{Id: 3, ParentId: &parentId, Content: "reply3"},
	}

	got := toCommentTree(flat)

	if len(got) != 1 {
		t.Fatalf("expected 1 top-level comment, got %d", len(got))
	}
	replies := got[0].Replies
	wantOrder := []types.CommentId{4, 2, 3}
	if len(replies) != len(wantOrder) {
		t.Fatalf("expected %d replies, got %d", len(wantOrder), len(replies))
	}
	for i, wantId := range wantOrder {
		if replies[i].Id != wantId {
			t.Fatalf("expected reply order %v at index %d, got id %v (replies follow flat-slice order, not id order)", wantOrder, i, replies[i].Id)
		}
	}
}

func TestToCommentTree_MultipleTopLevelInterleavedWithReplies(t *testing.T) {
	p1 := types.CommentId(1)
	p2 := types.CommentId(2)
	flat := []*models.Comment{
		{Id: 1, Content: "parent1"},
		{Id: 20, ParentId: &p2, Content: "reply-to-parent2"},
		{Id: 2, Content: "parent2"},
		{Id: 10, ParentId: &p1, Content: "reply-to-parent1"},
	}

	got := toCommentTree(flat)

	if len(got) != 2 {
		t.Fatalf("expected 2 top-level comments, got %d", len(got))
	}
	if got[0].Id != 1 || got[1].Id != 2 {
		t.Fatalf("expected top-level order [1 2], got [%v %v]", got[0].Id, got[1].Id)
	}
	if len(got[0].Replies) != 1 || got[0].Replies[0].Id != 10 {
		t.Fatalf("expected parent1's only reply to be 10, got %#v", got[0].Replies)
	}
	if len(got[1].Replies) != 1 || got[1].Replies[0].Id != 20 {
		t.Fatalf("expected parent2's only reply to be 20, got %#v", got[1].Replies)
	}
}

func TestToCommentTree_OrphanReplyDropped(t *testing.T) {
	missingParent := types.CommentId(999)
	flat := []*models.Comment{
		{Id: 1, Content: "parent"},
		{Id: 2, ParentId: &missingParent, Content: "orphan"},
	}

	got := toCommentTree(flat)

	if len(got) != 1 {
		t.Fatalf("expected 1 top-level comment (an orphan reply must not surface there), got %d", len(got))
	}
	if len(got[0].Replies) != 0 {
		t.Fatalf("expected the orphan reply to be silently dropped, got %#v", got[0].Replies)
	}
}

// TestToCommentTree_ReplyToAReply_MapperIsPermissive documents that
// toCommentTree itself has no nesting-depth guard - it will happily render
// arbitrarily deep trees if fed one. The "flat nesting only" rule enforced
// when creating a comment lives solely in CommentManager.CreateComment.
func TestToCommentTree_ReplyToAReply_MapperIsPermissive(t *testing.T) {
	aId := types.CommentId(1)
	bId := types.CommentId(2)
	flat := []*models.Comment{
		{Id: aId, Content: "A"},
		{Id: bId, ParentId: &aId, Content: "B"},
		{Id: 3, ParentId: &bId, Content: "C"},
	}

	got := toCommentTree(flat)

	if len(got) != 1 || got[0].Id != aId {
		t.Fatalf("expected A as the sole top-level comment, got %#v", got)
	}
	if len(got[0].Replies) != 1 || got[0].Replies[0].Id != bId {
		t.Fatalf("expected A's only reply to be B, got %#v", got[0].Replies)
	}
	bReplies := got[0].Replies[0].Replies
	if len(bReplies) != 1 || bReplies[0].Id != 3 {
		t.Fatalf("expected B's only reply to be C, got %#v", bReplies)
	}
}

func TestDisplayName(t *testing.T) {
	cases := []struct {
		name        string
		isAnonymous bool
		authorName  *string
		want        string
	}{
		{"anonymous true overrides present name", true, strPtr("Alice"), constants.AnonymousDisplayName},
		{"not anonymous, nil author name", false, nil, constants.AnonymousDisplayName},
		{"not anonymous, empty string author name", false, strPtr(""), constants.AnonymousDisplayName},
		{"not anonymous, normal name", false, strPtr("Bob"), "Bob"},
		{"not anonymous, name not trimmed by displayName", false, strPtr("  Bob  "), "  Bob  "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &models.Comment{IsAnonymous: tc.isAnonymous, AuthorName: tc.authorName}
			got := displayName(c)
			if got != tc.want {
				t.Fatalf("displayName() = %q, want %q", got, tc.want)
			}
		})
	}
}
