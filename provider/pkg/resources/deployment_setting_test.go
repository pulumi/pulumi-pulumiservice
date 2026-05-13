package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type getDeploymentSettingsFunc func() (*pulumiapi.DeploymentSettings, error)

type DeploymentSettingsClientMock struct {
	getDeploymentSettingsFunc getDeploymentSettingsFunc
}

func (c *DeploymentSettingsClientMock) CreateDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
	_ pulumiapi.DeploymentSettings,
) (*pulumiapi.DeploymentSettings, error) {
	return nil, nil
}

func (c *DeploymentSettingsClientMock) UpdateDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
	_ pulumiapi.DeploymentSettings,
) (*pulumiapi.DeploymentSettings, error) {
	return nil, nil
}

func (c *DeploymentSettingsClientMock) GetDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
) (*pulumiapi.DeploymentSettings, error) {
	return c.getDeploymentSettingsFunc()
}

func (c *DeploymentSettingsClientMock) DeleteDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
) error {
	return nil
}

func buildDeploymentSettingsClientMock(
	getDeploymentSettingsFunc getDeploymentSettingsFunc,
) *DeploymentSettingsClientMock {
	return &DeploymentSettingsClientMock{
		getDeploymentSettingsFunc,
	}
}

func TestDeploymentSettings(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) { return nil, nil },
		)

		provider := PulumiServiceDeploymentSettingsResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) {
				return &pulumiapi.DeploymentSettings{
					Operation:     &pulumiapi.OperationContext{},
					GitHub:        &pulumiapi.DeploymentSettingsGitHub{},
					SourceContext: &pulumiapi.SourceContext{},
					Executor:      &pulumiapi.ExecutorContext{},
				}, nil
			},
		)

		provider := PulumiServiceDeploymentSettingsResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "abc/def/123")
	})
}

func TestDeploymentSettingsRoundtrip(t *testing.T) {
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			CacheOptions: &pulumiapi.CacheOptions{
				Enable: true,
			},
		}}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
}

func TestDeploymentSettingsVcsRoundtrip(t *testing.T) {
	deployPR := int64(1)
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			Vcs: pulumiapi.DeploymentSettingsVCSAzureDevOpsBuilder{
				DeploymentSettingsVCSBuilder: pulumiapi.DeploymentSettingsVCSBuilder{
					Repository:          "my-org/my-repo",
					InstallationID:      "129444790",
					DeployCommits:       true,
					PreviewPullRequests: true,
					PullRequestTemplate: false,
					Paths:               []string{"infra/**"},
					DeployPullRequest:   &deployPR,
				},
			}.Build(),
		},
	}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
}
