// Copyright 2026, Pulumi Corporation.

package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOidcIssuer(t *testing.T) {
	t.Run("not found returns nil", func(t *testing.T) {
		c := startTestServerMulti(t, func(*http.Request) (int, any) {
			return http.StatusNotFound, ErrorResponse{StatusCode: http.StatusNotFound, Message: "not found"}
		})
		issuer, err := c.GetOidcIssuer(t.Context(), "an-org", "an-issuer")
		assert.NoError(t, err)
		assert.Nil(t, issuer, "a 404 should report the issuer as missing")
	})

	t.Run("server unreachable", func(t *testing.T) {
		_, err := newUnreachableClient(t).GetOidcIssuer(t.Context(), "an-org", "an-issuer")
		assert.Error(t, err, "a request with no response should fail the read")
	})
}
