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

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestStackResourceID(t *testing.T) {
	id := stackResourceID(pulumiapi.StackIdentifier{
		OrgName:     gcMyOrg,
		ProjectName: gcMyProject,
		StackName:   "my-stack",
	})
	assert.Equal(t, "my-org/my-project/my-stack", id)
}

func TestSplitStackResourceID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		org, project, stack, err := splitStackResourceID("my-org/my-project/my-stack")
		require.NoError(t, err)
		assert.Equal(t, gcMyOrg, org)
		assert.Equal(t, gcMyProject, project)
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
