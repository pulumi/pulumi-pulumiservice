package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type getDeploymentSettingsFunc func() (*pulumiapi.DeploymentSettings, error)

type DeploymentSettingsClientMock struct {
	getDeploymentSettingsFunc getDeploymentSettingsFunc
}

func (c *DeploymentSettingsClientMock) CreateDeploymentSettings(ctx context.Context, stack pulumiapi.StackName, ds pulumiapi.DeploymentSettings) error {
	return nil
}
func (c *DeploymentSettingsClientMock) GetDeploymentSettings(ctx context.Context, stack pulumiapi.StackName) (*pulumiapi.DeploymentSettings, error) {
	return c.getDeploymentSettingsFunc()
}
func (c *DeploymentSettingsClientMock) DeleteDeploymentSettings(ctx context.Context, stack pulumiapi.StackName) error {
	return nil
}

func buildDeploymentSettingsClientMock(getDeploymentSettingsFunc getDeploymentSettingsFunc) *DeploymentSettingsClientMock {
	return &DeploymentSettingsClientMock{
		getDeploymentSettingsFunc,
	}
}

func TestDeploymentSettings(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) { return nil, nil },
		)

		provider := PulumiServiceDeploymentSettingsResource{}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(WithClient(mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) {
				return &pulumiapi.DeploymentSettings{
					OperationContext: &pulumiapi.OperationContext{},
					GitHub:           &pulumiapi.GitHubConfiguration{},
					SourceContext:    &apitype.SourceContext{},
					ExecutorContext:  &apitype.ExecutorContext{},
				}, nil
			},
		)

		provider := PulumiServiceDeploymentSettingsResource{}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(WithClient(mockedClient), &req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "abc/def/123")
	})
}
