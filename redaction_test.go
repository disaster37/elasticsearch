package elasticsearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactURL(t *testing.T) {
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://user:pass@host:9200/_search"))
	// Username-only userinfo (often an API key) must also be stripped.
	assert.Equal(t, "https://host:9200/", RedactURL("https://api-key-12345@host:9200/"))
	assert.Equal(t, "https://host:9200/_search", RedactURL("https://host:9200/_search"))
	// Unparseable input is returned unchanged.
	assert.Equal(t, "://bad", RedactURL("://bad"))
}

func TestRedactSensitiveHeaders(t *testing.T) {
	msg := "HEAD /_doc/1 HTTP/1.1\n\tAuthorization: ApiKey c2VjcmV0\n\tContent-Type: application/json\n\tCookie: session=abc\nresponse body"
	redacted := redactSensitiveHeaders(msg)
	assert.Contains(t, redacted, "Authorization: [REDACTED]")
	assert.Contains(t, redacted, "Cookie: [REDACTED]")
	assert.Contains(t, redacted, "Content-Type: application/json")
	assert.NotContains(t, redacted, "c2VjcmV0")
	assert.NotContains(t, redacted, "session=abc")
}

func TestRedactSensitiveHeadersNoColon(t *testing.T) {
	assert.Equal(t, "plain message", redactSensitiveHeaders("plain message"))
}
