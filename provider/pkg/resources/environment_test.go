package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/esc"
	"github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/asset"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type getEnvironmentFunc func(
	ctx context.Context, orgName string, envName string, version string, decrypt bool,
) (yaml []byte, etag string, revision int, err error)

type getEnvironmentRevisionTagFunc func(
	ctx context.Context, orgName, envName, tagName string,
) (*client.EnvironmentRevisionTag, error)

type EscClientMock struct {
	getEnvironmentFunc            getEnvironmentFunc
	getEnvironmentRevisionTagFunc getEnvironmentRevisionTagFunc
}

func (c *EscClientMock) GetEnvironment(
	ctx context.Context,
	orgName, _ /* projectName */, envName string,
	version string,
	decrypt bool,
) (yaml []byte, etag string, revision int, err error) {
	return c.getEnvironmentFunc(ctx, orgName, envName, version, decrypt)
}

func (c *EscClientMock) GetEnvironmentRevision(
	_ context.Context,
	_, _, _ string,
	_ int,
) (*client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentRevisionTag(
	ctx context.Context,
	orgName, _ /* projectName */, envName, tagName string,
) (*client.EnvironmentRevisionTag, error) {
	return c.getEnvironmentRevisionTagFunc(ctx, orgName, envName, tagName)
}

func (c *EscClientMock) GetRevisionNumber(
	_ context.Context,
	_, _, _, _ string,
) (int, error) {
	return 0, nil
}

func (c *EscClientMock) CheckYAMLEnvironment(
	context.Context,
	string,
	[]byte,
	...client.CheckYAMLOption,
) (*esc.Environment, []client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}

func (c *EscClientMock) CreateEnvironment(context.Context, string, string) error {
	return nil
}

func (c *EscClientMock) CreateEnvironmentWithProject(context.Context, string, string, string) error {
	return nil
}

func (c *EscClientMock) CloneEnvironment(
	context.Context,
	string,
	string,
	string,
	client.CloneEnvironmentRequest,
) error {
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

func (c *EscClientMock) GetOpenEnvironmentWithProject(
	context.Context,
	string,
	string,
	string,
	string,
) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetAnonymousOpenProperty(context.Context, string, string, string) (*esc.Value, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenProperty(context.Context, string, string, string, string, string) (*esc.Value, error) {
	return nil, nil
}

func (c *EscClientMock) GetPulumiAccountDetails(
	context.Context,
) (string, []string, *workspace.TokenInformation, error) {
	return "", nil, nil, nil
}

func (c *EscClientMock) ListEnvironments(context.Context, string) ([]client.OrgEnvironment, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) ListOrganizationEnvironments(
	context.Context,
	string,
	string,
) ([]client.OrgEnvironment, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) ListEnvironmentRevisions(
	_ context.Context,
	_, _, _ string,
	_ client.ListEnvironmentRevisionsOptions,
) ([]client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) ListEnvironmentRevisionTags(
	_ context.Context,
	_, _, _ string,
	_ client.ListEnvironmentRevisionTagsOptions,
) ([]client.EnvironmentRevisionTag, error) {
	return nil, nil
}

func (c *EscClientMock) OpenEnvironment(
	context.Context,
	string,
	string,
	string,
	string,
	time.Duration,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) OpenYAMLEnvironment(
	context.Context,
	string,
	[]byte,
	time.Duration,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) OpenEnvironmentDraft(
	context.Context,
	string,
	string,
	string,
	string,
	time.Duration,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) RotateEnvironment(
	context.Context,
	string,
	string,
	string,
	[]string,
) (*client.RotateEnvironmentResponse, []client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}

func (c *EscClientMock) SubmitChangeRequest(context.Context, string, string, *string) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironment(
	context.Context,
	string,
	string,
	[]byte,
	string,
) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentWithProject(
	context.Context,
	string,
	string,
	string,
	[]byte,
	string,
) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentDraft(
	context.Context,
	string,
	string,
	string,
	string,
	[]byte,
	string,
) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) CreateEnvironmentDraft(
	context.Context,
	string,
	string,
	string,
	[]byte,
	string,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) GetDefaultOrg(context.Context) (string, error) {
	return "", nil
}

func (c *EscClientMock) GetEnvironmentDraft(context.Context, string, string, string, string) ([]byte, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) CreateEnvironmentTag(
	context.Context,
	string,
	string,
	string,
	string,
	string,
) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentTag(
	context.Context,
	string,
	string,
	string,
	string,
) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) ListEnvironmentTags(
	context.Context,
	string,
	string,
	string,
	client.ListEnvironmentTagsOptions,
) ([]*client.EnvironmentTag, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) UpdateEnvironmentTag(
	context.Context,
	string,
	string,
	string,
	string,
	string,
	string,
	string,
) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) DeleteEnvironmentTag(context.Context, string, string, string, string) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironmentRevisionTag(
	_ context.Context,
	_, _, _, _ string,
	_ *int,
) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironmentWithRevision(
	_ context.Context,
	_, _, _ string,
	_ []byte,
	_ string,
) ([]client.EnvironmentDiagnostic, int, error) {
	return nil, 0, nil
}

func (c *EscClientMock) CreateEnvironmentRevisionTag(
	_ context.Context,
	_, _, _, _ string,
	_ *int,
) error {
	return nil
}

func (c *EscClientMock) DeleteEnvironmentRevisionTag(
	_ context.Context,
	_, _, _, _ string,
) error {
	return nil
}

func (c *EscClientMock) RetractEnvironmentRevision(
	_ context.Context,
	_, _, _ string,
	_ string,
	_ *int,
	_ string,
) error {
	return nil
}

func (c *EscClientMock) Insecure() bool {
	return false
}

func (c *EscClientMock) URL() string {
	return ""
}

func buildEscClientMock(
	getEnvironmentFunc getEnvironmentFunc,
	getEnvironmentRevisionTagFunc getEnvironmentRevisionTagFunc,
) *EscClientMock {
	return &EscClientMock{
		getEnvironmentFunc,
		getEnvironmentRevisionTagFunc,
	}
}

func TestEnvironmentCheck(t *testing.T) {
	mockedClient := buildEscClientMock(
		func(_ context.Context, _ string, _ string, _ string, _ bool) (yaml []byte, etag string, revision int, err error) {
			return nil, "", 0, fmt.Errorf("not found")
		},
		func(_ context.Context, _, _, _ string) (*client.EnvironmentRevisionTag, error) {
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
			func(_ context.Context, _ string, _ string, _ string, _ bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, fmt.Errorf("not found")
			},
			func(_ context.Context, _, _, _ string) (*client.EnvironmentRevisionTag, error) {
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
			func(_ context.Context, _ string, _ string, _ string, _ bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, nil
			},
			func(_ context.Context, _, _, _ string) (*client.EnvironmentRevisionTag, error) {
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
}

func TestEnvironmentDiff(t *testing.T) {
	mockedClient := buildEscClientMock(
		func(_ context.Context, _ string, _ string, _ string, _ bool) (yaml []byte, etag string, revision int, err error) {
			return nil, "", 0, nil
		},
		func(_ context.Context, _, _, _ string) (*client.EnvironmentRevisionTag, error) {
			return nil, nil
		},
	)
	provider := PulumiServiceEnvironmentResource{
		Client: mockedClient,
	}

	t.Run("No diff when inputs are identical", func(t *testing.T) {
		inputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env"),
			"yaml":         resource.NewStringProperty("values:\n  foo: bar"),
		}

		state, err := structpb.NewStruct(inputs.Mappable())
		require.NoError(t, err)

		req := &pulumirpc.DiffRequest{
			Id:        "org/default/env",
			Urn:       "urn:pulumi:dev::test::pulumiservice:index:Environment::testEnv",
			OldInputs: state,
			News:      state,
		}

		resp, err := provider.Diff(req)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
		assert.Empty(t, resp.DetailedDiff)
	})

	t.Run("Replace when organization changes", func(t *testing.T) {
		oldInputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org1"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env"),
			"yaml":         resource.NewStringProperty("values:\n  foo: bar"),
		}
		newInputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org2"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env"),
			"yaml":         resource.NewStringProperty("values:\n  foo: bar"),
		}

		oldState, err := structpb.NewStruct(oldInputs.Mappable())
		require.NoError(t, err)
		newState, err := structpb.NewStruct(newInputs.Mappable())
		require.NoError(t, err)

		req := &pulumirpc.DiffRequest{
			Id:        "org1/default/env",
			Urn:       "urn:pulumi:dev::test::pulumiservice:index:Environment::testEnv",
			OldInputs: oldState,
			News:      newState,
		}

		resp, err := provider.Diff(req)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)

		orgDiff, ok := resp.DetailedDiff["organization"]
		assert.True(t, ok, "organization should be in the detailed diff")
		assert.Equal(t, pulumirpc.PropertyDiff_UPDATE_REPLACE, orgDiff.Kind)
	})

	t.Run("Replace when name changes", func(t *testing.T) {
		oldInputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env1"),
			"yaml":         resource.NewStringProperty("values:\n  foo: bar"),
		}
		newInputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env2"),
			"yaml":         resource.NewStringProperty("values:\n  foo: bar"),
		}

		oldState, err := structpb.NewStruct(oldInputs.Mappable())
		require.NoError(t, err)
		newState, err := structpb.NewStruct(newInputs.Mappable())
		require.NoError(t, err)

		req := &pulumirpc.DiffRequest{
			Id:        "org/default/env1",
			Urn:       "urn:pulumi:dev::test::pulumiservice:index:Environment::testEnv",
			OldInputs: oldState,
			News:      newState,
		}

		resp, err := provider.Diff(req)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)

		nameDiff, ok := resp.DetailedDiff["name"]
		assert.True(t, ok, "name should be in the detailed diff")
		assert.Equal(t, pulumirpc.PropertyDiff_UPDATE_REPLACE, nameDiff.Kind)
	})

	t.Run("No replace when only yaml changes", func(t *testing.T) {
		oldInputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env"),
			"yaml":         resource.NewStringProperty("values:\n  foo: bar"),
		}
		newInputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env"),
			"yaml":         resource.NewStringProperty("values:\n  foo: baz"),
		}

		oldState, err := structpb.NewStruct(oldInputs.Mappable())
		require.NoError(t, err)
		newState, err := structpb.NewStruct(newInputs.Mappable())
		require.NoError(t, err)

		req := &pulumirpc.DiffRequest{
			Id:        "org/default/env",
			Urn:       "urn:pulumi:dev::test::pulumiservice:index:Environment::testEnv",
			OldInputs: oldState,
			News:      newState,
		}

		resp, err := provider.Diff(req)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)

		yamlDiff, ok := resp.DetailedDiff["yaml"]
		assert.True(t, ok, "yaml should be in the detailed diff")
		assert.Equal(t, pulumirpc.PropertyDiff_UPDATE, yamlDiff.Kind, "yaml change should be UPDATE, not REPLACE")
	})

	t.Run("No diff when provider version changes but inputs are same", func(t *testing.T) {
		// This is the core regression test: when upgrading the provider version,
		// the diff should show no changes if the user's inputs haven't changed.
		inputs := resource.PropertyMap{
			"organization": resource.NewStringProperty("org"),
			"project":      resource.NewStringProperty("default"),
			"name":         resource.NewStringProperty("env"),
			"yaml":         resource.NewStringProperty("values:\n  foo: bar"),
		}

		state, err := structpb.NewStruct(inputs.Mappable())
		require.NoError(t, err)

		req := &pulumirpc.DiffRequest{
			Id:        "org/default/env",
			Urn:       "urn:pulumi:dev::test::pulumiservice:index:Environment::testEnv",
			OldInputs: state,
			News:      state,
		}

		resp, err := provider.Diff(req)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
		assert.Empty(t, resp.DetailedDiff)
	})
}

func TestEnvironmentVersionTag(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildEscClientMock(
			func(_ context.Context, _ string, _ string, _ string, _ bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, nil
			},
			func(_ context.Context, _, _, _ string) (*client.EnvironmentRevisionTag, error) {
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
			func(_ context.Context, _ string, _ string, _ string, _ bool) (yaml []byte, etag string, revision int, err error) {
				return nil, "", 0, nil
			},
			func(_ context.Context, _, _, _ string) (*client.EnvironmentRevisionTag, error) {
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
