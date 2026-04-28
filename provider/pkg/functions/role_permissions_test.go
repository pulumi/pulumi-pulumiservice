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

// assertScopedPermissionsShape verifies the outer envelope produced by every
// scoped-permissions helper is a single-entry Group → Condition(Equal(left,
// right)) → Allow(perms), expressed in the user-facing `kind` form. Returns
// the literal sub-tree for caller-specific assertions.
func assertScopedPermissionsShape(
	t *testing.T,
	got map[string]interface{},
	expectedExpressionKind, expectedLiteralKind string,
	expectedIdentity string,
	expectedPermissions []string,
) map[string]interface{} {
	t.Helper()

	assert.Equal(t, "group", got["kind"])

	entries, ok := got["entries"].([]interface{})
	require.True(t, ok, "entries should be []interface{}")
	require.Len(t, entries, 1)

	condition, ok := entries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "condition", condition["kind"])

	cond, ok := condition["condition"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "equal", cond["kind"])

	left, ok := cond["left"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, expectedExpressionKind, left["kind"])

	right, ok := cond["right"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, expectedLiteralKind, right["kind"])
	assert.Equal(t, expectedIdentity, right["identity"])

	subNode, ok := condition["subNode"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "allow", subNode["kind"])

	rawPerms, ok := subNode["permissions"].([]interface{})
	require.True(t, ok)
	gotPerms := make([]string, len(rawPerms))
	for i, p := range rawPerms {
		gotPerms[i], _ = p.(string)
	}
	assert.Equal(t, expectedPermissions, gotPerms)

	return right
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
		assertScopedPermissionsShape(
			t, resp.Output.Permissions,
			kindExpressionEnvironment,
			kindLiteralEnvironment,
			"env-uuid-1",
			[]string{"environment:read", "environment:open"},
		)
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
		assertScopedPermissionsShape(
			t, resp.Output.Permissions,
			kindExpressionStack,
			kindLiteralStack,
			"stack-id-1",
			[]string{"stack:read"},
		)
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
		assertScopedPermissionsShape(
			t, resp.Output.Permissions,
			kindExpressionInsightsAccount,
			kindLiteralInsightsAccount,
			"acct-1",
			[]string{"insights-account:read"},
		)
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
