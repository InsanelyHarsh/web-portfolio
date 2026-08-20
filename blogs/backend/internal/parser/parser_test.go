package parser

import (
	"strings"
	"testing"
)

func TestMarkdownToHTML_BasicParagraph(t *testing.T) {
	out := string(MarkdownToHTML([]byte("Hello **world**.")))

	if !strings.Contains(out, "<strong>world</strong>") {
		t.Fatalf("expected bold text to render, got: %s", out)
	}
}

func TestMarkdownToHTML_FencedCodeKeepsLanguageClass(t *testing.T) {
	md := "```go\nfmt.Println(\"hi\")\n```"

	out := string(MarkdownToHTML([]byte(md)))

	if !strings.Contains(out, `<code class="language-go">`) {
		t.Fatalf("expected fenced code block to keep its language-go class, got: %s", out)
	}
}

func TestMarkdownToHTML_HeadingKeepsID(t *testing.T) {
	out := string(MarkdownToHTML([]byte("## Getting Started")))

	if !strings.Contains(out, `id="getting-started"`) {
		t.Fatalf("expected heading to keep its generated id, got: %s", out)
	}
}

func TestMarkdownToHTML_TableRenders(t *testing.T) {
	md := "| A | B |\n| - | - |\n| 1 | 2 |\n"

	out := string(MarkdownToHTML([]byte(md)))

	if !strings.Contains(out, "<table>") {
		t.Fatalf("expected markdown table to render as <table>, got: %s", out)
	}
}

func TestMarkdownToHTML_StripsScriptTag(t *testing.T) {
	md := "Hello <script>alert('xss')</script> world"

	out := string(MarkdownToHTML([]byte(md)))

	if strings.Contains(out, "<script") {
		t.Fatalf("expected <script> tag to be stripped, got: %s", out)
	}
}

func TestMarkdownToHTML_StripsOnClickHandler(t *testing.T) {
	md := `<div onclick="alert('xss')">click me</div>`

	out := string(MarkdownToHTML([]byte(md)))

	if strings.Contains(out, "onclick") {
		t.Fatalf("expected onclick handler to be stripped, got: %s", out)
	}
}

func TestMarkdownToHTML_StripsJavascriptHref(t *testing.T) {
	md := `[click me](javascript:alert('xss'))`

	out := string(MarkdownToHTML([]byte(md)))

	if strings.Contains(out, "javascript:") {
		t.Fatalf("expected javascript: href to be stripped, got: %s", out)
	}
}

func TestMarkdownToPlainText_StripsMarkup(t *testing.T) {
	out := string(MarkdownToPlainText([]byte("## Heading\n\nHello **world**.")))

	if strings.Contains(out, "<") || strings.Contains(out, "#") || strings.Contains(out, "*") {
		t.Fatalf("expected all markup stripped, got: %s", out)
	}
	if !strings.Contains(out, "Heading") || !strings.Contains(out, "Hello world.") {
		t.Fatalf("expected text content preserved, got: %s", out)
	}
}

func TestMarkdownToPlainText_StripsScriptTag(t *testing.T) {
	md := "Hello <script>alert('xss')</script> world"

	out := string(MarkdownToPlainText([]byte(md)))

	if strings.Contains(out, "<script") || strings.Contains(out, "alert") {
		t.Fatalf("expected <script> tag and its contents to be stripped, got: %s", out)
	}
}

func TestMarkdownToHTML_StripsHashnodeImageAlignHint(t *testing.T) {
	md := `![](https://cdn.hashnode.com/res/hashnode/image/upload/v1738449767153/8f42482b-bd2e-42f4-a359-47b78c6d2f05.png align="center")`

	out := string(MarkdownToHTML([]byte(md)))

	wantSrc := `src="https://cdn.hashnode.com/res/hashnode/image/upload/v1738449767153/8f42482b-bd2e-42f4-a359-47b78c6d2f05.png"`
	if !strings.Contains(out, wantSrc) {
		t.Fatalf("expected image to keep its src with the align hint stripped, got: %s", out)
	}
	if strings.Contains(out, "align") {
		t.Fatalf("expected align hint to be gone, got: %s", out)
	}
}

func TestMarkdownToHTML_NormalImageSrcUnaffected(t *testing.T) {
	md := `![alt text](https://example.com/a.png)`

	out := string(MarkdownToHTML([]byte(md)))

	if !strings.Contains(out, `src="https://example.com/a.png"`) {
		t.Fatalf("expected ordinary image src to render unchanged, got: %s", out)
	}
}

func TestMarkdownToHTML_StripsArbitraryClass(t *testing.T) {
	md := `<p class="not-a-language-tag">text</p>`

	out := string(MarkdownToHTML([]byte(md)))

	if strings.Contains(out, "class=") {
		t.Fatalf("expected non-language class on <p> to be stripped, got: %s", out)
	}
}
