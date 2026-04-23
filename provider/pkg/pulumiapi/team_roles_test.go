package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testTeamRolesOrg  = "an-organization"
	testTeamRolesTeam = "a-team"
)

func TestAssignRoleToTeam(t *testing.T) {
	t.Run("enables then puts", func(t *testing.T) {
		calls := 0
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			calls++
			switch calls {
			case 1:
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/orgs/an-organization/teams/a-team/enable-team-roles", r.URL.Path)
				return 200, nil
			case 2:
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/orgs/an-organization/teams/a-team/roles/"+testRoleID, r.URL.Path)
				return 204, nil
			}
			t.Fatalf("unexpected call %d", calls)
			return 0, nil
		})
		assert.NoError(t, c.AssignRoleToTeam(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID))
		assert.Equal(t, 2, calls)
	})

	t.Run("already enabled", func(t *testing.T) {
		calls := 0
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			calls++
			switch calls {
			case 1:
				return 409, ErrorResponse{Message: "already enabled"}
			case 2:
				assert.Equal(t, http.MethodPost, r.Method)
				return 204, nil
			}
			t.Fatalf("unexpected call %d", calls)
			return 0, nil
		})
		assert.NoError(t, c.AssignRoleToTeam(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID))
		assert.Equal(t, 2, calls)
	})
}

func TestRemoveRoleFromTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team/roles/" + testRoleID,
			ResponseCode:      204,
		})
		assert.NoError(t, c.RemoveRoleFromTeam(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID))
	})

	t.Run("not found swallowed", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team/roles/" + testRoleID,
			ResponseCode:      404,
			ResponseBody:      ErrorResponse{Message: "not found"},
		})
		assert.NoError(t, c.RemoveRoleFromTeam(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID))
	})
}

func TestListAndGetTeamRoles(t *testing.T) {
	roles := listTeamRolesResponse{Roles: []TeamRoleRef{
		{ID: testRoleID, Name: "devops"},
		{ID: "other", Name: "other-role"},
	}}

	t.Run("list", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team/roles",
			ResponseCode:      200,
			ResponseBody:      roles,
		})
		got, err := c.ListTeamRoles(ctx, testTeamRolesOrg, testTeamRolesTeam)
		assert.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("get found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team/roles",
			ResponseCode:      200,
			ResponseBody:      roles,
		})
		got, err := c.GetTeamRole(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, "devops", got.Name)
		}
	})

	t.Run("get missing", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team/roles",
			ResponseCode:      200,
			ResponseBody:      listTeamRolesResponse{Roles: []TeamRoleRef{}},
		})
		got, err := c.GetTeamRole(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}
