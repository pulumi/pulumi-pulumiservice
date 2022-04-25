package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTeams(t *testing.T) {
	orgName := "an-organization"
	member1 := TeamMember{
		Name: "member1",
	}
	member2 := TeamMember{
		Name: "member2",
	}
	team1 := Team{
		Type:        "pulumi",
		Name:        "team1",
		DisplayName: "Team 1",
		Description: "Team 1 description",
		Members:     []TeamMember{member1, member2},
	}
	teams := Teams{
		Teams: []Team{team1},
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams",
			ResponseCode:      200,
			ResponseBody:      teams,
		})
		defer cleanup()
		teamsList, err := c.ListTeams(ctx, orgName)
		assert.NoError(t, err)
		assert.Equal(t, teams.Teams, teamsList)
	})
	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams",
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		teamsList, err := c.ListTeams(ctx, orgName)
		assert.Nil(t, teamsList, "if list teams has error, no teams object should be returned")
		assert.EqualError(t, err, `failed to list teams for "an-organization": 401 API error: unauthorized`)
	})
}

func TestGetTeam(t *testing.T) {
	orgName := "an-organization"
	teamName := "a-team"
	team := Team{
		Type:        "pulumi",
		Name:        teamName,
		DisplayName: "Team 1",
		Description: "Team 1 description",
		Members: []TeamMember{
			{
				Name: "alice",
			},
		},
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ResponseCode:      200,
			ResponseBody:      team,
		})
		defer cleanup()
		actualTeam, err := c.GetTeam(ctx, orgName, teamName)
		assert.NoError(t, err)
		assert.Equal(t, &team, actualTeam)
	})
	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ResponseCode:      401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		team, err := c.GetTeam(ctx, orgName, teamName)
		assert.Nil(t, team, "team should be nil since error was returned")
		assert.EqualError(t, err, "failed to get team: 401 API error: unauthorized")
	})
}

func TestCreateTeam(t *testing.T) {
	orgName := "an-organization"
	teamName := "a-team"
	teamType := "pulumi"
	displayName := "A Team"
	description := "The A Team"
	team := Team{
		Type:        teamType,
		Name:        teamName,
		DisplayName: displayName,
		Description: description,
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/pulumi",
			ExpectedReqBody: createTeamRequest{
				Organization: orgName,
				TeamType:     teamType,
				Name:         teamName,
				DisplayName:  displayName,
				Description:  description,
			},
			ResponseBody: team,
			ResponseCode: 201,
		})
		defer cleanup()
		actualTeam, err := c.CreateTeam(ctx, orgName, teamName, teamType, displayName, description)
		assert.NoError(t, err)
		assert.Equal(t, team, *actualTeam)
	})
	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/pulumi",
			ExpectedReqBody: createTeamRequest{
				Organization: orgName,
				TeamType:     teamType,
				Name:         teamName,
				DisplayName:  displayName,
				Description:  description,
			},
			ResponseCode: 401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		team, err := c.CreateTeam(ctx, orgName, teamName, teamType, displayName, description)
		assert.Nil(t, team, "team should be nil since error was returned")
		assert.EqualError(t, err, "failed to create team: 401 API error: unauthorized")
	})
}

func TestAddMemberToTeam(t *testing.T) {
	orgName := "an-organization"
	teamName := "a-team"
	userName := "a-user"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: updateTeamMembershipRequest{
				MemberAction: "add",
				Member:       userName,
			},
			ResponseCode: 204,
		})
		defer cleanup()
		assert.NoError(t, c.AddMemberToTeam(ctx, orgName, teamName, userName))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: updateTeamMembershipRequest{
				MemberAction: "add",
				Member:       userName,
			},
			ResponseCode: 401,
			ResponseBody: errorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		assert.EqualError(t,
			c.AddMemberToTeam(ctx, orgName, teamName, userName),
			"failed to update team membership: 401 API error: unauthorized",
		)
	})
}
