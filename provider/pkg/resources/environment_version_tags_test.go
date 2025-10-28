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
	"errors"
	"testing"
	"time"

	"github.com/pulumi/esc"
	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock function types for EnvironmentVersionTag
type getEnvVersionTagFunc func(ctx context.Context, org, proj, env, tag string) (*esc_client.EnvironmentRevisionTag, error)
type createEnvVersionTagFunc func(ctx context.Context, org, proj, env, tag string, rev *int) error
type updateEnvVersionTagFunc func(ctx context.Context, org, proj, env, tag string, rev *int) error
type deleteEnvVersionTagFunc func(ctx context.Context, org, proj, env, tag string) error

// EscClientMockVersionTag mocks the esc_client.Client interface
type EscClientMockVersionTag struct {
	getFunc    getEnvVersionTagFunc
	createFunc createEnvVersionTagFunc
	updateFunc updateEnvVersionTagFunc
	deleteFunc deleteEnvVersionTagFunc
}

func (c *EscClientMockVersionTag) GetEnvironmentRevisionTag(ctx context.Context, org, proj, env, tag string) (*esc_client.EnvironmentRevisionTag, error) {
	if c.getFunc != nil {
		return c.getFunc(ctx, org, proj, env, tag)
	}
	return nil, nil
}

func (c *EscClientMockVersionTag) CreateEnvironmentRevisionTag(ctx context.Context, org, proj, env, tag string, rev *int) error {
	if c.createFunc != nil {
		return c.createFunc(ctx, org, proj, env, tag, rev)
	}
	return nil
}

func (c *EscClientMockVersionTag) UpdateEnvironmentRevisionTag(ctx context.Context, org, proj, env, tag string, rev *int) error {
	if c.updateFunc != nil {
		return c.updateFunc(ctx, org, proj, env, tag, rev)
	}
	return nil
}

func (c *EscClientMockVersionTag) DeleteEnvironmentRevisionTag(ctx context.Context, org, proj, env, tag string) error {
	if c.deleteFunc != nil {
		return c.deleteFunc(ctx, org, proj, env, tag)
	}
	return nil
}

// Implement other required esc_client.Client methods as no-ops
func (c *EscClientMockVersionTag) GetEnvironment(ctx context.Context, orgName, projectName, envName string, version string, decrypt bool) (yaml []byte, etag string, revision int, err error) {
	return nil, "", 0, nil
}
func (c *EscClientMockVersionTag) CreateEnvironment(ctx context.Context, org, proj string) error {
	return nil
}
func (c *EscClientMockVersionTag) UpdateEnvironment(ctx context.Context, org, proj string, yaml []byte, tag string) ([]esc_client.EnvironmentDiagnostic, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) DeleteEnvironment(ctx context.Context, org, proj, env string) error {
	return nil
}
func (c *EscClientMockVersionTag) OpenEnvironment(ctx context.Context, org, proj, env string, duration string, timeout time.Duration) (string, []esc_client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}
func (c *EscClientMockVersionTag) CheckYAMLEnvironment(context.Context, string, []byte, ...esc_client.CheckYAMLOption) (*esc.Environment, []esc_client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}
func (c *EscClientMockVersionTag) CreateEnvironmentWithProject(context.Context, string, string, string) error {
	return nil
}
func (c *EscClientMockVersionTag) CloneEnvironment(context.Context, string, string, string, esc_client.CloneEnvironmentRequest) error {
	return nil
}
func (c *EscClientMockVersionTag) EnvironmentExists(context.Context, string, string, string) (bool, error) {
	return false, nil
}
func (c *EscClientMockVersionTag) GetAnonymousOpenEnvironment(context.Context, string, string) (*esc.Environment, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) GetOpenEnvironment(context.Context, string, string, string) (*esc.Environment, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) GetOpenEnvironmentWithProject(context.Context, string, string, string, string) (*esc.Environment, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) GetAnonymousOpenProperty(context.Context, string, string, string) (*esc.Value, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) GetOpenProperty(context.Context, string, string, string, string, string) (*esc.Value, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) GetPulumiAccountDetails(context.Context) (string, []string, *workspace.TokenInformation, error) {
	return "", nil, nil, nil
}
func (c *EscClientMockVersionTag) ListEnvironments(context.Context, string) ([]esc_client.OrgEnvironment, string, error) {
	return nil, "", nil
}
func (c *EscClientMockVersionTag) ListOrganizationEnvironments(context.Context, string, string) ([]esc_client.OrgEnvironment, string, error) {
	return nil, "", nil
}
func (c *EscClientMockVersionTag) GetEnvironmentRevision(ctx context.Context, orgName, projectName, envName string, revision int) (*esc_client.EnvironmentRevision, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) GetRevisionNumber(ctx context.Context, orgName, projectName, envName, version string) (int, error) {
	return 0, nil
}
func (c *EscClientMockVersionTag) ListEnvironmentRevisions(ctx context.Context, orgName, projectName, envName string, options esc_client.ListEnvironmentRevisionsOptions) ([]esc_client.EnvironmentRevision, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) ListEnvironmentRevisionTags(ctx context.Context, orgName, projectName, envName string, options esc_client.ListEnvironmentRevisionTagsOptions) ([]esc_client.EnvironmentRevisionTag, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) OpenYAMLEnvironment(context.Context, string, []byte, time.Duration) (string, []esc_client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}
func (c *EscClientMockVersionTag) UpdateEnvironmentWithProject(context.Context, string, string, string, []byte, string) ([]esc_client.EnvironmentDiagnostic, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) UpdateEnvironmentWithRevision(ctx context.Context, orgName, projectName, envName string, yaml []byte, etag string) ([]esc_client.EnvironmentDiagnostic, int, error) {
	return nil, 0, nil
}
func (c *EscClientMockVersionTag) CreateEnvironmentTag(context.Context, string, string, string, string, string) (*esc_client.EnvironmentTag, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) GetEnvironmentTag(context.Context, string, string, string, string) (*esc_client.EnvironmentTag, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) ListEnvironmentTags(context.Context, string, string, string, esc_client.ListEnvironmentTagsOptions) ([]*esc_client.EnvironmentTag, string, error) {
	return nil, "", nil
}
func (c *EscClientMockVersionTag) UpdateEnvironmentTag(context.Context, string, string, string, string, string, string, string) (*esc_client.EnvironmentTag, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) DeleteEnvironmentTag(context.Context, string, string, string, string) error {
	return nil
}
func (c *EscClientMockVersionTag) CreateEnvironmentDraft(context.Context, string, string, string, []byte, string) (string, []esc_client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}
func (c *EscClientMockVersionTag) GetDefaultOrg(context.Context) (string, error) {
	return "", nil
}
func (c *EscClientMockVersionTag) GetEnvironmentDraft(context.Context, string, string, string, string) ([]byte, string, error) {
	return nil, "", nil
}
func (c *EscClientMockVersionTag) OpenEnvironmentDraft(context.Context, string, string, string, string, time.Duration) (string, []esc_client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}
func (c *EscClientMockVersionTag) UpdateEnvironmentDraft(context.Context, string, string, string, string, []byte, string) ([]esc_client.EnvironmentDiagnostic, error) {
	return nil, nil
}
func (c *EscClientMockVersionTag) DeleteEnvironmentDraft(context.Context, string, string, string, string) error {
	return nil
}
func (c *EscClientMockVersionTag) PublishEnvironmentDraft(context.Context, string, string, string, string) (int, error) {
	return 0, nil
}
func (c *EscClientMockVersionTag) RotateEnvironment(context.Context, string, string, string, []string) (*esc_client.RotateEnvironmentResponse, []esc_client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}
func (c *EscClientMockVersionTag) RetractEnvironmentRevision(ctx context.Context, orgName, projectName, envName string, version string, replacement *int, reason string) error {
	return nil
}
func (c *EscClientMockVersionTag) SubmitChangeRequest(ctx context.Context, orgName string, changeRequestID string, description *string) error {
	return nil
}
func (c *EscClientMockVersionTag) Insecure() bool {
	return false
}
func (c *EscClientMockVersionTag) URL() string {
	return ""
}

// TestEnvVersionTag_Read_NotFound tests Read when tag not found (nil response)
func TestEnvVersionTag_Read_NotFound(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		getFunc: func(ctx context.Context, org, proj, env, tag string) (*esc_client.EnvironmentRevisionTag, error) {
			return nil, nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/test-proj/test-env/v1.0",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

// TestEnvVersionTag_Read_NotFound_404Error tests Read when 404 error is returned
func TestEnvVersionTag_Read_NotFound_404Error(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		getFunc: func(ctx context.Context, org, proj, env, tag string) (*esc_client.EnvironmentRevisionTag, error) {
			return nil, errors.New("API error: 404 Not Found")
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/test-proj/test-env/v1.0",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

// TestEnvVersionTag_Read_Found tests Read when tag is found
func TestEnvVersionTag_Read_Found(t *testing.T) {
	revision := 42
	mockClient := &EscClientMockVersionTag{
		getFunc: func(ctx context.Context, org, proj, env, tag string) (*esc_client.EnvironmentRevisionTag, error) {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "test-proj", proj)
			assert.Equal(t, "test-env", env)
			assert.Equal(t, "v1.0", tag)
			return &esc_client.EnvironmentRevisionTag{
				Name:     "v1.0",
				Revision: revision,
			}, nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/test-proj/test-env/v1.0",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/test-proj/test-env/v1.0", resp.Id)
	assert.NotNil(t, resp.Properties)
}

// TestEnvVersionTag_Read_LegacyID tests Read with legacy 3-part ID format (org/env/tag)
func TestEnvVersionTag_Read_LegacyID(t *testing.T) {
	revision := 10
	mockClient := &EscClientMockVersionTag{
		getFunc: func(ctx context.Context, org, proj, env, tag string) (*esc_client.EnvironmentRevisionTag, error) {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "default", proj) // Should use default project
			assert.Equal(t, "test-env", env)
			assert.Equal(t, "v1.0", tag)
			return &esc_client.EnvironmentRevisionTag{
				Name:     "v1.0",
				Revision: revision,
			}, nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/test-env/v1.0", // 3-part legacy ID
		Urn: "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Read_ModernID tests Read with modern 4-part ID format (org/proj/env/tag)
func TestEnvVersionTag_Read_ModernID(t *testing.T) {
	revision := 10
	mockClient := &EscClientMockVersionTag{
		getFunc: func(ctx context.Context, org, proj, env, tag string) (*esc_client.EnvironmentRevisionTag, error) {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "my-project", proj)
			assert.Equal(t, "test-env", env)
			assert.Equal(t, "v1.0", tag)
			return &esc_client.EnvironmentRevisionTag{
				Name:     "v1.0",
				Revision: revision,
			}, nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/my-project/test-env/v1.0", // 4-part modern ID
		Urn: "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Read_InvalidID tests Read with malformed ID
func TestEnvVersionTag_Read_InvalidID(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	req := &pulumirpc.ReadRequest{
		Id:  "invalid/id", // Wrong part count (2 parts)
		Urn: "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
	}

	resp, err := provider.Read(req)

	// Should return error for invalid ID format
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestEnvVersionTag_Create_Success tests successful creation
func TestEnvVersionTag_Create_Success(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		createFunc: func(ctx context.Context, org, proj, env, tag string, rev *int) error {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "test-proj", proj)
			assert.Equal(t, "test-env", env)
			assert.Equal(t, "v1.0", tag)
			assert.NotNil(t, rev)
			assert.Equal(t, 42, *rev)
			return nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-org/test-proj/test-env/v1.0", resp.Id)
}

// TestEnvVersionTag_Create_DefaultProject tests creation with default project
func TestEnvVersionTag_Create_DefaultProject(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		createFunc: func(ctx context.Context, org, proj, env, tag string, rev *int) error {
			assert.Equal(t, "default", proj) // Should use default when not provided
			return nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		// project not provided
		"environment": resource.NewStringProperty("test-env"),
		"tagName":     resource.NewStringProperty("v1.0"),
		"revision":    resource.NewNumberProperty(10),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Create_RevisionZero tests creation with revision=0
func TestEnvVersionTag_Create_RevisionZero(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		createFunc: func(ctx context.Context, org, proj, env, tag string, rev *int) error {
			assert.NotNil(t, rev)
			assert.Equal(t, 0, *rev)
			return nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(0),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Update_Success tests successful update
func TestEnvVersionTag_Update_Success(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		updateFunc: func(ctx context.Context, org, proj, env, tag string, rev *int) error {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "test-proj", proj)
			assert.Equal(t, "test-env", env)
			assert.Equal(t, "v1.0", tag)
			assert.NotNil(t, rev)
			assert.Equal(t, 50, *rev)
			return nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(50),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "test-org/test-proj/test-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Update_ChangeRevision tests changing revision number
func TestEnvVersionTag_Update_ChangeRevision(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		updateFunc: func(ctx context.Context, org, proj, env, tag string, rev *int) error {
			assert.Equal(t, 100, *rev) // New revision
			return nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(100),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "test-org/test-proj/test-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Delete_Success tests successful deletion
func TestEnvVersionTag_Delete_Success(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		deleteFunc: func(ctx context.Context, org, proj, env, tag string) error {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "test-proj", proj)
			assert.Equal(t, "test-env", env)
			assert.Equal(t, "v1.0", tag)
			return nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	// Create properties struct for Delete
	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DeleteRequest{
		Id:         "test-org/test-proj/test-env/v1.0",
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Properties: inputsStruct,
	}

	resp, err := provider.Delete(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Delete_UsesProperties tests that Delete uses properties from req.GetProperties()
func TestEnvVersionTag_Delete_UsesProperties(t *testing.T) {
	mockClient := &EscClientMockVersionTag{
		deleteFunc: func(ctx context.Context, org, proj, env, tag string) error {
			// Verify the values come from properties, not ID parsing
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "test-proj", proj)
			assert.Equal(t, "test-env", env)
			assert.Equal(t, "v1.0", tag)
			return nil
		},
	}

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: mockClient,
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DeleteRequest{
		Id:         "different-id", // ID doesn't matter, properties do
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Properties: inputsStruct,
	}

	resp, err := provider.Delete(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestEnvVersionTag_Diff_OrganizationChange tests that organization change triggers replacement
func TestEnvVersionTag_Diff_OrganizationChange(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("old-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("new-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "old-org/test-proj/test-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "organization")
}

// TestEnvVersionTag_Diff_ProjectChange tests that project change triggers replacement
func TestEnvVersionTag_Diff_ProjectChange(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("old-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("new-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/old-proj/test-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "project")
}

// TestEnvVersionTag_Diff_EnvironmentChange tests that environment change triggers replacement
func TestEnvVersionTag_Diff_EnvironmentChange(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("old-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("new-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/test-proj/old-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "environment")
}

// TestEnvVersionTag_Diff_TagNameChange tests that tagName change triggers replacement
func TestEnvVersionTag_Diff_TagNameChange(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v2.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/test-proj/test-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "tagName")
}

// TestEnvVersionTag_Diff_RevisionChange tests that revision change does not trigger replacement
func TestEnvVersionTag_Diff_RevisionChange(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(100), // Changed
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/test-proj/test-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotContains(t, resp.Replaces, "revision")
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
}

// TestEnvVersionTag_Diff_NoChanges tests diff with no changes
func TestEnvVersionTag_Diff_NoChanges(t *testing.T) {
	t.Skip("TODO(#587): Skipping until StandardDiff false change detection is fixed - see https://github.com/pulumi/pulumi-pulumiservice/issues/587")

	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	state, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/test-proj/test-env/v1.0",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		Olds: state,
		News: state,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
}

// TestEnvVersionTag_Check_AllRequiredFields tests validation of all required fields
func TestEnvVersionTag_Check_AllRequiredFields(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		"revision":     resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}

// TestEnvVersionTag_Check_MissingOrganization tests validation failure when organization missing
func TestEnvVersionTag_Check_MissingOrganization(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	inputs := resource.PropertyMap{
		// organization missing
		"project":     resource.NewStringProperty("test-proj"),
		"environment": resource.NewStringProperty("test-env"),
		"tagName":     resource.NewStringProperty("v1.0"),
		"revision":    resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	assert.Contains(t, resp.Failures[0].Reason, "organization")
}

// TestEnvVersionTag_Check_MissingProject tests validation failure when project missing
func TestEnvVersionTag_Check_MissingProject(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		// project missing
		"environment": resource.NewStringProperty("test-env"),
		"tagName":     resource.NewStringProperty("v1.0"),
		"revision":    resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	found := false
	for _, failure := range resp.Failures {
		if failure.Property == "project" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have failure for project property")
}

// TestEnvVersionTag_Check_MissingEnvironment tests validation failure when environment missing
func TestEnvVersionTag_Check_MissingEnvironment(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		// environment missing
		"tagName":  resource.NewStringProperty("v1.0"),
		"revision": resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	found := false
	for _, failure := range resp.Failures {
		if failure.Property == "environment" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have failure for environment property")
}

// TestEnvVersionTag_Check_MissingTagName tests validation failure when tagName missing
func TestEnvVersionTag_Check_MissingTagName(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		// tagName missing
		"revision": resource.NewNumberProperty(42),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	found := false
	for _, failure := range resp.Failures {
		if failure.Property == "tagName" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have failure for tagName property")
}

// TestEnvVersionTag_Check_MissingRevision tests validation failure when revision missing
func TestEnvVersionTag_Check_MissingRevision(t *testing.T) {
	provider := PulumiServiceEnvironmentVersionTagResource{
		Client: &EscClientMockVersionTag{},
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"environment":  resource.NewStringProperty("test-env"),
		"tagName":      resource.NewStringProperty("v1.0"),
		// revision missing
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:EnvironmentVersionTag::testTag",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	found := false
	for _, failure := range resp.Failures {
		if failure.Property == "revision" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have failure for revision property")
}
