// Copyright 2016-2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock function types for TemplateSource
type getTemplateSourceFunc func(ctx context.Context, orgName, templateID string) (*pulumiapi.TemplateSourceResponse, error)
type createTemplateSourceFunc func(ctx context.Context, orgName string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error)
type updateTemplateSourceFunc func(ctx context.Context, orgName, templateID string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error)
type deleteTemplateSourceFunc func(ctx context.Context, orgName, templateID string) error

// ClientMockTemplateSource mocks the Client for TemplateSource operations
type ClientMockTemplateSource struct {
	getTemplateSourceFunc    getTemplateSourceFunc
	createTemplateSourceFunc createTemplateSourceFunc
	updateTemplateSourceFunc updateTemplateSourceFunc
	deleteTemplateSourceFunc deleteTemplateSourceFunc
}

func (c *ClientMockTemplateSource) GetTemplateSource(ctx context.Context, orgName, templateID string) (*pulumiapi.TemplateSourceResponse, error) {
	if c.getTemplateSourceFunc != nil {
		return c.getTemplateSourceFunc(ctx, orgName, templateID)
	}
	return nil, nil
}

func (c *ClientMockTemplateSource) CreateTemplateSource(ctx context.Context, orgName string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
	if c.createTemplateSourceFunc != nil {
		return c.createTemplateSourceFunc(ctx, orgName, req)
	}
	return nil, nil
}

func (c *ClientMockTemplateSource) UpdateTemplateSource(ctx context.Context, orgName, templateID string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
	if c.updateTemplateSourceFunc != nil {
		return c.updateTemplateSourceFunc(ctx, orgName, templateID, req)
	}
	return nil, nil
}

func (c *ClientMockTemplateSource) DeleteTemplateSource(ctx context.Context, orgName, templateID string) error {
	if c.deleteTemplateSourceFunc != nil {
		return c.deleteTemplateSourceFunc(ctx, orgName, templateID)
	}
	return nil
}

func stringPtr(s string) *string {
	return &s
}

// TestTemplateSource_Read_NotFound tests Read when GetTemplateSource returns nil
// TODO: This test has incorrect mock setup - mock is assigned to _ but provider uses nil client
func TestTemplateSource_Read_NotFound(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup that needs fixing")

	_ = &ClientMockTemplateSource{
		getTemplateSourceFunc: func(ctx context.Context, orgName, templateID string) (*pulumiapi.TemplateSourceResponse, error) {
			return nil, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil), // Will be accessed via mockClient in actual implementation
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/template-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

// TestTemplateSource_Read_Found tests Read when resource is found
// TODO: This test also has incorrect mock setup - mock is assigned to _ but provider uses nil client
func TestTemplateSource_Read_Found(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup that needs fixing")

	_ = &ClientMockTemplateSource{
		getTemplateSourceFunc: func(ctx context.Context, orgName, templateID string) (*pulumiapi.TemplateSourceResponse, error) {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "template-123", templateID)
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      "my-template",
				SourceURL: "https://github.com/org/repo",
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/template-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/template-123", resp.Id)
	assert.NotNil(t, resp.Properties)
}

// TestTemplateSource_Read_WithDestination tests Read with destination URL
func TestTemplateSource_Read_WithDestination(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		getTemplateSourceFunc: func(ctx context.Context, orgName, templateID string) (*pulumiapi.TemplateSourceResponse, error) {
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      "my-template",
				SourceURL: "https://github.com/org/repo",
				Destination: &pulumiapi.CreateTemplateSourceRequestDestination{
					URL: stringPtr("https://example.com"),
				},
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/template-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/template-123", resp.Id)
	assert.NotNil(t, resp.Properties)

	// Verify destination is in properties
	propMap, err := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{SkipNulls: true})
	require.NoError(t, err)
	assert.True(t, propMap["destination"].HasValue())
}

// TestTemplateSource_Read_WithoutDestination tests Read without destination
func TestTemplateSource_Read_WithoutDestination(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		getTemplateSourceFunc: func(ctx context.Context, orgName, templateID string) (*pulumiapi.TemplateSourceResponse, error) {
			return &pulumiapi.TemplateSourceResponse{
				Id:          "template-123",
				Name:        "my-template",
				SourceURL:   "https://github.com/org/repo",
				Destination: nil,
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/template-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/template-123", resp.Id)
}

// TestTemplateSource_Read_InvalidID tests Read with malformed ID
func TestTemplateSource_Read_InvalidID(t *testing.T) {
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	req := &pulumirpc.ReadRequest{
		Id:  "invalid-id-no-slash",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
	}

	resp, err := provider.Read(req)

	// Should return error for invalid ID format
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestTemplateSource_Create_Success tests successful creation
func TestTemplateSource_Create_Success(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		createTemplateSourceFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "my-template", req.Name)
			assert.Equal(t, "https://github.com/org/repo", req.SourceURL)
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      req.Name,
				SourceURL: req.SourceURL,
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-org/template-123", resp.Id)
}

// TestTemplateSource_Create_WithDestination tests creation with destination URL
func TestTemplateSource_Create_WithDestination(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		createTemplateSourceFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
			assert.NotNil(t, req.Destination)
			assert.Equal(t, "https://example.com", *req.Destination.URL)
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      req.Name,
				SourceURL: req.SourceURL,
				Destination: &pulumiapi.CreateTemplateSourceRequestDestination{
					URL: req.Destination.URL,
				},
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	destinationMap := resource.PropertyMap{
		"url": resource.NewStringProperty("https://example.com"),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
		"destination":      resource.NewObjectProperty(destinationMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestTemplateSource_Create_WithoutDestination tests creation without destination
func TestTemplateSource_Create_WithoutDestination(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		createTemplateSourceFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
			assert.Nil(t, req.Destination)
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      req.Name,
				SourceURL: req.SourceURL,
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestTemplateSource_Create_APIError tests handling of API errors during creation
func TestTemplateSource_Create_APIError(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		createTemplateSourceFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
			return nil, assert.AnError
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestTemplateSource_Update_Success tests successful update
func TestTemplateSource_Update_Success(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		updateTemplateSourceFunc: func(ctx context.Context, orgName, templateID string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "template-123", templateID)
			assert.Equal(t, "updated-name", req.Name)
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      req.Name,
				SourceURL: req.SourceURL,
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("updated-name"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:         "test-org/template-123",
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		News:       inputsStruct,
		Olds:       inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestTemplateSource_Update_ChangeDestination tests updating destination URL
func TestTemplateSource_Update_ChangeDestination(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		updateTemplateSourceFunc: func(ctx context.Context, orgName, templateID string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
			assert.NotNil(t, req.Destination)
			assert.Equal(t, "https://new-destination.com", *req.Destination.URL)
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      req.Name,
				SourceURL: req.SourceURL,
				Destination: req.Destination,
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	destinationMap := resource.PropertyMap{
		"url": resource.NewStringProperty("https://new-destination.com"),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
		"destination":      resource.NewObjectProperty(destinationMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:         "test-org/template-123",
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		News:       inputsStruct,
		Olds:       inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestTemplateSource_Update_RemoveDestination tests removing destination (set to nil)
func TestTemplateSource_Update_RemoveDestination(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		updateTemplateSourceFunc: func(ctx context.Context, orgName, templateID string, req pulumiapi.CreateTemplateSourceRequest) (*pulumiapi.TemplateSourceResponse, error) {
			assert.Nil(t, req.Destination)
			return &pulumiapi.TemplateSourceResponse{
				Id:        "template-123",
				Name:      req.Name,
				SourceURL: req.SourceURL,
			}, nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:         "test-org/template-123",
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		News:       inputsStruct,
		Olds:       inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestTemplateSource_Update_InvalidID tests update with malformed ID
func TestTemplateSource_Update_InvalidID(t *testing.T) {
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:         "invalid-id",
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		News:       inputsStruct,
		Olds:       inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestTemplateSource_Delete_Success tests successful deletion
func TestTemplateSource_Delete_Success(t *testing.T) {
	t.Skip("TODO: Skipping - test has incorrect mock setup (mock is created but not used)")
	_ = &ClientMockTemplateSource{
		deleteTemplateSourceFunc: func(ctx context.Context, orgName, templateID string) error {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "template-123", templateID)
			return nil
		},
	}

	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	req := &pulumirpc.DeleteRequest{
		Id:  "test-org/template-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
	}

	resp, err := provider.Delete(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestTemplateSource_Delete_InvalidID tests deletion with malformed ID
func TestTemplateSource_Delete_InvalidID(t *testing.T) {
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	req := &pulumirpc.DeleteRequest{
		Id:  "invalid-id",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
	}

	resp, err := provider.Delete(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestTemplateSource_Diff_OrganizationChange tests that changing organization triggers replacement
func TestTemplateSource_Diff_OrganizationChange(t *testing.T) {
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	oldInputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("old-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("new-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "old-org/template-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "organizationName")
}

// TestTemplateSource_Diff_NameChange tests that changing name does not trigger replacement
func TestTemplateSource_Diff_NameChange(t *testing.T) {
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	oldInputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("old-name"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("new-name"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/template-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotContains(t, resp.Replaces, "sourceName")
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
}

// TestTemplateSource_Diff_NoChanges tests diff with no changes
func TestTemplateSource_Diff_NoChanges(t *testing.T) {
	t.Skip("TODO(#587): Skipping until StandardDiff false change detection is fixed - see https://github.com/pulumi/pulumi-pulumiservice/issues/587")
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	state, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/template-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		Olds: state,
		News: state,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
}

// TestTemplateSource_Check_ValidInputs tests Check with valid inputs
func TestTemplateSource_Check_ValidInputs(t *testing.T) {
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}

// TestTemplateSource_Check_AllFieldTypes tests Check with all property types
func TestTemplateSource_Check_AllFieldTypes(t *testing.T) {
	provider := PulumiServiceTemplateSourceResource{
		Client: (*pulumiapi.Client)(nil),
	}

	destinationMap := resource.PropertyMap{
		"url": resource.NewStringProperty("https://example.com"),
	}

	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("test-org"),
		"sourceName":       resource.NewStringProperty("my-template"),
		"sourceURL":        resource.NewStringProperty("https://github.com/org/repo"),
		"destination":      resource.NewObjectProperty(destinationMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:TemplateSource::testTemplate",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}
