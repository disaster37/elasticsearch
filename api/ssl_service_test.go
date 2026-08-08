package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSslCertificates(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		jsonResponse(w, 200, `[{"path":"/certs/cert.p12","subject_dn":"CN=example","serial_number":"123","has_private_key":true}]`)
	}))
	defer srv.Close()
	svc := NewSslService(restyClient(srv), testLogger())
	resp, err := svc.Certificates(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "GET", gotMethod)
	assert.Equal(t, "/_ssl/certificates", gotPath)
	require.Len(t, resp, 1)
	assert.Equal(t, "/certs/cert.p12", resp[0].Path)
	assert.Equal(t, "CN=example", resp[0].SubjectDN)
	assert.Equal(t, "123", resp[0].SerialNumber)
	assert.True(t, resp[0].HasPrivateKey)
}

func TestUnitSslCertificatesError(t *testing.T) {
	srv := errServer(500)
	defer srv.Close()
	svc := NewSslService(restyClient(srv), testLogger())
	_, err := svc.Certificates(context.Background())
	assert.Error(t, err)
}

func TestUnitSslCertificatesNetworkError(t *testing.T) {
	svc := NewSslService(deadClient(), testLogger())
	_, err := svc.Certificates(context.Background())
	assert.Error(t, err)
}
