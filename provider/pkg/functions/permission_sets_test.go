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

func testSet(id, name, defaultIdentifier, resourceType string, details apitype.PermissionDescriptor,
) apitype.PermissionDescriptorRecord {
	return apitype.PermissionDescriptorRecord{
		PermissionDescriptorBase: apitype.PermissionDescriptorBase{
			Name:         name,
			Description:  name + " description",
			ResourceType: resourceType,
			UxPurpose:    apitype.PermissionDescriptorUXPurposeSet,
			Details:      details,
		},
		ID:                id,
		DefaultIdentifier: defaultIdentifier,
	}
}

func compose(ids ...string) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorComposeBuilder{PermissionDescriptors: ids}.Build()
}

func stubPermissionSets(t *testing.T, sets ...apitype.PermissionDescriptorRecord) *orgRolesClientMock {
	return &orgRolesClientMock{
		listOrgRolesFunc: func(
			_ context.Context,
			orgName string,
			uxPurpose apitype.PermissionDescriptorUXPurpose,
		) ([]apitype.PermissionDescriptorRecord, error) {
			assert.Equal(t, testFunctionRolesOrgName, orgName)
			assert.Equal(t, apitype.PermissionDescriptorUXPurposeSet, uxPurpose)
			return sets, nil
		},
	}
}

func TestPermissionSetLookups(t *testing.T) {
	t.Parallel()

	stackWrite := testSet("set-1", "Stack Write", "stack-write", entityTypeStack,
		allow(permStackRead, permStackAnnotationsUpdate))
	deployers := testSet("set-2", "Deployers", "", entityTypeStack, group(
		allow("stack_deployment:create"),
		compose("set-1"),
	))
	envRead := testSet("set-3", "Environment Read", "environment-read", entityTypeEnvironment,
		allow(permEnvironmentRead))
	acctRead := testSet("set-4", "Account Read", "insights-account-read", entityTypeInsightsAccount,
		allow(permInsightsAccountRead))
	readOnly := testSet("set-5", "Read Only", "org-settings-read-only", entityTypeOrganization,
		allow("team:read"))
	policy := stackWrite
	policy.ID, policy.UxPurpose = "policy-1", apitype.PermissionDescriptorUXPurposePolicy
	sets := []apitype.PermissionDescriptorRecord{policy, stackWrite, deployers, envRead, acctRead, readOnly}
	ctx := func(t *testing.T) context.Context {
		return config.WithMockClient(t.Context(), stubPermissionSets(t, sets...))
	}

	t.Run("getStackPermissionSet finds a built-in set by identifier", func(t *testing.T) {
		t.Parallel()
		resp, err := GetStackPermissionSetFunction{}.Invoke(ctx(t), infer.FunctionRequest[GetStackPermissionSetInput]{
			Input: GetStackPermissionSetInput{
				OrganizationName: testFunctionRolesOrgName, BuiltIn: new(BuiltInStackPermissionSetWrite),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, PermissionSetOutput{
			PermissionSetId: "set-1",
			Name:            "Stack Write",
			Description:     "Stack Write description",
			Scopes:          []string{permStackRead, permStackAnnotationsUpdate},
			BuiltIn:         new("stack-write"),
		}, resp.Output)
	})

	t.Run("getStackPermissionSet finds a custom set by case-insensitive name and expands the sets it contains",
		func(t *testing.T) {
			t.Parallel()
			resp, err := GetStackPermissionSetFunction{}.Invoke(ctx(t), infer.FunctionRequest[GetStackPermissionSetInput]{
				Input: GetStackPermissionSetInput{OrganizationName: testFunctionRolesOrgName, Name: new(" deployers ")},
			})
			require.NoError(t, err)
			assert.Equal(t, "set-2", resp.Output.PermissionSetId)
			assert.Nil(t, resp.Output.BuiltIn)
			assert.Equal(t, []string{"stack_deployment:create", permStackRead, permStackAnnotationsUpdate},
				resp.Output.Scopes)
		})

	t.Run("getEnvironmentPermissionSet", func(t *testing.T) {
		t.Parallel()
		resp, err := GetEnvironmentPermissionSetFunction{}.Invoke(ctx(t),
			infer.FunctionRequest[GetEnvironmentPermissionSetInput]{Input: GetEnvironmentPermissionSetInput{
				OrganizationName: testFunctionRolesOrgName, BuiltIn: new(BuiltInEnvironmentPermissionSetRead),
			}})
		require.NoError(t, err)
		assert.Equal(t, []string{permEnvironmentRead}, resp.Output.Scopes)
	})

	t.Run("getAccountPermissionSet", func(t *testing.T) {
		t.Parallel()
		resp, err := GetAccountPermissionSetFunction{}.Invoke(ctx(t),
			infer.FunctionRequest[GetAccountPermissionSetInput]{Input: GetAccountPermissionSetInput{
				OrganizationName: testFunctionRolesOrgName, BuiltIn: new(BuiltInAccountPermissionSetRead),
			}})
		require.NoError(t, err)
		assert.Equal(t, []string{permInsightsAccountRead}, resp.Output.Scopes)
	})

	t.Run("getOrganizationPermissionSet", func(t *testing.T) {
		t.Parallel()
		resp, err := GetOrganizationPermissionSetFunction{}.Invoke(ctx(t),
			infer.FunctionRequest[GetOrganizationPermissionSetInput]{Input: GetOrganizationPermissionSetInput{
				OrganizationName: testFunctionRolesOrgName, BuiltIn: new(BuiltInOrganizationPermissionSetReadOnly),
			}})
		require.NoError(t, err)
		assert.Equal(t, []string{"team:read"}, resp.Output.Scopes)
	})

	t.Run("scopes the provider doesn't recognize are left out and reported", func(t *testing.T) {
		t.Parallel()
		withNew := testSet("set-6", "Stack Write", "stack-write", entityTypeStack,
			allow(permStackRead, "stack_brand_new:read"))
		out, unknown, err := stackPermissionSets.lookup(
			config.WithMockClient(t.Context(), stubPermissionSets(t, withNew)),
			testFunctionRolesOrgName, new("stack-write"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{permStackRead}, out.Scopes)
		assert.Equal(t, []string{"stack_brand_new:read"}, unknown)
		assert.Equal(t, `Permission set "Stack Write" contains 1 scope(s) that this version of the `+
			"pulumiservice provider doesn't recognize: stack_brand_new:read. Roles built from this set won't "+
			"grant them. Every other scope in the set is granted as usual. Upgrade the pulumiservice provider "+
			"to grant them.", unknownScopesWarning(out.Name, unknown))
	})

	for _, tt := range []struct {
		name    string
		kind    permissionSetKind
		sets    []apitype.PermissionDescriptorRecord
		builtIn *string
		setName *string
		err     string
	}{
		{"neither builtIn nor name", stackPermissionSets, sets, nil, nil,
			"exactly one of `builtIn` or `name` must be set"},
		{"both builtIn and name", stackPermissionSets, sets, new("stack-write"), new("Stack Write"),
			"exactly one of `builtIn` or `name` must be set"},
		{"no set with the name", stackPermissionSets, sets, nil, new("Nope"),
			`organization "test-org" has no stack permission set named "Nope"`},
		{"a set of another entity type", stackPermissionSets, sets, nil, new("Environment Read"),
			`permission set "Environment Read" in organization "test-org" applies to environment, not stack`},
		{"policies are not permission sets", stackPermissionSets, []apitype.PermissionDescriptorRecord{policy},
			new("stack-write"), nil, `has no built-in stack permission set "stack-write"`},
		{"two sets with the name", stackPermissionSets, []apitype.PermissionDescriptorRecord{deployers, deployers},
			nil, new("Deployers"), "has 2 of stack permission set named"},
		{"conditional set", stackPermissionSets, []apitype.PermissionDescriptorRecord{
			testSet("set-7", "Tagged", "", entityTypeStack,
				condition(stackHasTag(testTagKeyTeam), allow(permStackRead))),
		}, nil, new("Tagged"), `permission set "Tagged": it contains a PermissionDescriptorCondition`},
		{"set that contains itself", stackPermissionSets, []apitype.PermissionDescriptorRecord{
			testSet("set-8", "Loop", "", entityTypeStack, compose("set-9")),
			testSet("set-9", "Pool", "", entityTypeStack, compose("set-8")),
		}, nil, new("Loop"), `it contains itself through permission set "set-8"`},
		{"set that contains a missing set", stackPermissionSets, []apitype.PermissionDescriptorRecord{
			testSet("set-10", "Dangling", "", entityTypeStack, compose("set-404")),
		}, nil, new("Dangling"), `"set-404", which wasn't found`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := tt.kind.lookup(config.WithMockClient(t.Context(), stubPermissionSets(t, tt.sets...)),
				testFunctionRolesOrgName, tt.builtIn, tt.setName)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}
