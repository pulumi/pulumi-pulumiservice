package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddMemberToOrg(t *testing.T) {
	userName := "a-user"
	orgName := "an-organization"
	role := "admin"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ExpectedReqBody: addMemberToOrgReq{
				Role: role,
			},
			ResponseCode: 201,
		})
		defer cleanup()
		err := c.AddMemberToOrg(ctx, userName, orgName, role)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ExpectedReqBody: addMemberToOrgReq{
				Role: role,
			},
			ResponseCode: 401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		err := c.AddMemberToOrg(ctx, userName, orgName, role)
		assert.EqualError(t, err, "failed to add member to org: 401 API error: unauthorized")
	})
}

func TestListOrgMembers(t *testing.T) {
	userName := "a-user"
	orgName := "an-organization"
	role := "admin"
	members := Members{
		Members: []Member{
			{
				User: User{
					Name: userName,
				},
				Role: role,
			},
		},
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/members",
			ResponseCode:      200,
			ResponseBody:      members,
		})
		defer cleanup()
		actualMembers, err := c.ListOrgMembers(ctx, orgName)
		assert.NoError(t, err)
		assert.Equal(t, members, *actualMembers)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/members",
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		actualMembers, err := c.ListOrgMembers(ctx, orgName)
		assert.Nil(t, actualMembers, "members should be null since error was returned")
		assert.EqualError(t, err, "failed to list organization members: 401 API error: unauthorized")
	})
}

func TestDeleteMemberFromOrg(t *testing.T) {
	userName := "a-user"
	orgName := "an-organization"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ResponseCode:      201,
		})
		defer cleanup()
		err := c.DeleteMemberFromOrg(ctx, orgName, userName)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		err := c.DeleteMemberFromOrg(ctx, orgName, userName)
		assert.EqualError(t, err, "failed to delete member from org: 401 API error: unauthorized")
	})
}
