// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// testFutureScope stands in for a scope Pulumi Cloud added after this
// provider's enum was generated.
const (
	testFutureScope = "stack:time_travel"
	testSetID       = "acme/set-1"
)

type permissionSetClientMock struct {
	config.Client
	create func(org string, req pulumiapi.PermissionSetRequest) (*pulumiapi.PermissionSet, error)
	get    func(org, id string) (*pulumiapi.PermissionSet, error)
	update func(org, id string, req pulumiapi.PermissionSetRequest) (*pulumiapi.PermissionSet, error)
}

func (m *permissionSetClientMock) CreatePermissionSet(
	_ context.Context, org string, req pulumiapi.PermissionSetRequest,
) (*pulumiapi.PermissionSet, error) {
	return m.create(org, req)
}

func (m *permissionSetClientMock) GetPermissionSet(
	_ context.Context, org, id string,
) (*pulumiapi.PermissionSet, error) {
	return m.get(org, id)
}

func (m *permissionSetClientMock) UpdatePermissionSet(
	_ context.Context, org, id string, req pulumiapi.PermissionSetRequest,
) (*pulumiapi.PermissionSet, error) {
	return m.update(org, id, req)
}

func TestRbacPermissionSetCreateSendsAdditionalScopes(t *testing.T) {
	t.Parallel()
	mock := &permissionSetClientMock{
		create: func(org string, req pulumiapi.PermissionSetRequest) (*pulumiapi.PermissionSet, error) {
			assert.Equal(t, gcAcme, org)
			assert.Equal(t, string(RbacResourceTypeStack), req.ResourceType)
			// Unknown scopes pass through untouched; duplicates collapse.
			assert.Equal(t, []string{string(RbacScopeStackRead), string(RbacScopeStackWrite), testFutureScope}, req.Permissions)
			return &pulumiapi.PermissionSet{ID: "set-1", Version: 1}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)

	resp, err := (&RbacPermissionSet{}).Create(ctx, infer.CreateRequest[RbacPermissionSetInput]{
		Inputs: RbacPermissionSetInput{
			OrganizationName:      gcAcme,
			Name:                  "Deployer",
			ResourceType:          RbacResourceTypeStack,
			Permissions:           []RbacScope{RbacScopeStackRead, RbacScopeStackWrite},
			AdditionalPermissions: []string{testFutureScope, string(RbacScopeStackRead)},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, testSetID, resp.ID)
	assert.Equal(t, "set-1", resp.Output.PermissionSetId)
}

func TestRbacPermissionSetRead(t *testing.T) {
	t.Parallel()
	stored := &pulumiapi.PermissionSet{
		ID:           "set-1",
		Name:         "Deployer",
		ResourceType: string(RbacResourceTypeStack),
		UxPurpose:    "set",
		DetailsType:  wireAllow,
		Version:      3,
		Permissions: []string{
			string(RbacScopeStackWrite), testFutureScope, string(RbacScopeStackRead), string(RbacScopeStackDelete),
		},
	}
	mock := &permissionSetClientMock{
		get: func(_, _ string) (*pulumiapi.PermissionSet, error) { return stored, nil },
	}
	ctx := config.WithMockClient(context.Background(), mock)

	t.Run("splits scopes and keeps the user's layout", func(t *testing.T) {
		resp, err := (&RbacPermissionSet{}).Read(ctx, infer.ReadRequest[RbacPermissionSetInput, RbacPermissionSetState]{
			ID: testSetID,
			Inputs: RbacPermissionSetInput{
				Permissions: []RbacScope{RbacScopeStackRead, RbacScopeStackWrite},
				// The user kept a known scope in additionalPermissions; leave it there.
				AdditionalPermissions: []string{testFutureScope, string(RbacScopeStackDelete)},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, []RbacScope{RbacScopeStackRead, RbacScopeStackWrite}, resp.Inputs.Permissions)
		assert.Equal(t, []string{testFutureScope, string(RbacScopeStackDelete)}, resp.Inputs.AdditionalPermissions)
		assert.Equal(t, RbacResourceTypeStack, resp.Inputs.ResourceType)
		assert.Nil(t, resp.Inputs.Description)
		assert.Equal(t, 3, resp.State.Version)
	})

	t.Run("import puts unknown scopes in additionalPermissions", func(t *testing.T) {
		resp, err := (&RbacPermissionSet{}).Read(ctx, infer.ReadRequest[RbacPermissionSetInput, RbacPermissionSetState]{
			ID: testSetID,
		})
		require.NoError(t, err)
		assert.Equal(t, "acme", resp.Inputs.OrganizationName)
		assert.Equal(t, []RbacScope{RbacScopeStackWrite, RbacScopeStackRead, RbacScopeStackDelete}, resp.Inputs.Permissions)
		assert.Equal(t, []string{testFutureScope}, resp.Inputs.AdditionalPermissions)
	})

	t.Run("rejects descriptors that are not permission sets", func(t *testing.T) {
		mock := &permissionSetClientMock{get: func(_, _ string) (*pulumiapi.PermissionSet, error) {
			return &pulumiapi.PermissionSet{
				ID:          "role-1",
				UxPurpose:   string(apitype.PermissionDescriptorUXPurposeRole),
				DetailsType: wireCompose,
			}, nil
		}}
		_, err := (&RbacPermissionSet{}).Read(config.WithMockClient(context.Background(), mock),
			infer.ReadRequest[RbacPermissionSetInput, RbacPermissionSetState]{ID: "acme/role-1"})
		assert.ErrorContains(t, err, "not a permission set")
	})

	t.Run("gone", func(t *testing.T) {
		mock := &permissionSetClientMock{get: func(_, _ string) (*pulumiapi.PermissionSet, error) { return nil, nil }}
		resp, err := (&RbacPermissionSet{}).Read(config.WithMockClient(context.Background(), mock),
			infer.ReadRequest[RbacPermissionSetInput, RbacPermissionSetState]{ID: testSetID})
		require.NoError(t, err)
		assert.Empty(t, resp.ID)
	})
}

func TestRbacPermissionSetCheck(t *testing.T) {
	t.Parallel()
	check := func(t *testing.T, inputs map[string]property.Value) map[string]string {
		t.Helper()
		base := map[string]property.Value{
			gcOrganizationName: property.New(gcAcme),
			gcName:             property.New("Deployer"),
			gcResourceType:     property.New(string(RbacResourceTypeStack)),
		}
		for k, v := range inputs {
			base[k] = v
		}
		resp, err := (&RbacPermissionSet{}).Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(base),
		})
		require.NoError(t, err)
		failures := map[string]string{}
		for _, f := range resp.Failures {
			failures[f.Property] = f.Reason
		}
		return failures
	}

	t.Run("valid", func(t *testing.T) {
		assert.Empty(t, check(t, map[string]property.Value{
			gcPermissions:           property.New([]property.Value{property.New(string(RbacScopeStackRead))}),
			gcAdditionalPermissions: property.New([]property.Value{property.New(testFutureScope)}),
		}))
	})

	t.Run("scope for another resource type", func(t *testing.T) {
		failures := check(t, map[string]property.Value{
			gcPermissions: property.New([]property.Value{
				property.New(string(RbacScopeStackRead)), property.New(string(RbacScopeEnvironmentRead)),
			}),
		})
		assert.Contains(t, failures["permissions[1]"], "applies to: environment")
	})

	t.Run("nothing granted", func(t *testing.T) {
		failures := check(t, map[string]property.Value{
			gcPermissions: property.New([]property.Value{}),
		})
		assert.Contains(t, failures, gcPermissions)
	})

	t.Run("unknown scope in permissions is rejected by the enum", func(t *testing.T) {
		failures := check(t, map[string]property.Value{
			gcPermissions: property.New([]property.Value{property.New(testFutureScope)}),
		})
		assert.NotEmpty(t, failures)
	})

	t.Run("computed permissions skip validation", func(t *testing.T) {
		assert.Empty(t, check(t, map[string]property.Value{
			gcPermissions: property.New(property.Computed),
		}))
	})
}
