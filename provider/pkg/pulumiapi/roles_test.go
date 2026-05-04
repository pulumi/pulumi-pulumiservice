package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/apitype"
)

const (
	testRoleOrgName = "an-organization"
	testRoleID      = "11111111-2222-3333-4444-555555555555"
)

// testRoleDetails is the wire-shape JSON the server speaks. Tests build a
// typed PermissionDescriptor from it via the generated unmarshaller — same
// path the resource layer uses.
const testRoleDetailsJSON = `{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`

func mustParseDetails(t *testing.T) apitype.PermissionDescriptor {
	t.Helper()
	var d apitype.PermissionDescriptor
	require.NoError(t, apitype.UnmarshalJSONPermissionDescriptor([]byte(testRoleDetailsJSON), &d))
	require.NotNil(t, d)
	return d
}

func TestCreateRole(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		details := mustParseDetails(t)
		resp := apitype.PermissionDescriptorRecord{
			PermissionDescriptorBase: apitype.PermissionDescriptorBase{
				Name:         "read-only",
				Description:  "read only access",
				ResourceType: "global",
				UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
			},
			ID:    testRoleID,
			OrgID: "org-id",
			Version: 1,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/roles",
			ExpectedReqBody: apitype.PermissionDescriptorBase{
				Name:         "read-only",
				Description:  "read only access",
				ResourceType: "global",
				UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
				Details:      details,
			},
			ResponseCode: 200,
			ResponseBody: resp,
		})

		got, err := c.CreateRole(ctx, testRoleOrgName, apitype.PermissionDescriptorBase{
			Name:         "read-only",
			Description:  "read only access",
			ResourceType: "global",
			UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
			Details:      details,
		})
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, testRoleID, got.ID)
			assert.Equal(t, "read-only", got.Name)
		}
	})

	t.Run("empty details rejected", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{ResponseCode: 200})
		_, err := c.CreateRole(ctx, testRoleOrgName, apitype.PermissionDescriptorBase{
			Name:         "r",
			ResourceType: "global",
			UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
		})
		assert.EqualError(t, err, "role permissions details must not be empty")
	})
}

func TestGetRole(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/roles/" + testRoleID,
			ResponseCode:      200,
			ResponseBody: apitype.PermissionDescriptorRecord{
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{Name: "read-only"},
				ID:                       testRoleID,
				Version:                  2,
			},
		})
		got, err := c.GetRole(ctx, testRoleOrgName, testRoleID)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, int32(2), got.Version)
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
	details := mustParseDetails(t)

	c := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPatch,
		ExpectedReqPath:   "/api/orgs/an-organization/roles/" + testRoleID,
		ExpectedReqBody:   updateRoleReqWire{Name: &name, Description: &desc, Details: details},
		ResponseCode:      200,
		ResponseBody: apitype.PermissionDescriptorRecord{
			PermissionDescriptorBase: apitype.PermissionDescriptorBase{Name: name, Description: desc},
			ID:                       testRoleID,
			Version:                  3,
		},
	})

	got, err := c.UpdateRole(ctx, testRoleOrgName, testRoleID, &name, &desc, details)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, int32(3), got.Version)
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

