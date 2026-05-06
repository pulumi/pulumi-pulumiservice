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

	assert.Equal(t, "PermissionDescriptorCondition", got["__type"],
		"top-level __type must be PermissionDescriptorCondition")

	cond, ok := got["condition"].(map[string]interface{})
	require.True(t, ok, "condition must be a map; got %T", got["condition"])
	assert.Equal(t, "PermissionExpressionEqual", cond["__type"])

	left, ok := cond["left"].(map[string]interface{})
	require.True(t, ok, "condition.left must be a map; got %T", cond["left"])
	assert.Equal(t, expectedExpressionType, left["__type"])

	right, ok := cond["right"].(map[string]interface{})
	require.True(t, ok, "condition.right must be a map; got %T", cond["right"])
	assert.Equal(t, expectedLiteralType, right["__type"])
	assert.Equal(t, expectedIdentity, right["identity"])

	sub, ok := got["subNode"].(map[string]interface{})
	require.True(t, ok, "subNode must be a map; got %T", got["subNode"])
	assert.Equal(t, "PermissionDescriptorAllow", sub["__type"])

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
					Permissions: []string{"stack:read", "environment:open"},
				},
			},
		)
		require.NoError(t, err)
		got := resp.Output.Permissions
		assert.Equal(t, "PermissionDescriptorAllow", got["__type"])
		// Permissions list passes through verbatim.
		perms, ok := got["permissions"].([]interface{})
		require.True(t, ok)
		assert.Equal(t, []interface{}{"stack:read", "environment:open"}, perms)
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
					Permissions:       []string{"insights_account:read"},
				},
			},
		)
		require.NoError(t, err)
		assertScopedConditionShape(
			t, resp.Output.Permissions,
			"PermissionExpressionInsightsAccount",
			"PermissionLiteralExpressionInsightsAccount",
			"acct-1",
			[]string{"insights_account:read"},
		)
	})

	t.Run("rejects empty insightsAccountId", func(t *testing.T) {
		t.Parallel()
		_, err := BuildInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[BuildInsightsAccountScopedPermissionsInput]{
				Input: BuildInsightsAccountScopedPermissionsInput{
					Permissions: []string{"insights_account:read"},
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
