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

func TestEnvironmentVersionTagID(t *testing.T) {
	t.Run("formats with all parts", func(t *testing.T) {
		assert.Equal(t,
			"my-org/my-proj/my-env/my-tag",
			environmentVersionTagID("my-org", "my-proj", "my-env", "my-tag"),
		)
	})
}

func TestSplitEnvironmentVersionTagID(t *testing.T) {
	t.Run("four-part id", func(t *testing.T) {
		org, project, env, tag, err := splitEnvironmentVersionTagID("my-org/my-proj/my-env/my-tag")
		require.NoError(t, err)
		assert.Equal(t, "my-org", org)
		assert.Equal(t, "my-proj", project)
		assert.Equal(t, "my-env", env)
		assert.Equal(t, "my-tag", tag)
	})

	t.Run("legacy three-part id assumes default project", func(t *testing.T) {
		org, project, env, tag, err := splitEnvironmentVersionTagID("my-org/my-env/my-tag")
		require.NoError(t, err)
		assert.Equal(t, "my-org", org)
		assert.Equal(t, defaultProject, project)
		assert.Equal(t, "my-env", env)
		assert.Equal(t, "my-tag", tag)
	})

	t.Run("malformed id", func(t *testing.T) {
		_, _, _, _, err := splitEnvironmentVersionTagID("just-one")
		require.Error(t, err)
	})

	t.Run("too many parts", func(t *testing.T) {
		_, _, _, _, err := splitEnvironmentVersionTagID("a/b/c/d/e")
		require.Error(t, err)
	})
}

func TestEnvironmentVersionTagLegacyStateMigration(t *testing.T) {
	t.Run("migrates legacy state with __inputs", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			"__inputs": property.New(property.NewMap(map[string]property.Value{
				"organization": property.New("my-org"),
				"project":      property.New("my-proj"),
				"environment":  property.New("my-env"),
				"tagName":      property.New("my-tag"),
				"revision":     property.New(float64(7)),
			})),
			"organization": property.New("my-org"),
			"project":      property.New("my-proj"),
			"environment":  property.New("my-env"),
			"tagName":      property.New("my-tag"),
			"revision":     property.New(float64(7)),
		})

		got, err := migrateEnvironmentVersionTagLegacyState(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &EnvironmentVersionTagState{
			EnvironmentVersionTagInput: EnvironmentVersionTagInput{
				Organization: "my-org",
				Project:      "my-proj",
				Environment:  "my-env",
				TagName:      "my-tag",
				Revision:     7,
			},
		}, got.Result)
	})

	t.Run("migrates legacy state without project (defaults to 'default')", func(t *testing.T) {
		// Pre-infer code permitted state without an explicit project; the
		// schema default of "default" had to be applied during decode.
		legacy := property.NewMap(map[string]property.Value{
			"__inputs":     property.New(property.NewMap(map[string]property.Value{})),
			"organization": property.New("my-org"),
			"environment":  property.New("my-env"),
			"tagName":      property.New("my-tag"),
			"revision":     property.New(float64(3)),
		})

		got, err := migrateEnvironmentVersionTagLegacyState(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &EnvironmentVersionTagState{
			EnvironmentVersionTagInput: EnvironmentVersionTagInput{
				Organization: "my-org",
				Project:      defaultProject,
				Environment:  "my-env",
				TagName:      "my-tag",
				Revision:     3,
			},
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			"organization": property.New("my-org"),
			"project":      property.New("my-proj"),
			"environment":  property.New("my-env"),
			"tagName":      property.New("my-tag"),
			"revision":     property.New(float64(1)),
		})

		got, err := migrateEnvironmentVersionTagLegacyState(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result)
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&EnvironmentVersionTag{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}
