// Copyright 2016-2026, Pulumi Corporation.  All rights reserved.

package apiclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.Handler) (*CloudClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &CloudClient{
		BaseURL:  srv.URL,
		Executor: srv.Client().Do,
	}, srv
}

func TestCloudClient_CreateRequest_BadBaseURL(t *testing.T) {
	c := &CloudClient{BaseURL: "://bad"}
	_, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.Error(t, err)
}

func TestCloudClient_CreateRequest_PathParams(t *testing.T) {
	c, srv := newTestClient(t, http.NotFoundHandler())
	c.BaseURL = srv.URL

	req, err := c.createRequest(context.Background(), http.MethodGet,
		"/api/orgs/{org}/stacks/{stack}",
		map[string]any{"org": "acme corp", "stack": "prod/123"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "/api/orgs/acme%20corp/stacks/prod%2F123", req.URL.RequestURI())
}

func TestCloudClient_CreateRequest_QueryParams(t *testing.T) {
	c, srv := newTestClient(t, http.NotFoundHandler())
	c.BaseURL = srv.URL

	limit := 10
	tags := []string{"a", "b"}
	var nilStr *string
	req, err := c.createRequest(context.Background(), http.MethodGet, "/api/things", nil,
		map[string]any{"limit": &limit, "tags": &tags, "filter": nilStr, "skip": nil})
	require.NoError(t, err)

	q := req.URL.Query()
	assert.Equal(t, "10", q.Get("limit"))
	assert.Equal(t, "a,b", q.Get("tags"))
	assert.False(t, q.Has("filter"), "nil pointer should be omitted")
	assert.False(t, q.Has("skip"), "nil any should be omitted")
}

func TestCloudClient_CreateRequest_QueryParamMustBePointer(t *testing.T) {
	c, srv := newTestClient(t, http.NotFoundHandler())
	c.BaseURL = srv.URL

	_, err := c.createRequest(context.Background(), http.MethodGet, "/api/x", nil,
		map[string]any{"v": "raw-string"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a pointer")
}

func TestCloudClient_CreateRequestWithBody(t *testing.T) {
	c, srv := newTestClient(t, http.NotFoundHandler())
	c.BaseURL = srv.URL

	req, err := c.createRequestWithBody(context.Background(), http.MethodPost, "/api/x", nil, nil,
		map[string]string{"hello": "world"})
	require.NoError(t, err)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"hello":"world"}`, string(body))
}

func TestCloudClient_Invoke_SetsDefaultAcceptHeader(t *testing.T) {
	var seen string
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))

	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)
	require.NoError(t, c.invoke(req, nil))
	assert.Equal(t, "application/vnd.pulumi+8", seen)
}

func TestCloudClient_Invoke_AppliesCustomHeaders(t *testing.T) {
	var got http.Header
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))

	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)

	headers := []http.Header{{
		"Accept":   []string{"text/plain"},
		"X-Custom": []string{"one", "two"},
	}}
	require.NoError(t, c.invoke(req, headers))

	assert.Equal(t, "text/plain", got.Get("Accept"))
	assert.Equal(t, []string{"one", "two"}, got.Values("X-Custom"))
}

func TestCloudClient_Invoke_NgrokBypassHeader(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("ngrok-skip-browser-warning")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := &CloudClient{
		BaseURL:  srv.URL + "/ngrok-fake",
		Executor: srv.Client().Do,
	}
	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)
	require.NoError(t, c.invoke(req, nil))
	assert.Equal(t, "true", got)
}

func TestCloudClient_Invoke_ExecutorError(t *testing.T) {
	c := &CloudClient{
		BaseURL: "http://example.invalid",
		Executor: func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		},
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"http://example.invalid/x", nil)
	require.NoError(t, err)

	err = c.invoke(req, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "performing HTTP request")
	assert.Contains(t, err.Error(), "network down")
}

func TestCloudClient_Invoke_ParsesStructuredErrorBody(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code":409,"message":"already exists"}`))
	}))

	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)

	err = c.invoke(req, nil)
	require.Error(t, err)
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusConflict, apiErr.HTTPStatusCode())
	assert.Equal(t, "already exists", apiErr.ResponseMessage())
}

func TestCloudClient_Invoke_NonJSONErrorBody(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`server fell over`))
	}))

	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)

	err = c.invoke(req, nil)
	require.Error(t, err)
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusInternalServerError, apiErr.HTTPStatusCode())
	assert.Equal(t, "server fell over", apiErr.ResponseMessage())
}

func TestCloudClient_Invoke_JSONErrorMissingCode(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{"unrelated":"shape"}`))
	}))

	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)

	err = c.invoke(req, nil)
	require.Error(t, err)
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusTeapot, apiErr.HTTPStatusCode())
}

func TestCloudClient_InvokeWithResponse(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "hello body")
	}))

	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)

	body, err := c.invokeWithResponse(req, nil)
	require.NoError(t, err)
	assert.Equal(t, "hello body", string(body))
}

func TestCloudClient_InvokeWithStreamingResponse_NoContent(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req, err := c.createRequest(context.Background(), http.MethodGet, "/x", nil, nil)
	require.NoError(t, err)

	_, err = c.invokeWithStreamingResponse(req, nil)
	require.Error(t, err)
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.True(t, apiErr.IsNoContent())
}

func TestCloudClient_CreateRequestWithRawBody(t *testing.T) {
	c, srv := newTestClient(t, http.NotFoundHandler())
	c.BaseURL = srv.URL

	raw := []byte("plain text")
	req, err := c.createRequestWithRawBody(context.Background(), http.MethodPost, "/x", nil, nil, raw)
	require.NoError(t, err)
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, "plain text", string(body))
	assert.Empty(t, req.Header.Get("Content-Type"), "raw body should not auto-set Content-Type")
}

func TestCloudClient_CreateRequest_PathUnescapeError(t *testing.T) {
	c, srv := newTestClient(t, http.NotFoundHandler())
	c.BaseURL = srv.URL

	_, err := c.createRequest(context.Background(), http.MethodGet,
		"/api/%ZZ-bad", nil, nil)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid URL escape") ||
		strings.Contains(err.Error(), "%ZZ"))
}
