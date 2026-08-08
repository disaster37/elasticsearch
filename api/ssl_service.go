package api

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// SslCertificate describes a single SSL certificate.
type SslCertificate struct {
	Path          string `json:"path"`
	SubjectDN     string `json:"subject_dn"`
	SerialNumber  string `json:"serial_number"`
	HasPrivateKey bool   `json:"has_private_key"`
	Alias         string `json:"alias,omitempty"`
	IssuerDN      string `json:"issuer_dn,omitempty"`
	Version       string `json:"version,omitempty"`
	NotBefore     string `json:"not_before,omitempty"`
	NotAfter      string `json:"not_after,omitempty"`
}

// SslCertificatesResponse is the response from GET /_ssl/certificates.
type SslCertificatesResponse []SslCertificate

// SslService provides access to the SSL APIs.
type SslService interface {
	Certificates(ctx context.Context) (SslCertificatesResponse, error)
}

// DefaultSslService is the default SslService implementation.
type DefaultSslService struct {
	client *resty.Client
	logger *logrus.Entry
}

// NewSslService creates a new SslService.
func NewSslService(client *resty.Client, logger *logrus.Entry) SslService {
	return &DefaultSslService{client: client, logger: logger.WithField("service", "ssl")}
}

func (s *DefaultSslService) Certificates(ctx context.Context) (SslCertificatesResponse, error) {
	var result SslCertificatesResponse
	return result, doReq(s.logger, ctx, s.client, "GET", "/_ssl/certificates", nil, nil, &result)
}
