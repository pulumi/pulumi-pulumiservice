package pulumiapi

import (
	"net/http"
	"testing"
	"time"

	"github.com/pgavlin/fx/v2"
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
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams",
			ResponseCode:      200,
			ResponseBody:      teams,
		})
		teamsList, err := c.ListTeams(ctx, orgName)
		assert.NoError(t, err)
		assert.Equal(t, teams.Teams, teamsList)
	})
	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
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
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ResponseCode:      200,
			ResponseBody:      team,
		})
		actualTeam, err := c.GetTeam(ctx, orgName, teamName)
		assert.NoError(t, err)
		assert.Equal(t, &team, actualTeam)
	})
	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		team, err := c.GetTeam(ctx, orgName, teamName)
		assert.Nil(t, team, "team should be nil since error was returned")
		assert.EqualError(t, err, "failed to get team: 401 API error: unauthorized")
	})

	t.Run("404", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})
		team, err := c.GetTeam(ctx, orgName, teamName)
		assert.Nil(t, team, "team should be nil since 404 was returned")
		assert.Nil(t, err, "err should be nil since 404 was returned")
	})
}

func TestCreateTeam(t *testing.T) {
	orgName := "an-organization"
	teamName := "a-team"
	displayName := "A Team"
	description := "The A Team"
	t.Run("Happy Path (pulumi team)", func(t *testing.T) {
		expected := Team{
			Type:        "pulumi",
			Name:        teamName,
			DisplayName: displayName,
			Description: description,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/pulumi",
			ExpectedReqBody: createTeamRequest{
				Organization: orgName,
				TeamType:     "pulumi",
				Name:         teamName,
				DisplayName:  displayName,
				Description:  description,
			},
			ResponseBody: expected,
			ResponseCode: 201,
		})
		actualTeam, err := c.CreateTeam(ctx, orgName, teamName, "pulumi", displayName, description, 0)
		assert.NoError(t, err)
		assert.Equal(t, expected, *actualTeam)
	})
	t.Run("Happy Path (github team)", func(t *testing.T) {
		expected := Team{
			Type:        "github",
			Name:        teamName,
			DisplayName: displayName,
			Description: description,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/github",
			ExpectedReqBody: createTeamRequest{
				TeamType:     "github",
				Organization: orgName,
				GitHubTeamID: 1,
			},
			ResponseBody: expected,
			ResponseCode: 201,
		})
		actualTeam, err := c.CreateTeam(ctx, orgName, "", "github", "", "", 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, *actualTeam)
	})
	t.Run("Error (pulumi team)", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/pulumi",
			ExpectedReqBody: createTeamRequest{
				Organization: orgName,
				TeamType:     "pulumi",
				Name:         teamName,
				DisplayName:  displayName,
				Description:  description,
			},
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		team, err := c.CreateTeam(ctx, orgName, teamName, "pulumi", displayName, description, 0)
		assert.Nil(t, team, "team should be nil since error was returned")
		assert.EqualError(t, err, "failed to create team: 401 API error: unauthorized")
	})
	t.Run("Error (github team missing ID)", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{})
		_, err := c.CreateTeam(ctx, orgName, "", "github", "", "", 0)
		assert.EqualError(t, err, "github teams require a githubTeamId")
	})
	t.Run("Error (invalid team type)", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{})
		_, err := c.CreateTeam(ctx, orgName, "", "foo", "", "", 0)
		assert.EqualError(t, err, "teamtype must be one of [github pulumi], got \"foo\"")
	})
	t.Run("Error (invalid team name for pulumi team)", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{})
		_, err := c.CreateTeam(ctx, orgName, "", "pulumi", "", "", 0)
		assert.EqualError(t, err, "teamname must not be empty")
	})
	t.Run("Error (missing org name)", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{})
		_, err := c.CreateTeam(ctx, "", "", "pulumi", "", "", 0)
		assert.EqualError(t, err, "orgname must not be empty")
	})
}

func TestAddMemberToTeam(t *testing.T) {
	orgName := "an-organization"
	teamName := "a-team"
	userName := "a-user"
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: updateTeamMembershipRequest{
				MemberAction: "add",
				Member:       userName,
			},
			ResponseCode: 204,
		})
		assert.NoError(t, c.AddMemberToTeam(ctx, orgName, teamName, userName))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: updateTeamMembershipRequest{
				MemberAction: "add",
				Member:       userName,
			},
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		assert.EqualError(t,
			c.AddMemberToTeam(ctx, orgName, teamName, userName),
			"failed to update team membership: 401 API error: unauthorized",
		)
	})
}

func TestAddStackPermission(t *testing.T) {
	teamName := "a-team"
	stack := StackIdentifier{
		OrgName:     "an-organization",
		ProjectName: "a-project",
		StackName:   "a-stack",
	}
	permission := 101
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: addStackPermissionRequest{
				AddStackPermission: AddStackPermission{
					ProjectName: stack.ProjectName,
					StackName:   stack.StackName,
					Permission:  permission,
				},
			},
			ResponseCode: 204,
		})
		assert.NoError(t, c.AddStackPermission(ctx, stack, teamName, permission))
	})
}

func TestRemoveStackPermission(t *testing.T) {
	teamName := "a-team"
	stack := StackIdentifier{
		OrgName:     "an-organization",
		ProjectName: "a-project",
		StackName:   "a-stack",
	}
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: removeStackPermissionRequest{
				RemoveStackPermission: RemoveStackPermission{
					ProjectName: stack.ProjectName,
					StackName:   stack.StackName,
				},
			},
			ResponseCode: 204,
		})
		assert.NoError(t, c.RemoveStackPermission(ctx, stack, teamName))
	})
}

func TestAddEnvironmentPermission(t *testing.T) {
	teamName := "a-team"
	permission := "admin"
	organization := "an-organization"
	project := "a-project"
	environment := "an-environment"
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: addEnvironmentSettingsRequest{
				AddEnvironmentPermission: AddEnvironmentPermission{
					ProjectName:     project,
					EnvName:         environment,
					Permission:      permission,
					MaxOpenDuration: fx.Some(Duration(15 * time.Minute)),
				},
			},
			ResponseCode: 204,
		})
		assert.NoError(t, c.AddEnvironmentSettings(ctx, CreateTeamEnvironmentSettingsRequest{
			TeamEnvironmentSettingsRequest: TeamEnvironmentSettingsRequest{
				Organization: organization,
				Project:      project,
				Environment:  environment,
				Team:         teamName,
			},
			Permission:      permission,
			MaxOpenDuration: fx.Some(Duration(15 * time.Minute)),
		}))
	})
}

func TestRemoveEnvironmentPermission(t *testing.T) {
	teamName := "a-team"
	organization := "an-organization"
	project := "a-project"
	environment := "an-environment"
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/teams/a-team",
			ExpectedReqBody: removeEnvironmentPermissionRequest{
				RemoveEnvironment: RemoveEnvironmentPermission{
					ProjectName: project,
					EnvName:     environment,
				},
			},
			ResponseCode: 204,
		})
		assert.NoError(t, c.RemoveEnvironmentSettings(ctx, TeamEnvironmentSettingsRequest{
			Organization: organization,
			Team:         teamName,
			Project:      project,
			Environment:  environment,
		}))
	})
}
