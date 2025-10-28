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

type createAccessTokenFunc func(ctx context.Context, description string) (*pulumiapi.AccessToken, error)
type deleteAccessTokenFunc func(ctx context.Context, tokenId string) error
type getAccessTokenFunc func(ctx context.Context, id string) (*pulumiapi.AccessToken, error)

type AccessTokenClientMock struct {
	createAccessTokenFunc createAccessTokenFunc
	deleteAccessTokenFunc deleteAccessTokenFunc
	getAccessTokenFunc    getAccessTokenFunc
}

func (c *AccessTokenClientMock) CreateAccessToken(ctx context.Context, description string) (*pulumiapi.AccessToken, error) {
	return c.createAccessTokenFunc(ctx, description)
}

func (c *AccessTokenClientMock) DeleteAccessToken(ctx context.Context, tokenId string) error {
	return c.deleteAccessTokenFunc(ctx, tokenId)
}

func (c *AccessTokenClientMock) GetAccessToken(ctx context.Context, id string) (*pulumiapi.AccessToken, error) {
	return c.getAccessTokenFunc(ctx, id)
}

func buildAccessTokenClientMock(
	createFunc createAccessTokenFunc,
	deleteFunc deleteAccessTokenFunc,
	getFunc getAccessTokenFunc,
) *AccessTokenClientMock {
	return &AccessTokenClientMock{
		createAccessTokenFunc: createFunc,
		deleteAccessTokenFunc: deleteFunc,
		getAccessTokenFunc:    getFunc,
	}
}

func TestAccessToken(t *testing.T) {
	t.Run("Read when resource not found", func(t *testing.T) {
		mockedClient := buildAccessTokenClientMock(
			nil,
			nil,
			func(ctx context.Context, id string) (*pulumiapi.AccessToken, error) {
				return nil, nil
			},
		)

		provider := &PulumiServiceAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "at-missing",
			Urn: "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.Id)
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when resource found", func(t *testing.T) {
		tokenValue := "pul-1234567890abcdef"
		description := "test access token"

		inputMap := resource.PropertyMap{
			"description": resource.NewPropertyValue(description),
		}

		// Create properties with value as secret
		outputMap := resource.PropertyMap{
			"__inputs":    resource.NewObjectProperty(inputMap),
			"description": resource.NewPropertyValue(description),
			"value":       resource.MakeSecret(resource.NewPropertyValue(tokenValue)),
		}
		existingProps, _ := plugin.MarshalProperties(outputMap, plugin.MarshalOptions{})

		mockedClient := buildAccessTokenClientMock(
			nil,
			nil,
			func(ctx context.Context, id string) (*pulumiapi.AccessToken, error) {
				return &pulumiapi.AccessToken{
					ID:          "at-abc123",
					Description: description,
					TokenValue:  "", // Token value comes from properties
				}, nil
			},
		)

		provider := &PulumiServiceAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:         "at-abc123",
			Urn:        "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
			Properties: existingProps,
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, "at-abc123", resp.Id)
		assert.NotNil(t, resp.Properties)
		assert.NotNil(t, resp.Inputs)
	})

	t.Run("Read with API error", func(t *testing.T) {
		mockedClient := buildAccessTokenClientMock(
			nil,
			nil,
			func(ctx context.Context, id string) (*pulumiapi.AccessToken, error) {
				return nil, errors.New("API error")
			},
		)

		provider := &PulumiServiceAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "at-abc123",
			Urn: "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
		}

		_, err := provider.Read(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("Create successful", func(t *testing.T) {
		description := "new token"
		tokenValue := "pul-xyz789"
		tokenId := "at-new123"

		inputMap := resource.PropertyMap{
			"description": resource.NewPropertyValue(description),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildAccessTokenClientMock(
			func(ctx context.Context, desc string) (*pulumiapi.AccessToken, error) {
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

		provider := &PulumiServiceAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
			Properties: inputProps,
		}

		resp, err := provider.Create(&req)

		assert.NoError(t, err)
		assert.Equal(t, tokenId, resp.Id)
		assert.NotNil(t, resp.Properties)

		// Verify properties contain the secret value
		props, _ := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{KeepSecrets: true})
		assert.True(t, props["value"].IsSecret())
		assert.Equal(t, description, props["description"].StringValue())
	})

	t.Run("Create with API error", func(t *testing.T) {
		description := "new token"

		inputMap := resource.PropertyMap{
			"description": resource.NewPropertyValue(description),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		mockedClient := buildAccessTokenClientMock(
			func(ctx context.Context, desc string) (*pulumiapi.AccessToken, error) {
				return nil, errors.New("create failed")
			},
			nil,
			nil,
		)

		provider := &PulumiServiceAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.CreateRequest{
			Urn:        "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
			Properties: inputProps,
		}

		_, err := provider.Create(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
	})

	t.Run("Delete successful", func(t *testing.T) {
		tokenId := "at-abc123"

		mockedClient := buildAccessTokenClientMock(
			nil,
			func(ctx context.Context, id string) error {
				assert.Equal(t, tokenId, id)
				return nil
			},
			nil,
		)

		provider := &PulumiServiceAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:  tokenId,
			Urn: "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
		}

		_, err := provider.Delete(&req)

		assert.NoError(t, err)
	})

	t.Run("Delete with API error", func(t *testing.T) {
		tokenId := "at-abc123"

		mockedClient := buildAccessTokenClientMock(
			nil,
			func(ctx context.Context, id string) error {
				return errors.New("delete failed")
			},
			nil,
		)

		provider := &PulumiServiceAccessTokenResource{
			Client: mockedClient,
		}

		req := pulumirpc.DeleteRequest{
			Id:  tokenId,
			Urn: "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
		}

		_, err := provider.Delete(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("Diff with no changes", func(t *testing.T) {
		description := "test token"

		inputMap := resource.PropertyMap{
			"description": resource.NewPropertyValue(description),
		}

		oldMap := resource.PropertyMap{
			"__inputs":    resource.NewObjectProperty(inputMap),
			"description": resource.NewPropertyValue(description),
		}

		olds, _ := plugin.MarshalProperties(oldMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceAccessTokenResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
			Id:  "at-abc123",
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
			"description": resource.NewPropertyValue("old description"),
		}

		oldMap := resource.PropertyMap{
			"__inputs":    resource.NewObjectProperty(oldInputMap),
			"description": resource.NewPropertyValue("old description"),
		}

		newInputMap := resource.PropertyMap{
			"description": resource.NewPropertyValue("new description"),
		}

		olds, _ := plugin.MarshalProperties(oldMap, plugin.MarshalOptions{})
		news, _ := plugin.MarshalProperties(newInputMap, plugin.MarshalOptions{})

		provider := PulumiServiceAccessTokenResource{}

		req := pulumirpc.DiffRequest{
			Urn: "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
			Id:  "at-abc123",
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
			"description": resource.NewPropertyValue("test token"),
		}
		inputProps, _ := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})

		provider := PulumiServiceAccessTokenResource{}

		req := pulumirpc.CheckRequest{
			Urn:  "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
			News: inputProps,
		}

		resp, err := provider.Check(&req)

		assert.NoError(t, err)
		assert.Equal(t, inputProps, resp.Inputs)
		assert.Nil(t, resp.Failures)
	})

	t.Run("Update returns error", func(t *testing.T) {
		provider := PulumiServiceAccessTokenResource{}

		req := pulumirpc.UpdateRequest{
			Id:  "at-abc123",
			Urn: "urn:pulumi:test::test::pulumiservice:index:AccessToken::test",
		}

		_, err := provider.Update(&req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected call to update")
	})
}
