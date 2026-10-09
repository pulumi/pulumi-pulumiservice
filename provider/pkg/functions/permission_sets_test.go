// Copyright 2026, Pulumi Corporation.
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

package functions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

type permissionSetsClientMock struct {
	config.Client
	sets []apitype.PermissionDescriptorRecord
}

func (c *permissionSetsClientMock) ListOrgRoles(
	_ context.Context, _, uxPurpose string,
) ([]apitype.PermissionDescriptorRecord, error) {
	if uxPurpose != string(apitype.PermissionDescriptorUXPurposeSet) {
		return nil, nil
	}
	return c.sets, nil
}

const (
	testOrgName      = "org"
	testResourceType = "stack"
	testSetID        = "set-1"
)

func TestGetOrganizationPermissionSet(t *testing.T) {
	t.Parallel()

	stackRead := apitype.PermissionDescriptorRecord{
		PermissionDescriptorBase: apitype.PermissionDescriptorBase{
			Name:         "Stack Read",
			Description:  "Read stacks.",
			ResourceType: testResourceType,
			Details:      allow(permStackRead),
		},
		ID:                testSetID,
		DefaultIdentifier: testStackReadSetID,
	}
	custom := apitype.PermissionDescriptorRecord{
		PermissionDescriptorBase: apitype.PermissionDescriptorBase{
			Name: "Deployer", ResourceType: testResourceType, Details: compose([]string{testSetID}),
		},
		ID: "set-2",
	}
	dup := apitype.PermissionDescriptorRecord{
		PermissionDescriptorBase: apitype.PermissionDescriptorBase{Name: "Deployer", ResourceType: testResourceType},
		ID:                       "set-3",
	}

	invoke := func(t *testing.T, sets []apitype.PermissionDescriptorRecord, in GetOrganizationPermissionSetInput) (
		GetOrganizationPermissionSetOutput, error,
	) {
		ctx := config.WithMockClient(t.Context(), &permissionSetsClientMock{sets: sets})
		resp, err := GetOrganizationPermissionSetFunction{}.Invoke(ctx,
			infer.FunctionRequest[GetOrganizationPermissionSetInput]{Input: in})
		return resp.Output, err
	}

	t.Run("by default identifier", func(t *testing.T) {
		t.Parallel()
		got, err := invoke(t, []apitype.PermissionDescriptorRecord{stackRead, custom},
			GetOrganizationPermissionSetInput{OrganizationName: testOrgName, DefaultIdentifier: ptr(testStackReadSetID)})
		require.NoError(t, err)
		assert.Equal(t, GetOrganizationPermissionSetOutput{
			PermissionSetId:   testSetID,
			Name:              "Stack Read",
			Description:       "Read stacks.",
			ResourceType:      testResourceType,
			DefaultIdentifier: ptr(testStackReadSetID),
			Permissions:       []string{permStackRead},
		}, got)
	})

	t.Run("by name, non-flat set", func(t *testing.T) {
		t.Parallel()
		got, err := invoke(t, []apitype.PermissionDescriptorRecord{stackRead, custom},
			GetOrganizationPermissionSetInput{OrganizationName: testOrgName, Name: ptr("Deployer")})
		require.NoError(t, err)
		assert.Equal(t, "set-2", got.PermissionSetId)
		assert.Nil(t, got.DefaultIdentifier)
		assert.Empty(t, got.Permissions)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, []apitype.PermissionDescriptorRecord{stackRead},
			GetOrganizationPermissionSetInput{OrganizationName: testOrgName, Name: ptr("Nope")})
		require.ErrorContains(t, err, `no permission set with name "Nope" found in organization "org"`)
	})

	t.Run("ambiguous name", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, []apitype.PermissionDescriptorRecord{custom, dup},
			GetOrganizationPermissionSetInput{OrganizationName: testOrgName, Name: ptr("Deployer")})
		require.ErrorContains(t, err, "2 permission sets with name")
	})

	t.Run("requires exactly one selector", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, nil, GetOrganizationPermissionSetInput{OrganizationName: testOrgName})
		require.ErrorContains(t, err, "exactly one of `defaultIdentifier` or `name`")
		_, err = invoke(t, nil, GetOrganizationPermissionSetInput{
			OrganizationName: testOrgName, Name: ptr("a"), DefaultIdentifier: ptr("b"),
		})
		require.ErrorContains(t, err, "exactly one of `defaultIdentifier` or `name`")
	})
}
