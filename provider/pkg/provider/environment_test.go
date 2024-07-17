package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pulumi/esc"
	"github.com/pulumi/esc/cmd/esc/cli/client"
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

func (c *EscClientMock) GetEnvironment(ctx context.Context, orgName string, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
	return c.getEnvironmentFunc(ctx, orgName, envName, version, decrypt)
}

func (c *EscClientMock) GetEnvironmentRevision(ctx context.Context, orgName, envName string, revision int) (*client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentRevisionTag(ctx context.Context, orgName, envName, tagName string) (*client.EnvironmentRevisionTag, error) {
	return c.getEnvironmentRevisionTagFunc(ctx, orgName, envName, tagName)
}

func (c *EscClientMock) GetRevisionNumber(ctx context.Context, orgName, envName, version string) (int, error) {
	return 0, nil
}

func (c *EscClientMock) CheckYAMLEnvironment(context.Context, string, []byte) (*esc.Environment, []client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}

func (c *EscClientMock) CreateEnvironment(context.Context, string, string) error {
	return nil
}

func (c *EscClientMock) DeleteEnvironment(context.Context, string, string) error {
	return nil
}

func (c *EscClientMock) GetOpenEnvironment(context.Context, string, string, string) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenProperty(context.Context, string, string, string, string) (*esc.Value, error) {
	return nil, nil
}

func (c *EscClientMock) GetPulumiAccountDetails(context.Context) (string, []string, *workspace.TokenInformation, error) {
	return "", nil, nil, nil
}

func (c *EscClientMock) ListEnvironments(context.Context, string, string) ([]client.OrgEnvironment, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) ListEnvironmentRevisions(ctx context.Context, orgName string, envName string, options client.ListEnvironmentRevisionsOptions) ([]client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) ListEnvironmentRevisionTags(ctx context.Context, orgName string, envName string, options client.ListEnvironmentRevisionTagsOptions) ([]client.EnvironmentRevisionTag, error) {
	return nil, nil
}

func (c *EscClientMock) OpenEnvironment(context.Context, string, string, string, time.Duration) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) OpenYAMLEnvironment(context.Context, string, []byte, time.Duration) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) UpdateEnvironment(context.Context, string, string, []byte, string) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentRevisionTag(ctx context.Context, orgName, envName, tagName string, revision *int) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironmentWithRevision(ctx context.Context, orgName string, envName string, yaml []byte, etag string) ([]client.EnvironmentDiagnostic, int, error) {
	return nil, 0, nil
}

func (c *EscClientMock) CreateEnvironmentRevisionTag(ctx context.Context, orgName, envName, tagName string, revision *int) error {
	return nil
}

func (c *EscClientMock) DeleteEnvironmentRevisionTag(ctx context.Context, orgName, envName, tagName string) error {
	return nil
}

func (c *EscClientMock) RetractEnvironmentRevision(ctx context.Context, orgName string, envName string, version string, replacement *int, reason string) error {
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
			client: mockedClient,
		}

		input := PulumiServiceEnvironmentInput{
			OrgName: "org",
			EnvName: "env",
			Yaml:    []byte("test-environment"),
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
			client: mockedClient,
		}

		input := PulumiServiceEnvironmentInput{
			OrgName: "org",
			EnvName: "project",
			Yaml:    []byte("test-environment"),
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
			client: mockedClient,
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
			client: mockedClient,
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
