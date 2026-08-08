package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestRedactionNoCredentialLeakInLogsAndError is a regression test for the
// CRITICAL CWE-532 finding: credentials embedded in the base URL must never
// leak into log output or into returned error values when a request fails.
//
// It spins up a test server, points a resty client at a base URL carrying
// basic-auth userinfo ("user:secretpassword"), triggers a 400 error, and
// asserts that neither the captured logrus output nor the returned error's
// .Error() string contains the secret password.
func TestRedactionNoCredentialLeakInLogsAndError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trigger the logAndReturnError path with a 400.
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"type":"bad_request","reason":"nope"},"status":400}`))
	}))
	defer srv.Close()

	// Build a logger that writes to an in-memory buffer so we can inspect
	// what would have been emitted.
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	entry := logrus.NewEntry(log)

	// Construct a resty client whose base URL carries credentials. resty
	// parses the userinfo out of the base URL and applies it as basic auth,
	// but resp.Request.URL still contains the raw userinfo until redacted.
	client := restyClient(srv)
	client.SetBaseURL(strings.Replace(srv.URL, "127.0.0.1", "user:secretpassword@127.0.0.1", 1))

	svc := NewDocumentService(client, entry)

	// A GET that the server answers with 400 exercises logAndReturnError.
	_, err := svc.Get(context.Background(), &GetRequest{Index: "idx", Id: "1"})

	// The returned error must not leak the password.
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "secretpassword",
		"returned error must not contain the secret password")

	// The captured log output must not leak the password either.
	assert.NotContains(t, buf.String(), "secretpassword",
		"log output must not contain the secret password")
}

// TestRedactionNoCredentialLeakInUnmarshalError is a regression test for the
// wrapUnmarshalError path: when the response body is not valid JSON, the
// returned wrapped error and the log output must both redact the URL.
func TestRedactionNoCredentialLeakInUnmarshalError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 200 OK but with a body that is not valid JSON, forcing the
		// json.Unmarshal -> wrapUnmarshalError path.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json at all`))
	}))
	defer srv.Close()

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	entry := logrus.NewEntry(log)

	client := restyClient(srv)
	client.SetBaseURL(strings.Replace(srv.URL, "127.0.0.1", "user:secretpassword@127.0.0.1", 1))

	svc := NewDocumentService(client, entry)

	_, err := svc.Get(context.Background(), &GetRequest{Index: "idx", Id: "1"})

	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "secretpassword",
		"returned unmarshal error must not contain the secret password")
	assert.NotContains(t, buf.String(), "secretpassword",
		"log output for unmarshal error must not contain the secret password")
}
