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

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestAgentPoolSplitID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		org, name, id, err := splitAgentPoolID("my-org/my-pool/abc-123")
		require.NoError(t, err)
		assert.Equal(t, "my-org", org)
		assert.Equal(t, "my-pool", name)
		assert.Equal(t, "abc-123", id)
	})

	t.Run("too few parts", func(t *testing.T) {
		_, _, _, err := splitAgentPoolID("my-org/my-pool")
		require.Error(t, err)
	})

	t.Run("too many parts", func(t *testing.T) {
		_, _, _, err := splitAgentPoolID("a/b/c/d")
		require.Error(t, err)
	})
}

func TestAgentPoolLegacyStateMigration(t *testing.T) {
	t.Run("migrates legacy state with __inputs", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			"__inputs": property.New(property.NewMap(map[string]property.Value{
				"name":             property.New("test-pool"),
				"organizationName": property.New("my-org"),
			})),
			"agentPoolID":      property.New("api-id-123"),
			"name":             property.New("test-pool"),
			"organizationName": property.New("my-org"),
			"description":      property.New("a description"),
			"forceDestroy":     property.New(true),
			"tokenValue":       property.New("token-secret"),
		})

		got, err := migrateAgentPoolLegacyState(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &AgentPoolState{
			AgentPoolInput: AgentPoolInput{
				OrganizationName: "my-org",
				Name:             "test-pool",
				Description:      "a description",
				ForceDestroy:     true,
			},
			AgentPoolID: "api-id-123",
			TokenValue:  "token-secret",
		}, got.Result)
	})

	t.Run("migrates state with misnamed agentPoolID key only", func(t *testing.T) {
		// Pre-infer code wrote `agentPoolID` (capital D) even though the
		// schema declared `agentPoolId`; the migration must rename the key.
		legacy := property.NewMap(map[string]property.Value{
			"agentPoolID":      property.New("api-id-456"),
			"name":             property.New("pool"),
			"organizationName": property.New("org"),
			"tokenValue":       property.New("tok"),
		})

		got, err := migrateAgentPoolLegacyState(t.Context(), legacy)
		require.NoError(t, err)
		require.NotNil(t, got.Result)
		assert.Equal(t, "api-id-456", got.Result.AgentPoolID)
	})

	t.Run("preserves secret-marked tokenValue", func(t *testing.T) {
		// Secrets in property.Value are flagged on the value itself, not
		// wrapped — IsString() returns true on a secret string.
		legacy := property.NewMap(map[string]property.Value{
			"__inputs":         property.New(property.NewMap(map[string]property.Value{})),
			"agentPoolID":      property.New("id"),
			"name":             property.New("pool"),
			"organizationName": property.New("org"),
			"tokenValue":       property.New("tok-secret").WithSecret(true),
		})

		got, err := migrateAgentPoolLegacyState(t.Context(), legacy)
		require.NoError(t, err)
		require.NotNil(t, got.Result)
		assert.Equal(t, "tok-secret", got.Result.TokenValue)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			"agentPoolId":      property.New("api-id"),
			"name":             property.New("pool"),
			"organizationName": property.New("org"),
			"tokenValue":       property.New("tok"),
		})

		got, err := migrateAgentPoolLegacyState(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result)
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&AgentPool{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}
