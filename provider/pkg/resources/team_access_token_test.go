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

// State written by the pre-infer TeamAccessToken implementation embedded a
// duplicate `__inputs` map alongside the real outputs. After migrating to
// infer, decoding such state fails with "Unrecognized field '__inputs'" unless
// migrateTeamAccessTokenLegacyInputs runs first.
func TestTeamAccessTokenLegacyInputsMigration(t *testing.T) {
	t.Run("migrates legacy state", func(t *testing.T) {
		legacyInputs := property.NewMap(map[string]property.Value{
			"name":             property.New("test-token"),
			"organizationName": property.New("my-org"),
			"teamName":         property.New("my-team"),
			"description":      property.New("example team token"),
		})
		legacy := property.NewMap(map[string]property.Value{
			"__inputs":         property.New(legacyInputs),
			"name":             property.New("test-token"),
			"organizationName": property.New("my-org"),
			"teamName":         property.New("my-team"),
			"description":      property.New("example team token"),
			"value":            property.New("tok-secret-value"),
		})

		got, err := migrateTeamAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		desc := "example team token"
		assert.Equal(t, &TeamAccessTokenState{
			TeamAccessTokenInput: TeamAccessTokenInput{
				Name:             "test-token",
				OrganizationName: "my-org",
				TeamName:         "my-team",
				Description:      &desc,
			},
			Value: "tok-secret-value",
		}, got.Result)
	})

	t.Run("migrates legacy state without description", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			"__inputs":         property.New(property.NewMap(nil)),
			"name":             property.New("test-token"),
			"organizationName": property.New("my-org"),
			"teamName":         property.New("my-team"),
			"value":            property.New("tok-secret-value"),
		})

		got, err := migrateTeamAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &TeamAccessTokenState{
			TeamAccessTokenInput: TeamAccessTokenInput{
				Name:             "test-token",
				OrganizationName: "my-org",
				TeamName:         "my-team",
			},
			Value: "tok-secret-value",
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			"name":             property.New("test-token"),
			"organizationName": property.New("my-org"),
			"teamName":         property.New("my-team"),
			"value":            property.New("tok-secret-value"),
		})

		got, err := migrateTeamAccessTokenLegacyInputs(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result, "migration must not fire when __inputs is absent")
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&TeamAccessToken{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}
