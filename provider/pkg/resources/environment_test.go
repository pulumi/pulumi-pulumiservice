package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pulumi/esc"
	"github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/asset"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type getEnvironmentFunc func(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error)
type getEnvironmentRevisionTagFunc func(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error)

type EscClientMock struct {
	getEnvironmentFunc            getEnvironmentFunc
	getEnvironmentRevisionTagFunc getEnvironmentRevisionTagFunc
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

func (c *EscClientMock) CreateEnvironmentWithProject(context.Context, string, string, string) error {
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

func (c *EscClientMock) ListEnvironments(context.Context, string, string) ([]client.OrgEnvironment, string, error) {
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

func (c *EscClientMock) UpdateEnvironment(context.Context, string, string, []byte, string) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentWithProject(context.Context, string, string, string, []byte, string) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
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
		getEnvironmentFunc,
		getEnvironmentRevisionTagFunc,
	}
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

		propertyMap := util.ToPropertyMap(input)
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

		propertyMap := util.ToPropertyMap(input)
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
