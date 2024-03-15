package pulumiapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteAgentPool(t *testing.T) {
	orgName := "anOrg"
	agentPoolId := "abcdegh"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/agent-pools/" + agentPoolId,
			ResponseCode:      204,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteAgentPool(teamCtx, agentPoolId, orgName))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/anOrg/agent-pools/" + agentPoolId,
			ResponseCode:      404,
			ResponseBody: errorResponse{
				StatusCode: 404,
				Message:    "agent pool not found",
			},
		})
		defer cleanup()
		assert.EqualError(t,
			c.DeleteAgentPool(teamCtx, agentPoolId, orgName),
			`failed to delete agent pool "abcdegh": 404 API error: agent pool not found`,
		)
	})

}

func TestCreateAgentPool(t *testing.T) {
	orgName := "anOrg"
	name := "anAgentPool"
	desc := "agent pool description"

	t.Run("Happy Path", func(t *testing.T) {
		resp := createTokenResponse{
			ID:         "token_id",
			TokenValue: "secret",
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqBody: createUpdateAgentPoolRequest{
				Description: desc,
				Name:        name,
			},
			ExpectedReqPath: "/api/orgs/anOrg/agent-pools",
			ResponseCode:    201,
			ResponseBody:    resp,
		})
		defer cleanup()
		token, err := c.CreateAgentPool(teamCtx, orgName, name, desc)
		assert.NoError(t, err)
		assert.Equal(t, &AgentPool{
			ID:          resp.ID,
			Name:        name,
			Description: desc,
			TokenValue:  resp.TokenValue,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/anOrg/agent-pools",
			ExpectedReqBody: createUpdateAgentPoolRequest{
				Description: desc,
				Name:        name,
			},
			ResponseCode: 401,
			ResponseBody: errorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		defer cleanup()
		token, err := c.CreateAgentPool(teamCtx, orgName, name, desc)
		assert.Nil(t, token, "agent pool should be nil")
		assert.EqualError(t,
			err,
			`failed to create agent pool: 401 API error: unauthorized`,
		)
	})
}

func TestGetAgentPool(t *testing.T) {
	id := "uuid"
	name := "Pool 1"
	desc := "agent pool description"
	org := "anOrg"
	created := 123
	t.Run("Happy Path", func(t *testing.T) {
		resp := listAgentPoolResponse{
			AgentPools: []agentPoolResponse{
				{
					ID:          id,
					Name:        name,
					Description: desc,
					Created:     created,
				},
				{
					ID:          "other",
					Name:        "Pool 2",
					Description: desc,
					Created:     created,
				},
			},
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqBody:   nil,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/agent-pools", org),
			ResponseCode:      200,
			ResponseBody:      resp,
		})
		defer cleanup()
		token, err := c.GetAgentPool(ctx, id, org)
		assert.NoError(t, err)
		assert.Equal(t, &AgentPool{
			ID:          id,
			Name:        name,
			Description: desc,
		}, token)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/agent-pools", org),
			ExpectedReqBody:   nil,
			ResponseCode:      401,
			ResponseBody: errorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		defer cleanup()
		token, err := c.GetAgentPool(ctx, id, org)
		assert.Nil(t, token, "agent pool should be nil")
		assert.EqualError(t,
			err,
			`failed to list agent pools: 401 API error: unauthorized`,
		)
	})
}
