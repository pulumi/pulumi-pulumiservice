package pulumiapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestDeleteAccessToken(t *testing.T) {
	tokenId := "abcdegh"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/user/tokens/" + tokenId,
			ResponseCode:      204,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteAccessToken(ctx, tokenId))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/user/tokens/" + tokenId,
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "token not found",
			},
		})
		defer cleanup()
		assert.EqualError(t,
			c.DeleteAccessToken(ctx, tokenId),
			`failed to delete access token "abcdegh": 404 API error: token not found`,
		)
	})

}

func TestCreateAccessToken(t *testing.T) {
	desc := "token description"
	t.Run("Happy Path", func(t *testing.T) {
		resp := createTokenResponse{
			ID:         "token_id",
			TokenValue: "secret",
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createTokenRequest{
				Description: desc,
			},
			ExpectedReqPath: "/api/user/tokens",
			ResponseCode:    201,
			ResponseBody:    resp,
		})
		defer cleanup()
		token, err := c.CreateAccessToken(ctx, desc)
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
			ExpectedReqPath:   "/api/user/tokens",
			ExpectedReqBody: createTokenRequest{
				Description: desc,
			},
			ResponseCode: 401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		defer cleanup()
		token, err := c.CreateAccessToken(ctx, desc)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to create access token: 401 API error: unauthorized`,
		)
	})
}

func TestGetAccessToken(t *testing.T) {
	id := "uuid"
	desc := "token description"
	lastUsed := 123
	t.Run("Happy Path", func(t *testing.T) {
		resp := listTokenResponse{
			Tokens: []accessTokenResponse{
				accessTokenResponse{
					ID:          id,
					Description: desc,
					LastUsed:    lastUsed,
				},
				accessTokenResponse{
					ID:          "other",
					Description: desc,
					LastUsed:    lastUsed,
				},
			},
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqBody:   nil,
			ExpectedReqPath:   "/api/user/tokens",
			ResponseCode:      200,
			ResponseBody:      resp,
		})
		defer cleanup()
		token, err := c.GetAccessToken(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          id,
			Description: desc,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/user/tokens",
			ExpectedReqBody:   nil,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		defer cleanup()
		token, err := c.GetAccessToken(ctx, id)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to list access tokens: 401 API error: unauthorized`,
		)
	})
}
