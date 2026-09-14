package constants

const (
	CommentMaxLength    = 5000
	AuthorNameMaxLength = 100
)

const AnonymousDisplayName = "Anonymous"

// EmailPattern is the HTML5/WHATWG input[type=email] validation regex - a
// pragmatic format check (not full RFC 5322) applied to an author's email
// when one is given. It's optional even for non-anonymous comments and is
// never returned by the API (see dtos.Comment).
const EmailPattern = `^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
