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
)

func TestEnvironmentResourceID(t *testing.T) {
	assert.Equal(t, "org/proj/env", environmentResourceID("org", "proj", "env"))
}

func TestSplitEnvironmentID(t *testing.T) {
	t.Run("canonical three-part", func(t *testing.T) {
		org, proj, env, err := splitEnvironmentID("org/proj/env")
		require.NoError(t, err)
		assert.Equal(t, "org", org)
		assert.Equal(t, "proj", proj)
		assert.Equal(t, "env", env)
	})

	t.Run("legacy two-part", func(t *testing.T) {
		org, proj, env, err := splitEnvironmentID("org/env")
		require.NoError(t, err)
		assert.Equal(t, "org", org)
		assert.Equal(t, defaultProject, proj)
		assert.Equal(t, "env", env)
	})

	t.Run("malformed single segment", func(t *testing.T) {
		_, _, _, err := splitEnvironmentID("just-one")
		require.Error(t, err)
	})

	t.Run("malformed too many segments", func(t *testing.T) {
		_, _, _, err := splitEnvironmentID("a/b/c/d")
		require.Error(t, err)
	})
}

func TestProjectOrDefault(t *testing.T) {
	assert.Equal(t, defaultProject, projectOrDefault(""))
	assert.Equal(t, "my-project", projectOrDefault("my-project"))
}
