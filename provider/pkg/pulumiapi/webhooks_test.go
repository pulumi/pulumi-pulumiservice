package pulumiapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWebhook(t *testing.T) {
	webhookName := "a-webhook"
	orgName := "an-organization"
	displayName := "A Webhook"
	payloadURL := "https://example.com/webhook"
	secret := "{...}"
	webhook := Webhook{
		Name:        webhookName,
		DisplayName: displayName,
		PayloadUrl:  payloadURL,
		Secret:      &secret,
		Active:      true,
	}
	createReq := WebhookRequest{
		OrganizationName: orgName,
		DisplayName:      displayName,
		PayloadURL:       payloadURL,
		Secret:           &secret,
		Active:           true,
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks",
			ExpectedReqBody:   createReq,
			ResponseCode:      201,
			ResponseBody:      webhook,
		})
		defer cleanup()
		actualWebhook, err := c.CreateWebhook(ctx, createReq)
		assert.NoError(t, err)
		assert.Equal(t, webhook, *actualWebhook)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks",
			ExpectedReqBody:   createReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		actualWebhook, err := c.CreateWebhook(ctx, createReq)
		assert.Nil(t, actualWebhook, "webhook should be nil since error was returned")
		assert.EqualError(t, err, "failed to create webhook: 401 API error: unauthorized")
	})
}

func TestListWebhooks(t *testing.T) {
	webhookName := "a-webhook"
	orgName := "an-organization"
	displayName := "A Webhook"
	payloadURL := "https://example.com/webhook"
	secret := "{...}"
	webhook := Webhook{
		Name:        webhookName,
		DisplayName: displayName,
		PayloadUrl:  payloadURL,
		Secret:      &secret,
		Active:      true,
	}
	webhooks := []Webhook{webhook}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks",
			ResponseCode:      200,
			ResponseBody:      webhooks,
		})
		defer cleanup()
		actualWebhooks, err := c.ListWebhooks(ctx, orgName, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, webhooks, actualWebhooks)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		actualWebhooks, err := c.ListWebhooks(ctx, orgName, nil, nil, nil)
		assert.Nil(t, actualWebhooks, "webhooks should be nil since error was returned")
		assert.EqualError(t, err, "failed to list webhooks: 401 API error: unauthorized")
	})
}

func TestGetWebhook(t *testing.T) {
	webhookName := "a-webhook"
	orgName := "an-organization"
	displayName := "A Webhook"
	payloadURL := "https://example.com/webhook"
	secret := "{...}"
	webhook := Webhook{
		Name:        webhookName,
		DisplayName: displayName,
		PayloadUrl:  payloadURL,
		Secret:      &secret,
		Active:      true,
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks/a-webhook",
			ResponseCode:      200,
			ResponseBody:      webhook,
		})
		defer cleanup()
		actualWebhook, err := c.GetWebhook(ctx, orgName, nil, nil, nil, webhookName)
		assert.NoError(t, err)
		assert.Equal(t, webhook, *actualWebhook)
	})

	t.Run("Error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks/a-webhook",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		actualWebhook, err := c.GetWebhook(ctx, orgName, nil, nil, nil, webhookName)
		assert.Nil(t, actualWebhook, "webhooks should be nil since error was returned")
		assert.EqualError(t, err, "failed to get webhook: 401 API error: unauthorized")
	})

	t.Run("404", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks/a-webhook",
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})
		defer cleanup()
		actualWebhook, err := c.GetWebhook(ctx, orgName, nil, nil, nil, webhookName)
		assert.Nil(t, actualWebhook, "webhook should be nil since error was returned")
		assert.Nil(t, err, "err should be nil since error was returned")
	})
}

func TestUpdateWebhook(t *testing.T) {
	webhookName := "a-webhook"
	orgName := "an-organization"
	displayName := "A Webhook"
	payloadURL := "https://example.com/webhook"
	secret := "{...}"
	webhook := Webhook{
		Name:        webhookName,
		DisplayName: displayName,
		PayloadUrl:  payloadURL,
		Secret:      &secret,
		Active:      true,
	}
	updateReq := UpdateWebhookRequest{
		Name: webhookName,
		WebhookRequest: WebhookRequest{
			OrganizationName: orgName,
			DisplayName:      displayName,
			PayloadURL:       payloadURL,
			Secret:           &secret,
			Active:           true,
		},
	}
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks/a-webhook",
			ExpectedReqBody:   updateReq,
			ResponseCode:      201,
			ResponseBody:      webhook,
		})
		defer cleanup()
		response, err := c.UpdateWebhook(ctx, updateReq)
		assert.NoError(t, err)
		assert.EqualValues(t, webhook, *response)
	})

	t.Run("error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks/a-webhook",
			ExpectedReqBody:   updateReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		_, err := c.UpdateWebhook(ctx, updateReq)
		assert.EqualError(t, err, "failed to update webhook: 401 API error: unauthorized")
	})
}

func TestDeleteWebhook(t *testing.T) {
	webhookName := "a-webhook"
	orgName := "an-organization"
	t.Run("Happy Path", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks/a-webhook",
			ResponseCode:      201,
		})
		defer cleanup()
		err := c.DeleteWebhook(ctx, orgName, nil, nil, nil, webhookName)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		c, cleanup := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/orgs/an-organization/hooks/a-webhook",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		defer cleanup()
		err := c.DeleteWebhook(ctx, orgName, nil, nil, nil, webhookName)
		assert.EqualError(t, err, "failed to delete webhook: 401 API error: unauthorized")
	})
}
