package terminal

import (
	"net/url"
	"slices"
)

const unsafeURLSubstitution = "#"

func sanitizeURL(s string) string {
	url, err := url.Parse(s)
	if err != nil {
		return unsafeURLSubstitution
	}

	// relative URLs (no scheme) are permitted
	if url.Scheme == "" {
		return url.String()
	}

	// deny-list known-XSS-dangerous URL schemes for <a href=""> etc.
	// An allow-list would be preferable, but we don't know what URL schemes are being legitimately
	// used in the wild, so that would be a breaking change, and likely require configurability.
	disallowedSchemes := []string{"javascript"}
	if slices.Contains(disallowedSchemes, url.Scheme) {
		return unsafeURLSubstitution
	}

	// default allow
	return url.String()
}

// strictURL reports whether s is an absolute http(s) URL, the only links
// WithStrict keeps.
func strictURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
