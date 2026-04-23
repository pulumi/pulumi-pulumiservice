package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testMemberUserName = "a-user"
	testMemberRole     = "admin"
	testMemberOrgName  = "an-organization"
)

func TestAddMemberToOrg(t *testing.T) {
	userName := testMemberUserName
	orgName := testMemberOrgName
	role := testMemberRole
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ExpectedReqBody: addMemberToOrgReq{
				Role: role,
			},
			ResponseCode: 201,
		})
		err := c.AddMemberToOrg(ctx, userName, orgName, role)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ExpectedReqBody: addMemberToOrgReq{
				Role: role,
			},
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.AddMemberToOrg(ctx, userName, orgName, role)
		assert.EqualError(t, err, "failed to add member to org: 401 API error: unauthorized")
	})
}

func TestListOrgMembers(t *testing.T) {
	userName := testMemberUserName
	orgName := testMemberOrgName
	role := testMemberRole
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
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/members",
			ResponseCode:      200,
			ResponseBody:      members,
		})
		actualMembers, err := c.ListOrgMembers(ctx, orgName)
		assert.NoError(t, err)
		assert.Equal(t, members, *actualMembers)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/members",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		actualMembers, err := c.ListOrgMembers(ctx, orgName)
		assert.Nil(t, actualMembers, "members should be null since error was returned")
		assert.EqualError(t, err, "failed to list organization members: 401 API error: unauthorized")
	})
}

func TestDeleteMemberFromOrg(t *testing.T) {
	userName := testMemberUserName
	orgName := testMemberOrgName
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ResponseCode:      201,
		})
		err := c.DeleteMemberFromOrg(ctx, orgName, userName)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.DeleteMemberFromOrg(ctx, orgName, userName)
		assert.EqualError(t, err, "failed to delete member from org: 401 API error: unauthorized")
	})
}

func TestUpdateOrgMemberRole(t *testing.T) {
	orgName := testMemberOrgName
	userName := testMemberUserName

	t.Run("built-in role", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ExpectedReqBody:   updateMemberRoleReq{Role: "admin"},
			ResponseCode:      204,
		})
		assert.NoError(t, c.UpdateOrgMemberRole(ctx, orgName, userName, "admin", nil))
	})

	t.Run("custom role", func(t *testing.T) {
		roleID := "role-123"
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/members/a-user",
			ExpectedReqBody:   updateMemberRoleReq{FGARoleID: &roleID},
			ResponseCode:      204,
		})
		assert.NoError(t, c.UpdateOrgMemberRole(ctx, orgName, userName, "", &roleID))
	})

	t.Run("neither role nor roleID", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{ResponseCode: 200})
		err := c.UpdateOrgMemberRole(ctx, orgName, userName, "", nil)
		assert.EqualError(t, err, "one of role or fgaRoleID must be set")
	})

	t.Run("invalid built-in", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{ResponseCode: 200})
		err := c.UpdateOrgMemberRole(ctx, orgName, userName, "superadmin", nil)
		assert.EqualError(t, err, "role must be one of: admin, member, billing-manager")
	})
}

func TestListOrgMembersPagination(t *testing.T) {
	orgName := testMemberOrgName
	second := "tok2"

	call := 0
	pages := []Members{
		{
			Members: []Member{
				{User: User{Name: "alice"}, Role: "member"},
				{User: User{Name: "bob"}, Role: "member"},
			},
			ContinuationToken: &second,
		},
		{
			Members: []Member{
				{User: User{Name: "carol"}, Role: "admin"},
			},
		},
	}

	c := startTestServerMulti(t, func(r *http.Request) (int, any) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/orgs/an-organization/members", r.URL.Path)
		assert.Equal(t, "backend", r.URL.Query().Get("type"))
		if call == 0 {
			assert.Equal(t, "", r.URL.Query().Get("continuationToken"))
		} else {
			assert.Equal(t, "tok2", r.URL.Query().Get("continuationToken"))
		}
		body := pages[call]
		call++
		return 200, body
	})

	got, err := c.ListOrgMembers(ctx, orgName)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Len(t, got.Members, 3)
		assert.Equal(t, "alice", got.Members[0].User.Name)
		assert.Equal(t, "carol", got.Members[2].User.Name)
	}
	assert.Equal(t, 2, call)
}

func TestGetOrgMember(t *testing.T) {
	orgName := testMemberOrgName
	members := Members{Members: []Member{
		{User: User{Name: "alice", GithubLogin: "alice"}, Role: "admin"},
		{User: User{Name: "bob", GithubLogin: "bob"}, Role: "member"},
	}}

	t.Run("found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/members",
			ResponseCode:      200,
			ResponseBody:      members,
		})
		got, err := c.GetOrgMember(ctx, orgName, "bob")
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, "member", got.Role)
		}
	})

	t.Run("not found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/members",
			ResponseCode:      200,
			ResponseBody:      members,
		})
		got, err := c.GetOrgMember(ctx, orgName, "nobody")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}
