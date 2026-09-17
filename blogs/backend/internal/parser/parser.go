package parser

import (
	"regexp"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	mdparser "github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

// mdExtensions mirrors gomarkdown's own default (parser.New()) but
// additionally turns on AutoHeadingIDs: CommonExtensions' HeadingIDs bit only
// recognizes an explicit "{#id}" written by the author, it does not derive an
// id from heading text, so headings would render without an id for anchor
// links unless AutoHeadingIDs is enabled too.
const mdExtensions = mdparser.CommonExtensions | mdparser.AutoHeadingIDs

// htmlFlags extends gomarkdown's own default (html.CommonFlags) with
// LazyLoadImages, deferring offscreen <img> loads since posts can carry
// several inline screenshots/diagrams. Link safety (nofollow/noreferrer/
// noopener/target="_blank" on external links) is handled by htmlPolicy
// instead of gomarkdown's own link flags: gomarkdown decides "external"
// with a naive byte-prefix check on the href, whereas bluemonday parses the
// URL, and bluemonday's variants correctly restrict themselves to
// fully-qualified links so in-page heading anchors stay untouched. See
// newSanitizePolicy.
const htmlFlags = html.CommonFlags | html.LazyLoadImages

// htmlPolicy is built once and reused across calls: bluemonday policies are
// safe for concurrent use once constructed, and building one involves
// configuring dozens of allow-rules, which is wasteful to repeat on every
// MarkdownToHTML call.
var htmlPolicy = newSanitizePolicy()

// newSanitizePolicy extends bluemonday's UGCPolicy with the allowances
// needed to preserve output that gomarkdown's configured parser/renderer
// (see mdExtensions, htmlFlags) produces, plus link-safety hardening:
//
//   - FencedCode renders a language tag as class="language-xxx" on <code>,
//     which UGCPolicy strips (it disallows "class" entirely). Highlighting
//     that class lets client-side syntax highlighters pick the language.
//   - HeadingIDs renders id="..." on h1-h6 for anchor links, which UGCPolicy
//     also strips (headings take no attributes).
//   - LazyLoadImages renders loading="lazy" on <img>, which UGCPolicy also
//     strips (images take no attributes beyond src/alt/title).
//   - UGCPolicy defaults to RequireNoFollowOnLinks(true), which stamps
//     rel="nofollow" onto every <a> unconditionally, including our own
//     in-page heading anchors. That's swapped for the FullyQualified-link
//     variants (plus AddTargetBlankToFullyQualifiedLinks, which in turn
//     makes bluemonday add rel="noopener" itself) so only links leaving the
//     post get nofollow/noreferrer/noopener/target="_blank", leaving
//     same-page anchors untouched.
//
// The class/id/loading allowances are pattern-bounded so only the safe,
// expected shape of value survives; anything else continues to be stripped
// exactly as before.
func newSanitizePolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	p.AllowAttrs("class").
		Matching(regexp.MustCompile(`^language-[\w+.#-]{1,32}$`)).
		OnElements("code")

	p.AllowAttrs("id").
		Matching(regexp.MustCompile(`^[\w-]{1,64}$`)).
		OnElements("h1", "h2", "h3", "h4", "h5", "h6")

	p.AllowAttrs("loading").
		Matching(regexp.MustCompile(`^(?:lazy|eager)$`)).
		OnElements("img")

	p.RequireNoFollowOnLinks(false)
	p.RequireNoFollowOnFullyQualifiedLinks(true)
	p.RequireNoReferrerOnFullyQualifiedLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)

	return p
}

// stripPolicy removes every HTML tag, leaving plain text. Built once and
// reused, like htmlPolicy.
var stripPolicy = bluemonday.StrictPolicy()

// hashnodeImageAttrsPattern matches an image destination followed by one or
// more trailing key="value" tokens glued inside the same parens, e.g.
// Hashnode's markdown export writes image alignment as
// "![](https://.../img.png align=\"center\")" instead of using a proper
// CommonMark title. Because the opening quote there isn't preceded by
// whitespace, gomarkdown doesn't recognize it as a title and instead folds
// the whole "align=\"center\"" tail into the URL itself, producing a
// destination containing a literal space; bluemonday's sanitizer then
// rejects any non-data: URL containing whitespace and strips the src
// attribute entirely.
//
// The `+` on the trailing group requires at least one such token to match,
// so ordinary images with no trailing junk are left untouched.
var hashnodeImageAttrsPattern = regexp.MustCompile(`(!\[[^\]]*\]\(\S+)(?:\s+\w+="[^"]*")+\)`)

// stripHashnodeImageAttrs rewrites "(url attr=\"value\" ...)"-style trailing
// tokens out of image destinations, leaving just "(url)", before the
// markdown reaches gomarkdown. See hashnodeImageAttrsPattern for why this is
// needed.
func stripHashnodeImageAttrs(md []byte) []byte {
	return hashnodeImageAttrsPattern.ReplaceAll(md, []byte("$1)"))
}

// MarkdownToPlainText renders markdown to HTML (reusing MarkdownToHTML's
// sanitized pipeline) then strips all remaining tags, leaving plain text
// suitable for previews/excerpts.
func MarkdownToPlainText(md []byte) []byte {
	html := MarkdownToHTML(md)
	return stripPolicy.SanitizeBytes(html)
}

// MarkdownToHTML renders markdown source to HTML and sanitizes the result so
// it is safe to embed directly in a page. Markdown source may come from an
// untrusted or semi-trusted author, so sanitization is not optional: it
// strips scripts, event handlers, and other unsafe markup while preserving
// the HTML gomarkdown's default extensions produce, including fenced code
// language classes and heading anchor ids.
func MarkdownToHTML(md []byte) []byte {
	md = stripHashnodeImageAttrs(md)

	// A parser is stateful and must not be reused across Parse() calls, so a
	// fresh one is created per call. The renderer is likewise stateful (it
	// tracks heading ids seen so far to dedupe collisions within a single
	// document), so it too is built fresh per call; only the sanitize policy
	// is shared.
	p := mdparser.NewWithExtensions(mdExtensions)
	renderer := html.NewRenderer(html.RendererOptions{Flags: htmlFlags})
	maybeUnsafeHTML := markdown.ToHTML(md, p, renderer)
	return htmlPolicy.SanitizeBytes(maybeUnsafeHTML)
}
