// Copyright 2016-2026, Pulumi Corporation.  All rights reserved.

package apiclient

import (
	"fmt"
	"net/http"
)

// APIError represents an HTTP API error response, preserving the status code,
// message, and response headers so callers can programmatically detect
// well-known conditions (204, 404, 409, etc.) and inspect headers (e.g. Retry-After).
type APIError struct {
	statusCode int
	message    string
	header     http.Header
}

// NewAPIError creates a new APIError with the given status code, message, and optional response headers.
func NewAPIError(statusCode int, message string, header http.Header) *APIError {
	return &APIError{statusCode: statusCode, message: message, header: header}
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.statusCode, e.message)
	}
	if e.statusCode != 0 {
		if text := http.StatusText(e.statusCode); text != "" {
			return fmt.Sprintf("HTTP %d: %s", e.statusCode, text)
		}
	}
	return "api error"
}

// HTTPStatusCode returns the HTTP status code of the API response.
func (e *APIError) HTTPStatusCode() int {
	return e.statusCode
}

// ResponseMessage returns the error message from the API response.
func (e *APIError) ResponseMessage() string {
	return e.message
}

// ResponseHeader returns the HTTP response headers.
func (e *APIError) ResponseHeader() http.Header {
	return e.header
}

// IsNotFound reports whether the error represents an HTTP 404 Not Found response.
func (e *APIError) IsNotFound() bool {
	return e.statusCode == http.StatusNotFound
}

// IsConflict reports whether the error represents an HTTP 409 Conflict response.
func (e *APIError) IsConflict() bool {
	return e.statusCode == http.StatusConflict
}

// IsNoContent reports whether the response had HTTP 204 No Content status.
// This is used by invokeWithStreamingResponse when a caller expected a body
// but the server returned none.
func (e *APIError) IsNoContent() bool {
	return e.statusCode == http.StatusNoContent
}
