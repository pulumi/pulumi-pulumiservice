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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
)

const (
	testOrgSetID   = "org-set"
	testStackSetID = "stack-set"
	testEnvSetID   = "env-set"
	testStackUUID  = "20dede8f-f399-4deb-bd19-174410f209c6"
	testEnvUUID    = "58dd45c3-23fd-40b0-9b7e-3db51e09877b"
)

// testSetTypes stands in for the organization's permission sets.
func testSetTypes(id string) (RbacResourceType, error) {
	switch id {
	case testStackSetID:
		return RbacResourceTypeStack, nil
	case testEnvSetID:
		return RbacResourceTypeEnvironment, nil
	default:
		return RbacResourceTypeGlobal, nil
	}
}

// Shorthands for building descriptor trees with the Cloud SDK.

func stackExpr() apitype.PermissionContextExpression {
	return apitype.PermissionExpressionStackBuilder{}.Build()
}

func envExpr() apitype.PermissionContextExpression {
	return apitype.PermissionExpressionEnvironmentBuilder{}.Build()
}

func condition(c apitype.PermissionBooleanExpression, sub apitype.PermissionDescriptor) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorConditionBuilder{Condition: c, SubNode: sub}.Build()
}

func equal(l, r apitype.PermissionExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionEqualBuilder{Left: l, Right: r}.Build()
}

func tagEquals(ctx apitype.PermissionContextExpression, key, value string) apitype.PermissionBooleanExpression {
	return equal(
		apitype.PermissionExpressionTagBuilder{Context: ctx, Key: key}.Build(),
		apitype.PermissionLiteralExpressionStringBuilder{Value: value}.Build(),
	)
}

func hasTag(ctx apitype.PermissionContextExpression, key string) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionHasTagBuilder{Context: ctx, Key: key}.Build()
}

func not(n apitype.PermissionBooleanExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionNotBuilder{
		PermissionBooleanExpressionUnaryBuilder: apitype.PermissionBooleanExpressionUnaryBuilder{Node: n},
	}.Build()
}

func and(l, r apitype.PermissionBooleanExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionAndBuilder{
		PermissionBooleanExpressionBinaryBuilder: apitype.PermissionBooleanExpressionBinaryBuilder{Left: l, Right: r},
	}.Build()
}

func group(entries ...apitype.PermissionDescriptor) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorGroupBuilder{Entries: entries}.Build()
}

func allow(scopes ...apitype.RbacPermission) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorAllowBuilder{Permissions: scopes}.Build()
}

// consoleRoleTree mirrors the shapes the Pulumi Cloud console writes, captured
// from roles created in the console: org-level access and "all stacks" share
// one unconditional Compose, then one Condition per entity or tag rule.
func consoleRoleTree() apitype.PermissionDescriptor {
	return group(
		compose([]string{testOrgSetID, testStackSetID}),
		condition(
			equal(stackExpr(), apitype.PermissionLiteralExpressionStackBuilder{Identity: testStackUUID}.Build()),
			compose([]string{testStackSetID})),
		condition(
			equal(envExpr(), apitype.PermissionLiteralExpressionEnvironmentBuilder{Identity: testEnvUUID}.Build()),
			compose([]string{testEnvSetID})),
		condition(
			and(and(tagEquals(stackExpr(), "Owner", "DevTeam"), not(tagEquals(stackExpr(), "env", "prod"))),
				hasTag(stackExpr(), "team")),
			compose([]string{testStackSetID})),
	)
}

func consoleRoleModel() RbacRoleCore {
	return RbacRoleCore{
		OrganizationPermissionSetIds: []string{testOrgSetID},
		EntityRules: []RbacEntityRule{
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{All: pointerTo(true)}},
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{Id: pointerTo(testStackUUID)}},
			{PermissionSetIds: []string{testEnvSetID}, Environment: &RbacEntitySelector{Id: pointerTo(testEnvUUID)}},
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{Tags: []RbacTagCondition{
				{Key: "Owner", Value: pointerTo("DevTeam")},
				{Key: "env", Value: pointerTo("prod"), Operator: pointerTo(RbacTagOperatorNotEquals)},
				{Key: "team"},
			}}},
		},
	}
}

func TestBuildPolicyDetailsMatchesConsole(t *testing.T) {
	t.Parallel()
	details, err := buildPolicyDetails(consoleRoleModel())
	require.NoError(t, err)
	assert.Equal(t, consoleRoleTree(), details)
}

func TestParsePolicyDetailsRoundTrip(t *testing.T) {
	t.Parallel()
	org, rules, err := parsePolicyDetails(consoleRoleTree(), testSetTypes)
	require.NoError(t, err)

	want := consoleRoleModel()
	assert.Equal(t, want.OrganizationPermissionSetIds, org)
	// Tag values come back as explicit pointers and default operators as nil;
	// compare through the canonical form, which normalizes both.
	assert.Equal(t, canonicalRoleKey(want.OrganizationPermissionSetIds, want.EntityRules),
		canonicalRoleKey(org, rules))
	require.Len(t, rules, 4)
	assert.Equal(t, pointerTo(true), rules[0].Stack.All, "all-stacks rule comes from the unconditional Compose")
}

func TestParsePolicyDetailsEnvironmentTagRule(t *testing.T) {
	t.Parallel()
	_, rules, err := parsePolicyDetails(
		group(condition(not(hasTag(envExpr(), "sensitive")), compose([]string{testEnvSetID}))),
		testSetTypes)
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.NotNil(t, rules[0].Environment)
	assert.Equal(t, []RbacTagCondition{{Key: "sensitive", Operator: pointerTo(RbacTagOperatorNotEquals)}},
		rules[0].Environment.Tags)
}

func TestParsePolicyDetailsUnrepresentable(t *testing.T) {
	t.Parallel()
	or := apitype.PermissionExpressionOrBuilder{
		PermissionBooleanExpressionBinaryBuilder: apitype.PermissionBooleanExpressionBinaryBuilder{
			Left: hasTag(stackExpr(), "a"), Right: hasTag(stackExpr(), "b"),
		},
	}.Build()
	cases := map[string]apitype.PermissionDescriptor{
		"raw allow":          allow(apitype.RbacPermissionStackRead),
		"allow inside group": group(allow(apitype.RbacPermissionStackRead)),
		"or condition":       group(condition(or, compose([]string{testStackSetID}))),
		"mixed tag contexts": group(condition(
			and(hasTag(stackExpr(), "a"), hasTag(envExpr(), "b")), compose([]string{testStackSetID}))),
		"condition granting allow": group(condition(hasTag(stackExpr(), "a"), allow(apitype.RbacPermissionStackRead))),
	}
	for name, details := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := parsePolicyDetails(details, testSetTypes)
			assert.True(t, errors.Is(err, errUnrepresentable), "got %v", err)
		})
	}
}

func TestBuildPolicyDetailsDedupesUnconditionalSets(t *testing.T) {
	t.Parallel()
	details, err := buildPolicyDetails(RbacRoleCore{
		OrganizationPermissionSetIds: []string{testOrgSetID},
		EntityRules: []RbacEntityRule{
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{All: pointerTo(true)}},
			{PermissionSetIds: []string{testStackSetID}, Stack: &RbacEntitySelector{All: pointerTo(true)}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, group(compose([]string{testOrgSetID, testStackSetID})), details)
}

func TestCanonicalRoleKeyIgnoresLayout(t *testing.T) {
	t.Parallel()
	a := consoleRoleModel()
	b := consoleRoleModel()
	slicesReverse(b.EntityRules)
	b.EntityRules[0].Stack.Tags[0].Operator = pointerTo(RbacTagOperatorEquals) // explicit default
	assert.Equal(t, canonicalRoleKey(a.OrganizationPermissionSetIds, a.EntityRules),
		canonicalRoleKey(b.OrganizationPermissionSetIds, b.EntityRules))

	b.EntityRules[0].Stack.Tags[0].Value = pointerTo("OtherTeam")
	assert.NotEqual(t, canonicalRoleKey(a.OrganizationPermissionSetIds, a.EntityRules),
		canonicalRoleKey(b.OrganizationPermissionSetIds, b.EntityRules))
}

func slicesReverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
