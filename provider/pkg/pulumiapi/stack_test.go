package pulumiapi

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateStack(t *testing.T) {
	s := StackIdentifier{
		OrgName:     "organization",
		ProjectName: "project",
		StackName:   "stack",
	}
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s", s.OrgName, s.ProjectName),
			ExpectedReqBody:   CreateStackRequest{StackName: s.StackName},
			ResponseCode:      http.StatusNoContent,
		})
		assert.NoError(t, c.CreateStack(ctx, s, nil))
	})

	t.Run("With service-backed config", func(t *testing.T) {
		cfg := StackConfig{Environment: "default/prod-cfg@3"}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s", s.OrgName, s.ProjectName),
			ExpectedReqBody:   CreateStackRequest{StackName: s.StackName, Config: &cfg},
			ResponseCode:      http.StatusNoContent,
		})
		assert.NoError(t, c.CreateStack(ctx, s, &cfg))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s", s.OrgName, s.ProjectName),
			ResponseCode:      http.StatusUnauthorized,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.CreateStack(ctx, s, nil)
		assert.EqualError(t, err, "failed to create stack 'organization/project/stack': 401 API error: unauthorized")
	})
}

func TestDeleteStack(t *testing.T) {
	s := StackIdentifier{
		OrgName:     "organization",
		ProjectName: "project",
		StackName:   "stack",
	}
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s", s.OrgName, s.ProjectName, s.StackName),
			ResponseCode:      http.StatusNoContent,
		})
		assert.NoError(t, c.DeleteStack(ctx, s, false, false))
	})

	t.Run("Force destroy", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod:   http.MethodDelete,
			ExpectedReqPath:     fmt.Sprintf("/api/stacks/%s/%s/%s", s.OrgName, s.ProjectName, s.StackName),
			ExpectedQueryParams: url.Values{"force": {"true"}},
			ResponseCode:        http.StatusNoContent,
		})
		assert.NoError(t, c.DeleteStack(ctx, s, true, false))
	})

	t.Run("Preserve environment", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod:   http.MethodDelete,
			ExpectedReqPath:     fmt.Sprintf("/api/stacks/%s/%s/%s", s.OrgName, s.ProjectName, s.StackName),
			ExpectedQueryParams: url.Values{"preserveEnvironment": {"true"}},
			ResponseCode:        http.StatusNoContent,
		})
		assert.NoError(t, c.DeleteStack(ctx, s, false, true))
	})

	t.Run("Force and preserve", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s", s.OrgName, s.ProjectName, s.StackName),
			ExpectedQueryParams: url.Values{
				"force":               {"true"},
				"preserveEnvironment": {"true"},
			},
			ResponseCode: http.StatusNoContent,
		})
		assert.NoError(t, c.DeleteStack(ctx, s, true, true))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/organization/project/stack",
			ResponseCode:      http.StatusUnauthorized,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		assert.EqualError(t, c.DeleteStack(ctx, s, false, false),
			"failed to delete stack: 401 API error: unauthorized")
	})
}

func TestStackConfigClient(t *testing.T) {
	s := StackIdentifier{
		OrgName:     "organization",
		ProjectName: "project",
		StackName:   "stack",
	}
	t.Run("Get returns nil on 404", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/config", s.OrgName, s.ProjectName, s.StackName),
			ResponseCode:      http.StatusNotFound,
		})
		got, err := c.GetStackConfig(ctx, s)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Get returns config", func(t *testing.T) {
		body := StackConfig{Environment: "default/prod-cfg@3"}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/config", s.OrgName, s.ProjectName, s.StackName),
			ResponseCode:      http.StatusOK,
			ResponseBody:      body,
		})
		got, err := c.GetStackConfig(ctx, s)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, body, *got)
	})

	t.Run("Set sends PUT", func(t *testing.T) {
		body := StackConfig{Environment: "default/prod-cfg"}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/config", s.OrgName, s.ProjectName, s.StackName),
			ExpectedReqBody:   body,
			ResponseCode:      http.StatusOK,
			ResponseBody:      body,
		})
		assert.NoError(t, c.SetStackConfig(ctx, s, body))
	})

	t.Run("Delete sends DELETE", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/config", s.OrgName, s.ProjectName, s.StackName),
			ResponseCode:      http.StatusNoContent,
		})
		assert.NoError(t, c.DeleteStackConfig(ctx, s))
	})

	t.Run("Delete tolerates 404", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/config", s.OrgName, s.ProjectName, s.StackName),
			ResponseCode:      http.StatusNotFound,
		})
		assert.NoError(t, c.DeleteStackConfig(ctx, s))
	})
}

func TestParseEnvRef(t *testing.T) {
	cases := []struct {
		in          string
		wantProject string
		wantName    string
		wantVersion string
	}{
		{"default/prod-cfg", "default", "prod-cfg", ""},
		{"default/prod-cfg@3", "default", "prod-cfg", "3"},
		{"default/prod-cfg:prod", "default", "prod-cfg", "prod"},
		{"my-proj/env-name@stable", "my-proj", "env-name", "stable"},
		{"legacy-env", "default", "legacy-env", ""},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			p, n, v := ParseEnvRef(tc.in)
			assert.Equal(t, tc.wantProject, p)
			assert.Equal(t, tc.wantName, n)
			assert.Equal(t, tc.wantVersion, v)
		})
	}
}

func TestFormatEnvRef(t *testing.T) {
	assert.Equal(t, "default/prod-cfg", FormatEnvRef("default", "prod-cfg", ""))
	assert.Equal(t, "default/prod-cfg@3", FormatEnvRef("default", "prod-cfg", "3"))
	assert.Equal(t, "my-proj/env-name@stable", FormatEnvRef("my-proj", "env-name", "stable"))
}
