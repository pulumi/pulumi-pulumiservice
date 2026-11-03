// Copyright 2026, Pulumi Corporation.

package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
)

const (
	assignmentsOrg  = "my-org"
	assignmentsList = "/api/orgs/my-org/assignments/list"
	assignmentsPut  = "/api/orgs/my-org/assignments/update"
)

func stackEntity(id string) apitype.EntityInstance {
	return apitype.EntityInstance{Context: apitype.PermissionExpressionStackBuilder{}.Build(), ID: id}
}

func TestListEntityAssignments(t *testing.T) {
	t.Run("sends the entity and returns the grants", func(t *testing.T) {
		stackSet := "stack-read"
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   assignmentsList,
			ExpectedReqBody:   stackEntity("stack-id"),
			ResponseCode:      http.StatusOK,
			ResponseBody: apitype.EntityAssignmentsResponse{
				Assignments: []apitype.EntityAssignment{{
					Principal:     apitype.AssignmentPrincipal{User: &apitype.UserInfo{GitHubLogin: "alice"}},
					PermissionSet: apitype.PermissionSet{ID: "set-id", DefaultIdentifier: &stackSet},
					Source:        apitype.AssignmentSourceAssignment,
				}},
				AvailablePermissionSets: []apitype.PermissionSet{{ID: "set-id", DefaultIdentifier: &stackSet}},
			},
		})
		resp, err := c.ListEntityAssignments(t.Context(), assignmentsOrg, stackEntity("stack-id"))
		require.NoError(t, err)
		require.Len(t, resp.Assignments, 1)
		assert.Equal(t, "alice", resp.Assignments[0].Principal.User.GitHubLogin)
	})

	t.Run("missing entity returns nil", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   assignmentsList,
			ResponseCode:      http.StatusNotFound,
		})
		resp, err := c.ListEntityAssignments(t.Context(), assignmentsOrg, stackEntity("gone"))
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("rejects empty arguments", func(t *testing.T) {
		c := &Client{}
		_, err := c.ListEntityAssignments(t.Context(), "", stackEntity("id"))
		assert.Error(t, err)
		_, err = c.ListEntityAssignments(t.Context(), assignmentsOrg, stackEntity(""))
		assert.Error(t, err)
	})
}

func TestUpdateEntityAssignment(t *testing.T) {
	t.Run("grant sends the permission set", func(t *testing.T) {
		req := apitype.UpdateEntityAssignmentRequest{
			EntityInstance:  stackEntity("stack-id"),
			UserID:          "alice",
			PermissionSetID: "set-id",
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   assignmentsPut,
			ExpectedReqBody:   req,
			ResponseCode:      http.StatusOK,
			ResponseBody:      apitype.UpdateEntityAssignmentResponse{},
		})
		_, err := c.UpdateEntityAssignment(t.Context(), assignmentsOrg, req)
		assert.NoError(t, err)
	})

	t.Run("revoke omits the permission set", func(t *testing.T) {
		req := apitype.UpdateEntityAssignmentRequest{EntityInstance: stackEntity("stack-id"), UserID: "alice"}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   assignmentsPut,
			ExpectedReqBody: map[string]any{
				"context": map[string]any{"__type": "PermissionExpressionStack"},
				"id":      "stack-id",
				"userId":  "alice",
			},
			ResponseCode: http.StatusOK,
			ResponseBody: apitype.UpdateEntityAssignmentResponse{},
		})
		_, err := c.UpdateEntityAssignment(t.Context(), assignmentsOrg, req)
		assert.NoError(t, err)
	})

	t.Run("API error carries the status code", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   assignmentsPut,
			ResponseCode:      http.StatusNotFound,
		})
		_, err := c.UpdateEntityAssignment(t.Context(), assignmentsOrg, apitype.UpdateEntityAssignmentRequest{
			EntityInstance: stackEntity("stack-id"), UserID: "alice",
		})
		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, GetErrorStatusCode(err))
	})

	t.Run("rejects empty arguments", func(t *testing.T) {
		c := &Client{}
		_, err := c.UpdateEntityAssignment(t.Context(), assignmentsOrg, apitype.UpdateEntityAssignmentRequest{
			EntityInstance: stackEntity("id"),
		})
		assert.Error(t, err)
	})
}
