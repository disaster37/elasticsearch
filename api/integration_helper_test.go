//go:build integration

package api_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/disaster37/elasticsearch/v9"
	"github.com/sirupsen/logrus"
)

// integrationClient builds a Client from the ELASTICSEARCH_* environment
// variables. It panics with guidance when the URL is unset so integration runs
// fail fast instead of silently skipping.
func integrationClient(t *testing.T) elasticsearch.Client {
	t.Helper()
	url := os.Getenv("ELASTICSEARCH_URL")
	if url == "" {
		t.Fatalf("ELASTICSEARCH_URL is required for integration tests")
	}
	username := os.Getenv("ELASTICSEARCH_USERNAME")
	password := os.Getenv("ELASTICSEARCH_PASSWORD")
	logger := logrus.NewEntry(logrus.StandardLogger())
	c, err := elasticsearch.New(&elasticsearch.Config{
		Addresses:         []string{url},
		Username:          username,
		Password:          password,
		TLSSkipVerify:     true,
		AllowInsecureHTTP: true,
	}, logger)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

// uniqueIndex returns a unique index name for a test.
func uniqueIndex(name string) string {
	return fmt.Sprintf("test-%s-%d", name, time.Now().UnixNano())
}

func nowNano() int64 { return time.Now().UnixNano() }
