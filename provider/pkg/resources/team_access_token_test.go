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

type createTeamAccessTokenFunc func(ctx context.Context, name, orgName, teamName, description string) (*pulumiapi.AccessToken, error)
type deleteTeamAccessTokenFunc func(ctx context.Context, tokenId, orgName, teamName string) error
type getTeamAccessTokenFunc func(ctx context.Context, tokenId, orgName, teamName string) (*pulumiapi.AccessToken, error)

type TeamAccessTokenClientMock struct {
	createTeamAccessTokenFunc createTeamAccessTokenFunc
	deleteTeamAccessTokenFunc deleteTeamAccessTokenFunc
	getTeamAccessTokenFunc    getTeamAccessTokenFunc
}

func (c *TeamAccessTokenClientMock) CreateTeamAccessToken(ctx context.Context, name, orgName, teamName, description string) (*pulumiapi.AccessToken, error) {
	return c.createTeamAccessTokenFunc(ctx, name, orgName, teamName, description)
}

func (c *TeamAccessTokenClientMock) DeleteTeamAccessToken(ctx context.Context, tokenId, orgName, teamName string) error {
	return c.deleteTeamAccessTokenFunc(ctx, tokenId, orgName, teamName)
}

func (c *TeamAccessTokenClientMock) GetTeamAccessToken(ctx context.Context, tokenId, orgName, teamName string) (*pulumiapi.AccessToken, error) {
	return c.getTeamAccessTokenFunc(ctx, tokenId, orgName, teamName)
}

func buildTeamAccessTokenClientMock(
	createFunc createTeamAccessTokenFunc,
	deleteFunc deleteTeamAccessTokenFunc,
	getFunc getTeamAccessTokenFunc,
) *TeamAccessTokenClientMock {
	return &TeamAccessTokenClientMock{
		createTeamAccessTokenFunc: createFunc,
		deleteTeamAccessTokenFunc: deleteFunc,
		getTeamAccessTokenFunc:    getFunc,
	}
}

func TestTeamAccessToken(t *testing.T) {
	t.Run("Read when token not found", func(t *testing.T) {
		mockedClient := buildTeamAccessTokenClientMock(
			nil,
			nil,
			func(ctx context.Context, tokenId, orgName, teamName string) (*pulumiapi.AccessToken, error) {
				return nil, nil
			},
		)

		provider := PulumiServiceTeamAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-team/my-token/tat-123",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.Id)
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when token found", func(t *testing.T) {
		tokenValue := "pul-team-987654321fedcba"
		description := "test team token"

		inputMap := resource.PropertyMap{
			"name":             resource.NewPropertyValue("my-token"),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue(description),
		}

		outputMap := inputMap.Copy()
		outputMap["__inputs"] = resource.NewObjectProperty(inputMap)
		outputMap["value"] = resource.MakeSecret(resource.NewPropertyValue(tokenValue))

		existingProps, _ := plugin.MarshalProperties(outputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamAccessTokenClientMock(
			nil,
			nil,
			func(ctx context.Context, tokenId, orgName, teamName string) (*pulumiapi.AccessToken, error) {
				assert.Equal(t, "tat-123", tokenId)
				assert.Equal(t, "test-org", orgName)
				assert.Equal(t, "test-team", teamName)
				return &pulumiapi.AccessToken{
					ID:          "tat-123",
					Description: description,
					TokenValue:  "", // Token value comes from properties
				}, nil
			},
		)

		provider := PulumiServiceTeamAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:         "test-org/test-team/my-token/tat-123",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
			Properties: existingProps,
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-team/my-token/tat-123", resp.Id)
		assert.NotNil(t, resp.Properties)
		assert.NotNil(t, resp.Inputs)
	})

	t.Run("Read with API error", func(t *testing.T) {
		mockedClient := buildTeamAccessTokenClientMock(
			nil,
			nil,
			func(ctx context.Context, tokenId, orgName, teamName string) (*pulumiapi.AccessToken, error) {
				return nil, errors.New("API error")
			},
		)

		provider := PulumiServiceTeamAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "test-org/test-team/my-token/tat-123",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("Read with invalid ID format", func(t *testing.T) {
		provider := PulumiServiceTeamAccessTokenResource{
			Client: &pulumiapi.Client{},
		}

		req := pulumirpc.ReadRequest{
			Id:  "invalid/id",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid, must be in format 'organization/teamName/tokenName/tokenId'")
	})

	t.Run("Create successful", func(t *testing.T) {
		name := "my-token"
		orgName := "test-org"
		teamName := "test-team"
		description := "new team token"
		tokenValue := "pul-xyz789"
		tokenId := "tat-new123"

		inputMap := resource.PropertyMap{
			"name":             resource.NewPropertyValue(name),
			"organizationName": resource.NewPropertyValue(orgName),
			"teamName":         resource.NewPropertyValue(teamName),
			"description":      resource.NewPropertyValue(description),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamAccessTokenClientMock(
			func(ctx context.Context, n, org, team, desc string) (*pulumiapi.AccessToken, error) {
				assert.Equal(t, name, n)
				assert.Equal(t, orgName, org)
				assert.Equal(t, teamName, team)
				assert.Equal(t, description, desc)
				return &pulumiapi.AccessToken{
					ID:          tokenId,
					TokenValue:  tokenValue,
					Description: desc,
				}, nil
			},
			nil,
			nil,
		)

		provider := PulumiServiceTeamAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
			Properties: inputProps,
		}

		resp, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.Equal(t, "test-org/test-team/my-token/tat-new123", resp.Id)
		assert.NotNil(t, resp.Properties)

		// Verify properties contain the secret value
		props, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{KeepSecrets: true})
		assert.True(t, props["value"].IsSecret())
		assert.Equal(t, description, props["description"].StringValue())
	})

	t.Run("Create with API error", func(t *testing.T) {
		name := "my-token"
		inputMap := resource.PropertyMap{
			"name":             resource.NewPropertyValue(name),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue("desc"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildTeamAccessTokenClientMock(
			func(ctx context.Context, n, org, team, desc string) (*pulumiapi.AccessToken, error) {
				return nil, errors.New("create failed")
			},
			nil,
			nil,
		)

		provider := PulumiServiceTeamAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
			Properties: inputProps,
		}

		_, err := provider.Create(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error creating access token")
		assert.Contains(t, err.Error(), name)
	})

	t.Run("Delete successful", func(t *testing.T) {
		tokenId := "test-org/test-team/my-token/tat-123"

		mockedClient := buildTeamAccessTokenClientMock(
			nil,
			func(ctx context.Context, id, orgName, teamName string) error {
				assert.Equal(t, "tat-123", id)
				assert.Equal(t, "test-org", orgName)
				assert.Equal(t, "test-team", teamName)
				return nil
			},
			nil,
		)

		provider := PulumiServiceTeamAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:  tokenId,
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
		}

		resp, err := provider.Delete(&req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Delete with API error", func(t *testing.T) {
		tokenId := "test-org/test-team/my-token/tat-123"

		mockedClient := buildTeamAccessTokenClientMock(
			nil,
			func(ctx context.Context, id, orgName, teamName string) error {
				return errors.New("delete failed")
			},
			nil,
		)

		provider := PulumiServiceTeamAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:  tokenId,
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
		}

		_, err := provider.Delete(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("Diff with no changes", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"name":             resource.NewPropertyValue("my-token"),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue("desc"),
		}

		oldMap := resource.PropertyMap{
			"__inputs":         resource.NewObjectProperty(inputMap),
			"name":             resource.NewPropertyValue("my-token"),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue("desc"),
		}

		olds, _ := plugin.MarshalProperties(oldMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceTeamAccessTokenResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
			Id:  "test-org/test-team/my-token/tat-123",
			Olds: olds,
			News: news,
		}

		resp, err := provider.Diff(&req)

		assert.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
		assert.Empty(t, resp.Replaces)
	})

	t.Run("Diff with property changes", func(t *testing.T) {
		oldInputMap := resource.PropertyMap{
			"name":             resource.NewPropertyValue("my-token"),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue("old description"),
		}

		oldMap := resource.PropertyMap{
			"__inputs":         resource.NewObjectProperty(oldInputMap),
			"name":             resource.NewPropertyValue("my-token"),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue("old description"),
		}

		newInputMap := resource.PropertyMap{
			"name":             resource.NewPropertyValue("my-token"),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue("new description"),
		}

		olds, _ := plugin.MarshalProperties(oldMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(newInputMap, plugin.MarshalOptions{})

		provider := PulumiServiceTeamAccessTokenResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
			Id:  "test-org/test-team/my-token/tat-123",
			Olds: olds,
			News: news,
		}

		resp, err := provider.Diff(&req)

		assert.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.Contains(t, resp.Replaces, "description")
	})

	t.Run("Check with valid input", func(t *testing.T) {
		inputMap := resource.PropertyMap{
			"name":             resource.NewPropertyValue("my-token"),
			"organizationName": resource.NewPropertyValue("test-org"),
			"teamName":         resource.NewPropertyValue("test-team"),
			"description":      resource.NewPropertyValue("desc"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceTeamAccessTokenResource{}

		req := pulumirpc.CheckRequest{
			Urn:  "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
			News: inputProps,
		}

		resp, err := provider.Check(&req)

		assert.NoError(t, err)
		assert.Equal(t, inputProps, resp.Inputs)
		assert.Nil(t, resp.Failures)
	})

	t.Run("Update returns error", func(t *testing.T) {
		provider := PulumiServiceTeamAccessTokenResource{}

		req := pulumirpc.UpdateRequest{
			Id:  "test-org/test-team/my-token/tat-123",
			Urn: "urn:pulumi:test::test::pulumiservice:index:TeamAccessToken::test",
		}

		_, err := provider.Update(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected call to update")
	})
}
