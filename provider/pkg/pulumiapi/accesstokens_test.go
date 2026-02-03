package pulumiapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

const (
	testTokenID          = "abcdegh"
	testTokenUUID        = "uuid"
	testTokenDescription = "token description"
)

func TestDeleteAccessToken(t *testing.T) {
	tokenID := testTokenID
	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/user/tokens/" + tokenID,
			ResponseCode:      204,
		})
		assert.NoError(t, c.DeleteAccessToken(ctx, tokenID))
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/user/tokens/" + tokenID,
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "token not found",
			},
		})
		assert.EqualError(t,
			c.DeleteAccessToken(ctx, tokenID),
			`failed to delete access token "abcdegh": 404 API error: token not found`,
		)
	})

}

func TestCreateAccessToken(t *testing.T) {
	desc := testTokenDescription
	t.Run("Happy Path", func(t *testing.T) {
		resp := createTokenResponse{
			ID:         "token_id",
			TokenValue: "secret",
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createTokenRequest{
				Description: desc,
			},
			ExpectedReqPath: "/api/user/tokens",
			ResponseCode:    201,
			ResponseBody:    resp,
		})
		token, err := c.CreateAccessToken(ctx, desc)
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
		token, err := c.CreateAccessToken(ctx, desc)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to create access token: 401 API error: unauthorized`,
		)
	})
}

func TestGetAccessToken(t *testing.T) {
	id := testTokenUUID
	desc := testTokenDescription
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
			ExpectedReqPath:   "/api/user/tokens",
			ResponseCode:      200,
			ResponseBody:      resp,
		})
		token, err := c.GetAccessToken(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, &AccessToken{
			ID:          id,
			Description: desc,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/user/tokens",
			ExpectedReqBody:   nil,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		token, err := c.GetAccessToken(ctx, id)
		assert.Nil(t, token, "token should be nil")
		assert.EqualError(t,
			err,
			`failed to list access tokens: 401 API error: unauthorized`,
		)
	})
}
