// Copyright 2026, Pulumi Corporation.

package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTemplateSources(t *testing.T) {
	sources := []TemplateSourceResponse{
		{ID: "id-1", IsValid: true, Name: "bootstrap-12345", SourceURL: "https://github.com/pulumi/pulumi"},
		{ID: "id-2", IsValid: true, Name: "other", SourceURL: "https://github.com/pulumi/examples"},
	}

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/" + testOrgName + "/templates/sources",
			ResponseCode:      200,
			ResponseBody:      ListResponse{Sources: sources},
		})
		actual, err := c.ListTemplateSources(ctx, testOrgName)
		assert.NoError(t, err)
		assert.Equal(t, sources, actual)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/" + testOrgName + "/templates/sources",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		_, err := c.ListTemplateSources(ctx, testOrgName)
		assert.EqualError(t, err,
			"failed to list template sources in org anOrg due to error: 401 API error: unauthorized")
	})
}

func TestGetTemplateSource(t *testing.T) {
	sources := []TemplateSourceResponse{
		{ID: "id-1", Name: "bootstrap-12345", SourceURL: "https://github.com/pulumi/pulumi"},
	}
	newClient := func(t *testing.T) *Client {
		return startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/" + testOrgName + "/templates/sources",
			ResponseCode:      200,
			ResponseBody:      ListResponse{Sources: sources},
		})
	}

	t.Run("Found", func(t *testing.T) {
		actual, err := newClient(t).GetTemplateSource(ctx, testOrgName, "id-1")
		assert.NoError(t, err)
		assert.Equal(t, &sources[0], actual)
	})

	t.Run("Not Found", func(t *testing.T) {
		actual, err := newClient(t).GetTemplateSource(ctx, testOrgName, "missing")
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})
}
