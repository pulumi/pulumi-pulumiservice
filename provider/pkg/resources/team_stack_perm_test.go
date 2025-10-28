// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

type addStackPermissionFunc func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string, permission int) error
type removeStackPermissionFunc func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) error
type getTeamStackPermissionFunc func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) (*int, error)

type TeamStackPermissionClientMock struct {
	TeamClientMock
	addStackPermissionFunc    addStackPermissionFunc
	removeStackPermissionFunc removeStackPermissionFunc
	getTeamStackPermissionFunc getTeamStackPermissionFunc
}

func (c *TeamStackPermissionClientMock) AddStackPermission(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string, permission int) error {
	if c.addStackPermissionFunc != nil {
		return c.addStackPermissionFunc(ctx, stack, teamName, permission)
	}
	return nil
}

func (c *TeamStackPermissionClientMock) RemoveStackPermission(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) error {
	if c.removeStackPermissionFunc != nil {
		return c.removeStackPermissionFunc(ctx, stack, teamName)
	}
	return nil
}

func (c *TeamStackPermissionClientMock) GetTeamStackPermission(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) (*int, error) {
	if c.getTeamStackPermissionFunc != nil {
		return c.getTeamStackPermissionFunc(ctx, stack, teamName)
	}
	return nil, nil
}

func buildTeamStackPermissionClientMock(
	addFunc addStackPermissionFunc,
	removeFunc removeStackPermissionFunc,
	getFunc getTeamStackPermissionFunc,
) *TeamStackPermissionClientMock {
	return &TeamStackPermissionClientMock{
		addStackPermissionFunc:    addFunc,
		removeStackPermissionFunc: removeFunc,
		getTeamStackPermissionFunc: getFunc,
	}
}

func TestTeamStackPermission(t *testing.T) {
	t.Run("Read when permission not found", func(t *testing.T) {
		mockedClient := buildTeamStackPermissionClientMock(
			nil,
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) (*int, error) {
				return nil, nil
			},
		)

		provider := TeamStackPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-project/dev/developers",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.Id)
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when permission found", func(t *testing.T) {
		permission := 100

		mockedClient := buildTeamStackPermissionClientMock(
			nil,
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) (*int, error) {
				assert.Equal(t, "test-org", stack.OrgName)
				assert.Equal(t, "test-project", stack.ProjectName)
				assert.Equal(t, "dev", stack.StackName)
				assert.Equal(t, "developers", teamName)
				return &permission, nil
			},
		)

		provider := TeamStackPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-project/dev/developers",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-project/dev/developers", resp.Id)
		assert.NotNil(t, resp.Properties)

		// Verify properties
		props, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{})
		assert.Equal(t, "test-org", props["organization"].StringValue())
		assert.Equal(t, "test-project", props["project"].StringValue())
		assert.Equal(t, "dev", props["stack"].StringValue())
		assert.Equal(t, "developers", props["team"].StringValue())
		assert.Equal(t, float64(100), props["permission"].NumberValue())
	})

	t.Run("Read with API error", func(t *testing.T) {
		mockedClient := buildTeamStackPermissionClientMock(
			nil,
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) (*int, error) {
				return nil, errors.New("API error")
			},
		)

		provider := TeamStackPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-project/dev/developers",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get team stack permission")
	})

	t.Run("Read with legacy ID format", func(t *testing.T) {
		provider := TeamStackPermissionResource{
			Client: &TeamStackPermissionClientMock{},
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-project/developers", // 3 parts instead of 4
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TeamStackPermission resources created before v0.17.0")
		assert.Contains(t, err.Error(), "do not support refresh")
	})

	t.Run("Create successful", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(100),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamStackPermissionClientMock(
			func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string, permission int) error {
				assert.Equal(t, "test-org", stack.OrgName)
				assert.Equal(t, "test-project", stack.ProjectName)
				assert.Equal(t, "dev", stack.StackName)
				assert.Equal(t, "developers", teamName)
				assert.Equal(t, 100, permission)
				return nil
			},
			nil,
			nil,
		)

		provider := TeamStackPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
			Properties: inputProps,
		}

		resp, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-project/dev/developers", resp.Id)
		assert.NotNil(t, resp.Properties)
	})

	t.Run("Create with API error", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(100),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamStackPermissionClientMock(
			func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string, permission int) error {
				return errors.New("create failed")
			},
			nil,
			nil,
		)

		provider := TeamStackPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
			Properties: inputProps,
		}

		_, err := provider.Create(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
	})

	t.Run("Delete successful", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(100),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamStackPermissionClientMock(
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) error {
				assert.Equal(t, "test-org", stack.OrgName)
				assert.Equal(t, "test-project", stack.ProjectName)
				assert.Equal(t, "dev", stack.StackName)
				assert.Equal(t, "developers", teamName)
				return nil
			},
			nil,
		)

		provider := TeamStackPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:         "test-org/test-project/dev/developers",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
			Properties: inputProps,
		}

		resp, err := provider.Delete(&req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Delete with API error", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(100),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamStackPermissionClientMock(
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, teamName string) error {
				return errors.New("delete failed")
			},
			nil,
		)

		provider := TeamStackPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:         "test-org/test-project/dev/developers",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
			Properties: inputProps,
		}

		_, err := provider.Delete(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("Diff with no changes", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(100),
		}

		olds, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := TeamStackPermissionResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
			Id:  "test-org/test-project/dev/developers",
			Olds: olds,
			News: news,
		}

		resp, err := provider.Diff(&req)

		assert.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
	})

	t.Run("Diff with property changes", func(t *testing.T) {
		oldMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(100),
		}

		newMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(200),
		}

		olds, _ := plugin.MarshalProperties(oldMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(newMap, plugin.MarshalOptions{})

		provider := TeamStackPermissionResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
			Id:  "test-org/test-project/dev/developers",
			Olds: olds,
			News: news,
		}

		resp, err := provider.Diff(&req)

		assert.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.NotEmpty(t, resp.Replaces)
	})

	t.Run("Check with valid input", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"project":      resource.NewPropertyValue("test-project"),
			"stack":        resource.NewPropertyValue("dev"),
			"team":         resource.NewPropertyValue("developers"),
			"permission":   resource.NewPropertyValue(100),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := TeamStackPermissionResource{}

		req := pulumirpc.CheckRequest{
			Urn:  "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
			News: inputProps,
		}

		resp, err := provider.Check(&req)

		assert.NoError(t, err)
		assert.Equal(t, inputProps, resp.Inputs)
	})

	t.Run("Update returns error", func(t *testing.T) {
		provider := TeamStackPermissionResource{}

		req := pulumirpc.UpdateRequest{
			Id:  "test-org/test-project/dev/developers",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamStackPermission::test",
		}

		_, err := provider.Update(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected call to update")
	})
}
