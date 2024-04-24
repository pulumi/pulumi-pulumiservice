package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type getAgentPoolFunc func() (*pulumiapi.AgentPool, error)

type AgentPoolClientMock struct {
	getAgentPoolFunc getAgentPoolFunc
}

func (c *AgentPoolClientMock) GetAgentPool(ctx context.Context, agentPoolId, orgName string) (*pulumiapi.AgentPool, error) {
	return c.getAgentPoolFunc()
}
func (c *AgentPoolClientMock) CreateAgentPool(ctx context.Context, name, orgName, description string) (*pulumiapi.AgentPool, error) {
	return nil, nil
}
func (c *AgentPoolClientMock) UpdateAgentPool(ctx context.Context, agentPoolId, name, orgName, description string) error {
	return nil
}
func (c *AgentPoolClientMock) DeleteAgentPool(ctx context.Context, agentPoolId, orgName string) error {
	return nil
}

func buildAgentPoolClientMock(getAgentPoolFunc getAgentPoolFunc) *AgentPoolClientMock {
	return &AgentPoolClientMock{
		getAgentPoolFunc,
	}
}

func TestAgentPool(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildAgentPoolClientMock(
			func() (*pulumiapi.AgentPool, error) { return nil, nil },
		)

		provider := PulumiServiceAgentPoolResource{}

		req := pulumirpc.ReadRequest{
			Id:  "org/abc/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(context.WithValue(context.Background(),
			TestClientKey, mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildAgentPoolClientMock(
			func() (*pulumiapi.AgentPool, error) {
				return &pulumiapi.AgentPool{
					Name:        "test",
					Description: "test agent pool description",
				}, nil
			},
		)

		provider := PulumiServiceAgentPoolResource{}

		req := pulumirpc.ReadRequest{
			Id:  "org/abc/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(context.WithValue(context.Background(),
			TestClientKey, mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "org/abc/123")
	})
}
