package pulumiapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

<<<<<<< HEAD
<<<<<<< HEAD
func TestCreateStackTags(t *testing.T) {
=======
func TestSetStackTags(t *testing.T) {
>>>>>>> d06708e (Add tests for api client library)
=======
func TestCreateStackTags(t *testing.T) {
>>>>>>> 932b63e (rename SetStackTags to CreateStackTag to better match model)
	tagName := "tagName"
	tagValue := "tagValue"
	tag := StackTag{
		Name:  tagName,
		Value: tagValue,
	}
<<<<<<< HEAD
<<<<<<< HEAD
	stackName := StackName{
=======
	stack := StackName{
>>>>>>> d06708e (Add tests for api client library)
=======
	stackName := StackName{
>>>>>>> 932b63e (rename SetStackTags to CreateStackTag to better match model)
		OrgName:     "organization",
		ProjectName: "project",
		StackName:   "stack",
	}
<<<<<<< HEAD
<<<<<<< HEAD
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stackName.OrgName, stackName.ProjectName, stackName.StackName),
=======
	tagMap := map[string]string{
		tagName: tagValue,
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stack.OrgName, stack.ProjectName, stack.StackName),
>>>>>>> d06708e (Add tests for api client library)
=======
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stackName.OrgName, stackName.ProjectName, stackName.StackName),
>>>>>>> 932b63e (rename SetStackTags to CreateStackTag to better match model)
			ExpectedReqBody:   tag,
			ResponseCode:      http.StatusNoContent,
		})
		defer cleanup()
<<<<<<< HEAD
<<<<<<< HEAD
		assert.NoError(t, c.CreateTag(ctx, stackName, tag))
=======
		assert.NoError(t, c.SetTags(ctx, stack, tagMap))
>>>>>>> d06708e (Add tests for api client library)
=======
		assert.NoError(t, c.CreateTag(ctx, stackName, tag))
>>>>>>> 932b63e (rename SetStackTags to CreateStackTag to better match model)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
<<<<<<< HEAD
<<<<<<< HEAD
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stackName.OrgName, stackName.ProjectName, stackName.StackName),
=======
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stack.OrgName, stack.ProjectName, stack.StackName),
>>>>>>> d06708e (Add tests for api client library)
=======
			ExpectedReqPath:   fmt.Sprintf("/api/stacks/%s/%s/%s/tags", stackName.OrgName, stackName.ProjectName, stackName.StackName),
>>>>>>> 932b63e (rename SetStackTags to CreateStackTag to better match model)
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
<<<<<<< HEAD
<<<<<<< HEAD
		err := c.CreateTag(ctx, stackName, tag)
=======
		err := c.SetTags(ctx, stack, tagMap)
>>>>>>> d06708e (Add tests for api client library)
=======
		err := c.CreateTag(ctx, stackName, tag)
>>>>>>> 932b63e (rename SetStackTags to CreateStackTag to better match model)
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
