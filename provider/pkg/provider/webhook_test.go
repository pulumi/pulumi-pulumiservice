package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type getWebhookFunc func() (*pulumiapi.Webhook, error)

type WebhookClientMock struct {
	getWebhookFunc getWebhookFunc
}

func (c *WebhookClientMock) GetWebhook(ctx context.Context, orgName string, projectName, stackName *string, webhookName string) (*pulumiapi.Webhook, error) {
	return c.getWebhookFunc()
}

func (c *WebhookClientMock) CreateWebhook(ctx context.Context, req pulumiapi.WebhookRequest) (*pulumiapi.Webhook, error) {
	return nil, nil
}

func (c *WebhookClientMock) ListWebhooks(ctx context.Context, orgName string, projectName, stackName *string) ([]pulumiapi.Webhook, error) {
	return nil, nil
}

func (c *WebhookClientMock) UpdateWebhook(ctx context.Context, req pulumiapi.UpdateWebhookRequest) error {
	return nil
}

func (c *WebhookClientMock) DeleteWebhook(ctx context.Context, orgName string, projectName, stackName *string, name string) error {
	return nil
}

func buildWebhookClientMock(getWebhookFunc getWebhookFunc) *WebhookClientMock {
	return &WebhookClientMock{
		getWebhookFunc,
	}
}

func TestWebhook(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildWebhookClientMock(
			func() (*pulumiapi.Webhook, error) { return nil, nil },
		)

		provider := PulumiServiceWebhookResource{}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/ghi/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(WithClient(mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildWebhookClientMock(
			func() (*pulumiapi.Webhook, error) {
				return &pulumiapi.Webhook{
					Active:      true,
					DisplayName: "test webhook",
					PayloadUrl:  "https://example.com/webhook",
					Name:        "test-webhook",
				}, nil
			},
		)

		provider := PulumiServiceWebhookResource{}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/ghi/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(WithClient(mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "abc/def/ghi/123")
	})
}
