// Copyright 2016-2026, Pulumi Corporation.
package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestStackConfigEnvironmentRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("managed round trip", func(t *testing.T) {
		env := &StackConfigEnvironment{Managed: true}
		pm := env.toPropertyMap()
		got := parseStackConfigEnvironment(resource.NewObjectProperty(pm))
		require.NotNil(t, got)
		assert.True(t, got.Managed)
	})

	t.Run("nil and managed=false are not set", func(t *testing.T) {
		var env *StackConfigEnvironment
		assert.False(t, env.IsSet())
		assert.False(t, (&StackConfigEnvironment{}).IsSet())
		assert.False(t, (&StackConfigEnvironment{Managed: false}).IsSet())
	})
}

func TestEnvRefForCreate(t *testing.T) {
	t.Parallel()
	stack := pulumiapi.StackIdentifier{
		OrgName: "acme", ProjectName: "website", StackName: "production",
	}
	assert.Equal(t, "website/production", envRefForCreate(stack))
}

func TestStackResourceDiff(t *testing.T) {
	t.Parallel()
	res := &PulumiServiceStackResource{}

	base := func() resource.PropertyMap {
		return resource.PropertyMap{
			"organizationName": resource.NewStringProperty("acme"),
			"projectName":      resource.NewStringProperty("website"),
			"stackName":        resource.NewStringProperty("production"),
		}
	}

	diff := func(olds, news resource.PropertyMap) *pulumirpc.DiffResponse {
		oldS, err := plugin.MarshalProperties(olds, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		newS, err := plugin.MarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		resp, err := res.Diff(&pulumirpc.DiffRequest{OldInputs: oldS, News: newS})
		require.NoError(t, err)
		return resp
	}

	t.Run("no diff when identical", func(t *testing.T) {
		resp := diff(base(), base())
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
	})

	t.Run("toggle managed on forces replace", func(t *testing.T) {
		olds := base()
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"managed": resource.NewBoolProperty(true),
		})
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got := resp.DetailedDiff["configEnvironment"]
		require.NotNil(t, got)
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})

	t.Run("toggle managed off forces replace", func(t *testing.T) {
		olds := base()
		olds["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"managed": resource.NewBoolProperty(true),
		})
		news := base()
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got := resp.DetailedDiff["configEnvironment"]
		require.NotNil(t, got)
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})

	t.Run("computed managed forces replace at preview", func(t *testing.T) {
		// At preview, managed may arrive wired to another resource's
		// Output<T>. parseStackConfigEnvironment skips computed values, so
		// without the conservative override Diff would report "no change"
		// and let users approve a plan that misses the destructive replace
		// the apply will actually require.
		olds := base()
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"managed": resource.MakeComputed(resource.NewBoolProperty(false)),
		})
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got := resp.DetailedDiff["configEnvironment"]
		require.NotNil(t, got)
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})

	t.Run("project name change forces replace", func(t *testing.T) {
		olds := base()
		news := base()
		news["projectName"] = resource.NewStringProperty("other")
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got, ok := resp.DetailedDiff["projectName"]
		require.True(t, ok)
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})
}
