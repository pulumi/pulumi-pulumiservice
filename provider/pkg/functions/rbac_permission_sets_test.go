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

package functions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/resources"
)

type permissionSetListMock struct {
	config.Client
	sets []apitype.PermissionDescriptorRecord
}

func (m *permissionSetListMock) ListOrgRoles(
	_ context.Context, _, uxPurpose string,
) ([]apitype.PermissionDescriptorRecord, error) {
	if uxPurpose != string(apitype.PermissionDescriptorUXPurposeSet) {
		return nil, nil
	}
	return m.sets, nil
}

func permissionSet(
	id, name string, rt resources.RbacResourceType, defaultID string, scopes ...apitype.RbacPermission,
) apitype.PermissionDescriptorRecord {
	return apitype.PermissionDescriptorRecord{
		PermissionDescriptorBase: apitype.PermissionDescriptorBase{
			Name:         name,
			ResourceType: string(rt),
			UxPurpose:    apitype.PermissionDescriptorUXPurposeSet,
			Details:      apitype.PermissionDescriptorAllowBuilder{Permissions: scopes}.Build(),
		},
		ID:                id,
		DefaultIdentifier: defaultID,
	}
}

func permissionSetCtx() context.Context {
	return config.WithMockClient(context.Background(), &permissionSetListMock{
		sets: []apitype.PermissionDescriptorRecord{
			permissionSet("s1", "Stack Read", resources.RbacResourceTypeStack, "stack-read",
				apitype.RbacPermissionStackRead),
			permissionSet("s2", "Read Only", resources.RbacResourceTypeGlobal, "org-settings-read-only"),
			permissionSet("s3", "Deployer", resources.RbacResourceTypeStack, ""),
		},
	})
}

func TestGetRbacPermissionSet(t *testing.T) {
	t.Parallel()
	get := func(in GetRbacPermissionSetInput) (RbacPermissionSetInfo, error) {
		in.OrganizationName = "acme"
		resp, err := GetRbacPermissionSetFunction{}.Invoke(permissionSetCtx(),
			infer.FunctionRequest[GetRbacPermissionSetInput]{Input: in})
		return resp.Output, err
	}
	str := func(s string) *string { return &s }

	t.Run("by default identifier", func(t *testing.T) {
		got, err := get(GetRbacPermissionSetInput{DefaultIdentifier: str("stack-read")})
		require.NoError(t, err)
		assert.Equal(t, "s1", got.PermissionSetId)
		assert.Equal(t, resources.RbacResourceTypeStack, got.ResourceType)
		assert.Equal(t, []string{"stack:read"}, got.Permissions)
	})

	t.Run("by name", func(t *testing.T) {
		got, err := get(GetRbacPermissionSetInput{Name: str("Deployer")})
		require.NoError(t, err)
		assert.Equal(t, "s3", got.PermissionSetId)
		assert.Nil(t, got.DefaultIdentifier)
		assert.Equal(t, []string{}, got.Permissions)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := get(GetRbacPermissionSetInput{Name: str("Nope")})
		assert.ErrorContains(t, err, `name "Nope"`)
	})

	t.Run("needs exactly one selector", func(t *testing.T) {
		_, err := get(GetRbacPermissionSetInput{Name: str("Deployer"), DefaultIdentifier: str("stack-read")})
		assert.ErrorContains(t, err, "exactly one")
	})
}

func TestGetRbacPermissionSets(t *testing.T) {
	t.Parallel()
	stack := resources.RbacResourceTypeStack
	resp, err := GetRbacPermissionSetsFunction{}.Invoke(permissionSetCtx(),
		infer.FunctionRequest[GetRbacPermissionSetsInput]{Input: GetRbacPermissionSetsInput{
			OrganizationName: "acme", ResourceType: &stack,
		}})
	require.NoError(t, err)
	var names []string
	for _, s := range resp.Output.PermissionSets {
		names = append(names, s.Name)
	}
	assert.Equal(t, []string{"Deployer", "Stack Read"}, names)
}
