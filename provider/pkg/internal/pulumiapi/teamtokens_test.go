package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var teamCtx = context.Background()

func TestDeleteTeamAccessToken(t *testing.T) {
	orgName := "anOrg"
	teamName := "aTeam"
	tokenId := "abcdegh"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/teams/aTeam/tokens/" + tokenId,
			ResponseCode:      204,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteTeamAccessToken(teamCtx, tokenId, orgName, teamName))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/teams/aTeam/tokens/" + tokenId,
			ResponseCode:      404,
			ResponseBody: errorResponse{
				StatusCode: 404,
				Message:    "token not found",
			},
		})
		defer cleanup()
		assert.EqualError(t,
			c.DeleteTeamAccessToken(teamCtx, tokenId, orgName, teamName),
			`failed to delete access token "abcdegh": 404 API error: token not found`,
		)
	})

}

func TestCreateTeamAccessToken(t *testing.T) {
	orgName := "anOrg"
	teamName := "aTeam"
	desc := "token description"
	tokenName := "aTeamToken"
	t.Run("Happy Path", func(t *testing.T) {
		resp := createTokenResponse{
			ID:         "token_id",
			TokenValue: "secret",
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createTeamTokenRequest{
				Name:        tokenName,
				Description: desc,
			},
			ExpectedReqPath: "/api/orgs/anOrg/teams/aTeam/tokens",
			ResponseCode:    201,
			ResponseBody:    resp,
		})
		defer cleanup()
		token, err := c.CreateTeamAccessToken(teamCtx, tokenName, orgName, teamName, desc)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          resp.ID,
			TokenValue:  resp.TokenValue,
			Description: desc,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/anOrg/teams/aTeam/tokens",
			ExpectedReqBody: createTeamTokenRequest{
				Description: desc,
				Name:        tokenName,
			},
			ResponseCode: 401,
			ResponseBody: errorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		defer cleanup()
		token, err := c.CreateTeamAccessToken(teamCtx, tokenName, orgName, teamName, desc)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to create access token: 401 API error: unauthorized`,
		)
	})
}

func TestGetTeamAccessToken(t *testing.T) {
	id := "uuid"
	desc := "token description"
	org := "anOrg"
	team := "aTeam"
	lastUsed := 123
	t.Run("Happy Path", func(t *testing.T) {
		resp := listTokenResponse{
			Tokens: []accessTokenResponse{
				{
					ID:          id,
					Description: desc,
					LastUsed:    lastUsed,
				},
				{
					ID:          "other",
					Description: desc,
					LastUsed:    lastUsed,
				},
			},
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqBody:   nil,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/teams/%s/tokens", org, team),
			ResponseCode:      200,
			ResponseBody:      resp,
		})
		defer cleanup()
		token, err := c.GetTeamAccessToken(ctx, org, team, id)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          id,
			Description: desc,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/teams/%s/tokens", org, team),
			ExpectedReqBody:   nil,
			ResponseCode:      401,
			ResponseBody: errorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		defer cleanup()
		token, err := c.GetTeamAccessToken(ctx, org, team, id)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to list team access tokens: 401 API error: unauthorized`,
		)
	})
}
