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

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/urn"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

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

// TestStackDiffForceDestroy pins that forceDestroy never replaces the stack.
// Providers before v1.1.0 omitted a false forceDestroy from state, and v1.1.0
// through v1.3.0 declared the property replaceOnChanges, so the first update
// after upgrading scheduled a replacement of every existing stack.
func TestStackDiffForceDestroy(t *testing.T) {
	t.Parallel()

	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&Stack{})).
		Build()
	require.NoError(t, err)
	server, err := integration.NewServer(t.Context(), "pulumiservice", semver.MustParse("0.0.0"),
		integration.WithProvider(prov))
	require.NoError(t, err)

	identity := map[string]property.Value{
		"organizationName": property.New("org"),
		"projectName":      property.New("proj"),
		"stackName":        property.New("dev"),
	}
	withForceDestroy := func(v bool) property.Map {
		m := map[string]property.Value{}
		for k, val := range identity {
			m[k] = val
		}
		m["forceDestroy"] = property.New(v)
		return property.NewMap(m)
	}
	stackURN := urn.New("test", "proj", "", "pulumiservice:index:Stack", "s")
	const id = "org/proj/dev"

	t.Run("absent in old state, false in new inputs", func(t *testing.T) {
		resp, err := server.Diff(p.DiffRequest{
			ID:        id,
			Urn:       stackURN,
			State:     property.NewMap(identity),
			OldInputs: property.NewMap(identity),
			Inputs:    withForceDestroy(false),
		})
		require.NoError(t, err)
		for k, d := range resp.DetailedDiff {
			assert.NotContains(t, []p.DiffKind{p.AddReplace, p.UpdateReplace, p.DeleteReplace}, d.Kind, k)
		}
	})

	t.Run("flipped", func(t *testing.T) {
		resp, err := server.Diff(p.DiffRequest{
			ID:        id,
			Urn:       stackURN,
			State:     withForceDestroy(false),
			OldInputs: withForceDestroy(false),
			Inputs:    withForceDestroy(true),
		})
		require.NoError(t, err)
		require.True(t, resp.HasChanges)
		require.Contains(t, resp.DetailedDiff, "forceDestroy")
		assert.Equal(t, p.Update, resp.DetailedDiff["forceDestroy"].Kind)
	})

	t.Run("identity still replaces", func(t *testing.T) {
		news := withForceDestroy(false)
		resp, err := server.Diff(p.DiffRequest{
			ID:        id,
			Urn:       stackURN,
			State:     withForceDestroy(false),
			OldInputs: withForceDestroy(false),
			Inputs:    news.Set("stackName", property.New("prod")),
		})
		require.NoError(t, err)
		require.Contains(t, resp.DetailedDiff, "stackName")
		assert.Equal(t, p.UpdateReplace, resp.DetailedDiff["stackName"].Kind)
	})
}
