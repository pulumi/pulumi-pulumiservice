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

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestStackResourceID(t *testing.T) {
	id := stackResourceID(pulumiapi.StackIdentifier{
		OrgName:     "my-org",
		ProjectName: "my-project",
		StackName:   "my-stack",
	})
	assert.Equal(t, "my-org/my-project/my-stack", id)
}

func TestSplitStackResourceID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		org, project, stack, err := splitStackResourceID("my-org/my-project/my-stack")
		require.NoError(t, err)
		assert.Equal(t, "my-org", org)
		assert.Equal(t, "my-project", project)
		assert.Equal(t, "my-stack", stack)
	})

	t.Run("too few parts", func(t *testing.T) {
		_, _, _, err := splitStackResourceID("my-org/my-project")
		require.Error(t, err)
	})

	t.Run("too many parts", func(t *testing.T) {
		_, _, _, err := splitStackResourceID("a/b/c/d")
		require.Error(t, err)
	})
}

func TestStackLegacyStateMigration(t *testing.T) {
	t.Run("migrates legacy state with __inputs", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			"__inputs": property.New(property.NewMap(map[string]property.Value{
				"organizationName": property.New("my-org"),
				"projectName":      property.New("my-project"),
				"stackName":        property.New("my-stack"),
				"forceDestroy":     property.New(true),
			})),
			"organizationName": property.New("my-org"),
			"projectName":      property.New("my-project"),
			"stackName":        property.New("my-stack"),
			"forceDestroy":     property.New(true),
		})

		got, err := migrateStackLegacyState(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &StackState{
			StackInput: StackInput{
				OrganizationName: "my-org",
				ProjectName:      "my-project",
				StackName:        "my-stack",
				ForceDestroy:     true,
			},
		}, got.Result)
	})

	t.Run("migrates legacy state without forceDestroy", func(t *testing.T) {
		// The legacy code only wrote forceDestroy when it was true; absent
		// values should migrate to the Go zero value (false).
		legacy := property.NewMap(map[string]property.Value{
			"__inputs": property.New(property.NewMap(map[string]property.Value{
				"organizationName": property.New("my-org"),
				"projectName":      property.New("my-project"),
				"stackName":        property.New("my-stack"),
			})),
			"organizationName": property.New("my-org"),
			"projectName":      property.New("my-project"),
			"stackName":        property.New("my-stack"),
		})

		got, err := migrateStackLegacyState(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &StackState{
			StackInput: StackInput{
				OrganizationName: "my-org",
				ProjectName:      "my-project",
				StackName:        "my-stack",
				ForceDestroy:     false,
			},
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			"organizationName": property.New("my-org"),
			"projectName":      property.New("my-project"),
			"stackName":        property.New("my-stack"),
		})

		got, err := migrateStackLegacyState(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result)
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&Stack{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}
