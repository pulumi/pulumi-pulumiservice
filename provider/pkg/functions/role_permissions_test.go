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

// assertScopedAllowShape verifies a helper's output is the collapsed
// {kind: "allow", on: {entityType: identity}, permissions: [...]} shape.
func assertScopedAllowShape(
	t *testing.T,
	got map[string]interface{},
	expectedEntityType string,
	expectedIdentity string,
	expectedPermissions []string,
) {
	t.Helper()

	assert.Equal(t, "allow", got["kind"])

	on, ok := got["on"].(map[string]interface{})
	require.True(t, ok, "on should be map[string]interface{}; got %T", got["on"])
	require.Len(t, on, 1, "on must have exactly one key")
	assert.Equal(t, expectedIdentity, on[expectedEntityType],
		"on.%s should be %q", expectedEntityType, expectedIdentity)

	rawPerms, ok := got["permissions"].([]interface{})
	require.True(t, ok, "permissions should be []interface{}")
	gotPerms := make([]string, len(rawPerms))
	for i, p := range rawPerms {
		gotPerms[i], _ = p.(string)
	}
	assert.Equal(t, expectedPermissions, gotPerms)
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
		assertScopedAllowShape(
			t, resp.Output.Permissions,
			"environment", "env-uuid-1",
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
		assertScopedAllowShape(
			t, resp.Output.Permissions,
			"stack", "stack-id-1",
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
		assertScopedAllowShape(
			t, resp.Output.Permissions,
			"insightsAccount", "acct-1",
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
