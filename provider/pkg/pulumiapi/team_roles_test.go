package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testTeamRolesOrg  = testDeploymentSettingsOrgName
	testTeamRolesTeam = "a-team"
)

func TestAssignRoleToTeam(t *testing.T) {
	c := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPost,
		ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team/roles/" + testRoleID,
		ResponseCode:      204,
	})
	assert.NoError(t, c.AssignRoleToTeam(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID))
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
			ResponseBody:      ErrorResponse{Message: notFoundError},
		})
		assert.NoError(t, c.RemoveRoleFromTeam(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID))
	})
}

func TestListAndGetTeamRoles(t *testing.T) {
	roles := listTeamRolesResponse{Roles: []TeamRoleRef{
		{ID: testRoleID, Name: "devops"},
		{ID: otherValue, Name: "other-role"},
	}}

	t.Run("list", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   teamRolesPath,
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
			ExpectedReqPath:   teamRolesPath,
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
			ExpectedReqPath:   teamRolesPath,
			ResponseCode:      200,
			ResponseBody:      listTeamRolesResponse{Roles: []TeamRoleRef{}},
		})
		got, err := c.GetTeamRole(ctx, testTeamRolesOrg, testTeamRolesTeam, testRoleID)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}
