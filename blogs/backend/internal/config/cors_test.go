package config

import (
	"reflect"
	"testing"
)

func TestParseCommaSeparated(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "http://localhost:3000", []string{"http://localhost:3000"}},
		{
			"multiple with whitespace",
			" http://localhost:3000 , http://localhost:5501 ",
			[]string{"http://localhost:3000", "http://localhost:5501"},
		},
		{"drops empty entries", "http://a.com,,http://b.com,", []string{"http://a.com", "http://b.com"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseCommaSeparated(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParseCommaSeparated(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeOrigin_PlainOriginUnchanged(t *testing.T) {
	got := NormalizeOrigin("http://localhost:3000")
	if got != "http://localhost:3000" {
		t.Fatalf("expected plain origin to pass through unchanged, got: %s", got)
	}
}

// TestNormalizeOrigin_StripsPath is a regression test for the reported bug:
// a full page URL pasted into CORS_ALLOWED_ORIGIN instead of just its
// origin silently never matched a browser's Origin header.
func TestNormalizeOrigin_StripsPath(t *testing.T) {
	got := NormalizeOrigin("http://127.0.0.1:5501/blogs/web/blog_list.html")
	if got != "http://127.0.0.1:5501" {
		t.Fatalf("expected path to be stripped, got: %s", got)
	}
}

func TestNormalizeOrigin_StripsQueryAndFragment(t *testing.T) {
	got := NormalizeOrigin("http://localhost:5501/page?foo=bar#section")
	if got != "http://localhost:5501" {
		t.Fatalf("expected query/fragment to be stripped, got: %s", got)
	}
}

func TestNormalizeOrigin_UnparseableValuePassesThrough(t *testing.T) {
	raw := "not a url"
	got := NormalizeOrigin(raw)
	if got != raw {
		t.Fatalf("expected unparseable value to pass through unchanged, got: %s", got)
	}
}
