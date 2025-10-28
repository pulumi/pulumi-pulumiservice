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

type createStackFunc func(ctx context.Context, stack pulumiapi.StackIdentifier) error
type deleteStackFunc func(ctx context.Context, stack pulumiapi.StackIdentifier, forceDestroy bool) error
type stackExistsFunc func(ctx context.Context, stack pulumiapi.StackIdentifier) (bool, error)

type StackClientMock struct {
	createStackFunc createStackFunc
	deleteStackFunc deleteStackFunc
	stackExistsFunc stackExistsFunc
}

func (c *StackClientMock) CreateStack(ctx context.Context, stack pulumiapi.StackIdentifier) error {
	return c.createStackFunc(ctx, stack)
}

func (c *StackClientMock) DeleteStack(ctx context.Context, stack pulumiapi.StackIdentifier, forceDestroy bool) error {
	return c.deleteStackFunc(ctx, stack, forceDestroy)
}

func (c *StackClientMock) StackExists(ctx context.Context, stack pulumiapi.StackIdentifier) (bool, error) {
	return c.stackExistsFunc(ctx, stack)
}

func buildStackClientMock(
	createFunc createStackFunc,
	deleteFunc deleteStackFunc,
	existsFunc stackExistsFunc,
) *StackClientMock {
	return &StackClientMock{
		createStackFunc: createFunc,
		deleteStackFunc: deleteFunc,
		stackExistsFunc: existsFunc,
	}
}

func TestStack(t *testing.T) {
	t.Run("Read when stack not found", func(t *testing.T) {
		mockedClient := buildStackClientMock(
			nil,
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier) (bool, error) {
				return false, nil
			},
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-project/dev",
			Urn: "urn:pulumi:test::test::pulumiservice:index:Stack::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.Id)
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when stack found", func(t *testing.T) {
		mockedClient := buildStackClientMock(
			nil,
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier) (bool, error) {
				assert.Equal(t, "test-org", stack.OrgName)
				assert.Equal(t, "test-project", stack.ProjectName)
				assert.Equal(t, "dev", stack.StackName)
				return true, nil
			},
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-project/dev",
			Urn: "urn:pulumi:test::test::pulumiservice:index:Stack::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-project/dev", resp.Id)
		assert.NotNil(t, resp.Properties)

		// Verify properties
		props, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{})
		assert.Equal(t, "test-org", props["organizationName"].StringValue())
		assert.Equal(t, "test-project", props["projectName"].StringValue())
		assert.Equal(t, "dev", props["stackName"].StringValue())
	})

	t.Run("Read with API error", func(t *testing.T) {
		mockedClient := buildStackClientMock(
			nil,
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier) (bool, error) {
				return false, errors.New("API error")
			},
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-project/dev",
			Urn: "urn:pulumi:test::test::pulumiservice:index:Stack::test",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failure while checking if stack")
	})

	t.Run("Create successful", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})

		mockedClient := buildStackClientMock(
			func(ctx context.Context, stack pulumiapi.StackIdentifier) error {
				assert.Equal(t, "test-org", stack.OrgName)
				assert.Equal(t, "test-project", stack.ProjectName)
				assert.Equal(t, "dev", stack.StackName)
				return nil
			},
			nil,
			nil,
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			Properties: inputProps,
		}

		resp, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-project/dev", resp.Id)
		assert.NotNil(t, resp.Properties)
	})

	t.Run("Create with API error", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})

		mockedClient := buildStackClientMock(
			func(ctx context.Context, stack pulumiapi.StackIdentifier) error {
				return errors.New("create failed")
			},
			nil,
			nil,
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			Properties: inputProps,
		}

		_, err := provider.Create(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
	})

	t.Run("Delete successful", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
			"forceDestroy":     resource.NewPropertyValue(false),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})

		mockedClient := buildStackClientMock(
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, forceDestroy bool) error {
				assert.Equal(t, "test-org", stack.OrgName)
				assert.Equal(t, "test-project", stack.ProjectName)
				assert.Equal(t, "dev", stack.StackName)
				assert.False(t, forceDestroy)
				return nil
			},
			nil,
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:         "test-org/test-project/dev",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			Properties: inputProps,
		}

		resp, err := provider.Delete(&req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Delete with forceDestroy", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
			"forceDestroy":     resource.NewPropertyValue(true),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})

		mockedClient := buildStackClientMock(
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, forceDestroy bool) error {
				assert.True(t, forceDestroy)
				return nil
			},
			nil,
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:         "test-org/test-project/dev",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			Properties: inputProps,
		}

		resp, err := provider.Delete(&req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Delete with API error", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})

		mockedClient := buildStackClientMock(
			nil,
			func(ctx context.Context, stack pulumiapi.StackIdentifier, forceDestroy bool) error {
				return errors.New("delete failed")
			},
			nil,
		)

		provider := PulumiServiceStackResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:         "test-org/test-project/dev",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			Properties: inputProps,
		}

		_, err := provider.Delete(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("Diff with no changes", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
		}

		olds, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		news, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})

		provider := PulumiServiceStackResource{}

		req := pulumirpc.DiffRequest{
			Urn:       "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			Id:        "test-org/test-project/dev",
			OldInputs: olds,
			News:      news,
		}

		resp, err := provider.Diff(&req)

		assert.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
	})

	t.Run("Diff with property changes", func(t *testing.T) {
		oldMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
		}

		newMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("prod"),
		}

		olds, _ := plugin.MarshalProperties(oldMap, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
		news, _ := plugin.MarshalProperties(newMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})

		provider := PulumiServiceStackResource{}

		req := pulumirpc.DiffRequest{
			Urn:       "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			Id:        "test-org/test-project/dev",
			OldInputs: olds,
			News:      news,
		}

		resp, err := provider.Diff(&req)

		assert.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.True(t, resp.DeleteBeforeReplace)
		assert.True(t, resp.HasDetailedDiff)
		assert.NotEmpty(t, resp.DetailedDiff)
	})

	t.Run("Check with valid input", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organizationName": resource.NewPropertyValue("test-org"),
			"projectName":      resource.NewPropertyValue("test-project"),
			"stackName":        resource.NewPropertyValue("dev"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceStackResource{}

		req := pulumirpc.CheckRequest{
			Urn:  "urn:pulumi:test::test::pulumiservice:index:Stack::test",
			News: inputProps,
		}

		resp, err := provider.Check(&req)

		assert.NoError(t, err)
		assert.Equal(t, inputProps, resp.Inputs)
		assert.Nil(t, resp.Failures)
	})

	t.Run("Update returns error", func(t *testing.T) {
		provider := PulumiServiceStackResource{}

		req := pulumirpc.UpdateRequest{
			Id:  "test-org/test-project/dev",
			Urn: "urn:pulumi:test::test::pulumiservice:index:Stack::test",
		}

		_, err := provider.Update(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected call to update")
	})
}
