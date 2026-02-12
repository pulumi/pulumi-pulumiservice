package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type getWebhookFunc func() (*pulumiapi.Webhook, error)

type WebhookClientMock struct {
	getWebhookFunc getWebhookFunc
}

func (c *WebhookClientMock) GetWebhook(
	_ context.Context,
	_ string,
	_, _, _ *string,
	_ string,
) (*pulumiapi.Webhook, error) {
	return c.getWebhookFunc()
}

func (c *WebhookClientMock) CreateWebhook(
	_ context.Context,
	_ pulumiapi.WebhookRequest,
) (*pulumiapi.Webhook, error) {
	return nil, nil
}

func (c *WebhookClientMock) ListWebhooks(
	_ context.Context,
	_ string,
	_, _, _ *string,
) ([]pulumiapi.Webhook, error) {
	return nil, nil
}

func (c *WebhookClientMock) UpdateWebhook(
	_ context.Context,
	_ pulumiapi.UpdateWebhookRequest,
) (*pulumiapi.Webhook, error) {
	return nil, nil
}

func (c *WebhookClientMock) DeleteWebhook(
	_ context.Context,
	_ string,
	_, _, _ *string,
	_ string,
) error {
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

		provider := PulumiServiceWebhookResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/ghi/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

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
					PayloadURL:  "https://example.com/webhook",
					Name:        "test-webhook",
				}, nil
			},
		)

		provider := PulumiServiceWebhookResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/ghi/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "abc/def/ghi/123")
	})
}
