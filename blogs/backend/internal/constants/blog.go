package constants

// Excerpt limits used when building blog list previews: at most
// ExcerptMaxLines non-empty lines, joined and hard-capped at
// ExcerptMaxChars so a single very long line can't blow up the response.
const (
	ExcerptMaxLines = 2
	ExcerptMaxChars = 200
)
