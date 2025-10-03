package pulumiapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testOrgName = "anOrg"
)

func TestDeleteAgentPool(t *testing.T) {
	orgName := testOrgName
	agentPoolId := "abcdegh"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/" + testOrgName + "/agent-pools/" + agentPoolId,
			ResponseCode:      204,
		})
		defer cleanup()
		assert.NoError(t, c.DeleteAgentPool(teamCtx, agentPoolId, orgName, false))
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/" + testOrgName + "/agent-pools/" + agentPoolId,
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "agent pool not found",
			},
		})
		defer cleanup()
		assert.EqualError(t,
			c.DeleteAgentPool(teamCtx, agentPoolId, orgName, false),
			`failed to delete agent pool "abcdegh": 404 API error: agent pool not found`,
		)
	})

}

func TestCreateAgentPool(t *testing.T) {
	orgName := testOrgName
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
			ExpectedReqPath: "/api/orgs/" + testOrgName + "/agent-pools",
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
			ResponseBody: ErrorResponse{
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
	org := testOrgName
	t.Run("Happy Path", func(t *testing.T) {
		resp := AgentPool{
			ID:          id,
			Name:        name,
			Description: desc,
		}
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqBody:   nil,
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/agent-pools/%s", org, id),
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
			ExpectedReqPath:   fmt.Sprintf("/api/orgs/%s/agent-pools/%s", org, id),
			ExpectedReqBody:   nil,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				StatusCode: 401,
				Message:    "unauthorized",
			},
		})
		defer cleanup()
		token, err := c.GetAgentPool(ctx, id, org)
		assert.Nil(t, token, "agent pool should be nil")
		assert.EqualError(t,
			err,
			`failed to get agent pool: 401 API error: unauthorized`,
		)
	})
}
