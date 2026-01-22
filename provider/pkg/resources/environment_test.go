package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pulumi/esc"
	"github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/asset"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type getEnvironmentFunc func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error)
type getEnvironmentRevisionTagFunc func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error)
type createEnvironmentWithProjectFunc func(ctx context.Context, orgName, projectName, envName string) error
type updateEnvironmentWithRevisionFunc func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error)

type EscClientMock struct {
	getEnvironmentFunc                getEnvironmentFunc
	getEnvironmentRevisionTagFunc     getEnvironmentRevisionTagFunc
	createEnvironmentWithProjectFunc  createEnvironmentWithProjectFunc
	updateEnvironmentWithRevisionFunc updateEnvironmentWithRevisionFunc
}

func (c *EscClientMock) GetEnvironment(ctx context.Context, orgName, projectName, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
	return c.getEnvironmentFunc(ctx, orgName, envName, version, decrypt)
}

func (c *EscClientMock) GetEnvironmentRevision(ctx context.Context, orgName, projectName, envName string, revision int) (*client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentRevisionTag(ctx context.Context, orgName, projectName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
	return c.getEnvironmentRevisionTagFunc(ctx, orgName, envName, tagName)
}

func (c *EscClientMock) GetRevisionNumber(ctx context.Context, orgName, projectName, envName, version string) (int, error) {
	return 0, nil
}

func (c *EscClientMock) CheckYAMLEnvironment(context.Context, string, []byte, ...client.CheckYAMLOption) (*esc.Environment, []client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}

func (c *EscClientMock) CreateEnvironment(context.Context, string, string) error {
	return nil
}

func (c *EscClientMock) CreateEnvironmentWithProject(ctx context.Context, orgName, projectName, envName string) error {
	if c.createEnvironmentWithProjectFunc != nil {
		return c.createEnvironmentWithProjectFunc(ctx, orgName, projectName, envName)
	}
	return nil
}

func (c *EscClientMock) CloneEnvironment(context.Context, string, string, string, client.CloneEnvironmentRequest) error {
	return nil
}

func (c *EscClientMock) DeleteEnvironment(context.Context, string, string, string) error {
	return nil
}

func (c *EscClientMock) EnvironmentExists(context.Context, string, string, string) (bool, error) {
	return false, nil
}

func (c *EscClientMock) GetAnonymousOpenEnvironment(context.Context, string, string) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenEnvironment(context.Context, string, string, string) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenEnvironmentWithProject(context.Context, string, string, string, string) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetAnonymousOpenProperty(context.Context, string, string, string) (*esc.Value, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenProperty(context.Context, string, string, string, string, string) (*esc.Value, error) {
	return nil, nil
}

func (c *EscClientMock) GetPulumiAccountDetails(context.Context) (string, []string, *workspace.TokenInformation, error) {
	return "", nil, nil, nil
}

func (c *EscClientMock) ListEnvironments(context.Context, string) ([]client.OrgEnvironment, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) ListOrganizationEnvironments(context.Context, string, string) ([]client.OrgEnvironment, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) ListEnvironmentRevisions(ctx context.Context, orgName, projectName, envName string, options client.ListEnvironmentRevisionsOptions) ([]client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) ListEnvironmentRevisionTags(ctx context.Context, orgName, projectName, envName string, options client.ListEnvironmentRevisionTagsOptions) ([]client.EnvironmentRevisionTag, error) {
	return nil, nil
}

func (c *EscClientMock) OpenEnvironment(context.Context, string, string, string, string, time.Duration) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) OpenYAMLEnvironment(context.Context, string, []byte, time.Duration) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) OpenEnvironmentDraft(context.Context, string, string, string, string, time.Duration) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) RotateEnvironment(context.Context, string, string, string, []string) (*client.RotateEnvironmentResponse, []client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}

func (c *EscClientMock) SubmitChangeRequest(context.Context, string, string, *string) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironment(context.Context, string, string, []byte, string) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentWithProject(context.Context, string, string, string, []byte, string) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentDraft(context.Context, string, string, string, string, []byte, string) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) CreateEnvironmentDraft(context.Context, string, string, string, []byte, string) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) GetDefaultOrg(context.Context) (string, error) {
	return "", nil
}

func (c *EscClientMock) GetEnvironmentDraft(context.Context, string, string, string, string) ([]byte, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) CreateEnvironmentTag(context.Context, string, string, string, string, string) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentTag(context.Context, string, string, string, string) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) ListEnvironmentTags(context.Context, string, string, string, client.ListEnvironmentTagsOptions) ([]*client.EnvironmentTag, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) UpdateEnvironmentTag(context.Context, string, string, string, string, string, string, string) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) DeleteEnvironmentTag(context.Context, string, string, string, string) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironmentRevisionTag(ctx context.Context, orgName, projectName, envName, tagName string, revision *int) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironmentWithRevision(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
	if c.updateEnvironmentWithRevisionFunc != nil {
		return c.updateEnvironmentWithRevisionFunc(ctx, orgName, projectName, envName, yaml, etag)
	}
	return nil, 0, nil
}

func (c *EscClientMock) CreateEnvironmentRevisionTag(ctx context.Context, orgName, projectName, envName, tagName string, revision *int) error {
	return nil
}

func (c *EscClientMock) DeleteEnvironmentRevisionTag(ctx context.Context, orgName, projectName, envName, tagName string) error {
	return nil
}

func (c *EscClientMock) RetractEnvironmentRevision(ctx context.Context, orgName, projectName, envName string, version string, replacement *int, reason string) error {
	return nil
}

func (c *EscClientMock) Insecure() bool {
	return false
}

func (c *EscClientMock) URL() string {
	return ""
}

func buildEscClientMock(getEnvironmentFunc getEnvironmentFunc, getEnvironmentRevisionTagFunc getEnvironmentRevisionTagFunc) *EscClientMock {
	return &EscClientMock{
		getEnvironmentFunc:            getEnvironmentFunc,
		getEnvironmentRevisionTagFunc: getEnvironmentRevisionTagFunc,
	}
}

type EnvironmentSettingsClientMock struct {
	getSettingsFunc    func(ctx context.Context, orgName, projectName, envName string) (*pulumiapi.EnvironmentSettings, error)
	updateSettingsFunc func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error
}

func (m *EnvironmentSettingsClientMock) GetEnvironmentSettings(ctx context.Context, orgName, projectName, envName string) (*pulumiapi.EnvironmentSettings, error) {
	if m.getSettingsFunc != nil {
		return m.getSettingsFunc(ctx, orgName, projectName, envName)
	}
	return &pulumiapi.EnvironmentSettings{}, nil
}

func (m *EnvironmentSettingsClientMock) UpdateEnvironmentSettings(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
	if m.updateSettingsFunc != nil {
		return m.updateSettingsFunc(ctx, orgName, projectName, envName, req)
	}
	return nil
}

func TestEnvironmentCheck(t *testing.T) {
	mockedClient := buildEscClientMock(
		func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
			return nil, "", 0, fmt.Errorf("not found")
		},
		func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
			return nil, nil
		},
	)
	provider := PulumiServiceEnvironmentResource{
		Client: mockedClient,
	}

	envDef := `values:
foo: bar
`
	propertyMap := resource.PropertyMap{}
	propertyMap["organization"] = resource.NewPropertyValue("org")
	propertyMap["project"] = resource.NewPropertyValue("project")
	propertyMap["name"] = resource.NewPropertyValue("env")
	propertyMap["yaml"] = resource.NewAssetProperty(&asset.Asset{Text: envDef})

	t.Run("Check", func(t *testing.T) {
		properties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.CheckRequest{
			News: properties,
		}

		_, err := provider.Check(&req)
		assert.NoError(t, err)
	})

	t.Run("Check when yaml contains computed resource", func(t *testing.T) {
		propertyMap["yaml"] = resource.NewComputedProperty(resource.Computed{Element: resource.NewStringProperty("")})

		properties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.CheckRequest{
			News: properties,
		}

		_, err := provider.Check(&req)
		assert.NoError(t, err)
	})

	t.Run("Check when yaml is secret wrapping computed resource", func(t *testing.T) {
		// This tests the bug fix for issue #606:
		// Secret wraps a computed value (from Output.ApplyT), causing panic in Check()
		computedValue := resource.NewComputedProperty(resource.Computed{Element: resource.NewStringProperty("")})
		propertyMap["yaml"] = resource.MakeSecret(computedValue)

		properties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
				KeepSecrets:  true,
			},
		)
		req := pulumirpc.CheckRequest{
			News: properties,
		}

		_, err := provider.Check(&req)
		assert.NoError(t, err)
	})
}

func TestEnvironment(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, fmt.Errorf("not found")
			},
			func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
				return nil, nil
			},
		)

		provider := PulumiServiceEnvironmentResource{
			Client: mockedClient,
		}

		input := PulumiServiceEnvironmentInput{
			OrgName:     "org",
			ProjectName: "project",
			EnvName:     "env",
			Yaml: `values:
	foo: bar
`,
		}

		propertyMap, _ := input.ToPropertyMap()
		outputProperties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.ReadRequest{
			Id:         "org/env",
			Properties: outputProperties,
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, nil
			},
			func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
				return nil, nil
			},
		)

		provider := PulumiServiceEnvironmentResource{
			Client: mockedClient,
		}

		input := PulumiServiceEnvironmentInput{
			OrgName: "org",
			EnvName: "project",
			Yaml: `values:
	foo: bar
`,
		}

		propertyMap, _ := input.ToPropertyMap()
		outputProperties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.ReadRequest{
			Id:         "org/env",
			Properties: outputProperties,
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "org/env")
	})

	t.Run("Read with deletion protection enabled", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
				return []byte("values:\n  foo: bar"), "", 1, nil
			},
			func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
				return nil, nil
			},
		)

		settingsClient := &EnvironmentSettingsClientMock{
			getSettingsFunc: func(ctx context.Context, orgName, projectName, envName string) (*pulumiapi.EnvironmentSettings, error) {
				return &pulumiapi.EnvironmentSettings{DeletionProtected: true}, nil
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		req := pulumirpc.ReadRequest{
			Id: "org/project/env",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "org/project/env", resp.Id)

		properties, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{KeepSecrets: true})
		assert.True(t, properties["deletionProtected"].BoolValue())
	})

	t.Run("Read with deletion protection disabled", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
				return []byte("values:\n  foo: bar"), "", 1, nil
			},
			func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
				return nil, nil
			},
		)

		settingsClient := &EnvironmentSettingsClientMock{
			getSettingsFunc: func(ctx context.Context, orgName, projectName, envName string) (*pulumiapi.EnvironmentSettings, error) {
				return &pulumiapi.EnvironmentSettings{DeletionProtected: false}, nil
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		req := pulumirpc.ReadRequest{
			Id: "org/project/env",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "org/project/env", resp.Id)

		properties, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{KeepSecrets: true})
		assert.False(t, properties["deletionProtected"].BoolValue())
	})

	t.Run("Read returns error when settings client fails", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
				return []byte("values:\n  foo: bar"), "", 1, nil
			},
			func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
				return nil, nil
			},
		)

		settingsClient := &EnvironmentSettingsClientMock{
			getSettingsFunc: func(ctx context.Context, orgName, projectName, envName string) (*pulumiapi.EnvironmentSettings, error) {
				return nil, fmt.Errorf("settings API error")
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		req := pulumirpc.ReadRequest{
			Id: "org/project/env",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get environment settings")
	})
}

func TestEnvironmentCreate(t *testing.T) {
	t.Run("Create with deletion protection enabled", func(t *testing.T) {
		updateSettingsCalled := false
		var capturedDeletionProtected *bool

		mockedClient := &EscClientMock{
			createEnvironmentWithProjectFunc: func(ctx context.Context, orgName, projectName, envName string) error {
				return nil
			},
			updateEnvironmentWithRevisionFunc: func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
				return nil, 1, nil
			},
		}

		settingsClient := &EnvironmentSettingsClientMock{
			updateSettingsFunc: func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
				updateSettingsCalled = true
				capturedDeletionProtected = req.DeletionProtected
				return nil
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		input := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: true,
		}

		propertyMap, _ := input.ToPropertyMap()
		properties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.CreateRequest{
			Properties: properties,
		}

		resp, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.True(t, updateSettingsCalled)
		assert.NotNil(t, capturedDeletionProtected)
		assert.True(t, *capturedDeletionProtected)
		assert.Equal(t, "org/project/env", resp.Id)
	})

	t.Run("Create with deletion protection disabled", func(t *testing.T) {
		updateSettingsCalled := false

		mockedClient := &EscClientMock{
			createEnvironmentWithProjectFunc: func(ctx context.Context, orgName, projectName, envName string) error {
				return nil
			},
			updateEnvironmentWithRevisionFunc: func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
				return nil, 1, nil
			},
		}

		settingsClient := &EnvironmentSettingsClientMock{
			updateSettingsFunc: func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
				updateSettingsCalled = true
				return nil
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		input := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: false,
		}

		propertyMap, _ := input.ToPropertyMap()
		properties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.CreateRequest{
			Properties: properties,
		}

		_, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.False(t, updateSettingsCalled, "UpdateEnvironmentSettings should not be called when deletion protection is disabled")
	})

	t.Run("Create returns error when settings update fails", func(t *testing.T) {
		mockedClient := &EscClientMock{
			createEnvironmentWithProjectFunc: func(ctx context.Context, orgName, projectName, envName string) error {
				return nil
			},
			updateEnvironmentWithRevisionFunc: func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
				return nil, 1, nil
			},
		}

		settingsClient := &EnvironmentSettingsClientMock{
			updateSettingsFunc: func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
				return fmt.Errorf("settings update error")
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		input := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: true,
		}

		propertyMap, _ := input.ToPropertyMap()
		properties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.CreateRequest{
			Properties: properties,
		}

		_, err := provider.Create(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to enable deletion protection")
	})
}

func TestEnvironmentUpdate(t *testing.T) {
	t.Run("Update enables deletion protection", func(t *testing.T) {
		updateSettingsCalled := false
		var capturedDeletionProtected *bool

		mockedClient := &EscClientMock{
			updateEnvironmentWithRevisionFunc: func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
				return nil, 2, nil
			},
		}

		settingsClient := &EnvironmentSettingsClientMock{
			updateSettingsFunc: func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
				updateSettingsCalled = true
				capturedDeletionProtected = req.DeletionProtected
				return nil
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		oldInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: false,
		}
		newInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: true,
		}

		oldPropertyMap, _ := oldInput.ToPropertyMap()
		oldProperties, _ := plugin.MarshalProperties(
			oldPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		newPropertyMap, _ := newInput.ToPropertyMap()
		newProperties, _ := plugin.MarshalProperties(
			newPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)

		req := pulumirpc.UpdateRequest{
			Olds: oldProperties,
			News: newProperties,
		}

		_, err := provider.Update(&req)

		assert.NoError(t, err)
		assert.True(t, updateSettingsCalled)
		assert.NotNil(t, capturedDeletionProtected)
		assert.True(t, *capturedDeletionProtected)
	})

	t.Run("Update disables deletion protection", func(t *testing.T) {
		updateSettingsCalled := false
		var capturedDeletionProtected *bool

		mockedClient := &EscClientMock{
			updateEnvironmentWithRevisionFunc: func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
				return nil, 2, nil
			},
		}

		settingsClient := &EnvironmentSettingsClientMock{
			updateSettingsFunc: func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
				updateSettingsCalled = true
				capturedDeletionProtected = req.DeletionProtected
				return nil
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		oldInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: true,
		}
		newInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: false,
		}

		oldPropertyMap, _ := oldInput.ToPropertyMap()
		oldProperties, _ := plugin.MarshalProperties(
			oldPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		newPropertyMap, _ := newInput.ToPropertyMap()
		newProperties, _ := plugin.MarshalProperties(
			newPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)

		req := pulumirpc.UpdateRequest{
			Olds: oldProperties,
			News: newProperties,
		}

		_, err := provider.Update(&req)

		assert.NoError(t, err)
		assert.True(t, updateSettingsCalled)
		assert.NotNil(t, capturedDeletionProtected)
		assert.False(t, *capturedDeletionProtected)
	})

	t.Run("Update with no change to deletion protection", func(t *testing.T) {
		updateSettingsCalled := false

		mockedClient := &EscClientMock{
			updateEnvironmentWithRevisionFunc: func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
				return nil, 2, nil
			},
		}

		settingsClient := &EnvironmentSettingsClientMock{
			updateSettingsFunc: func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
				updateSettingsCalled = true
				return nil
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		oldInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: true,
		}
		newInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: baz",
			DeletionProtected: true,
		}

		oldPropertyMap, _ := oldInput.ToPropertyMap()
		oldProperties, _ := plugin.MarshalProperties(
			oldPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		newPropertyMap, _ := newInput.ToPropertyMap()
		newProperties, _ := plugin.MarshalProperties(
			newPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)

		req := pulumirpc.UpdateRequest{
			Olds: oldProperties,
			News: newProperties,
		}

		_, err := provider.Update(&req)

		assert.NoError(t, err)
		assert.False(t, updateSettingsCalled, "UpdateEnvironmentSettings should not be called when deletion protection is unchanged")
	})

	t.Run("Update returns error when settings update fails", func(t *testing.T) {
		mockedClient := &EscClientMock{
			updateEnvironmentWithRevisionFunc: func(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
				return nil, 2, nil
			},
		}

		settingsClient := &EnvironmentSettingsClientMock{
			updateSettingsFunc: func(ctx context.Context, orgName, projectName, envName string, req pulumiapi.UpdateEnvironmentSettingsRequest) error {
				return fmt.Errorf("settings update error")
			},
		}

		provider := PulumiServiceEnvironmentResource{
			Client:         mockedClient,
			SettingsClient: settingsClient,
		}

		oldInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: false,
		}
		newInput := PulumiServiceEnvironmentInput{
			OrgName:           "org",
			ProjectName:       "project",
			EnvName:           "env",
			Yaml:              "values:\n  foo: bar",
			DeletionProtected: true,
		}

		oldPropertyMap, _ := oldInput.ToPropertyMap()
		oldProperties, _ := plugin.MarshalProperties(
			oldPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		newPropertyMap, _ := newInput.ToPropertyMap()
		newProperties, _ := plugin.MarshalProperties(
			newPropertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)

		req := pulumirpc.UpdateRequest{
			Olds: oldProperties,
			News: newProperties,
		}

		_, err := provider.Update(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update deletion protection")
	})
}

func TestEnvironmentVersionTag(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, nil
			},
			func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
				return nil, nil
			},
		)

		provider := PulumiServiceEnvironmentVersionTagResource{
			Client: mockedClient,
		}

		input := PulumiServiceEnvironmentVersionTagInput{
			Organization: "org",
			Environment:  "env",
			TagName:      "tag",
			Revision:     1,
		}

		propertyMap := input.ToPropertyMap()
		outputProperties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.ReadRequest{
			Id:         "org/env/tag",
			Properties: outputProperties,
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, nil
			},
			func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
				return &client.EnvironmentRevisionTag{
					Revision: 1,
				}, nil
			},
		)

		provider := PulumiServiceEnvironmentVersionTagResource{
			Client: mockedClient,
		}

		input := PulumiServiceEnvironmentVersionTagInput{
			Organization: "org",
			Environment:  "env",
			TagName:      "tag",
			Revision:     1,
		}

		propertyMap := input.ToPropertyMap()
		outputProperties, _ := plugin.MarshalProperties(
			propertyMap,
			plugin.MarshalOptions{
				KeepUnknowns: true,
				SkipNulls:    true,
			},
		)
		req := pulumirpc.ReadRequest{
			Id:         "org/env/tag",
			Properties: outputProperties,
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "org/env/tag")
	})
}
