package pulumiapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testRoleOrgName = "an-organization"
	testRoleID      = "11111111-2222-3333-4444-555555555555"
)

var testRoleDetails = json.RawMessage(`{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`)

func TestCreateRole(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		resp := RoleDescriptor{
			ID:           testRoleID,
			Name:         "read-only",
			Description:  "read only access",
			ResourceType: "organization",
			UXPurpose:    "role",
			OrgID:        "org-id",
			Version:      1,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/roles",
			ExpectedReqBody: CreateRoleRequest{
				Name:         "read-only",
				Description:  "read only access",
				ResourceType: "organization",
				UXPurpose:    "role",
				Details:      testRoleDetails,
			},
			ResponseCode: 200,
			ResponseBody: resp,
		})

		got, err := c.CreateRole(ctx, testRoleOrgName, NewCreateRoleRequest(
			"read-only", "read only access", "organization", "role", testRoleDetails,
		))
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, testRoleID, got.ID)
			assert.Equal(t, "read-only", got.Name)
		}
	})

	t.Run("defaults resourceType and uxPurpose", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/roles",
			ExpectedReqBody: CreateRoleRequest{
				Name:         "r",
				ResourceType: "organization",
				UXPurpose:    "role",
				Details:      testRoleDetails,
			},
			ResponseCode: 200,
			ResponseBody: RoleDescriptor{ID: testRoleID, Name: "r"},
		})
		_, err := c.CreateRole(ctx, testRoleOrgName, NewCreateRoleRequest("r", "", "", "", testRoleDetails))
		assert.NoError(t, err)
	})

	t.Run("empty details rejected", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{ResponseCode: 200})
		_, err := c.CreateRole(ctx, testRoleOrgName, NewCreateRoleRequest("r", "", "", "", nil))
		assert.EqualError(t, err, "role permissions details should not be empty")
	})
}

func TestGetRole(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/roles/" + testRoleID,
			ResponseCode:      200,
			ResponseBody: RoleDescriptor{
				ID:      testRoleID,
				Name:    "read-only",
				Version: 2,
			},
		})
		got, err := c.GetRole(ctx, testRoleOrgName, testRoleID)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, 2, got.Version)
		}
	})

	t.Run("not found returns nil", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/roles/" + testRoleID,
			ResponseCode:      404,
			ResponseBody:      ErrorResponse{Message: "not found"},
		})
		got, err := c.GetRole(ctx, testRoleOrgName, testRoleID)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestUpdateRole(t *testing.T) {
	name := "new-name"
	desc := "new description"

	c := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPatch,
		ExpectedReqPath:   "/api/orgs/an-organization/roles/" + testRoleID,
		ExpectedReqBody:   updateRoleReq{Name: &name, Description: &desc, Details: testRoleDetails},
		ResponseCode:      200,
		ResponseBody:      RoleDescriptor{ID: testRoleID, Name: name, Description: desc, Version: 3},
	})

	got, err := c.UpdateRole(ctx, testRoleOrgName, testRoleID, &name, &desc, testRoleDetails)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, 3, got.Version)
	}
}

func TestListAvailableRoleScopes(t *testing.T) {
	t.Run("flattens wire shape", func(t *testing.T) {
		// Service returns map[string][]RbacScopeGroup where each group holds
		// rbacScope{name, metadata:{description}}. We exercise one bucket with
		// one group with two scopes.
		wire := map[string][]rbacScopeGroup{
			"stack": {
				{
					Name: "Stacks",
					Scopes: []rbacScope{
						{Name: "stack:read", Metadata: rbacScopeMetadata{Description: "Read stacks"}},
						{Name: "stack:update", Metadata: rbacScopeMetadata{Description: "Update stacks"}},
					},
				},
			},
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/roles/scopes",
			ResponseCode:      200,
			ResponseBody:      wire,
		})
		got, err := c.ListAvailableRoleScopes(ctx, testRoleOrgName)
		assert.NoError(t, err)
		if assert.Contains(t, got, "stack") && assert.Len(t, got["stack"], 1) {
			group := got["stack"][0]
			assert.Equal(t, "Stacks", group.Name)
			if assert.Len(t, group.Scopes, 2) {
				assert.Equal(t, "stack:read", group.Scopes[0].Name)
				assert.Equal(t, "Read stacks", group.Scopes[0].Description)
			}
		}
	})

	t.Run("unauthorised", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/roles/scopes",
			ResponseCode:      401,
			ResponseBody:      ErrorResponse{Message: "unauthorized"},
		})
		_, err := c.ListAvailableRoleScopes(ctx, testRoleOrgName)
		assert.ErrorContains(t, err, "failed to list available role scopes")
	})
}

func TestDeleteRole(t *testing.T) {
	t.Run("force flag sent", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod:   http.MethodDelete,
			ExpectedReqPath:     "/api/orgs/an-organization/roles/" + testRoleID,
			ExpectedQueryParams: map[string][]string{"force": {"true"}},
			ResponseCode:        204,
		})
		assert.NoError(t, c.DeleteRole(ctx, testRoleOrgName, testRoleID, true))
	})

	t.Run("not found swallowed", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/roles/" + testRoleID,
			ResponseCode:      404,
			ResponseBody:      ErrorResponse{Message: "not found"},
		})
		assert.NoError(t, c.DeleteRole(ctx, testRoleOrgName, testRoleID, false))
	})
}
