package types

import (
	"net/url"
	"strconv"
)

// boolStr returns "true" or "false" for a bool value.
func boolStr(b bool) string {
	return strconv.FormatBool(b)
}

// RedactURL removes sensitive userinfo from a URL string (e.g.,
// "https://user:pass@host/path" → "https://host/path"). It is more aggressive
// than net/url.Redacted, which preserves the username when no password is
// present — we remove the entire userinfo section because tokens, API keys,
// and short-lived credentials are often placed in the username slot alone
// (e.g. "https://api-key-12345@host"), and we must not leak them into logs or
// exported telemetry attributes. If the URL cannot be parsed it is returned
// unchanged.
func RedactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.User != nil {
		u.User = nil
	}
	return u.String()
}

// joinComma joins a string slice with a comma. Returns "" for an empty slice.
func joinComma(s []string) string {
	if len(s) == 0 {
		return ""
	}
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out
}
