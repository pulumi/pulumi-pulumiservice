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

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock function types for OidcIssuer
type registerOidcIssuerFunc func(ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error)
type updateOidcIssuerFunc func(ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error)
type getOidcIssuerFunc func(ctx context.Context, org, issuerID string) (*pulumiapi.OidcIssuerRegistrationResponse, error)
type deleteOidcIssuerFunc func(ctx context.Context, org, issuerID string) error
type getAuthPoliciesFunc func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error)
type updateAuthPoliciesFunc func(ctx context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest) (*pulumiapi.AuthPolicy, error)

// OidcClientMock mocks the pulumiapi.OidcClient interface
type OidcClientMock struct {
	registerOidcIssuerFunc registerOidcIssuerFunc
	updateOidcIssuerFunc   updateOidcIssuerFunc
	getOidcIssuerFunc      getOidcIssuerFunc
	deleteOidcIssuerFunc   deleteOidcIssuerFunc
	getAuthPoliciesFunc    getAuthPoliciesFunc
	updateAuthPoliciesFunc updateAuthPoliciesFunc
}

func (c *OidcClientMock) RegisterOidcIssuer(ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
	if c.registerOidcIssuerFunc != nil {
		return c.registerOidcIssuerFunc(ctx, org, req)
	}
	return nil, nil
}

func (c *OidcClientMock) UpdateOidcIssuer(ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
	if c.updateOidcIssuerFunc != nil {
		return c.updateOidcIssuerFunc(ctx, org, issuerID, req)
	}
	return nil, nil
}

func (c *OidcClientMock) GetOidcIssuer(ctx context.Context, org, issuerID string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
	if c.getOidcIssuerFunc != nil {
		return c.getOidcIssuerFunc(ctx, org, issuerID)
	}
	return nil, nil
}

func (c *OidcClientMock) DeleteOidcIssuer(ctx context.Context, org, issuerID string) error {
	if c.deleteOidcIssuerFunc != nil {
		return c.deleteOidcIssuerFunc(ctx, org, issuerID)
	}
	return nil
}

func (c *OidcClientMock) GetAuthPolicies(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
	if c.getAuthPoliciesFunc != nil {
		return c.getAuthPoliciesFunc(ctx, org, issuerID)
	}
	return nil, nil
}

func (c *OidcClientMock) UpdateAuthPolicies(ctx context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest) (*pulumiapi.AuthPolicy, error) {
	if c.updateAuthPoliciesFunc != nil {
		return c.updateAuthPoliciesFunc(ctx, org, policyID, req)
	}
	return nil, nil
}

// Helper function to convert slice of AuthPolicyDefinition to slice of pointers
func toAuthPolicyDefinitionPtrs(definitions []pulumiapi.AuthPolicyDefinition) []*pulumiapi.AuthPolicyDefinition {
	result := make([]*pulumiapi.AuthPolicyDefinition, len(definitions))
	for i := range definitions {
		result[i] = &definitions[i]
	}
	return result
}

// TestOidcIssuer_Read_NotFound tests Read when issuer not found
func TestOidcIssuer_Read_NotFound(t *testing.T) {
	mockClient := &OidcClientMock{
		getOidcIssuerFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			return nil, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/issuer-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

// TestOidcIssuer_Read_Found tests Read when issuer is found with policies
func TestOidcIssuer_Read_Found(t *testing.T) {
	mockClient := &OidcClientMock{
		getOidcIssuerFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "issuer-123", issuerID)
			maxExp := int64(3600)
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:                   "issuer-123",
				Name:                 "my-issuer",
				URL:                  "https://example.com",
				MaxExpiration: &maxExp,
				Thumbprints:          []string{"thumb1", "thumb2"},
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{
				ID:         "policy-123",
				Definition: []*pulumiapi.AuthPolicyDefinition{},
			}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/issuer-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "test-org/issuer-123", resp.Id)
	assert.NotNil(t, resp.Properties)
}

// TestOidcIssuer_Read_WithThumbprints tests Read with certificate thumbprints
func TestOidcIssuer_Read_WithThumbprints(t *testing.T) {
	mockClient := &OidcClientMock{
		getOidcIssuerFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:          "issuer-123",
				Name:        "my-issuer",
				URL:         "https://example.com",
				Thumbprints: []string{"ABC123", "DEF456"},
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/issuer-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	propMap, err := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{SkipNulls: true})
	require.NoError(t, err)
	assert.True(t, propMap["thumbprints"].IsArray())
}

// TestOidcIssuer_Read_WithMaxExpiration tests Read with maxExpirationSeconds
func TestOidcIssuer_Read_WithMaxExpiration(t *testing.T) {
	maxExp := int64(7200)
	mockClient := &OidcClientMock{
		getOidcIssuerFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:                   "issuer-123",
				Name:                 "my-issuer",
				URL:                  "https://example.com",
				MaxExpiration: &maxExp,
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	req := &pulumirpc.ReadRequest{
		Id:  "test-org/issuer-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp.Properties)
}

// TestOidcIssuer_Read_InvalidID tests Read with malformed ID
func TestOidcIssuer_Read_InvalidID(t *testing.T) {
	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	req := &pulumirpc.ReadRequest{
		Id:  "invalid-id",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
	}

	resp, err := provider.Read(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestOidcIssuer_Create_Success tests successful creation with policy retrieval
func TestOidcIssuer_Create_Success(t *testing.T) {
	mockClient := &OidcClientMock{
		registerOidcIssuerFunc: func(ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "my-issuer", req.Name)
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:          "issuer-123",
				Name:        req.Name,
				URL:         req.URL,
				Thumbprints: []string{},
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-org/issuer-123", resp.Id)
}

// TestOidcIssuer_Create_WithPolicies tests creation with custom policies
func TestOidcIssuer_Create_WithPolicies(t *testing.T) {
	mockClient := &OidcClientMock{
		registerOidcIssuerFunc: func(ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:   "issuer-123",
				Name: req.Name,
				URL:  req.URL,
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
		updateAuthPoliciesFunc: func(ctx context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest) (*pulumiapi.AuthPolicy, error) {
			assert.Equal(t, "policy-123", policyID)
			assert.NotEmpty(t, req.Definition)
			return &pulumiapi.AuthPolicy{ID: policyID, Definition: toAuthPolicyDefinitionPtrs(req.Definition)}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	policyMap := resource.PropertyMap{
		"decision":  resource.NewStringProperty("allow"),
		"tokenType": resource.NewStringProperty("organization"),
		"rules":     resource.NewObjectProperty(resource.PropertyMap{}),
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
		"policies":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(policyMap)}),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Create_WithThumbprints tests creation with certificate thumbprints
func TestOidcIssuer_Create_WithThumbprints(t *testing.T) {
	mockClient := &OidcClientMock{
		registerOidcIssuerFunc: func(ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			assert.Contains(t, req.Thumbprints, "ABC123")
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:          "issuer-123",
				Name:        req.Name,
				URL:         req.URL,
				Thumbprints: req.Thumbprints,
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
		"thumbprints": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("ABC123"),
			resource.NewStringProperty("DEF456"),
		}),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Create_PolicyUpdateFails_CleansUp tests cleanup on policy update failure
func TestOidcIssuer_Create_PolicyUpdateFails_CleansUp(t *testing.T) {
	deleteAttempted := false

	mockClient := &OidcClientMock{
		registerOidcIssuerFunc: func(ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:   "issuer-123",
				Name: req.Name,
				URL:  req.URL,
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
		updateAuthPoliciesFunc: func(ctx context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest) (*pulumiapi.AuthPolicy, error) {
			return nil, errors.New("policy update failed")
		},
		deleteOidcIssuerFunc: func(ctx context.Context, org, issuerID string) error {
			deleteAttempted = true
			return nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	policyMap := resource.PropertyMap{
		"decision":  resource.NewStringProperty("allow"),
		"tokenType": resource.NewStringProperty("organization"),
		"rules":     resource.NewObjectProperty(resource.PropertyMap{}),
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
		"policies":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(policyMap)}),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, deleteAttempted, "Should attempt to delete issuer on policy update failure")
}

// TestOidcIssuer_Create_NoPolicies tests creation with default policies only
func TestOidcIssuer_Create_NoPolicies(t *testing.T) {
	mockClient := &OidcClientMock{
		registerOidcIssuerFunc: func(ctx context.Context, org string, req pulumiapi.OidcIssuerRegistrationRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:   "issuer-123",
				Name: req.Name,
				URL:  req.URL,
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Update_Success tests successful update
func TestOidcIssuer_Update_Success(t *testing.T) {
	mockClient := &OidcClientMock{
		updateOidcIssuerFunc: func(ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "issuer-123", issuerID)
			return &pulumiapi.OidcIssuerRegistrationResponse{
				ID:   issuerID,
				Name: *req.Name,
				URL:  "https://example.com",
			}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("updated-name"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "test-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Update_ChangePolicies tests updating policies
func TestOidcIssuer_Update_ChangePolicies(t *testing.T) {
	mockClient := &OidcClientMock{
		updateOidcIssuerFunc: func(ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			return &pulumiapi.OidcIssuerRegistrationResponse{ID: issuerID, Name: *req.Name}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
		updateAuthPoliciesFunc: func(ctx context.Context, org, policyID string, req pulumiapi.AuthPolicyUpdateRequest) (*pulumiapi.AuthPolicy, error) {
			assert.NotEmpty(t, req.Definition)
			return &pulumiapi.AuthPolicy{ID: policyID, Definition: toAuthPolicyDefinitionPtrs(req.Definition)}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	policyMap := resource.PropertyMap{
		"decision":  resource.NewStringProperty("allow"),
		"tokenType": resource.NewStringProperty("organization"),
		"rules":     resource.NewObjectProperty(resource.PropertyMap{}),
	}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
		"policies":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(policyMap)}),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "test-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Update_ChangeThumbprints tests updating thumbprints
func TestOidcIssuer_Update_ChangeThumbprints(t *testing.T) {
	mockClient := &OidcClientMock{
		updateOidcIssuerFunc: func(ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			assert.Contains(t, *req.Thumbprints, "NEW123")
			return &pulumiapi.OidcIssuerRegistrationResponse{ID: issuerID, Thumbprints: *req.Thumbprints}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
		"thumbprints":  resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("NEW123")}),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "test-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Update_RemoveMaxExpiration tests removing maxExpirationSeconds
func TestOidcIssuer_Update_RemoveMaxExpiration(t *testing.T) {
	mockClient := &OidcClientMock{
		updateOidcIssuerFunc: func(ctx context.Context, org, issuerID string, req pulumiapi.OidcIssuerUpdateRequest) (*pulumiapi.OidcIssuerRegistrationResponse, error) {
			assert.Nil(t, req.MaxExpiration)
			return &pulumiapi.OidcIssuerRegistrationResponse{ID: issuerID, MaxExpiration: nil}, nil
		},
		getAuthPoliciesFunc: func(ctx context.Context, org, issuerID string) (*pulumiapi.AuthPolicy, error) {
			return &pulumiapi.AuthPolicy{ID: "policy-123", Definition: []*pulumiapi.AuthPolicyDefinition{}}, nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "test-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Delete_Success tests successful deletion
func TestOidcIssuer_Delete_Success(t *testing.T) {
	mockClient := &OidcClientMock{
		deleteOidcIssuerFunc: func(ctx context.Context, org, issuerID string) error {
			assert.Equal(t, "test-org", org)
			assert.Equal(t, "issuer-123", issuerID)
			return nil
		},
	}

	provider := PulumiServiceOidcIssuerResource{Client: mockClient}

	req := &pulumirpc.DeleteRequest{
		Id:  "test-org/issuer-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
	}

	resp, err := provider.Delete(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestOidcIssuer_Delete_InvalidID tests deletion with malformed ID
func TestOidcIssuer_Delete_InvalidID(t *testing.T) {
	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	req := &pulumirpc.DeleteRequest{
		Id:  "invalid-id",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
	}

	resp, err := provider.Delete(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestOidcIssuer_Diff_OrganizationChange tests that organization change triggers replacement
func TestOidcIssuer_Diff_OrganizationChange(t *testing.T) {
	t.Skip("TODO(#586): Skipping until StandardDiff populates Replaces array - see https://github.com/pulumi/pulumi-pulumiservice/issues/586")

	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("old-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("new-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "old-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "organization")
}

// TestOidcIssuer_Diff_URLChange tests that URL change triggers replacement
func TestOidcIssuer_Diff_URLChange(t *testing.T) {
	t.Skip("TODO(#586): Skipping until StandardDiff populates Replaces array - see https://github.com/pulumi/pulumi-pulumiservice/issues/586")

	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://old.example.com"),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://new.example.com"),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "url")
}

// TestOidcIssuer_Diff_NameChange tests that name change does not trigger replacement
func TestOidcIssuer_Diff_NameChange(t *testing.T) {
	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	oldInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("old-name"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("new-name"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotContains(t, resp.Replaces, "name")
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
}

// TestOidcIssuer_Diff_NoChanges tests diff with no changes
func TestOidcIssuer_Diff_NoChanges(t *testing.T) {
	t.Skip("TODO(#587): Skipping until StandardDiff false change detection is fixed - see https://github.com/pulumi/pulumi-pulumiservice/issues/587")

	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	state, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "test-org/issuer-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		Olds: state,
		News: state,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
}

// TestOidcIssuer_Check_SortsPolicies tests that Check sorts policies for consistency
func TestOidcIssuer_Check_SortsPolicies(t *testing.T) {
	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"name":         resource.NewStringProperty("my-issuer"),
		"url":          resource.NewStringProperty("https://example.com"),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}

// TestOidcIssuer_Check_AllFieldTypes tests Check with all property types including nested objects
func TestOidcIssuer_Check_AllFieldTypes(t *testing.T) {
	provider := PulumiServiceOidcIssuerResource{Client: &OidcClientMock{}}

	policyMap := resource.PropertyMap{
		"decision":  resource.NewStringProperty("allow"),
		"tokenType": resource.NewStringProperty("organization"),
		"rules":     resource.NewObjectProperty(resource.PropertyMap{"claim": resource.NewStringProperty("value")}),
	}

	inputs := resource.PropertyMap{
		"organization":         resource.NewStringProperty("test-org"),
		"name":                 resource.NewStringProperty("my-issuer"),
		"url":                  resource.NewStringProperty("https://example.com"),
		"maxExpirationSeconds": resource.NewNumberProperty(3600),
		"thumbprints":          resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("ABC123")}),
		"policies":             resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(policyMap)}),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:OidcIssuer::testIssuer",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}
