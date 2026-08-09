package parser

import (
	"regexp"

	"github.com/gomarkdown/markdown"
	mdparser "github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

// mdExtensions mirrors gomarkdown's own default (parser.New()) but
// additionally turns on AutoHeadingIDs: CommonExtensions' HeadingIDs bit only
// recognizes an explicit "{#id}" written by the author, it does not derive an
// id from heading text, so headings would render without an id for anchor
// links unless AutoHeadingIDs is enabled too.
const mdExtensions = mdparser.CommonExtensions | mdparser.AutoHeadingIDs

// htmlPolicy is built once and reused across calls: bluemonday policies are
// safe for concurrent use once constructed, and building one involves
// configuring dozens of allow-rules, which is wasteful to repeat on every
// MarkdownToHTML call.
var htmlPolicy = newSanitizePolicy()

// newSanitizePolicy extends bluemonday's UGCPolicy with a couple of narrow,
// pattern-bounded allowances needed to preserve output that gomarkdown's
// default (CommonExtensions) parser produces:
//
//   - FencedCode renders a language tag as class="language-xxx" on <code>,
//     which UGCPolicy strips (it disallows "class" entirely). Highlighting
//     that class lets client-side syntax highlighters pick the language.
//   - HeadingIDs renders id="..." on h1-h6 for anchor links, which UGCPolicy
//     also strips (headings take no attributes).
//
// Both allowances are pattern-bounded so only the safe, expected shape of
// value survives; anything else continues to be stripped exactly as before.
func newSanitizePolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	p.AllowAttrs("class").
		Matching(regexp.MustCompile(`^language-[\w+.#-]{1,32}$`)).
		OnElements("code")

	p.AllowAttrs("id").
		Matching(regexp.MustCompile(`^[\w-]{1,64}$`)).
		OnElements("h1", "h2", "h3", "h4", "h5", "h6")

	return p
}

// MarkdownToHTML renders markdown source to HTML and sanitizes the result so
// it is safe to embed directly in a page. Markdown source may come from an
// untrusted or semi-trusted author, so sanitization is not optional: it
// strips scripts, event handlers, and other unsafe markup while preserving
// the HTML gomarkdown's default extensions produce, including fenced code
// language classes and heading anchor ids.
func MarkdownToHTML(md []byte) []byte {
	// A parser is stateful and must not be reused across Parse() calls, so a
	// fresh one is created per call; only the sanitize policy is shared.
	p := mdparser.NewWithExtensions(mdExtensions)
	maybeUnsafeHTML := markdown.ToHTML(md, p, nil)
	return htmlPolicy.SanitizeBytes(maybeUnsafeHTML)
}
