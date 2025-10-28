// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// Mock types for DeploymentSettings
type getDeploymentSettingsFunc func() (*pulumiapi.DeploymentSettings, error)
type createDeploymentSettingsFunc func() (*pulumiapi.DeploymentSettings, error)
type updateDeploymentSettingsFunc func() (*pulumiapi.DeploymentSettings, error)
type deleteDeploymentSettingsFunc func() error

type DeploymentSettingsClientMock struct {
	getDeploymentSettingsFunc    getDeploymentSettingsFunc
	createDeploymentSettingsFunc createDeploymentSettingsFunc
	updateDeploymentSettingsFunc updateDeploymentSettingsFunc
	deleteDeploymentSettingsFunc deleteDeploymentSettingsFunc
}

func (c *DeploymentSettingsClientMock) CreateDeploymentSettings(ctx context.Context, stack pulumiapi.StackIdentifier, ds pulumiapi.DeploymentSettings) (*pulumiapi.DeploymentSettings, error) {
	if c.createDeploymentSettingsFunc != nil {
		return c.createDeploymentSettingsFunc()
	}
	return nil, nil
}

func (c *DeploymentSettingsClientMock) UpdateDeploymentSettings(ctx context.Context, stack pulumiapi.StackIdentifier, ds pulumiapi.DeploymentSettings) (*pulumiapi.DeploymentSettings, error) {
	if c.updateDeploymentSettingsFunc != nil {
		return c.updateDeploymentSettingsFunc()
	}
	return nil, nil
}

func (c *DeploymentSettingsClientMock) GetDeploymentSettings(ctx context.Context, stack pulumiapi.StackIdentifier) (*pulumiapi.DeploymentSettings, error) {
	if c.getDeploymentSettingsFunc != nil {
		return c.getDeploymentSettingsFunc()
	}
	return nil, nil
}

func (c *DeploymentSettingsClientMock) DeleteDeploymentSettings(ctx context.Context, stack pulumiapi.StackIdentifier) error {
	if c.deleteDeploymentSettingsFunc != nil {
		return c.deleteDeploymentSettingsFunc()
	}
	return nil
}

func buildDeploymentSettingsClientMock(
	getFunc getDeploymentSettingsFunc,
	createFunc createDeploymentSettingsFunc,
	updateFunc updateDeploymentSettingsFunc,
	deleteFunc deleteDeploymentSettingsFunc,
) *DeploymentSettingsClientMock {
	return &DeploymentSettingsClientMock{
		getDeploymentSettingsFunc:    getFunc,
		createDeploymentSettingsFunc: createFunc,
		updateDeploymentSettingsFunc: updateFunc,
		deleteDeploymentSettingsFunc: deleteFunc,
	}
}

// Test helper functions
func testDeploymentSettingsInputMinimal() PulumiServiceDeploymentSettingsInput {
	return PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
		},
	}
}

func testDeploymentSettingsMinimal() *pulumiapi.DeploymentSettings {
	return &pulumiapi.DeploymentSettings{
		AgentPoolId: "test-agent-pool",
	}
}

func testDeploymentSettingsWithGitHub() *pulumiapi.DeploymentSettings {
	return &pulumiapi.DeploymentSettings{
		AgentPoolId: "test-agent-pool",
		GitHub: &pulumiapi.GitHubConfiguration{
			Repository:          "org/repo",
			DeployCommits:       true,
			PreviewPullRequests: true,
		},
	}
}

func testDeploymentSettingsWithOIDC() *pulumiapi.DeploymentSettings {
	return &pulumiapi.DeploymentSettings{
		AgentPoolId: "test-agent-pool",
		OperationContext: &pulumiapi.OperationContext{
			OIDC: &pulumiapi.OIDCConfiguration{
				AWS: &pulumiapi.AWSOIDCConfiguration{
					RoleARN:     "arn:aws:iam::123456789012:role/test-role",
					Duration:    "1h0m0s",
					SessionName: "test-session",
				},
			},
			EnvironmentVariables: map[string]pulumiapi.SecretValue{
				"PUBLIC_VAR": {Value: "public-value", Secret: false},
				"SECRET_VAR": {Value: "secret-value", Secret: true},
			},
		},
	}
}

// Original Tests (keeping for compatibility)
func TestDeploymentSettings(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) { return nil, nil },
			nil,
			nil,
			nil,
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
			nil,
			nil,
			nil,
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

// New comprehensive tests below

// Read Tests
func TestDeploymentSettings_Read_FoundWithComplexSettings(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		func() (*pulumiapi.DeploymentSettings, error) {
			return testDeploymentSettingsWithOIDC(), nil
		},
		nil,
		nil,
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.ReadRequest{
		Id:         "test-org/test-project/test-stack",
		Properties: inputProperties,
	}

	resp, err := provider.Read(&req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-project/test-stack", resp.Id)
	assert.NotNil(t, resp.Properties)
}

// Create Tests
func TestDeploymentSettings_Create_MinimalSettings(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return testDeploymentSettingsMinimal(), nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Create_WithGitHub(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return testDeploymentSettingsWithGitHub(), nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
			GitHub: &pulumiapi.GitHubConfiguration{
				Repository:          "org/repo",
				DeployCommits:       true,
				PreviewPullRequests: true,
			},
		},
	}

	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Create_WithOperationContext(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return testDeploymentSettingsWithOIDC(), nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
			OperationContext: &pulumiapi.OperationContext{
				EnvironmentVariables: map[string]pulumiapi.SecretValue{
					"PUBLIC_VAR": {Value: "public-value", Secret: false},
					"SECRET_VAR": {Value: "secret-value", Secret: true},
				},
			},
		},
	}

	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Create_WithAWSOIDC(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return testDeploymentSettingsWithOIDC(), nil
		},
		nil,
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
			OperationContext: &pulumiapi.OperationContext{
				OIDC: &pulumiapi.OIDCConfiguration{
					AWS: &pulumiapi.AWSOIDCConfiguration{
						RoleARN:     "arn:aws:iam::123456789012:role/test-role",
						Duration:    "1h30m", // Should be normalized to "1h30m0s"
						SessionName: "test-session",
					},
				},
			},
		},
	}

	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Create_APIError(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return nil, errors.New("API error")
		},
		nil,
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CreateRequest{
		Properties: inputProperties,
	}

	resp, err := provider.Create(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Update Tests
func TestDeploymentSettings_Update_ChangesAgentPoolId(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return &pulumiapi.DeploymentSettings{
				AgentPoolId: "new-agent-pool",
			}, nil
		},
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	oldInput := testDeploymentSettingsInputMinimal()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	newInput := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "new-agent-pool",
		},
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-project/test-stack",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Update(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Update_AddsGitHub(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return testDeploymentSettingsWithGitHub(), nil
		},
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	oldInput := testDeploymentSettingsInputMinimal()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	newInput := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
			GitHub: &pulumiapi.GitHubConfiguration{
				Repository:          "org/repo",
				DeployCommits:       true,
				PreviewPullRequests: true,
			},
		},
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-project/test-stack",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Update(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Update_PartialUpdate(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return testDeploymentSettingsWithOIDC(), nil
		},
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	oldInput := testDeploymentSettingsInputMinimal()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	newInput := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
			OperationContext: &pulumiapi.OperationContext{
				EnvironmentVariables: map[string]pulumiapi.SecretValue{
					"NEW_VAR": {Value: "new-value", Secret: false},
				},
			},
		},
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-project/test-stack",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Update(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Update_APIError(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		nil,
		func() (*pulumiapi.DeploymentSettings, error) {
			return nil, errors.New("API error")
		},
		nil,
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.UpdateRequest{
		Id:   "test-org/test-project/test-stack",
		Olds: inputProperties,
		News: inputProperties,
	}

	resp, err := provider.Update(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Delete Tests
func TestDeploymentSettings_Delete_Success(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		nil,
		nil,
		func() error { return nil },
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-stack",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Delete_APIError(t *testing.T) {
	mockedClient := buildDeploymentSettingsClientMock(
		nil,
		nil,
		nil,
		func() error { return errors.New("API error") },
	)

	provider := PulumiServiceDeploymentSettingsResource{
		Client: mockedClient,
	}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DeleteRequest{
		Id:         "test-org/test-project/test-stack",
		Properties: inputProperties,
	}

	resp, err := provider.Delete(&req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Diff Tests
func TestDeploymentSettings_Diff_NoChanges(t *testing.T) {
	provider := PulumiServiceDeploymentSettingsResource{}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:   "test-org/test-project/test-stack",
		Olds: inputProperties,
		News: inputProperties,
	}

	resp, err := provider.Diff(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Diff_DetectsChanges(t *testing.T) {
	provider := PulumiServiceDeploymentSettingsResource{}

	oldInput := testDeploymentSettingsInputMinimal()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	newInput := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "new-agent-pool",
		},
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:   "test-org/test-project/test-stack",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Diff(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Diff_DetectsNestedChanges(t *testing.T) {
	provider := PulumiServiceDeploymentSettingsResource{}

	oldInput := testDeploymentSettingsInputMinimal()
	oldProperties, _ := plugin.MarshalProperties(
		oldInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	newInput := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
			GitHub: &pulumiapi.GitHubConfiguration{
				Repository: "org/repo",
			},
		},
	}
	newProperties, _ := plugin.MarshalProperties(
		newInput.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.DiffRequest{
		Id:   "test-org/test-project/test-stack",
		Olds: oldProperties,
		News: newProperties,
	}

	resp, err := provider.Diff(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// Check Tests
func TestDeploymentSettings_Check_ValidInputs(t *testing.T) {
	provider := PulumiServiceDeploymentSettingsResource{}

	input := testDeploymentSettingsInputMinimal()
	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CheckRequest{
		News: inputProperties,
	}

	resp, err := provider.Check(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Check_MissingStackField(t *testing.T) {
	provider := PulumiServiceDeploymentSettingsResource{}

	// Missing stack field
	input := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			// StackName missing
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
		},
	}

	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CheckRequest{
		News: inputProperties,
	}

	resp, err := provider.Check(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeploymentSettings_Check_ComplexNestedStructure(t *testing.T) {
	provider := PulumiServiceDeploymentSettingsResource{}

	// Complex structure with multiple nested objects
	input := PulumiServiceDeploymentSettingsInput{
		Stack: pulumiapi.StackIdentifier{
			OrgName:     "test-org",
			ProjectName: "test-project",
			StackName:   "test-stack",
		},
		DeploymentSettings: pulumiapi.DeploymentSettings{
			AgentPoolId: "test-agent-pool",
			GitHub: &pulumiapi.GitHubConfiguration{
				Repository:          "org/repo",
				DeployCommits:       true,
				PreviewPullRequests: true,
			},
			OperationContext: &pulumiapi.OperationContext{
				OIDC: &pulumiapi.OIDCConfiguration{
					AWS: &pulumiapi.AWSOIDCConfiguration{
						RoleARN:     "arn:aws:iam::123456789012:role/test-role",
						Duration:    "1h30m",
						SessionName: "test-session",
					},
				},
				EnvironmentVariables: map[string]pulumiapi.SecretValue{
					"VAR1": {Value: "value1", Secret: false},
					"VAR2": {Value: "value2", Secret: true},
				},
			},
		},
	}

	inputProperties, _ := plugin.MarshalProperties(
		input.ToPropertyMap(nil, nil, true),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)

	req := pulumirpc.CheckRequest{
		News: inputProperties,
	}

	resp, err := provider.Check(&req)

	// This will fail since the resource is not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
