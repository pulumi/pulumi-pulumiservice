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

type addEnvironmentSettingsFunc func(ctx context.Context, req pulumiapi.CreateTeamEnvironmentSettingsRequest) error
type removeEnvironmentSettingsFunc func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) error
type getTeamEnvironmentSettingsFunc func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) (*string, *pulumiapi.Duration, error)

type TeamEnvironmentPermissionClientMock struct {
	TeamClientMock
	addEnvironmentSettingsFunc    addEnvironmentSettingsFunc
	removeEnvironmentSettingsFunc removeEnvironmentSettingsFunc
	getTeamEnvironmentSettingsFunc getTeamEnvironmentSettingsFunc
}

func (c *TeamEnvironmentPermissionClientMock) AddEnvironmentSettings(ctx context.Context, req pulumiapi.CreateTeamEnvironmentSettingsRequest) error {
	if c.addEnvironmentSettingsFunc != nil {
		return c.addEnvironmentSettingsFunc(ctx, req)
	}
	return nil
}

func (c *TeamEnvironmentPermissionClientMock) RemoveEnvironmentSettings(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) error {
	if c.removeEnvironmentSettingsFunc != nil {
		return c.removeEnvironmentSettingsFunc(ctx, req)
	}
	return nil
}

func (c *TeamEnvironmentPermissionClientMock) GetTeamEnvironmentSettings(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) (*string, *pulumiapi.Duration, error) {
	if c.getTeamEnvironmentSettingsFunc != nil {
		return c.getTeamEnvironmentSettingsFunc(ctx, req)
	}
	return nil, nil, nil
}

func buildTeamEnvironmentPermissionClientMock(
	addFunc addEnvironmentSettingsFunc,
	removeFunc removeEnvironmentSettingsFunc,
	getFunc getTeamEnvironmentSettingsFunc,
) *TeamEnvironmentPermissionClientMock {
	return &TeamEnvironmentPermissionClientMock{
		addEnvironmentSettingsFunc:    addFunc,
		removeEnvironmentSettingsFunc: removeFunc,
		getTeamEnvironmentSettingsFunc: getFunc,
	}
}

func TestTeamEnvironmentPermission(t *testing.T) {
	t.Run("Read when permission not found", func(t *testing.T) {
		mockedClient := buildTeamEnvironmentPermissionClientMock(
			nil,
			nil,
			func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) (*string, *pulumiapi.Duration, error) {
				return nil, nil, nil
			},
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/ops-team/infrastructure+production",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.Id)
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when permission found", func(t *testing.T) {
		permission := "write"
		duration := pulumiapi.Duration(8 * 3600 * 1000000000) // 8 hours in nanoseconds

		mockedClient := buildTeamEnvironmentPermissionClientMock(
			nil,
			nil,
			func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) (*string, *pulumiapi.Duration, error) {
				assert.Equal(t, "test-org", req.Organization)
				assert.Equal(t, "ops-team", req.Team)
				assert.Equal(t, "infrastructure", req.Project)
				assert.Equal(t, "production", req.Environment)
				return &permission, &duration, nil
			},
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/ops-team/infrastructure+production",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/ops-team/infrastructure+production", resp.Id)
		assert.NotNil(t, resp.Properties)

		// Verify properties
		props, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{})
		assert.Equal(t, "test-org", props["organization"].StringValue())
		assert.Equal(t, "ops-team", props["team"].StringValue())
		assert.Equal(t, "infrastructure", props["project"].StringValue())
		assert.Equal(t, "production", props["environment"].StringValue())
		assert.Equal(t, "write", props["permission"].StringValue())
		assert.Equal(t, "8h0m0s", props["maxOpenDuration"].StringValue())
	})

	t.Run("Read with default project ID", func(t *testing.T) {
		permission := "read"

		mockedClient := buildTeamEnvironmentPermissionClientMock(
			nil,
			nil,
			func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) (*string, *pulumiapi.Duration, error) {
				assert.Equal(t, "test-org", req.Organization)
				assert.Equal(t, "ops-team", req.Team)
				assert.Equal(t, "default", req.Project)
				assert.Equal(t, "production", req.Environment)
				return &permission, nil, nil
			},
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/ops-team/production",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/ops-team/production", resp.Id)
		assert.NotNil(t, resp.Properties)

		// Verify properties
		props, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{})
		assert.Equal(t, "default", props["project"].StringValue())
	})

	t.Run("Read with API error", func(t *testing.T) {
		mockedClient := buildTeamEnvironmentPermissionClientMock(
			nil,
			nil,
			func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) (*string, *pulumiapi.Duration, error) {
				return nil, nil, errors.New("API error")
			},
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/ops-team/infrastructure+production",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get team environment permission")
	})

	t.Run("Create successful", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"team":         resource.NewPropertyValue("ops-team"),
			"project":      resource.NewPropertyValue("infrastructure"),
			"environment":  resource.NewPropertyValue("production"),
			"permission":   resource.NewPropertyValue("write"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamEnvironmentPermissionClientMock(
			func(ctx context.Context, req pulumiapi.CreateTeamEnvironmentSettingsRequest) error {
				assert.Equal(t, "test-org", req.Organization)
				assert.Equal(t, "ops-team", req.Team)
				assert.Equal(t, "infrastructure", req.Project)
				assert.Equal(t, "production", req.Environment)
				assert.Equal(t, "write", req.Permission)
				assert.Nil(t, req.MaxOpenDuration)
				return nil
			},
			nil,
			nil,
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			Properties: inputProps,
		}

		resp, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/ops-team/infrastructure+production", resp.Id)
		assert.NotNil(t, resp.Properties)
	})

	t.Run("Create with maxOpenDuration", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization":    resource.NewPropertyValue("test-org"),
			"team":            resource.NewPropertyValue("ops-team"),
			"project":         resource.NewPropertyValue("infrastructure"),
			"environment":     resource.NewPropertyValue("production"),
			"permission":      resource.NewPropertyValue("write"),
			"maxOpenDuration": resource.NewPropertyValue("8h0m0s"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamEnvironmentPermissionClientMock(
			func(ctx context.Context, req pulumiapi.CreateTeamEnvironmentSettingsRequest) error {
				assert.NotNil(t, req.MaxOpenDuration)
				return nil
			},
			nil,
			nil,
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			Properties: inputProps,
		}

		resp, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/ops-team/infrastructure+production", resp.Id)
		assert.NotNil(t, resp.Properties)
	})

	t.Run("Create with API error", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"team":         resource.NewPropertyValue("ops-team"),
			"project":      resource.NewPropertyValue("infrastructure"),
			"environment":  resource.NewPropertyValue("production"),
			"permission":   resource.NewPropertyValue("write"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamEnvironmentPermissionClientMock(
			func(ctx context.Context, req pulumiapi.CreateTeamEnvironmentSettingsRequest) error {
				return errors.New("create failed")
			},
			nil,
			nil,
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			Properties: inputProps,
		}

		_, err := provider.Create(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
	})

	t.Run("Delete successful", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"team":         resource.NewPropertyValue("ops-team"),
			"project":      resource.NewPropertyValue("infrastructure"),
			"environment":  resource.NewPropertyValue("production"),
			"permission":   resource.NewPropertyValue("write"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamEnvironmentPermissionClientMock(
			nil,
			func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) error {
				assert.Equal(t, "test-org", req.Organization)
				assert.Equal(t, "ops-team", req.Team)
				assert.Equal(t, "infrastructure", req.Project)
				assert.Equal(t, "production", req.Environment)
				return nil
			},
			nil,
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:         "test-org/ops-team/infrastructure+production",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			Properties: inputProps,
		}

		resp, err := provider.Delete(&req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Delete with API error", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"team":         resource.NewPropertyValue("ops-team"),
			"project":      resource.NewPropertyValue("infrastructure"),
			"environment":  resource.NewPropertyValue("production"),
			"permission":   resource.NewPropertyValue("write"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamEnvironmentPermissionClientMock(
			nil,
			func(ctx context.Context, req pulumiapi.TeamEnvironmentSettingsRequest) error {
				return errors.New("delete failed")
			},
			nil,
		)

		provider := PulumiServiceTeamEnvironmentPermissionResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:         "test-org/ops-team/infrastructure+production",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			Properties: inputProps,
		}

		_, err := provider.Delete(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("Diff with no changes", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"team":         resource.NewPropertyValue("ops-team"),
			"project":      resource.NewPropertyValue("infrastructure"),
			"environment":  resource.NewPropertyValue("production"),
			"permission":   resource.NewPropertyValue("write"),
		}

		olds, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceTeamEnvironmentPermissionResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			Id:  "test-org/ops-team/infrastructure+production",
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
			"team":         resource.NewPropertyValue("ops-team"),
			"project":      resource.NewPropertyValue("infrastructure"),
			"environment":  resource.NewPropertyValue("production"),
			"permission":   resource.NewPropertyValue("read"),
		}

		newMap := resource.PropertyMap{
			"organization": resource.NewPropertyValue("test-org"),
			"team":         resource.NewPropertyValue("ops-team"),
			"project":      resource.NewPropertyValue("infrastructure"),
			"environment":  resource.NewPropertyValue("production"),
			"permission":   resource.NewPropertyValue("write"),
		}

		olds, _ := plugin.MarshalProperties(oldMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(newMap, plugin.MarshalOptions{})

		provider := PulumiServiceTeamEnvironmentPermissionResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			Id:  "test-org/ops-team/infrastructure+production",
			Olds: olds,
			News: news,
		}

		resp, err := provider.Diff(&req)

		assert.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.True(t, resp.DeleteBeforeReplace)
		assert.NotEmpty(t, resp.Replaces)
	})

	t.Run("Check with valid input and valid duration", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization":    resource.NewPropertyValue("test-org"),
			"team":            resource.NewPropertyValue("ops-team"),
			"project":         resource.NewPropertyValue("infrastructure"),
			"environment":     resource.NewPropertyValue("production"),
			"permission":      resource.NewPropertyValue("write"),
			"maxOpenDuration": resource.NewPropertyValue("8h"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceTeamEnvironmentPermissionResource{}

		req := pulumirpc.CheckRequest{
			Urn:  "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			News: inputProps,
		}

		resp, err := provider.Check(&req)

		assert.NoError(t, err)
		assert.Nil(t, resp.Failures)

		// Verify the duration was normalized
		inputs, _ := plugin.UnmarshalProperties(resp.Inputs, plugin.MarshalOptions{})
		assert.Equal(t, "8h0m0s", inputs["maxOpenDuration"].StringValue())
	})

	t.Run("Check with invalid duration", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"organization":    resource.NewPropertyValue("test-org"),
			"team":            resource.NewPropertyValue("ops-team"),
			"project":         resource.NewPropertyValue("infrastructure"),
			"environment":     resource.NewPropertyValue("production"),
			"permission":      resource.NewPropertyValue("write"),
			"maxOpenDuration": resource.NewPropertyValue("invalid"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceTeamEnvironmentPermissionResource{}

		req := pulumirpc.CheckRequest{
			Urn:  "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
			News: inputProps,
		}

		resp, err := provider.Check(&req)

		assert.NoError(t, err)
		assert.NotNil(t, resp.Failures)
		assert.Len(t, resp.Failures, 1)
		assert.Equal(t, "maxOpenDuration", resp.Failures[0].Property)
		assert.Contains(t, resp.Failures[0].Reason, "malformed duration")
	})

	t.Run("Update returns error", func(t *testing.T) {
		provider := PulumiServiceTeamEnvironmentPermissionResource{}

		req := pulumirpc.UpdateRequest{
			Id:  "test-org/ops-team/infrastructure+production",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamEnvironmentPermission::test",
		}

		_, err := provider.Update(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected call to update")
	})
}
