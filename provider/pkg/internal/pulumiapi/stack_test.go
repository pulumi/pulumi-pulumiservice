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
	req := CreateStackRequest{
		StackName: s.StackName,
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s", s.OrgName, s.ProjectName),
			ExpectedReqBody:   req,
			ResponseCode:      http.StatusNoContent,
		})
		defer cleanup()
		assert.NoError(t, c.CreateStack(ctx, s))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s", s.OrgName, s.ProjectName),
			ResponseCode:      http.StatusUnauthorized,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		err := c.CreateStack(ctx, s)
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
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s", s.OrgName, s.ProjectName, s.StackName),
			ResponseCode:      http.StatusNoContent,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteStack(ctx, s, false))
	})

	t.Run("Force destroy", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod:   http.MethodDelete,
			ExpectedReqPath:     fmt.Sprintf("/api/stacks/%s/%s/%s", s.OrgName, s.ProjectName, s.StackName),
			ExpectedQueryParams: url.Values{"forceDestroy": {"true"}},
			ResponseCode:        http.StatusNoContent,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteStack(ctx, s, true))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/organization/project/stack",
			ResponseCode:      http.StatusUnauthorized,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		assert.EqualError(t, c.DeleteStack(ctx, s, false), "failed to delete stack: 401 API error: unauthorized")
	})
}
