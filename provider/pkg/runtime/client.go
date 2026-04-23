// Copyright 2016-2026, Pulumi Corporation.
//
// client.go — authenticated HTTP client for the Pulumi Cloud REST API.
// One client instance is shared across all CRUD dispatches.

package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the Pulumi Cloud HTTP client. Thin wrapper over net/http that
// adds auth, source identification, and JSON marshaling.
type Client struct {
	BaseURL    string // e.g. "https://api.pulumi.com"
	AccessToken string // Bearer token
	HTTP       *http.Client
}

// NewClient constructs a Client with sensible defaults. A custom HTTP
// client can be supplied via the Transport field after construction if
// callers need TLS pinning, proxy settings, etc.
func NewClient(baseURL, accessToken string) *Client {
	if baseURL == "" {
		baseURL = "https://api.pulumi.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL:    baseURL,
		AccessToken: accessToken,
		HTTP: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Request is one HTTP call against the Pulumi Cloud API. The dispatcher
// constructs these from metadata + inputs and the client executes them.
// Either Body (JSON-encoded) or RawBody (sent verbatim) may be set; not
// both. RawBody paired with ContentType supports endpoints that take
// non-JSON payloads, e.g. ESC Environment's YAML PATCH.
type Request struct {
	Method      string            // uppercase HTTP verb
	Path        string            // already-expanded path (`/api/orgs/acme-corp/...`)
	Query       url.Values        // query parameters
	Body        interface{}       // marshaled as JSON if non-nil
	RawBody     []byte            // raw bytes sent verbatim (mutually exclusive with Body)
	ContentType string            // Content-Type override; defaults to application/json
	Headers     map[string]string // extra headers beyond the defaults
}

// Response carries the decoded JSON body plus metadata for error diagnostics.
type Response struct {
	StatusCode int
	Body       []byte
	Parsed     interface{} // lazily decoded in Call
}

// Call executes a single Request. Non-2xx responses surface as errors with
// the server's body attached for diagnostics (Pulumi Cloud returns structured
// error payloads we want the user to see).
func (c *Client) Call(ctx context.Context, req Request) (*Response, error) {
	u := c.BaseURL + req.Path
	if len(req.Query) > 0 {
		u = u + "?" + req.Query.Encode()
	}

	var bodyReader io.Reader
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	switch {
	case req.RawBody != nil:
		bodyReader = bytes.NewReader(req.RawBody)
	case req.Body != nil:
		raw, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	// Standard Pulumi Cloud headers — must match what the hand-written v1
	// provider and the Pulumi CLI send. See CLAUDE.md for the rationale.
	httpReq.Header.Set("Authorization", "token "+c.AccessToken)
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Accept", "application/vnd.pulumi+8")
	httpReq.Header.Set("X-Pulumi-Source", "provider")
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", req.Method, req.Path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	out := &Response{StatusCode: resp.StatusCode, Body: body}
	if resp.StatusCode >= 400 {
		return out, &HTTPError{
			Method:     req.Method,
			URL:        u,
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}
	return out, nil
}

// HTTPError is a non-2xx response. The body is frequently a structured
// Pulumi Cloud error object; we keep it raw so callers can render whichever
// fields are most useful for the user.
type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	bodyExcerpt := strings.TrimSpace(string(e.Body))
	const maxLen = 512
	if len(bodyExcerpt) > maxLen {
		bodyExcerpt = bodyExcerpt[:maxLen] + "…"
	}
	return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.URL, e.StatusCode, bodyExcerpt)
}

// Decode parses the response body into the provided value. Returns nil if
// the response body is empty (some Pulumi Cloud endpoints return 204).
func (r *Response) Decode(v interface{}) error {
	if len(r.Body) == 0 {
		return nil
	}
	return json.Unmarshal(r.Body, v)
}
