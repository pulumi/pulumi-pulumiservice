package pulumiapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetStackTags(t *testing.T) {
	tagName := "tagName"
	tagValue := "tagValue"
	tag := StackTag{
		Name:  tagName,
		Value: tagValue,
	}
	stack := StackName{
		OrgName:     "organization",
		ProjectName: "project",
		StackName:   "stack",
	}
	tagMap := map[string]string{
		tagName: tagValue,
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stack.OrgName, stack.ProjectName, stack.StackName),
			ExpectedReqBody:   tag,
			ResponseCode:      http.StatusNoContent,
		})
		defer cleanup()
		assert.NoError(t, c.SetTags(ctx, stack, tagMap))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stack.OrgName, stack.ProjectName, stack.StackName),
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		err := c.SetTags(ctx, stack, tagMap)
		assert.EqualError(t, err, "failed to create tag (tagName=tagValue): 401 API error: unauthorized")
	})
}

func TestDeleteStackTags(t *testing.T) {
	stackName := StackName{
		OrgName:     "organization",
		ProjectName: "project",
		StackName:   "stack",
	}
	tagName := "tagName"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/organization/project/stack/tags/tagName",
			ResponseCode:      204,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteStackTag(ctx, stackName, tagName))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/organization/project/stack/tags/tagName",
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		assert.EqualError(t, c.DeleteStackTag(ctx, stackName, tagName), "failed to make request: 401 API error: unauthorized")
	})
}
