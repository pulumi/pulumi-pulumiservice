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

func TestParseTemplateSourceID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		org, id, err := parseTemplateSourceID("my-org/my-template-id")
		require.NoError(t, err)
		assert.Equal(t, gcMyOrg, org)
		assert.Equal(t, "my-template-id", id)
	})

	t.Run("invalid", func(t *testing.T) {
		_, _, err := parseTemplateSourceID("just-one-part")
		require.Error(t, err)
	})
}

func TestTemplateSourceLegacyStateMigration(t *testing.T) {
	t.Run("migrates legacy state with __inputs", func(t *testing.T) {
		destURL := "https://github.com/pulumi/pulumi"
		legacy := property.NewMap(map[string]property.Value{
			gcInputs: property.New(property.NewMap(map[string]property.Value{
				gcOrganizationName: property.New(gcMyOrg),
				gcSourceName:       property.New("bootstrap"),
				gcSourceURL:        property.New("https://github.com/pulumi/pulumi"),
			})),
			gcOrganizationName: property.New(gcMyOrg),
			gcSourceName:       property.New("bootstrap"),
			gcSourceURL:        property.New("https://github.com/pulumi/pulumi"),
			"destination": property.New(property.NewMap(map[string]property.Value{
				"url": property.New(destURL),
			})),
		})

		got, err := migrateTemplateSourceLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &TemplateSourceState{
			TemplateSourceInput: TemplateSourceInput{
				OrganizationName: gcMyOrg,
				SourceName:       "bootstrap",
				SourceURL:        "https://github.com/pulumi/pulumi",
				Destination:      &TemplateSourceDestination{URL: &destURL},
			},
		}, got.Result)
	})

	t.Run("migrates legacy state without destination", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			gcInputs:           property.New(property.NewMap(map[string]property.Value{})),
			gcOrganizationName: property.New(gcMyOrg),
			gcSourceName:       property.New("bootstrap"),
			gcSourceURL:        property.New("https://example.com"),
		})

		got, err := migrateTemplateSourceLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		require.NotNil(t, got.Result)
		assert.Nil(t, got.Result.Destination)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			gcOrganizationName: property.New(gcMyOrg),
			gcSourceName:       property.New("bootstrap"),
			gcSourceURL:        property.New("https://example.com"),
		})

		got, err := migrateTemplateSourceLegacyInputs(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result)
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&TemplateSource{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}
