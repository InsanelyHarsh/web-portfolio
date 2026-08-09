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

func TestMarkdownToHTML_StripsArbitraryClass(t *testing.T) {
	md := `<p class="not-a-language-tag">text</p>`

	out := string(MarkdownToHTML([]byte(md)))

	if strings.Contains(out, "class=") {
		t.Fatalf("expected non-language class on <p> to be stripped, got: %s", out)
	}
}
