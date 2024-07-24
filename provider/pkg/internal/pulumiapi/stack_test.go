package pulumiapi

import (
	"fmt"
	"net/http"
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
			ResponseCode:      401,
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
			ExpectedReqPath:   "/api/stacks/organization/project/stack",
			ResponseCode:      204,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteStack(ctx, s, false))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/organization/project/stack",
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		assert.EqualError(t, c.DeleteStack(ctx, s, false), "failed to delete stack: 401 API error: unauthorized")
	})
}
