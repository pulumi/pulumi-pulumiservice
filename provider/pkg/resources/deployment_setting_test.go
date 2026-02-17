package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
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
					OperationContext: &pulumiapi.OperationContext{},
					GitHub:           &pulumiapi.GitHubConfiguration{},
					SourceContext:    &pulumiapi.SourceContext{},
					ExecutorContext:  &apitype.ExecutorContext{},
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

func TestDeploymentSettingsRoleRoundtrip(t *testing.T) {
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			OperationContext: &pulumiapi.OperationContext{
				Role: &pulumiapi.DeploymentRole{
					ID: "role-123",
				},
			},
		},
	}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
}

func TestDeploymentSettingsRoleWithOtherContextRoundtrip(t *testing.T) {
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			OperationContext: &pulumiapi.OperationContext{
				PreRunCommands: []string{"echo hello"},
				Role: &pulumiapi.DeploymentRole{
					ID: "role-456",
				},
				Options: &pulumiapi.OperationContextOptions{
					SkipInstallDependencies: true,
				},
			},
		},
	}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
}

func TestCheckPreservesUnknownsDuringPreview(t *testing.T) {
	keepUnknowns := plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true}
	provider := PulumiServiceDeploymentSettingsResource{}

	t.Run("top-level unknown passes validation", func(t *testing.T) {
		inputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("test-org"),
			"project":      resource.NewStringProperty("test-project"),
			"stack":        resource.MakeComputed(resource.NewStringProperty("")),
		}

		news, err := plugin.MarshalProperties(inputs, keepUnknowns)
		assert.NoError(t, err)

		resp, err := provider.Check(&pulumirpc.CheckRequest{News: news})
		assert.NoError(t, err)
		assert.Empty(t, resp.Failures)

		outputs, err := plugin.UnmarshalProperties(resp.Inputs, keepUnknowns)
		assert.NoError(t, err)
		assert.True(t, outputs["stack"].IsComputed())
	})

	t.Run("nested unknown in operationContext is preserved", func(t *testing.T) {
		inputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("test-org"),
			"project":      resource.NewStringProperty("test-project"),
			"stack":        resource.NewStringProperty("test-stack"),
			"operationContext": resource.NewObjectProperty(resource.PropertyMap{
				"role": resource.NewObjectProperty(resource.PropertyMap{
					"id": resource.MakeComputed(resource.NewStringProperty("")),
				}),
			}),
		}

		news, err := plugin.MarshalProperties(inputs, keepUnknowns)
		assert.NoError(t, err)

		resp, err := provider.Check(&pulumirpc.CheckRequest{News: news})
		assert.NoError(t, err)
		assert.Empty(t, resp.Failures)

		outputs, err := plugin.UnmarshalProperties(resp.Inputs, keepUnknowns)
		assert.NoError(t, err)
		oc := outputs["operationContext"].ObjectValue()
		role := oc["role"].ObjectValue()
		assert.True(t, role["id"].IsComputed())
	})

	t.Run("secrets are preserved alongside unknowns", func(t *testing.T) {
		inputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("test-org"),
			"project":      resource.NewStringProperty("test-project"),
			"stack":        resource.MakeSecret(resource.NewStringProperty("secret-stack")),
		}

		news, err := plugin.MarshalProperties(inputs, keepUnknowns)
		assert.NoError(t, err)

		resp, err := provider.Check(&pulumirpc.CheckRequest{News: news})
		assert.NoError(t, err)
		assert.Empty(t, resp.Failures)

		outputs, err := plugin.UnmarshalProperties(resp.Inputs, keepUnknowns)
		assert.NoError(t, err)
		assert.True(t, outputs["stack"].IsSecret())
		assert.Equal(t, "secret-stack", outputs["stack"].SecretValue().Element.StringValue())
	})
}
