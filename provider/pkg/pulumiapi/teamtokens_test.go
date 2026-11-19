package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var teamCtx = context.Background()

const (
	testTeamTokenTeamName = "aTeam"
	testTeamTokenOrgName  = "anOrg"
)

func TestDeleteTeamAccessToken(t *testing.T) {
	orgName := testTeamTokenOrgName
	teamName := testTeamTokenTeamName
	tokenID := "abcdegh"
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/teams/aTeam/tokens/" + tokenID,
			ResponseCode:      204,
		})
		assert.NoError(t, c.DeleteTeamAccessToken(teamCtx, tokenID, orgName, teamName))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/teams/aTeam/tokens/" + tokenID,
			ResponseCode:      500,
			ResponseBody: ErrorResponse{
				StatusCode: 500,
				Message:    tokenNotFoundError,
			},
		})
		assert.EqualError(t,
			c.DeleteTeamAccessToken(teamCtx, tokenID, orgName, teamName),
			`failed to delete access token "abcdegh": 500 API error: token not found`,
		)
	})

}

func TestCreateTeamAccessToken(t *testing.T) {
	orgName := testTeamTokenOrgName
	teamName := "aTeam"
	desc := "token description"
	tokenName := "aTeamToken"
	t.Run("Happy Path", func(t *testing.T) {
		resp := createTokenResponse{
			ID:         tokenIDKey,
			TokenValue: secretKey,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createTeamTokenRequest{
				Name:        tokenName,
				Description: desc,
			},
			ExpectedReqPath: "/api/orgs/anOrg/teams/aTeam/tokens",
			ResponseCode:    201,
			ResponseBody:    resp,
		})
		token, err := c.CreateTeamAccessToken(teamCtx, tokenName, orgName, teamName, desc, 0)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          resp.ID,
			TokenValue:  resp.TokenValue,
			Description: desc,
		}, token)
	})

	t.Run("With expiry", func(t *testing.T) {
		const apiPath = "/api/orgs/anOrg/teams/aTeam/tokens"
		const expires int64 = 1814374800
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createTeamTokenRequest{
				Name:        tokenName,
				Description: desc,
				Expires:     expires,
			},
			ExpectedReqPath: apiPath,
			ResponseCode:    201,
			ResponseBody:    createTokenResponse{ID: tokenIDKey, TokenValue: secretKey},
		})
		_, err := c.CreateTeamAccessToken(teamCtx, tokenName, orgName, teamName, desc, expires)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/anOrg/teams/aTeam/tokens",
			ExpectedReqBody: createTeamTokenRequest{
				Description: desc,
				Name:        tokenName,
			},
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    unauthorizedError,
			},
		})
		token, err := c.CreateTeamAccessToken(teamCtx, tokenName, orgName, teamName, desc, 0)
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
	org := testOrgName
	team := "aTeam"
	lastUsed := 123
	t.Run("Happy Path", func(t *testing.T) {
		resp := listTokenResponse{
			Tokens: []accessTokenResponse{
				{
					ID:          id,
					Description: desc,
					LastUsed:    lastUsed,
					Expires:     1814374800,
				},
				{
					ID:          otherValue,
					Description: desc,
					LastUsed:    lastUsed,
				},
			},
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqBody:   nil,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/teams/%s/tokens", org, team),
			ResponseCode:      200,
			ResponseBody:      resp,
		})
		token, err := c.GetTeamAccessToken(ctx, id, org, team)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          id,
			Description: desc,
			Expires:     1814374800,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/teams/%s/tokens", org, team),
			ExpectedReqBody:   nil,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    unauthorizedError,
			},
		})
		token, err := c.GetTeamAccessToken(ctx, id, org, team)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to list team access tokens: 401 API error: unauthorized`,
		)
	})
}
