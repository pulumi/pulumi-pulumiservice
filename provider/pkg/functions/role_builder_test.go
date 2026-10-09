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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
)

const (
	testOrgReadSetID   = "org-read"
	testStackReadSetID = "stack-read"
	testAccountID      = "acct-1"
	testTagKeyTeam     = "team"
)

func ptr[T any](v T) *T { return &v }

func group(entries ...apitype.PermissionDescriptor) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorGroupBuilder{Entries: entries}.Build()
}

func allow(scopes ...apitype.RbacPermission) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorAllowBuilder{Permissions: scopes}.Build()
}

func condition(c apitype.PermissionBooleanExpression, sub apitype.PermissionDescriptor) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorConditionBuilder{Condition: c, SubNode: sub}.Build()
}

func stackTagEquals(key, value string) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionEqualBuilder{
		Left: apitype.PermissionExpressionTagBuilder{
			Context: apitype.PermissionExpressionStackBuilder{}.Build(), Key: key,
		}.Build(),
		Right: apitype.PermissionLiteralExpressionStringBuilder{Value: value}.Build(),
	}.Build()
}

func not(e apitype.PermissionBooleanExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionNotBuilder{
		PermissionBooleanExpressionUnaryBuilder: apitype.PermissionBooleanExpressionUnaryBuilder{Node: e},
	}.Build()
}

func and(l, r apitype.PermissionBooleanExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionAndBuilder{
		PermissionBooleanExpressionBinaryBuilder: apitype.PermissionBooleanExpressionBinaryBuilder{Left: l, Right: r},
	}.Build()
}

func invokeBuildRolePermissions(t *testing.T, in BuildRolePermissionsInput) (map[string]any, error) {
	t.Helper()
	resp, err := BuildRolePermissionsFunction{}.Invoke(t.Context(),
		infer.FunctionRequest[BuildRolePermissionsInput]{Input: in})
	return resp.Output.Permissions, err
}

func setRef(id, resourceType string) RolePermissionSetRef {
	return RolePermissionSetRef{PermissionSetId: id, ResourceType: resourceType}
}

func TestBuildRolePermissions(t *testing.T) {
	t.Parallel()

	envAllowMap, err := descriptorToSDKMap(condition(
		apitype.PermissionExpressionEqualBuilder{
			Left:  apitype.PermissionExpressionEnvironmentBuilder{}.Build(),
			Right: apitype.PermissionLiteralExpressionEnvironmentBuilder{Identity: "env-1"}.Build(),
		}.Build(),
		allow(permEnvironmentRead),
	))
	require.NoError(t, err)

	tests := []struct {
		name     string
		in       BuildRolePermissionsInput
		expected apitype.PermissionDescriptor
	}{
		{
			name: "organization-level sets and scopes",
			in: BuildRolePermissionsInput{
				OrganizationPermissionSets: []RolePermissionSetRef{setRef(testOrgReadSetID, "global")},
				OrganizationScopes:         []RbacOrganizationScope{RbacOrganizationScopeTeamRead},
			},
			expected: group(compose([]string{testOrgReadSetID}), allow("team:read")),
		},
		{
			name: "all-entity rules fold into the unconditional entries, deduplicated",
			in: BuildRolePermissionsInput{
				OrganizationPermissionSets: []RolePermissionSetRef{setRef(testOrgReadSetID, "global")},
				StackRules: []RoleStackRule{{
					RoleRuleSelector: RoleRuleSelector{All: ptr(true)},
					PermissionSets:   []RolePermissionSetRef{setRef(testStackReadSetID, "stack")},
				}},
				EnvironmentRules: []RoleEnvironmentRule{{
					RoleRuleSelector: RoleRuleSelector{All: ptr(true)},
					Scopes:           []RbacEnvironmentScope{RbacEnvironmentScopeEnvironmentRead},
				}},
			},
			expected: group(compose([]string{testOrgReadSetID, testStackReadSetID}), allow(permEnvironmentRead)),
		},
		{
			name: "one stack by id, granting sets and scopes",
			in: BuildRolePermissionsInput{
				StackRules: []RoleStackRule{{
					RoleRuleSelector: RoleRuleSelector{Id: ptr("stack-1")},
					Scopes:           []RbacStackScope{RbacStackScopeStackRead},
					PermissionSets:   []RolePermissionSetRef{setRef("deployer", "stack")},
				}},
			},
			expected: group(condition(
				apitype.PermissionExpressionEqualBuilder{
					Left:  apitype.PermissionExpressionStackBuilder{}.Build(),
					Right: apitype.PermissionLiteralExpressionStackBuilder{Identity: "stack-1"}.Build(),
				}.Build(),
				group(compose([]string{"deployer"}), allow(permStackRead)),
			)),
		},
		{
			name: "one insights account by id",
			in: BuildRolePermissionsInput{
				InsightsAccountRules: []RoleInsightsAccountRule{{
					RoleRuleSelector: RoleRuleSelector{Id: ptr(testAccountID)},
					Scopes:           []RbacInsightsAccountScope{RbacInsightsAccountScopeInsightsAccountRead},
				}},
			},
			expected: group(condition(
				apitype.PermissionExpressionEqualBuilder{
					Left:  apitype.PermissionExpressionInsightsAccountBuilder{}.Build(),
					Right: apitype.PermissionLiteralExpressionInsightsAccountBuilder{Identity: testAccountID}.Build(),
				}.Build(),
				allow(permInsightsAccountRead),
			)),
		},
		{
			name: "tag equals, not-equals, has-tag, and AND",
			in: BuildRolePermissionsInput{
				StackRules: []RoleStackRule{
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []RoleTagCondition{
							{Key: testTagKeyTeam, Value: ptr("platform")},
						}},
						PermissionSets: []RolePermissionSetRef{setRef("d", "stack")},
					},
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []RoleTagCondition{
							{Key: "env", Value: ptr("prod"), Operator: ptr(RoleTagOperatorNotEquals)},
						}},
						PermissionSets: []RolePermissionSetRef{setRef("d", "stack")},
					},
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []RoleTagCondition{{Key: "cost-center"}}},
						PermissionSets:   []RolePermissionSetRef{setRef("d", "stack")},
					},
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []RoleTagCondition{
							{Key: testTagKeyTeam, Value: ptr("platform"), Operator: ptr(RoleTagOperatorEquals)},
							{Key: "env", Value: ptr("prod"), Operator: ptr(RoleTagOperatorNotEquals)},
						}},
						PermissionSets: []RolePermissionSetRef{setRef("d", "stack")},
					},
				},
			},
			expected: group(
				condition(stackTagEquals(testTagKeyTeam, "platform"), compose([]string{"d"})),
				condition(not(stackTagEquals("env", "prod")), compose([]string{"d"})),
				condition(apitype.PermissionExpressionHasTagBuilder{
					Context: apitype.PermissionExpressionStackBuilder{}.Build(), Key: "cost-center",
				}.Build(), compose([]string{"d"})),
				condition(and(stackTagEquals(testTagKeyTeam, "platform"), not(stackTagEquals("env", "prod"))),
					compose([]string{"d"})),
			),
		},
		{
			name: "additional entries from other helpers are appended",
			in: BuildRolePermissionsInput{
				OrganizationPermissionSets: []RolePermissionSetRef{setRef(testOrgReadSetID, "global")},
				AdditionalEntries:          []map[string]any{envAllowMap},
			},
			expected: group(compose([]string{testOrgReadSetID}), condition(
				apitype.PermissionExpressionEqualBuilder{
					Left:  apitype.PermissionExpressionEnvironmentBuilder{}.Build(),
					Right: apitype.PermissionLiteralExpressionEnvironmentBuilder{Identity: "env-1"}.Build(),
				}.Build(),
				allow(permEnvironmentRead),
			)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := invokeBuildRolePermissions(t, tt.in)
			require.NoError(t, err)
			want, err := descriptorToSDKMap(tt.expected)
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestBuildRolePermissionsErrors(t *testing.T) {
	t.Parallel()

	all := RoleRuleSelector{All: ptr(true)}
	tests := []struct {
		name string
		in   BuildRolePermissionsInput
		err  string
	}{
		{"empty input", BuildRolePermissionsInput{}, "at least one of"},
		{"rule without grants", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{RoleRuleSelector: all}},
		}, "stackRules[0]: at least one of `scopes` or `permissionSets`"},
		{"empty selector", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{Scopes: []RbacStackScope{RbacStackScopeStackRead}}},
		}, "stackRules[0]: exactly one of `all`, `id`, or `tags`"},
		{"all combined with id", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{All: ptr(true), Id: ptr("x")},
				Scopes:           []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "exactly one of `all`, `id`, or `tags`"},
		{"id combined with tags", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{Id: ptr("x"), Tags: []RoleTagCondition{{Key: "k"}}},
				Scopes:           []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "exactly one of `all`, `id`, or `tags`"},
		{"empty tag key", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{Tags: []RoleTagCondition{{Key: " "}}},
				Scopes:           []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "tags[0]: `key` must not be empty"},
		{"unknown operator", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{Tags: []RoleTagCondition{{
					Key: "k", Value: ptr("v"), Operator: ptr(RoleTagOperator("contains")),
				}}},
				Scopes: []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "unknown operator"},
		// The fail-open case verified against Pulumi Cloud: an `all` rule is
		// encoded without a condition, so an environment scope in a stack
		// rule would grant it on every environment.
		{"environment scope in an all-stacks rule", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: all,
				Scopes:           []RbacStackScope{RbacStackScope("environment:read")},
			}},
		}, "stackRules[0].scopes: `environment:read` is not a stack scope"},
		{"stack scope in an environment rule", BuildRolePermissionsInput{
			EnvironmentRules: []RoleEnvironmentRule{{
				RoleRuleSelector: RoleRuleSelector{Id: ptr("env-1")},
				Scopes:           []RbacEnvironmentScope{RbacEnvironmentScope("stack:write")},
			}},
		}, "environmentRules[0].scopes: `stack:write` is not an environment scope"},
		{"stack scope at organization level", BuildRolePermissionsInput{
			OrganizationScopes: []RbacOrganizationScope{RbacOrganizationScope("stack:read")},
		}, "organizationScopes: `stack:read` is not an organization scope"},
		{"unknown scope", BuildRolePermissionsInput{
			InsightsAccountRules: []RoleInsightsAccountRule{{
				RoleRuleSelector: all,
				Scopes:           []RbacInsightsAccountScope{RbacInsightsAccountScope("insights_account:fly")},
			}},
		}, "is not an Insights account scope"},
		{"stack permission set in an all-environments rule", BuildRolePermissionsInput{
			EnvironmentRules: []RoleEnvironmentRule{{
				RoleRuleSelector: all,
				PermissionSets:   []RolePermissionSetRef{setRef("stack-write", "stack")},
			}},
		}, "environmentRules[0].permissionSets: [0]: permission set \"stack-write\" is for `stack`, not `environment`"},
		{"stack permission set at organization level", BuildRolePermissionsInput{
			OrganizationPermissionSets: []RolePermissionSetRef{setRef("stack-write", "stack")},
		}, "organizationPermissionSets: [0]: permission set \"stack-write\" is for `stack`, not `global`"},
		{"permission set without an id", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: all,
				PermissionSets:   []RolePermissionSetRef{setRef("", "stack")},
			}},
		}, "`permissionSetId` must not be empty"},
		{"additional entry without __type", BuildRolePermissionsInput{
			AdditionalEntries: []map[string]any{{"permissions": []any{permStackRead}}},
		}, "additionalEntries[0]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := invokeBuildRolePermissions(t, tt.in)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

func TestBuildComposePermissions(t *testing.T) {
	t.Parallel()

	invoke := func(ids ...string) (map[string]any, error) {
		resp, err := BuildComposePermissionsFunction{}.Invoke(t.Context(),
			infer.FunctionRequest[BuildComposePermissionsInput]{Input: BuildComposePermissionsInput{Ids: ids}})
		return resp.Output.Permissions, err
	}

	got, err := invoke("policy-1", "policy-2", "policy-1")
	require.NoError(t, err)
	want, err := descriptorToSDKMap(compose([]string{"policy-1", "policy-2"}))
	require.NoError(t, err)
	assert.Equal(t, want, got)

	_, err = invoke()
	require.ErrorContains(t, err, "`ids` must not be empty")
	_, err = invoke("policy-1", " ")
	require.ErrorContains(t, err, "ids[1] must not be empty")
}
