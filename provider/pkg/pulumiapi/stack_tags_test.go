package pulumiapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	projectKey      = "project"
	organizationKey = "organization"
)

func TestCreateStackTags(t *testing.T) {
	tagName := "tagName"
	tagValue := "tagValue"
	tag := StackTag{
		Name:  tagName,
		Value: tagValue,
	}
	stackName := StackIdentifier{
		OrgName:     organizationKey,
		ProjectName: projectKey,
		StackName:   stackKey,
	}
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath: fmt.Sprintf(
				"/api/stacks/%s/%s/%s/tags",
				stackName.OrgName,
				stackName.ProjectName,
				stackName.StackName,
			),
			ExpectedReqBody: tag,
			ResponseCode:    http.StatusNoContent,
		})
		assert.NoError(t, c.CreateStackTag(ctx, stackName, tag))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath: fmt.Sprintf(
				"/api/stacks/%s/%s/%s/tags",
				stackName.OrgName,
				stackName.ProjectName,
				stackName.StackName,
			),
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				Message: unauthorizedError,
			},
		})
		err := c.CreateStackTag(ctx, stackName, tag)
		assert.EqualError(t, err, "failed to create tag (tagName=tagValue): 401 API error: unauthorized")
	})
}

func TestDeleteStackTags(t *testing.T) {
	stackName := StackIdentifier{
		OrgName:     organizationKey,
		ProjectName: projectKey,
		StackName:   stackKey,
	}
	tagName := "tagName"
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/organization/project/stack/tags/tagName",
			ResponseCode:      204,
		})
		assert.NoError(t, c.DeleteStackTag(ctx, stackName, tagName))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/stacks/organization/project/stack/tags/tagName",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: unauthorizedError,
			},
		})
		assert.EqualError(
			t,
			c.DeleteStackTag(ctx, stackName, tagName),
			"failed to make request: 401 API error: unauthorized",
		)
	})
}
