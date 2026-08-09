package config

import (
	"net/url"
	"strings"
)

// ParseCommaSeparated splits a comma-separated env var value into trimmed,
// non-empty parts. Empty/blank input yields nil, so callers can distinguish
// "unset" from "explicitly empty" and apply a default.
func ParseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// NormalizeOrigin reduces a raw CORS origin down to scheme://host(:port),
// discarding any path/query/fragment a caller might paste by mistake (e.g.
// a full page URL instead of the page's origin). A browser's Origin request
// header never includes a path, so a configured origin with one would
// silently never match and CORS would appear to just not work. Returns the
// input unchanged if it doesn't parse as a URL with both a scheme and host.
func NormalizeOrigin(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return raw
	}
	return u.Scheme + "://" + u.Host
}
