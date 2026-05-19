package pulumiapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testOrgTokenOrgName     = "anOrg"
	testOrgTokenID          = "abcdegh"
	testOrgTokenDescription = "token description"
)

func TestDeleteOrgAccessToken(t *testing.T) {
	orgName := testOrgTokenOrgName
	tokenID := testOrgTokenID
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/tokens/" + tokenID,
			ResponseCode:      204,
		})
		assert.NoError(t, c.DeleteOrgAccessToken(teamCtx, tokenID, orgName))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/tokens/" + tokenID,
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "token not found",
			},
		})
		assert.EqualError(t,
			c.DeleteOrgAccessToken(teamCtx, tokenID, orgName),
			fmt.Sprintf(`failed to delete access token "%s": 404 API error: token not found`, testOrgTokenID),
		)
	})

}

func TestCreateOrgAccessToken(t *testing.T) {
	orgName := testOrgTokenOrgName
	name := "anOrgToken"
	desc := testOrgTokenDescription

	t.Run("Happy Path", func(t *testing.T) {
		resp := createTokenResponse{
			ID:         "token_id",
			TokenValue: "secret",
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createOrgTokenRequest{
				Description: desc,
				Name:        name,
			},
			ExpectedReqPath: "/api/orgs/anOrg/tokens",
			ResponseCode:    201,
			ResponseBody:    resp,
		})
		token, err := c.CreateOrgAccessToken(teamCtx, name, orgName, desc, false)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          resp.ID,
			TokenValue:  resp.TokenValue,
			Description: desc,
		}, token)
	})

	t.Run("Admin token", func(t *testing.T) {
		resp := createTokenResponse{
			ID:         "token_id",
			TokenValue: "secret",
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createOrgTokenRequest{
				Description: desc,
				Name:        name,
				Admin:       true,
			},
			ExpectedReqPath: "/api/orgs/anOrg/tokens",
			ResponseCode:    201,
			ResponseBody:    resp,
		})
		token, err := c.CreateOrgAccessToken(teamCtx, name, orgName, desc, true)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          resp.ID,
			TokenValue:  resp.TokenValue,
			Description: desc,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/anOrg/tokens",
			ExpectedReqBody: createOrgTokenRequest{
				Description: desc,
				Name:        name,
			},
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		token, err := c.CreateOrgAccessToken(teamCtx, name, orgName, desc, false)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to create access token: 401 API error: unauthorized`,
		)
	})
}

func TestGetOrgAccessToken(t *testing.T) {
	id := "uuid"
	desc := testOrgTokenDescription
	org := testOrgTokenOrgName
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
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqBody:   nil,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/tokens", org),
			ResponseCode:      200,
			ResponseBody:      resp,
		})
		token, err := c.GetOrgAccessToken(ctx, id, org)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          id,
			Description: desc,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/tokens", org),
			ExpectedReqBody:   nil,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		token, err := c.GetOrgAccessToken(ctx, id, org)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to list org access tokens: 401 API error: unauthorized`,
		)
	})
}
