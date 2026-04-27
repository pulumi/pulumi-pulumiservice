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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// assertScopedPermissionsShape verifies the outer envelope produced by every
// scoped-permissions helper is a single-entry Group → Condition(Equal(left,
// right)) → Allow(perms), and returns the literal sub-tree for caller-specific
// assertions.
func assertScopedPermissionsShape(
	t *testing.T,
	got map[string]interface{},
	expectedExpression, expectedLiteral string,
	expectedIdentity string,
	expectedPermissions []string,
) map[string]interface{} {
	t.Helper()

	assert.Equal(t, "PermissionDescriptorGroup", got["__type"])

	entries, ok := got["entries"].([]interface{})
	require.True(t, ok, "entries should be []interface{}")
	require.Len(t, entries, 1)

	condition, ok := entries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "PermissionDescriptorCondition", condition["__type"])

	cond, ok := condition["condition"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "PermissionExpressionEqual", cond["__type"])

	left, ok := cond["left"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, expectedExpression, left["__type"])

	right, ok := cond["right"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, expectedLiteral, right["__type"])
	assert.Equal(t, expectedIdentity, right["identity"])

	subNode, ok := condition["subNode"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "PermissionDescriptorAllow", subNode["__type"])

	rawPerms, ok := subNode["permissions"].([]interface{})
	require.True(t, ok)
	gotPerms := make([]string, len(rawPerms))
	for i, p := range rawPerms {
		gotPerms[i], _ = p.(string)
	}
	assert.Equal(t, expectedPermissions, gotPerms)

	return right
}

func TestScopedPermissionsDescriptor_RoundTripsThroughJSON(t *testing.T) {
	t.Parallel()

	d := scopedPermissionsDescriptor(
		"PermissionExpressionEnvironment",
		"PermissionLiteralExpressionEnvironment",
		"env-uuid-1",
		[]string{"environment:read"},
	)

	// Producing the descriptor via map[string]interface{} is only useful if it
	// survives a round-trip through json.Marshal — that's how the role
	// resource ships it to the service.
	raw, err := json.Marshal(d)
	require.NoError(t, err)
	parsed := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(raw, &parsed))

	assertScopedPermissionsShape(
		t, parsed,
		"PermissionExpressionEnvironment",
		"PermissionLiteralExpressionEnvironment",
		"env-uuid-1",
		[]string{"environment:read"},
	)
}

func TestGetEnvironmentScopedPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := GetEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetEnvironmentScopedPermissionsInput]{
				Input: GetEnvironmentScopedPermissionsInput{
					EnvironmentID: "env-uuid-1",
					Permissions:   []string{"environment:read", "environment:open"},
				},
			},
		)
		require.NoError(t, err)
		assertScopedPermissionsShape(
			t, resp.Output.Permissions,
			"PermissionExpressionEnvironment",
			"PermissionLiteralExpressionEnvironment",
			"env-uuid-1",
			[]string{"environment:read", "environment:open"},
		)
		assertPermissionsJSONMatches(t, resp.Output.PermissionsJson, resp.Output.Permissions)
	})

	t.Run("rejects empty environmentId", func(t *testing.T) {
		t.Parallel()
		_, err := GetEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetEnvironmentScopedPermissionsInput]{
				Input: GetEnvironmentScopedPermissionsInput{
					Permissions: []string{"environment:read"},
				},
			},
		)
		assert.ErrorContains(t, err, "environmentId")
	})

	t.Run("rejects empty permissions", func(t *testing.T) {
		t.Parallel()
		_, err := GetEnvironmentScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetEnvironmentScopedPermissionsInput]{
				Input: GetEnvironmentScopedPermissionsInput{
					EnvironmentID: "env-uuid-1",
				},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}

// assertPermissionsJSONMatches confirms that the helper's `permissionsJson`
// output decodes back to the structured `permissions` Mapping. The JSON
// sibling is the workaround Python users round-trip through `json.loads` to
// dodge the SDK's `__`-prefix strip on invoke responses; if it ever drifts
// from the structured form the workaround quietly stops being equivalent.
func assertPermissionsJSONMatches(t *testing.T, gotJSON string, gotMap map[string]interface{}) {
	t.Helper()
	require.NotEmpty(t, gotJSON, "permissionsJson must be set")
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(gotJSON), &parsed))
	assert.Equal(t, gotMap, parsed, "permissionsJson must decode to the same descriptor as permissions")
}

func TestGetStackScopedPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := GetStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetStackScopedPermissionsInput]{
				Input: GetStackScopedPermissionsInput{
					StackID:     "stack-id-1",
					Permissions: []string{"stack:read"},
				},
			},
		)
		require.NoError(t, err)
		assertScopedPermissionsShape(
			t, resp.Output.Permissions,
			"PermissionExpressionStack",
			"PermissionLiteralExpressionStack",
			"stack-id-1",
			[]string{"stack:read"},
		)
		assertPermissionsJSONMatches(t, resp.Output.PermissionsJson, resp.Output.Permissions)
	})

	t.Run("rejects empty stackId", func(t *testing.T) {
		t.Parallel()
		_, err := GetStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetStackScopedPermissionsInput]{
				Input: GetStackScopedPermissionsInput{
					Permissions: []string{"stack:read"},
				},
			},
		)
		assert.ErrorContains(t, err, "stackId")
	})

	t.Run("rejects empty permissions", func(t *testing.T) {
		t.Parallel()
		_, err := GetStackScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetStackScopedPermissionsInput]{
				Input: GetStackScopedPermissionsInput{
					StackID: "stack-id-1",
				},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}

func TestGetInsightsAccountScopedPermissions(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp, err := GetInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetInsightsAccountScopedPermissionsInput]{
				Input: GetInsightsAccountScopedPermissionsInput{
					InsightsAccountID: "acct-1",
					Permissions:       []string{"insights-account:read"},
				},
			},
		)
		require.NoError(t, err)
		assertScopedPermissionsShape(
			t, resp.Output.Permissions,
			"PermissionExpressionInsightsAccount",
			"PermissionLiteralExpressionInsightsAccount",
			"acct-1",
			[]string{"insights-account:read"},
		)
		assertPermissionsJSONMatches(t, resp.Output.PermissionsJson, resp.Output.Permissions)
	})

	t.Run("rejects empty insightsAccountId", func(t *testing.T) {
		t.Parallel()
		_, err := GetInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetInsightsAccountScopedPermissionsInput]{
				Input: GetInsightsAccountScopedPermissionsInput{
					Permissions: []string{"insights-account:read"},
				},
			},
		)
		assert.ErrorContains(t, err, "insightsAccountId")
	})

	t.Run("rejects empty permissions", func(t *testing.T) {
		t.Parallel()
		_, err := GetInsightsAccountScopedPermissionsFunction{}.Invoke(
			context.Background(),
			infer.FunctionRequest[GetInsightsAccountScopedPermissionsInput]{
				Input: GetInsightsAccountScopedPermissionsInput{
					InsightsAccountID: "acct-1",
				},
			},
		)
		assert.ErrorContains(t, err, "permissions")
	})
}
