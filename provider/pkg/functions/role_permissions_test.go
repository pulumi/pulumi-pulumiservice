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

// assertScopedConditionShape verifies a helper's output is the
// `PermissionDescriptorCondition(Equal(Expression<E>, Literal<E>(id)),
// Allow(perms))` wire shape — the same shape Pulumi Cloud's REST API uses,
// modulo the `__type` → `kind` rename.
func assertScopedConditionShape(
	t *testing.T,
	got map[string]interface{},
	expectedExpressionKind string,
	expectedLiteralKind string,
	expectedIdentity string,
	expectedPermissions []string,
) {
	t.Helper()

	assert.Equal(t, "PermissionDescriptorCondition", got["kind"],
		"top-level kind must be PermissionDescriptorCondition")

	cond, ok := got["condition"].(map[string]interface{})
	require.True(t, ok, "condition must be a map; got %T", got["condition"])
	assert.Equal(t, "PermissionExpressionEqual", cond["kind"],
		"condition kind must be PermissionExpressionEqual")

	left, ok := cond["left"].(map[string]interface{})
	require.True(t, ok, "condition.left must be a map; got %T", cond["left"])
	assert.Equal(t, expectedExpressionKind, left["kind"])

	right, ok := cond["right"].(map[string]interface{})
	require.True(t, ok, "condition.right must be a map; got %T", cond["right"])
	assert.Equal(t, expectedLiteralKind, right["kind"])
	assert.Equal(t, expectedIdentity, right["identity"])

	sub, ok := got["subNode"].(map[string]interface{})
	require.True(t, ok, "subNode must be a map; got %T", got["subNode"])
	assert.Equal(t, "PermissionDescriptorAllow", sub["kind"])

	rawPerms, ok := sub["permissions"].([]interface{})
	require.True(t, ok, "subNode.permissions must be a list; got %T", sub["permissions"])
	gotPerms := make([]string, len(rawPerms))
	for i, p := range rawPerms {
		gotPerms[i], _ = p.(string)
	}
	assert.Equal(t, expectedPermissions, gotPerms)
}

// Each helper's output must be free of the wire-format `__type` discriminator
// — the SDK boundary uses `kind` only. Pulumi's Python SDK silently strips
// `__`-prefixed keys, so emitting `__type` would mean the field disappears at
// the language boundary and the role gets created with a malformed body.
// Recursive check, since the helpers emit nested expressions.
func assertNoUnderscoreType(t *testing.T, v interface{}) {
	t.Helper()
	switch x := v.(type) {
	case map[string]interface{}:
		_, has := x["__type"]
		assert.False(t, has, "helper output must not contain `__type`; the SDK boundary uses `kind`")
		for _, val := range x {
			assertNoUnderscoreType(t, val)
		}
	case []interface{}:
		for _, item := range x {
			assertNoUnderscoreType(t, item)
		}
	}
}

func TestBuildEnvironmentScopedPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := BuildEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildEnvironmentScopedPermissionsInput]{
				Input: BuildEnvironmentScopedPermissionsInput{
					EnvironmentID: "env-uuid-1",
					Permissions:   []string{"environment:read", "environment:open"},
				},
			},
		)
		require.NoError(t, err)
		assertScopedConditionShape(
			t, resp.Output.Permissions,
			"PermissionExpressionEnvironment",
			"PermissionLiteralExpressionEnvironment",
			"env-uuid-1",
			[]string{"environment:read", "environment:open"},
		)
		assertNoUnderscoreType(t, resp.Output.Permissions)
	})

	t.Run("rejects empty environmentId", func(t *testing.T) {
		t.Parallel()
		_, err := BuildEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildEnvironmentScopedPermissionsInput]{
				Input: BuildEnvironmentScopedPermissionsInput{
					Permissions: []string{"environment:read"},
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
					EnvironmentID: "env-uuid-1",
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
					StackID:     "stack-id-1",
					Permissions: []string{"stack:read"},
				},
			},
		)
		require.NoError(t, err)
		assertScopedConditionShape(
			t, resp.Output.Permissions,
			"PermissionExpressionStack",
			"PermissionLiteralExpressionStack",
			"stack-id-1",
			[]string{"stack:read"},
		)
		assertNoUnderscoreType(t, resp.Output.Permissions)
	})

	t.Run("rejects empty stackId", func(t *testing.T) {
		t.Parallel()
		_, err := BuildStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildStackScopedPermissionsInput]{
				Input: BuildStackScopedPermissionsInput{
					Permissions: []string{"stack:read"},
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
					StackID: "stack-id-1",
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
					InsightsAccountID: "acct-1",
					Permissions:       []string{"insights-account:read"},
				},
			},
		)
		require.NoError(t, err)
		assertScopedConditionShape(
			t, resp.Output.Permissions,
			"PermissionExpressionInsightsAccount",
			"PermissionLiteralExpressionInsightsAccount",
			"acct-1",
			[]string{"insights-account:read"},
		)
		assertNoUnderscoreType(t, resp.Output.Permissions)
	})

	t.Run("rejects empty insightsAccountId", func(t *testing.T) {
		t.Parallel()
		_, err := BuildInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildInsightsAccountScopedPermissionsInput]{
				Input: BuildInsightsAccountScopedPermissionsInput{
					Permissions: []string{"insights-account:read"},
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
					InsightsAccountID: "acct-1",
				},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}
