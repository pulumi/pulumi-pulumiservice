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

func TestEnvironmentVersionTagID(t *testing.T) {
	t.Run("formats with all parts", func(t *testing.T) {
		assert.Equal(t,
			"my-org/my-proj/my-env/my-tag",
			environmentVersionTagID(gcMyOrg, "my-proj", "my-env", "my-tag"),
		)
	})
}

func TestSplitEnvironmentVersionTagID(t *testing.T) {
	t.Run("four-part id", func(t *testing.T) {
		org, project, env, tag, err := splitEnvironmentVersionTagID("my-org/my-proj/my-env/my-tag")
		require.NoError(t, err)
		assert.Equal(t, gcMyOrg, org)
		assert.Equal(t, "my-proj", project)
		assert.Equal(t, "my-env", env)
		assert.Equal(t, "my-tag", tag)
	})

	t.Run("legacy three-part id assumes default project", func(t *testing.T) {
		org, project, env, tag, err := splitEnvironmentVersionTagID("my-org/my-env/my-tag")
		require.NoError(t, err)
		assert.Equal(t, gcMyOrg, org)
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
