package pulumiapi

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
)

const (
	testMemberUserName = "a-user"
	testMemberRole     = adminRole
	testMemberOrgName  = testDeploymentSettingsOrgName

	rosterBackend  = "backend"
	rosterFrontend = "frontend"
	memberUserPath = "/api/orgs/an-organization/members/a-user"
	orgRolesPath   = "/api/orgs/an-organization/roles"
	membersPath    = "/api/orgs/an-organization/members"
	bobUser        = "bob"
	aliceUser      = "alice"
)

func TestAddMemberToOrg(t *testing.T) {
	userName := testMemberUserName
	orgName := testMemberOrgName
	role := testMemberRole
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   memberUserPath,
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
			ExpectedReqPath:   memberUserPath,
			ExpectedReqBody: addMemberToOrgReq{
				Role: role,
			},
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				Message: unauthorizedError,
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

	t.Run("Happy Path", func(t *testing.T) {
		// Both rosters return the same single member; merge dedups to one.
		body := Members{Members: []Member{{User: User{Name: userName}, Role: role}}}
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, membersPath, r.URL.Path)
			return 200, body
		})
		got, err := c.ListOrgMembers(ctx, orgName)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Len(t, got.Members, 1)
			assert.Equal(t, userName, got.Members[0].User.Name)
		}
	})

	t.Run("Backend error fails the call", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   membersPath,
			ResponseCode:      401,
			ResponseBody:      ErrorResponse{Message: unauthorizedError},
		})
		got, err := c.ListOrgMembers(ctx, orgName)
		assert.Nil(t, got, "members should be null since backend error was returned")
		assert.EqualError(t, err, "failed to list organization members: 401 API error: unauthorized")
	})

	// Frontend errors are non-fatal: a transient frontend issue should not
	// gate the more-inclusive backend roster.
	t.Run("Frontend error is non-fatal", func(t *testing.T) {
		backendBody := Members{Members: []Member{{User: User{Name: aliceUser}, Role: adminRole}}}
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			switch r.URL.Query().Get("type") {
			case rosterBackend:
				return 200, backendBody
			case rosterFrontend:
				return 500, ErrorResponse{Message: "frontend down"}
			default:
				t.Fatalf("unexpected type=%q", r.URL.Query().Get("type"))
				return 500, nil
			}
		})
		got, err := c.ListOrgMembers(ctx, orgName)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Len(t, got.Members, 1, "backend result must be returned despite frontend failure")
			assert.Equal(t, aliceUser, got.Members[0].User.Name)
		}
	})

	// The backend roster emits a synthetic "org-as-user" entry for some
	// orgs — username equals the org name, no real created timestamp,
	// no fgaRole. Filter it out so `getOrganizationMembers` only
	// surfaces real human members.
	t.Run("Filters synthetic org-as-user entry", func(t *testing.T) {
		backendBody := Members{Members: []Member{
			{User: User{Name: aliceUser}, Role: memberRole},
			{User: User{Name: orgName, GithubLogin: orgName}, Role: memberRole},
		}}
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			switch r.URL.Query().Get("type") {
			case rosterBackend:
				return 200, backendBody
			case rosterFrontend:
				return 200, Members{Members: []Member{}}
			}
			t.Fatalf("unexpected type=%q", r.URL.Query().Get("type"))
			return 500, nil
		})
		got, err := c.ListOrgMembers(ctx, orgName)
		assert.NoError(t, err)
		require.NotNil(t, got)
		assert.Len(t, got.Members, 1, "synthetic org-as-user entry must be dropped")
		assert.Equal(t, aliceUser, got.Members[0].User.Name)
	})

	// Members appearing in both rosters are deduped; backend's record wins.
	// Members unique to either roster are included.
	t.Run("Merges and dedups", func(t *testing.T) {
		backendBody := Members{Members: []Member{
			{User: User{Name: aliceUser}, Role: adminRole, KnownToPulumi: false},
			{User: User{Name: bobUser}, Role: memberRole},
		}}
		frontendBody := Members{Members: []Member{
			{User: User{Name: aliceUser}, Role: adminRole, KnownToPulumi: true}, // dup
			{User: User{Name: "carol"}, Role: memberRole},                       // unique to frontend
		}}
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			switch r.URL.Query().Get("type") {
			case rosterBackend:
				return 200, backendBody
			case rosterFrontend:
				return 200, frontendBody
			}
			t.Fatalf("unexpected type=%q", r.URL.Query().Get("type"))
			return 500, nil
		})
		got, err := c.ListOrgMembers(ctx, orgName)
		assert.NoError(t, err)
		require.NotNil(t, got)
		assert.Len(t, got.Members, 3, "alice + bob + carol after dedup")
		// Backend wins on conflict — alice's KnownToPulumi must be false.
		var alice *Member
		for i, m := range got.Members {
			if m.User.Name == aliceUser {
				alice = &got.Members[i]
			}
		}
		require.NotNil(t, alice)
		assert.False(t, alice.KnownToPulumi,
			"backend record must win on dedup conflict")
	})

	t.Run("Takes the fine-grained role from the roster that has it", func(t *testing.T) {
		maintainer := &FGARole{ID: "role-1", Name: "Maintainer"}
		backendBody := Members{Members: []Member{
			{User: User{Name: aliceUser}, Role: memberRole, KnownToPulumi: true, FGARole: &FGARole{}},
			{User: User{Name: bobUser}, Role: memberRole, FGARole: &FGARole{}},
		}}
		frontendBody := Members{Members: []Member{
			{User: User{Name: aliceUser}, Role: memberRole, FGARole: maintainer},
		}}
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			if r.URL.Query().Get("type") == rosterBackend {
				return 200, backendBody
			}
			return 200, frontendBody
		})
		got, err := c.ListOrgMembers(ctx, orgName)
		require.NoError(t, err)
		require.Len(t, got.Members, 2)

		alice, bob := got.Members[0], got.Members[1]
		assert.Equal(t, maintainer, alice.FGARole)
		assert.True(t, alice.KnownToPulumi, "other fields stay as the backend reported them")
		assert.False(t, bob.HasFGARole(), "no roster has a role for bob, so the legacy role applies")
		assert.Equal(t, memberRole, bob.Role)
	})
}

func TestDeleteMemberFromOrg(t *testing.T) {
	userName := testMemberUserName
	orgName := testMemberOrgName
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   memberUserPath,
			ResponseCode:      201,
		})
		err := c.DeleteMemberFromOrg(ctx, orgName, userName)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   memberUserPath,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: unauthorizedError,
			},
		})
		err := c.DeleteMemberFromOrg(ctx, orgName, userName)
		assert.EqualError(t, err, "failed to delete member from org: 401 API error: unauthorized")
	})
}

func TestUpdateOrgMemberRole(t *testing.T) {
	orgName := testMemberOrgName
	userName := testMemberUserName

	for _, role := range []string{memberRole, adminRole, "billing-manager"} {
		t.Run(role+" is sent by name so the service resolves it", func(t *testing.T) {
			c := startTestServer(t, testServerConfig{
				ExpectedReqMethod: http.MethodPatch,
				ExpectedReqPath:   memberUserPath,
				ExpectedReqBody:   updateMemberRoleReq{Role: role},
				ResponseCode:      204,
			})
			assert.NoError(t, c.UpdateOrgMemberRole(ctx, orgName, userName, role, nil))
		})
	}

	for _, role := range []string{"reader", "builder", "maintainer"} {
		t.Run(role+" is resolved to its role ID, since the endpoint rejects it as a name", func(t *testing.T) {
			roles := struct {
				Roles []apitype.PermissionDescriptorRecord `json:"roles"`
			}{Roles: []apitype.PermissionDescriptorRecord{
				{ID: "other-id", DefaultIdentifier: "admin"},
				{ID: role + "-id", DefaultIdentifier: role},
			}}
			var patchBody []byte
			c := startTestServerMulti(t, func(r *http.Request) (int, any) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == orgRolesPath:
					assert.Equal(t, "role", r.URL.Query().Get("uxPurpose"))
					return 200, roles
				case r.Method == http.MethodPatch && r.URL.Path == memberUserPath:
					patchBody, _ = io.ReadAll(r.Body)
					return 204, nil
				default:
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
					return 500, nil
				}
			})
			assert.NoError(t, c.UpdateOrgMemberRole(ctx, orgName, userName, role, nil))
			assert.JSONEq(t, `{"fgaRoleId":"`+role+`-id"}`, string(patchBody))
		})
	}

	t.Run("a role missing from the organization is an error", func(t *testing.T) {
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			assert.Equal(t, http.MethodGet, r.Method)
			return 200, map[string]any{"roles": []any{}}
		})
		err := c.UpdateOrgMemberRole(ctx, orgName, userName, "reader", nil)
		assert.ErrorContains(t, err, `role "reader" not found in organization`)
	})

	t.Run("empty roleID sends the built-in role", func(t *testing.T) {
		empty := ""
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   memberUserPath,
			ExpectedReqBody:   updateMemberRoleReq{Role: memberRole},
			ResponseCode:      204,
		})
		assert.NoError(t, c.UpdateOrgMemberRole(ctx, orgName, userName, memberRole, &empty))
	})

	t.Run("custom role", func(t *testing.T) {
		roleID := "role-123"
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   memberUserPath,
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
		assert.EqualError(t, err, "role must be one of: admin, member, billing-manager, reader, builder, maintainer")
	})
}

// Backend roster pagination must follow continuationToken across pages.
// Frontend roster has no pagination per the API contract; this test
// returns it as a single empty page to keep focus on backend behavior.
func TestListOrgMembersPagination(t *testing.T) {
	orgName := testMemberOrgName
	second := "tok2"

	backendCall := 0
	backendPages := []Members{
		{
			Members: []Member{
				{User: User{Name: aliceUser}, Role: memberRole},
				{User: User{Name: bobUser}, Role: memberRole},
			},
			ContinuationToken: &second,
		},
		{
			Members: []Member{
				{User: User{Name: "carol"}, Role: adminRole},
			},
		},
	}

	c := startTestServerMulti(t, func(r *http.Request) (int, any) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, membersPath, r.URL.Path)
		switch r.URL.Query().Get("type") {
		case rosterBackend:
			if backendCall == 0 {
				assert.Equal(t, "", r.URL.Query().Get("continuationToken"))
			} else {
				assert.Equal(t, "tok2", r.URL.Query().Get("continuationToken"))
			}
			body := backendPages[backendCall]
			backendCall++
			return 200, body
		case rosterFrontend:
			return 200, Members{Members: []Member{}}
		}
		t.Fatalf("unexpected type=%q", r.URL.Query().Get("type"))
		return 500, nil
	})

	got, err := c.ListOrgMembers(ctx, orgName)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Len(t, got.Members, 3)
		assert.Equal(t, aliceUser, got.Members[0].User.Name)
		assert.Equal(t, "carol", got.Members[2].User.Name)
	}
	assert.Equal(t, 2, backendCall, "backend pagination must traverse both pages")
}

func TestGetOrgMember(t *testing.T) {
	orgName := testMemberOrgName
	members := Members{Members: []Member{
		{User: User{Name: aliceUser, GithubLogin: aliceUser}, Role: adminRole},
		{User: User{Name: bobUser, GithubLogin: bobUser}, Role: memberRole},
	}}

	t.Run("found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   membersPath,
			ResponseCode:      200,
			ResponseBody:      members,
		})
		got, err := c.GetOrgMember(ctx, orgName, bobUser)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, memberRole, got.Role)
		}
	})

	t.Run("takes the fine-grained role from the frontend roster", func(t *testing.T) {
		maintainer := &FGARole{ID: "role-1", Name: "Maintainer"}
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			if r.URL.Query().Get("type") == rosterBackend {
				return 200, Members{Members: []Member{
					{User: User{Name: bobUser, GithubLogin: bobUser}, Role: memberRole, FGARole: &FGARole{}},
				}}
			}
			return 200, Members{Members: []Member{
				{User: User{Name: bobUser, GithubLogin: bobUser}, Role: memberRole, FGARole: maintainer},
			}}
		})
		got, err := c.GetOrgMember(ctx, orgName, bobUser)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, maintainer, got.FGARole)
	})

	t.Run(notFoundError, func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   membersPath,
			ResponseCode:      200,
			ResponseBody:      members,
		})
		got, err := c.GetOrgMember(ctx, orgName, "nobody")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestBackendListingEmptyFGARoleObjectIsNotAFineGrainedRole(t *testing.T) {
	var m Member
	require.NoError(t, json.Unmarshal([]byte(`{"role":"admin","fgaRole":{}}`), &m))
	assert.False(t, m.HasFGARole())

	assert.False(t, Member{}.HasFGARole())
	assert.True(t, Member{FGARole: &FGARole{ID: "id", Name: "Admin"}}.HasFGARole())
}
