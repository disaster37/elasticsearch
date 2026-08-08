package api

import (
	"context"
	"fmt"
	"net/http"

	json "github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// buildRequest configures a resty Request with context, body, and query params.
// It is used internally by the doReq* and headReq helpers.
func buildRequest(client *resty.Client, ctx context.Context, body any, params interface{ ToMap() map[string]string }) (*resty.Request, error) {
	r := client.R().SetContext(ctx)
	if body != nil {
		nb, err := normalizeBody(body)
		if err != nil {
			return nil, err
		}
		r = r.SetBody(nb)
	}
	if params != nil {
		for k, v := range params.ToMap() {
			r.SetQueryParam(k, v)
		}
	}
	return r, nil
}

// doReq is a shared helper that performs an HTTP request, applies query params,
// optionally sets a body, and unmarshals the response into out (which must be a
// pointer). method is one of GET/POST/PUT/DELETE.
func doReq(logger *logrus.Entry, ctx context.Context, client *resty.Client, method, path string, body any, params interface{ ToMap() map[string]string }, out any) error {
	r, err := buildRequest(client, ctx, body, params)
	if err != nil {
		return err
	}
	resp, err := r.Execute(method, path)
	if err != nil {
		return wrapNetworkError(logger, err)
	}
	if resp.IsError() {
		return logAndReturnError(logger, resp)
	}
	if out != nil {
		if err := json.Unmarshal(resp.Body(), out); err != nil {
			return wrapUnmarshalError(logger, resp, err)
		}
	}
	return nil
}

// doReqRaw is like doReq but returns the raw response body as json.RawMessage.
func doReqRaw(logger *logrus.Entry, ctx context.Context, client *resty.Client, method, path string, body any, params interface{ ToMap() map[string]string }) (json.RawMessage, error) {
	r, err := buildRequest(client, ctx, body, params)
	if err != nil {
		return nil, err
	}
	resp, err := r.Execute(method, path)
	if err != nil {
		return nil, wrapNetworkError(logger, err)
	}
	if resp.IsError() {
		return nil, logAndReturnError(logger, resp)
	}
	return json.RawMessage(resp.Body()), nil
}

// doReqText is like doReq but returns the response body as a string.
func doReqText(logger *logrus.Entry, ctx context.Context, client *resty.Client, method, path string, body any, params interface{ ToMap() map[string]string }) (string, error) {
	r, err := buildRequest(client, ctx, body, params)
	if err != nil {
		return "", err
	}
	resp, err := r.Execute(method, path)
	if err != nil {
		return "", wrapNetworkError(logger, err)
	}
	if resp.IsError() {
		return "", logAndReturnError(logger, resp)
	}
	return string(resp.Body()), nil
}

// headReq performs a HEAD request and returns true for 200, false for 404,
// or a wrapped error for any other status. It accepts optional query params.
// SetDoNotParseResponse is used to prevent resty from attempting to read the
// (absent) response body on HEAD requests.
func headReq(logger *logrus.Entry, ctx context.Context, client *resty.Client, path string, params interface{ ToMap() map[string]string }) (bool, error) {
	r, err := buildRequest(client, ctx, nil, params)
	if err != nil {
		return false, err
	}
	resp, err := r.SetDoNotParseResponse(true).Head(path)
	if err != nil {
		return false, wrapNetworkError(logger, err)
	}
	defer func() { _ = resp.RawResponse.Body.Close() }()
	switch resp.StatusCode() {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("elasticsearch: unexpected status code %d", resp.StatusCode())
	}
}
