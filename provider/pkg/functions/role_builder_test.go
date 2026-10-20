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
	testAccountID  = "acct-1"
	testTagKeyTeam = "team"
	testTagKeyCost = "cost-center"
	testUnknownOp  = "unknown operator"
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

func stackHasTag(key string) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionHasTagBuilder{
		Context: apitype.PermissionExpressionStackBuilder{}.Build(), Key: key,
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

func TestBuildRolePermissions(t *testing.T) {
	t.Parallel()

	envAllowMap, err := toSDKMap(condition(
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
			name: "organization-level scopes",
			in: BuildRolePermissionsInput{
				OrganizationScopes: []RbacOrganizationScope{RbacOrganizationScopeTeamRead},
			},
			expected: group(allow("team:read")),
		},
		{
			name: "all-entity rules fold into the unconditional entries, deduplicated",
			in: BuildRolePermissionsInput{
				OrganizationScopes: []RbacOrganizationScope{RbacOrganizationScopeTeamRead},
				StackRules: []RoleStackRule{{
					RoleRuleSelector: RoleRuleSelector{All: ptr(true)},
					Scopes:           []RbacStackScope{RbacStackScopeStackRead},
				}},
				EnvironmentRules: []RoleEnvironmentRule{{
					RoleRuleSelector: RoleRuleSelector{All: ptr(true)},
					Scopes: []RbacEnvironmentScope{
						RbacEnvironmentScopeEnvironmentRead, RbacEnvironmentScopeEnvironmentRead,
					},
				}},
			},
			expected: group(allow("team:read", permStackRead, permEnvironmentRead)),
		},
		{
			name: "one stack by id",
			in: BuildRolePermissionsInput{
				StackRules: []RoleStackRule{{
					RoleRuleSelector: RoleRuleSelector{Id: ptr("stack-1")},
					Scopes:           []RbacStackScope{RbacStackScopeStackRead},
				}},
			},
			expected: group(condition(
				apitype.PermissionExpressionEqualBuilder{
					Left:  apitype.PermissionExpressionStackBuilder{}.Build(),
					Right: apitype.PermissionLiteralExpressionStackBuilder{Identity: "stack-1"}.Build(),
				}.Build(),
				allow(permStackRead),
			)),
		},
		{
			name: "one cloud account by id",
			in: BuildRolePermissionsInput{
				AccountRules: []RoleAccountRule{{
					RoleRuleSelector: RoleRuleSelector{Id: ptr(testAccountID)},
					Scopes:           []RbacAccountScope{RbacAccountScopeInsightsAccountRead},
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
			name: "tag equals, not-equals, exists, not-exists, and AND",
			in: BuildRolePermissionsInput{
				StackRules: []RoleStackRule{
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{
							{Key: testTagKeyTeam, Value: ptr("platform")},
						}},
						Scopes: []RbacStackScope{RbacStackScopeStackRead},
					},
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{
							{Key: "env", Value: ptr("prod"), Operator: ptr(TagOperatorNotEquals)},
						}},
						Scopes: []RbacStackScope{RbacStackScopeStackRead},
					},
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{
							{Key: testTagKeyCost, Operator: ptr(TagOperatorExists)},
						}},
						Scopes: []RbacStackScope{RbacStackScopeStackRead},
					},
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{
							{Key: testTagKeyCost, Operator: ptr(TagOperatorNotExists)},
						}},
						Scopes: []RbacStackScope{RbacStackScopeStackRead},
					},
					{
						RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{
							{Key: testTagKeyTeam, Value: ptr("platform"), Operator: ptr(TagOperatorEquals)},
							{Key: "env", Value: ptr("prod"), Operator: ptr(TagOperatorNotEquals)},
						}},
						Scopes: []RbacStackScope{RbacStackScopeStackRead},
					},
				},
			},
			expected: group(
				condition(stackTagEquals(testTagKeyTeam, "platform"), allow(permStackRead)),
				condition(not(stackTagEquals("env", "prod")), allow(permStackRead)),
				condition(stackHasTag(testTagKeyCost), allow(permStackRead)),
				condition(not(stackHasTag(testTagKeyCost)), allow(permStackRead)),
				condition(and(stackTagEquals(testTagKeyTeam, "platform"), not(stackTagEquals("env", "prod"))),
					allow(permStackRead)),
			),
		},
		{
			name: "additional entries from other helpers are appended",
			in: BuildRolePermissionsInput{
				OrganizationScopes: []RbacOrganizationScope{RbacOrganizationScopeTeamRead},
				AdditionalEntries:  []map[string]any{envAllowMap},
			},
			expected: group(allow("team:read"), condition(
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
			want, err := toSDKMap(tt.expected)
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
		}, "stackRules[0]: `scopes` must not be empty"},
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
				RoleRuleSelector: RoleRuleSelector{Id: ptr("x"), Tags: []TagCondition{{Key: "k", Value: ptr("v")}}},
				Scopes:           []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "exactly one of `all`, `id`, or `tags`"},
		{"empty tag key", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{{Key: " "}}},
				Scopes:           []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "tags[0]: `key` must not be empty"},
		{"equals without a value", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{{Key: testTagKeyCost}}},
				Scopes:           []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "tags[0]: `value` is required with the \"equals\" operator"},
		{"exists with a value", BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{{
					Key: testTagKeyCost, Value: ptr("v"), Operator: ptr(TagOperatorExists),
				}}},
				Scopes: []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, "tags[0]: `value` can't be set with the \"exists\" operator"},
		{testUnknownOp, BuildRolePermissionsInput{
			StackRules: []RoleStackRule{{
				RoleRuleSelector: RoleRuleSelector{Tags: []TagCondition{{
					Key: "k", Value: ptr("v"), Operator: ptr(TagOperator("contains")),
				}}},
				Scopes: []RbacStackScope{RbacStackScopeStackRead},
			}},
		}, testUnknownOp},
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
			AccountRules: []RoleAccountRule{{
				RoleRuleSelector: all,
				Scopes:           []RbacAccountScope{RbacAccountScope("insights_account:fly")},
			}},
		}, "is not a cloud account scope"},
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
