// Copyright 2016-2026, Pulumi Corporation.
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

	"github.com/pulumi/pulumi-go-provider/infer"
)

const (
	permEnvironmentOpen     = "environment:open"
	permEnvironmentRead     = "environment:read"
	permInsightsAccountRead = "insights_account:read"
	permStackRead           = "stack:read"

	keyType                  = "__type"
	keyPermissionDescriptors = "permissionDescriptors"
	typeCompose              = "PermissionDescriptorCompose"
	testEnvironmentID        = "env-uuid-1"
	testStackID              = "stack-id-1"
	testInsightsAccountID    = "acct-1"
	testSetID                = "set-1"
	testPolicyID             = "policy-1"
	testTagKey               = "team"
	entityEnvironment        = "environment"
)

// assertScopedConditionShape verifies a helper's output is the
// `PermissionDescriptorCondition(Equal(Expression<E>, Literal<E>(id)),
// Allow(perms))` shape. The wire format and SDK boundary share the
// `__type` discriminator at every level (Pulumi's Python SDK preserves
// `__`-prefixed keys as of pulumi/pulumi#22834).
func assertScopedConditionShape(
	t *testing.T,
	got map[string]interface{},
	expectedExpressionType string,
	expectedLiteralType string,
	expectedIdentity string,
	expectedPermissions []string,
) {
	t.Helper()

	assert.Equal(t, "PermissionDescriptorCondition", got[keyType],
		"top-level __type must be PermissionDescriptorCondition")

	cond, ok := got["condition"].(map[string]interface{})
	require.True(t, ok, "condition must be a map; got %T", got["condition"])
	assert.Equal(t, "PermissionExpressionEqual", cond[keyType])

	left, ok := cond["left"].(map[string]interface{})
	require.True(t, ok, "condition.left must be a map; got %T", cond["left"])
	assert.Equal(t, expectedExpressionType, left[keyType])

	right, ok := cond["right"].(map[string]interface{})
	require.True(t, ok, "condition.right must be a map; got %T", cond["right"])
	assert.Equal(t, expectedLiteralType, right[keyType])
	assert.Equal(t, expectedIdentity, right["identity"])

	sub, ok := got["subNode"].(map[string]interface{})
	require.True(t, ok, "subNode must be a map; got %T", got["subNode"])
	assert.Equal(t, "PermissionDescriptorAllow", sub[keyType])

	rawPerms, ok := sub["permissions"].([]interface{})
	require.True(t, ok, "subNode.permissions must be a list; got %T", sub["permissions"])
	gotPerms := make([]string, len(rawPerms))
	for i, p := range rawPerms {
		gotPerms[i], _ = p.(string)
	}
	assert.Equal(t, expectedPermissions, gotPerms)
}

func TestBuildAllowPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildAllowPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildAllowPermissionsInput]{
				Input: BuildAllowPermissionsInput{
					Permissions: []string{permStackRead, permEnvironmentOpen},
				},
			},
		)
		require.NoError(t, err)
		got := resp.Output.Permissions
		assert.Equal(t, "PermissionDescriptorAllow", got[keyType])
		// Permissions list passes through verbatim.
		perms, ok := got["permissions"].([]interface{})
		require.True(t, ok)
		assert.Equal(t, []interface{}{permStackRead, permEnvironmentOpen}, perms)
	})

	t.Run("rejects empty permissions", func(t *testing.T) {
		t.Parallel()
		_, err := BuildAllowPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildAllowPermissionsInput]{
				Input: BuildAllowPermissionsInput{},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}

func TestBuildEnvironmentScopedPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildEnvironmentScopedPermissionsInput]{
				Input: BuildEnvironmentScopedPermissionsInput{
					EnvironmentID: testEnvironmentID,
					Permissions:   []string{permEnvironmentRead, permEnvironmentOpen},
				},
			},
		)
		require.NoError(t, err)
		assertScopedConditionShape(
			t, resp.Output.Permissions,
			"PermissionExpressionEnvironment",
			"PermissionLiteralExpressionEnvironment",
			testEnvironmentID,
			[]string{permEnvironmentRead, permEnvironmentOpen},
		)
	})

	t.Run("rejects empty environmentId", func(t *testing.T) {
		t.Parallel()
		_, err := BuildEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildEnvironmentScopedPermissionsInput]{
				Input: BuildEnvironmentScopedPermissionsInput{
					Permissions: []string{permEnvironmentRead},
				},
			},
		)
		assert.ErrorContains(t, err, "environmentId")
	})

	t.Run("rejects empty permissions", func(t *testing.T) {
		t.Parallel()
		_, err := BuildEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildEnvironmentScopedPermissionsInput]{
				Input: BuildEnvironmentScopedPermissionsInput{
					EnvironmentID: testEnvironmentID,
				},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}

func TestBuildStackScopedPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildStackScopedPermissionsInput]{
				Input: BuildStackScopedPermissionsInput{
					StackID:     testStackID,
					Permissions: []string{permStackRead},
				},
			},
		)
		require.NoError(t, err)
		assertScopedConditionShape(
			t, resp.Output.Permissions,
			"PermissionExpressionStack",
			"PermissionLiteralExpressionStack",
			testStackID,
			[]string{permStackRead},
		)
	})

	t.Run("rejects empty stackId", func(t *testing.T) {
		t.Parallel()
		_, err := BuildStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildStackScopedPermissionsInput]{
				Input: BuildStackScopedPermissionsInput{
					Permissions: []string{permStackRead},
				},
			},
		)
		assert.ErrorContains(t, err, "stackId")
	})

	t.Run("rejects empty permissions", func(t *testing.T) {
		t.Parallel()
		_, err := BuildStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildStackScopedPermissionsInput]{
				Input: BuildStackScopedPermissionsInput{
					StackID: testStackID,
				},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}

func TestBuildInsightsAccountScopedPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildInsightsAccountScopedPermissionsInput]{
				Input: BuildInsightsAccountScopedPermissionsInput{
					InsightsAccountID: testInsightsAccountID,
					Permissions:       []string{permInsightsAccountRead},
				},
			},
		)
		require.NoError(t, err)
		assertScopedConditionShape(
			t, resp.Output.Permissions,
			"PermissionExpressionInsightsAccount",
			"PermissionLiteralExpressionInsightsAccount",
			testInsightsAccountID,
			[]string{permInsightsAccountRead},
		)
	})

	t.Run("rejects empty insightsAccountId", func(t *testing.T) {
		t.Parallel()
		_, err := BuildInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildInsightsAccountScopedPermissionsInput]{
				Input: BuildInsightsAccountScopedPermissionsInput{
					Permissions: []string{permInsightsAccountRead},
				},
			},
		)
		assert.ErrorContains(t, err, "insightsAccountId")
	})

	t.Run("rejects empty permissions", func(t *testing.T) {
		t.Parallel()
		_, err := BuildInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildInsightsAccountScopedPermissionsInput]{
				Input: BuildInsightsAccountScopedPermissionsInput{
					InsightsAccountID: testInsightsAccountID,
				},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}

func TestScopedPermissionsWithSetIDs(t *testing.T) {
	t.Parallel()

	t.Run("wraps a Compose of the sets", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildEnvironmentScopedPermissionsInput]{
				Input: BuildEnvironmentScopedPermissionsInput{
					EnvironmentID: testEnvironmentID,
					SetIDs:        []string{testSetID, "set-2"},
				},
			},
		)
		require.NoError(t, err)
		sub, ok := resp.Output.Permissions["subNode"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, typeCompose, sub[keyType])
		assert.Equal(t, []interface{}{testSetID, "set-2"}, sub[keyPermissionDescriptors])
	})

	t.Run("rejects both permissions and setIds", func(t *testing.T) {
		t.Parallel()
		_, err := BuildStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildStackScopedPermissionsInput]{
				Input: BuildStackScopedPermissionsInput{
					StackID:     testStackID,
					Permissions: []string{permStackRead},
					SetIDs:      []string{testSetID},
				},
			},
		)
		assert.ErrorContains(t, err, "exactly one of `permissions` or `setIds`")
	})

	t.Run("rejects an empty set ID", func(t *testing.T) {
		t.Parallel()
		_, err := BuildInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildInsightsAccountScopedPermissionsInput]{
				Input: BuildInsightsAccountScopedPermissionsInput{
					InsightsAccountID: testInsightsAccountID,
					SetIDs:            []string{""},
				},
			},
		)
		assert.ErrorContains(t, err, "setIds")
	})
}

func TestBuildTagConditionalPermissions(t *testing.T) {
	t.Parallel()

	t.Run("tag value gates on Equal(Tag, String)", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildTagConditionalPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildTagConditionalPermissionsInput]{
				Input: BuildTagConditionalPermissionsInput{
					EntityType: "stack",
					TagKey:     testTagKey,
					TagValue:   "platform",
					SetIDs:     []string{testSetID},
				},
			},
		)
		require.NoError(t, err)
		got := resp.Output.Permissions
		assert.Equal(t, "PermissionDescriptorCondition", got[keyType])
		assert.Equal(t, map[string]interface{}{
			keyType: "PermissionExpressionEqual",
			"left": map[string]interface{}{
				keyType:   "PermissionExpressionTag",
				"context": map[string]interface{}{keyType: "PermissionExpressionStack"},
				"key":     testTagKey,
			},
			"right": map[string]interface{}{
				keyType: "PermissionLiteralExpressionString",
				"value": "platform",
			},
		}, got["condition"])
		assert.Equal(t, map[string]interface{}{
			keyType:                  typeCompose,
			keyPermissionDescriptors: []interface{}{testSetID},
		}, got["subNode"])
	})

	t.Run("no tag value gates on HasTag", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildTagConditionalPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildTagConditionalPermissionsInput]{
				Input: BuildTagConditionalPermissionsInput{
					EntityType:  "insights-account",
					TagKey:      "owner",
					Permissions: []string{permInsightsAccountRead},
				},
			},
		)
		require.NoError(t, err)
		got := resp.Output.Permissions
		assert.Equal(t, map[string]interface{}{
			keyType:   "PermissionExpressionHasTag",
			"context": map[string]interface{}{keyType: "PermissionExpressionInsightsAccount"},
			"key":     "owner",
		}, got["condition"])
		sub, ok := got["subNode"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "PermissionDescriptorAllow", sub[keyType])
	})

	t.Run("rejects an unknown entity type", func(t *testing.T) {
		t.Parallel()
		_, err := BuildTagConditionalPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildTagConditionalPermissionsInput]{
				Input: BuildTagConditionalPermissionsInput{
					EntityType: testTagKey,
					TagKey:     "k",
					SetIDs:     []string{testSetID},
				},
			},
		)
		assert.ErrorContains(t, err, "entityType")
	})

	t.Run("rejects an empty tag key", func(t *testing.T) {
		t.Parallel()
		_, err := BuildTagConditionalPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildTagConditionalPermissionsInput]{
				Input: BuildTagConditionalPermissionsInput{
					EntityType: entityEnvironment,
					SetIDs:     []string{testSetID},
				},
			},
		)
		assert.ErrorContains(t, err, "tagKey")
	})

	t.Run("rejects a missing grant", func(t *testing.T) {
		t.Parallel()
		_, err := BuildTagConditionalPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildTagConditionalPermissionsInput]{
				Input: BuildTagConditionalPermissionsInput{
					EntityType: entityEnvironment,
					TagKey:     "k",
				},
			},
		)
		assert.ErrorContains(t, err, "exactly one of `permissions` or `setIds`")
	})
}

func TestBuildComposePermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildComposePermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildComposePermissionsInput]{
				Input: BuildComposePermissionsInput{PermissionDescriptorIDs: []string{testPolicyID}},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			keyType:                  typeCompose,
			keyPermissionDescriptors: []interface{}{testPolicyID},
		}, resp.Output.Permissions)
	})

	t.Run("rejects empty IDs", func(t *testing.T) {
		t.Parallel()
		_, err := BuildComposePermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildComposePermissionsInput]{
				Input: BuildComposePermissionsInput{PermissionDescriptorIDs: []string{testPolicyID, ""}},
			},
		)
		assert.ErrorContains(t, err, "permissionDescriptorIds")
	})
}

func TestBuildGroupPermissions(t *testing.T) {
	t.Parallel()

	entry := map[string]interface{}{
		keyType:                  typeCompose,
		keyPermissionDescriptors: []interface{}{testSetID},
	}

	t.Run("passes entries through verbatim", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildGroupPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildGroupPermissionsInput]{
				Input: BuildGroupPermissionsInput{Entries: []map[string]any{entry, entry}},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			keyType:   "PermissionDescriptorGroup",
			"entries": []map[string]any{entry, entry},
		}, resp.Output.Permissions)
	})

	t.Run("rejects no entries", func(t *testing.T) {
		t.Parallel()
		_, err := BuildGroupPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildGroupPermissionsInput]{Input: BuildGroupPermissionsInput{}},
		)
		assert.ErrorContains(t, err, "entries")
	})

	t.Run("rejects an entry that is not a descriptor", func(t *testing.T) {
		t.Parallel()
		_, err := BuildGroupPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildGroupPermissionsInput]{
				Input: BuildGroupPermissionsInput{Entries: []map[string]any{{"permissions": []any{"stack:read"}}}},
			},
		)
		assert.ErrorContains(t, err, "entries[0]")
	})
}
