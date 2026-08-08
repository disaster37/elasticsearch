package api

import (
	"fmt"

	json "github.com/goccy/go-json"
	"github.com/sirupsen/logrus"

	"github.com/disaster37/elasticsearch/v9/types"
	"github.com/go-resty/resty/v2"
)

func parseErrorResponse(resp *resty.Response) error {
	e := &types.ElasticsearchError{Status: resp.StatusCode()}
	if resp.Body() != nil {
		_ = json.Unmarshal(resp.Body(), e)
		if e.Status == 0 {
			e.Status = resp.StatusCode()
		}
	}
	return e
}

// logAndReturnError parses the Elasticsearch error response, logs it at Error
// level with the HTTP method, path and status code, and returns the parsed
// error. The request URL is redacted before logging so embedded credentials
// (e.g. "https://user:pass@host") never leak into log output.
func logAndReturnError(logger *logrus.Entry, resp *resty.Response) error {
	e := parseErrorResponse(resp)
	logger.WithFields(logrus.Fields{
		"http_method": resp.Request.Method,
		"http_path":   types.RedactURL(resp.Request.URL),
		"http_status": resp.StatusCode(),
	}).WithError(e).Error("Elasticsearch request failed")
	return e
}

// wrapNetworkError logs a network-level failure at Error level and returns the
// original error unchanged, preserving the error chain for errors.Is / errors.As.
func wrapNetworkError(logger *logrus.Entry, err error) error {
	logger.WithError(err).Error("HTTP request failed")
	return err
}

// wrapUnmarshalError logs a JSON deserialization failure at Error level,
// including a truncated response body for debugging, and returns a wrapped
// error with endpoint context (HTTP method + path). The request URL is
// redacted in BOTH the log fields and the returned error value so embedded
// credentials never leak into logs or error chains.
func wrapUnmarshalError(logger *logrus.Entry, resp *resty.Response, err error) error {
	body := string(resp.Body())
	if len(body) > 500 {
		body = body[:500] + "..."
	}
	redactedURL := types.RedactURL(resp.Request.URL)
	logger.WithFields(logrus.Fields{
		"http_method":   resp.Request.Method,
		"http_path":     redactedURL,
		"http_status":   resp.StatusCode(),
		"response_body": body,
	}).WithError(err).Error("Failed to unmarshal response")
	return fmt.Errorf("unmarshal response from %s %s: %w", resp.Request.Method, redactedURL, err)
}
