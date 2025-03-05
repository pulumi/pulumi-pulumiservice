package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type getAgentPoolFunc func() (*pulumiapi.AgentPool, error)
type deleteAgentPoolFunc func() error

type AgentPoolClientMock struct {
	getAgentPoolFunc    getAgentPoolFunc
	deleteAgentPoolFunc deleteAgentPoolFunc
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
func (c *AgentPoolClientMock) DeleteAgentPool(ctx context.Context, agentPoolId, orgName string, forceDestroy bool) error {
	return c.deleteAgentPoolFunc()
}

func buildAgentPoolClientMock(getAgentPoolFunc getAgentPoolFunc, deleteAgentPoolFunc deleteAgentPoolFunc) *AgentPoolClientMock {
	return &AgentPoolClientMock{
		getAgentPoolFunc,
		deleteAgentPoolFunc,
	}
}

func TestAgentPool(t *testing.T) {
	t.Run("Delete without force delete", func(t *testing.T) {
		errMsg := "Bad Request: 1 stacks have been configured to use deployment pool: \"beep-boop\". " +
			"Please change the stack deployment configuration before deleting this pool."
		mockedClient := buildAgentPoolClientMock(
			nil,
			func() error {
				return apitype.ErrorResponse{
					Code:    400,
					Message: errMsg,
				}
			},
		)

		provider := PulumiServiceAgentPoolResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:  "org/abc/beep-boop",
			Urn: "urn:beep-boop",
		}

		resp, err := provider.Delete(&req)

		assert.Error(t, err)
		assert.ErrorContains(t, err, errMsg)
		assert.Equal(t, resp, &emptypb.Empty{})
	})

	t.Run("Delete with force delete", func(t *testing.T) {
		mockedClient := buildAgentPoolClientMock(
			nil,
			func() error { return nil },
		)

		provider := PulumiServiceAgentPoolResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:  "org/abc/beep-boop",
			Urn: "urn:beep-boop",
			Properties: &structpb.Struct{Fields: map[string]*structpb.Value{
				"forceDestroy": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
			}},
		}

		resp, err := provider.Delete(&req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildAgentPoolClientMock(
			func() (*pulumiapi.AgentPool, error) { return nil, nil },
			nil,
		)

		provider := PulumiServiceAgentPoolResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "org/abc/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

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
			nil,
		)

		provider := PulumiServiceAgentPoolResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "org/abc/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "org/abc/123")
	})
}
