// Copyright 2016-2026, Pulumi Corporation.  All rights reserved.

package apiclient

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var e *APIError
		assert.Equal(t, "<nil>", e.Error())
	})

	t.Run("with message", func(t *testing.T) {
		e := NewAPIError(http.StatusBadRequest, "bad input", nil)
		assert.Equal(t, "HTTP 400: bad input", e.Error())
	})

	t.Run("status code with known text", func(t *testing.T) {
		e := NewAPIError(http.StatusNotFound, "", nil)
		assert.Equal(t, "HTTP 404: Not Found", e.Error())
	})

	t.Run("zero status, no message", func(t *testing.T) {
		e := NewAPIError(0, "", nil)
		assert.Equal(t, "api error", e.Error())
	})

	t.Run("unknown status code, no message", func(t *testing.T) {
		e := NewAPIError(799, "", nil)
		assert.Equal(t, "api error", e.Error())
	})
}

func TestAPIError_Accessors(t *testing.T) {
	header := http.Header{"Retry-After": []string{"30"}}
	e := NewAPIError(http.StatusConflict, "version mismatch", header)

	assert.Equal(t, http.StatusConflict, e.HTTPStatusCode())
	assert.Equal(t, "version mismatch", e.ResponseMessage())
	assert.Equal(t, "30", e.ResponseHeader().Get("Retry-After"))
}

func TestAPIError_StatusPredicates(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		isNotFound bool
		isConflict bool
		isNoContent bool
	}{
		{"404", http.StatusNotFound, true, false, false},
		{"409", http.StatusConflict, false, true, false},
		{"204", http.StatusNoContent, false, false, true},
		{"500", http.StatusInternalServerError, false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewAPIError(tc.status, "", nil)
			assert.Equal(t, tc.isNotFound, e.IsNotFound())
			assert.Equal(t, tc.isConflict, e.IsConflict())
			assert.Equal(t, tc.isNoContent, e.IsNoContent())
		})
	}
}
