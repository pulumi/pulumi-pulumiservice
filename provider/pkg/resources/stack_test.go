// Copyright 2016-2026, Pulumi Corporation.
package resources

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

func TestStackConfigEnvironmentParseAndPropertyMap(t *testing.T) {
	t.Parallel()

	t.Run("auto-only round trip", func(t *testing.T) {
		env := &StackConfigEnvironment{Auto: true}
		pm := env.toPropertyMap()
		got := parseStackConfigEnvironment(resource.NewObjectProperty(pm))
		require.NotNil(t, got)
		assert.True(t, got.Auto)
		assert.Empty(t, got.Project)
		assert.Empty(t, got.Environment)
		assert.Empty(t, got.Version)
	})

	t.Run("explicit env round trip", func(t *testing.T) {
		env := &StackConfigEnvironment{Project: "default", Environment: "prod-cfg", Version: "3"}
		pm := env.toPropertyMap()
		got := parseStackConfigEnvironment(resource.NewObjectProperty(pm))
		require.NotNil(t, got)
		assert.False(t, got.Auto)
		assert.Equal(t, "default", got.Project)
		assert.Equal(t, "prod-cfg", got.Environment)
		assert.Equal(t, "3", got.Version)
	})

	t.Run("nil and empty are not set", func(t *testing.T) {
		var env *StackConfigEnvironment
		assert.False(t, env.IsSet())
		assert.False(t, (&StackConfigEnvironment{}).IsSet())
	})
}

func TestEnvRefForCreate(t *testing.T) {
	t.Parallel()
	stack := pulumiapi.StackIdentifier{
		OrgName: "acme", ProjectName: "website", StackName: "production",
	}

	cases := []struct {
		name string
		env  StackConfigEnvironment
		want string
	}{
		{"auto no version", StackConfigEnvironment{Auto: true}, "website/production"},
		{"explicit env", StackConfigEnvironment{Environment: "prod-cfg"}, "default/prod-cfg"},
		{"explicit project+env", StackConfigEnvironment{Project: "shared", Environment: "prod-cfg"}, "shared/prod-cfg"},
		{
			"explicit env with version",
			StackConfigEnvironment{Project: "shared", Environment: "prod-cfg", Version: "stable"},
			"shared/prod-cfg@stable",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.env.envRefForCreate(stack))
		})
	}
}

func TestConfigEnvDiffKind(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		old, new    *StackConfigEnvironment
		wantChanged bool
		wantKind    plugin.DiffKind
	}{
		{"both unset", nil, nil, false, 0},
		{
			"toggle auto on forces replace",
			nil,
			&StackConfigEnvironment{Auto: true},
			true, plugin.DiffUpdateReplace,
		},
		{
			"toggle auto off forces replace",
			&StackConfigEnvironment{Auto: true},
			nil,
			true, plugin.DiffUpdateReplace,
		},
		{
			"toggle explicit-env on is in-place update",
			nil,
			&StackConfigEnvironment{Project: "default", Environment: "prod-cfg"},
			true, plugin.DiffUpdate,
		},
		{
			"toggle explicit-env off is in-place update",
			&StackConfigEnvironment{Project: "default", Environment: "prod-cfg"},
			nil,
			true, plugin.DiffUpdate,
		},
		{
			"auto flip",
			&StackConfigEnvironment{Auto: true},
			&StackConfigEnvironment{Project: "default", Environment: "prod-cfg"},
			true, plugin.DiffUpdateReplace,
		},
		{
			"auto-mode version-only change is in-place update",
			&StackConfigEnvironment{Auto: true, Version: "2"},
			&StackConfigEnvironment{Auto: true, Version: "3"},
			true, plugin.DiffUpdate,
		},
		{
			"env name change",
			&StackConfigEnvironment{Project: "default", Environment: "a"},
			&StackConfigEnvironment{Project: "default", Environment: "b"},
			true, plugin.DiffUpdateReplace,
		},
		{
			"env project change",
			&StackConfigEnvironment{Project: "p1", Environment: "x"},
			&StackConfigEnvironment{Project: "p2", Environment: "x"},
			true, plugin.DiffUpdateReplace,
		},
		{
			"version only change is in-place update",
			&StackConfigEnvironment{Project: "default", Environment: "prod-cfg", Version: "2"},
			&StackConfigEnvironment{Project: "default", Environment: "prod-cfg", Version: "3"},
			true, plugin.DiffUpdate,
		},
		{
			"no-op equality",
			&StackConfigEnvironment{Project: "default", Environment: "prod-cfg"},
			&StackConfigEnvironment{Project: "default", Environment: "prod-cfg"},
			false, 0,
		},
		{
			"omitted project resolves to default — no diff",
			&StackConfigEnvironment{Project: "default", Environment: "x"},
			&StackConfigEnvironment{Environment: "x"},
			false, 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kind, changed := configEnvDiffKind(tc.old, tc.new)
			assert.Equal(t, tc.wantChanged, changed)
			if changed {
				assert.Equal(t, tc.wantKind, kind)
			}
		})
	}
}

func TestStackResourceCheck(t *testing.T) {
	t.Parallel()
	res := &PulumiServiceStackResource{}

	check := func(news resource.PropertyMap) []*pulumirpc.CheckFailure {
		// KeepUnknowns ensures resource.MakeComputed(...) survives the
		// round-trip so the computed-value tests below exercise the IsComputed
		// path instead of seeing a stripped-out null.
		s, err := plugin.MarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		resp, err := res.Check(&pulumirpc.CheckRequest{News: s})
		require.NoError(t, err)
		return resp.Failures
	}

	base := func() resource.PropertyMap {
		return resource.PropertyMap{
			"organizationName": resource.NewStringProperty("acme"),
			"projectName":      resource.NewStringProperty("website"),
			"stackName":        resource.NewStringProperty("production"),
		}
	}

	t.Run("no configEnvironment is fine", func(t *testing.T) {
		assert.Empty(t, check(base()))
	})

	t.Run("auto only is fine", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto": resource.NewBoolProperty(true),
		})
		assert.Empty(t, check(m))
	})

	t.Run("explicit env is fine", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		assert.Empty(t, check(m))
	})

	t.Run("auto + environment is rejected", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto":        resource.NewBoolProperty(true),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		failures := check(m)
		require.Len(t, failures, 1)
		assert.Equal(t, "configEnvironment", failures[0].Property)
	})

	t.Run("auto + version is rejected", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto":    resource.NewBoolProperty(true),
			"version": resource.NewStringProperty("3"),
		})
		failures := check(m)
		require.Len(t, failures, 1)
		assert.Equal(t, "configEnvironment", failures[0].Property)
	})

	t.Run("neither auto nor environment is rejected", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"version": resource.NewStringProperty("3"),
		})
		failures := check(m)
		require.Len(t, failures, 1)
		assert.Equal(t, "configEnvironment", failures[0].Property)
	})

	t.Run("project without environment is rejected", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project": resource.NewStringProperty("default"),
		})
		failures := check(m)
		require.Len(t, failures, 1)
	})

	// Computed values arrive at preview when the user wires
	// `configEnvironment.environment` (or any sibling field) to another
	// resource's Output<T>. We must defer validation in those cases instead
	// of failing the program before apply.
	computedString := resource.MakeComputed(resource.NewStringProperty(""))
	computedBool := resource.MakeComputed(resource.NewBoolProperty(false))

	t.Run("computed environment defers required check", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"environment": computedString,
		})
		assert.Empty(t, check(m), "computed environment should not fail Check")
	})

	t.Run("computed auto defers required check", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto": computedBool,
		})
		assert.Empty(t, check(m), "computed auto should not fail Check")
	})

	t.Run("computed environment alongside auto=true defers mutex check", func(t *testing.T) {
		// Even though auto=true is concrete, environment may resolve to "" so
		// we don't yet know if it's actually set. Defer.
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto":        resource.NewBoolProperty(true),
			"environment": computedString,
		})
		assert.Empty(t, check(m), "computed environment with concrete auto should defer")
	})

	t.Run("computed version alongside auto=true defers version check", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto":    resource.NewBoolProperty(true),
			"version": computedString,
		})
		assert.Empty(t, check(m), "computed version with concrete auto should defer")
	})

	t.Run("concrete auto=true with concrete env still fails mutex check", func(t *testing.T) {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto":        resource.NewBoolProperty(true),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		failures := check(m)
		require.Len(t, failures, 1)
	})
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
		oldS, err := plugin.MarshalProperties(olds, plugin.MarshalOptions{})
		require.NoError(t, err)
		newS, err := plugin.MarshalProperties(news, plugin.MarshalOptions{})
		require.NoError(t, err)
		resp, err := res.Diff(&pulumirpc.DiffRequest{OldInputs: oldS, News: newS})
		require.NoError(t, err)
		return resp
	}

	t.Run("no diff when identical", func(t *testing.T) {
		resp := diff(base(), base())
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
	})

	t.Run("toggle auto on forces replace", func(t *testing.T) {
		olds := base()
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto": resource.NewBoolProperty(true),
		})
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		require.Contains(t, resp.DetailedDiff, "configEnvironment")
		got := resp.DetailedDiff["configEnvironment"]
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})

	t.Run("toggle explicit-env on is in-place update", func(t *testing.T) {
		olds := base()
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		require.Contains(t, resp.DetailedDiff, "configEnvironment")
		got := resp.DetailedDiff["configEnvironment"]
		assert.False(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})

	t.Run("auto flip forces replace", func(t *testing.T) {
		olds := base()
		olds["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto": resource.NewBoolProperty(true),
		})
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got := resp.DetailedDiff["configEnvironment"]
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})

	t.Run("version-only change is update, not replace", func(t *testing.T) {
		olds := base()
		olds["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
			"version":     resource.NewStringProperty("2"),
		})
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
			"version":     resource.NewStringProperty("3"),
		})
		resp := diff(olds, news)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got := resp.DetailedDiff["configEnvironment"]
		assert.False(t, plugin.DiffKind(got.Kind).IsReplace()) //nolint:gosec
	})

	t.Run("computed identity field forces replace at preview", func(t *testing.T) {
		// At preview, fields wired to another resource's Output<T> arrive
		// computed. parseStackConfigEnvironment skips them, so without the
		// conservative override Diff would report "no change" or only an
		// in-place update — letting users approve a plan that misses the
		// destructive replace the apply will actually require.
		olds := base()
		olds["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.MakeComputed(resource.NewStringProperty("")),
		})
		oldS, err := plugin.MarshalProperties(olds, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		newS, err := plugin.MarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		resp, err := res.Diff(&pulumirpc.DiffRequest{OldInputs: oldS, News: newS})
		require.NoError(t, err)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got := resp.DetailedDiff["configEnvironment"]
		require.NotNil(t, got)
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace(), //nolint:gosec
			"computed environment must surface as replace at preview")
	})

	t.Run("computed auto forces replace at preview", func(t *testing.T) {
		olds := base()
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"auto": resource.MakeComputed(resource.NewBoolProperty(false)),
		})
		oldS, err := plugin.MarshalProperties(olds, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		newS, err := plugin.MarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		resp, err := res.Diff(&pulumirpc.DiffRequest{OldInputs: oldS, News: newS})
		require.NoError(t, err)
		require.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		got := resp.DetailedDiff["configEnvironment"]
		require.NotNil(t, got)
		assert.True(t, plugin.DiffKind(got.Kind).IsReplace(), //nolint:gosec
			"computed auto must surface as replace at preview")
	})

	t.Run("computed version alone is in-place update", func(t *testing.T) {
		// version isn't an identity field — only computed project/environment/auto
		// require replace. A computed version still routes through the normal
		// diff classifier as an in-place update.
		olds := base()
		olds["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		news := base()
		news["configEnvironment"] = resource.NewObjectProperty(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
			"version":     resource.MakeComputed(resource.NewStringProperty("")),
		})
		oldS, err := plugin.MarshalProperties(olds, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		newS, err := plugin.MarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true})
		require.NoError(t, err)
		resp, err := res.Diff(&pulumirpc.DiffRequest{OldInputs: oldS, News: newS})
		require.NoError(t, err)
		// The diff may or may not surface depending on how the unknown
		// shakes out vs absence; what we *don't* want is a replace.
		if got, ok := resp.DetailedDiff["configEnvironment"]; ok {
			assert.False(t, plugin.DiffKind(got.Kind).IsReplace(), //nolint:gosec
				"computed version alone must not force replace")
		}
	})

	t.Run("project name change forces replace as before", func(t *testing.T) {
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

func TestStackConfigEnvironmentFromServer(t *testing.T) {
	t.Parallel()
	stack := pulumiapi.StackIdentifier{ProjectName: "website", StackName: "production"}

	t.Run("auto preserved when prior was auto and ref matches", func(t *testing.T) {
		got := stackConfigEnvironmentFromServer("website", "production", "3", stack,
			&StackConfigEnvironment{Auto: true})
		assert.True(t, got.Auto)
		assert.Equal(t, "3", got.Version)
	})

	t.Run("explicit form when prior is empty", func(t *testing.T) {
		got := stackConfigEnvironmentFromServer("default", "prod-cfg", "", stack, nil)
		assert.False(t, got.Auto)
		assert.Equal(t, "default", got.Project)
		assert.Equal(t, "prod-cfg", got.Environment)
	})

	t.Run("auto demoted when server reports a different env", func(t *testing.T) {
		got := stackConfigEnvironmentFromServer("default", "prod-cfg", "", stack,
			&StackConfigEnvironment{Auto: true})
		assert.False(t, got.Auto)
		assert.Equal(t, "default", got.Project)
		assert.Equal(t, "prod-cfg", got.Environment)
	})
}

// stackUpdateCall is a single observed HTTP call against the fake API. Tests
// use it to assert that Update routes to SetStackConfig vs DeleteStackConfig
// with the expected env reference, and that early-return guards short-circuit
// before any call leaves the resource.
type stackUpdateCall struct {
	method string
	path   string
	body   string
}

// startStackUpdateServer spins up an httptest server that records every call
// against the stack /config endpoints and returns a Client wired to it. We
// build this here rather than reusing pulumiapi/testutils_test.go because
// those helpers are package-private to pulumiapi.
func startStackUpdateServer(t *testing.T) (*pulumiapi.Client, *[]stackUpdateCall) {
	t.Helper()
	calls := &[]stackUpdateCall{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*calls = append(*calls, stackUpdateCall{method: r.Method, path: r.URL.Path, body: string(body)})
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	client, err := pulumiapi.NewClient(&http.Client{Timeout: 5 * time.Second}, "token", server.URL)
	require.NoError(t, err)
	return client, calls
}

func TestStackResourceUpdate(t *testing.T) {
	t.Parallel()

	stackID := pulumiapi.StackIdentifier{
		OrgName: "acme", ProjectName: "website", StackName: "production",
	}
	const configPath = "/api/stacks/acme/website/production/config"

	base := func() resource.PropertyMap {
		return resource.PropertyMap{
			"organizationName": resource.NewStringProperty(stackID.OrgName),
			"projectName":      resource.NewStringProperty(stackID.ProjectName),
			"stackName":        resource.NewStringProperty(stackID.StackName),
		}
	}

	// withConfigEnv returns a deep-cloned base() with configEnvironment set to env.
	// Callers mutate the resulting PropertyMap freely without touching base().
	withConfigEnv := func(env resource.PropertyMap) resource.PropertyMap {
		m := base()
		m["configEnvironment"] = resource.NewObjectProperty(env)
		return m
	}

	update := func(
		t *testing.T, client *pulumiapi.Client, olds, news resource.PropertyMap,
	) (*pulumirpc.UpdateResponse, error) {
		t.Helper()
		oldS, err := plugin.MarshalProperties(olds, plugin.MarshalOptions{})
		require.NoError(t, err)
		newS, err := plugin.MarshalProperties(news, plugin.MarshalOptions{})
		require.NoError(t, err)
		res := &PulumiServiceStackResource{Client: client}
		return res.Update(&pulumirpc.UpdateRequest{Olds: oldS, News: newS})
	}

	t.Run("identity change is rejected before any HTTP call", func(t *testing.T) {
		t.Parallel()
		client, calls := startStackUpdateServer(t)
		olds := base()
		news := base()
		news["projectName"] = resource.NewStringProperty("other")
		_, err := update(t, client, olds, news)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected stack identity change")
		assert.Empty(t, *calls, "Update must not call the API on identity change")
	})

	t.Run("forceDestroy change is rejected before any HTTP call", func(t *testing.T) {
		t.Parallel()
		client, calls := startStackUpdateServer(t)
		olds := base()
		news := base()
		news["forceDestroy"] = resource.NewBoolProperty(true)
		_, err := update(t, client, olds, news)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected stack identity change")
		assert.Empty(t, *calls)
	})

	t.Run("auto toggle is rejected before any HTTP call", func(t *testing.T) {
		t.Parallel()
		client, calls := startStackUpdateServer(t)
		olds := withConfigEnv(resource.PropertyMap{
			"auto": resource.NewBoolProperty(true),
		})
		news := withConfigEnv(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		_, err := update(t, client, olds, news)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auto toggle")
		assert.Empty(t, *calls)
	})

	// Defense-in-depth: Check rejects auto+version on the way in, but state
	// migrations and imports can bypass Check. Update must refuse rather than
	// PUT a versioned auto ref the server-side stack-create API would reject.
	t.Run("auto + version is rejected before any HTTP call", func(t *testing.T) {
		t.Parallel()
		client, calls := startStackUpdateServer(t)
		olds := withConfigEnv(resource.PropertyMap{
			"auto": resource.NewBoolProperty(true),
		})
		news := withConfigEnv(resource.PropertyMap{
			"auto":    resource.NewBoolProperty(true),
			"version": resource.NewStringProperty("3"),
		})
		_, err := update(t, client, olds, news)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version is not allowed with configEnvironment.auto=true")
		assert.Empty(t, *calls)
	})

	t.Run("link explicit env calls SetStackConfig with the right ref", func(t *testing.T) {
		t.Parallel()
		client, calls := startStackUpdateServer(t)
		olds := base()
		news := withConfigEnv(resource.PropertyMap{
			"project":     resource.NewStringProperty("shared"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		_, err := update(t, client, olds, news)
		require.NoError(t, err)
		require.Len(t, *calls, 1)
		got := (*calls)[0]
		assert.Equal(t, http.MethodPut, got.method)
		assert.Equal(t, configPath, got.path)
		var sentCfg pulumiapi.StackConfig
		require.NoError(t, json.Unmarshal([]byte(got.body), &sentCfg))
		assert.Equal(t, "shared/prod-cfg", sentCfg.Environment)
	})

	t.Run("version pin change re-PUTs the new ref", func(t *testing.T) {
		t.Parallel()
		client, calls := startStackUpdateServer(t)
		olds := withConfigEnv(resource.PropertyMap{
			"project":     resource.NewStringProperty("shared"),
			"environment": resource.NewStringProperty("prod-cfg"),
			"version":     resource.NewStringProperty("2"),
		})
		news := withConfigEnv(resource.PropertyMap{
			"project":     resource.NewStringProperty("shared"),
			"environment": resource.NewStringProperty("prod-cfg"),
			"version":     resource.NewStringProperty("3"),
		})
		_, err := update(t, client, olds, news)
		require.NoError(t, err)
		require.Len(t, *calls, 1)
		var sentCfg pulumiapi.StackConfig
		require.NoError(t, json.Unmarshal([]byte((*calls)[0].body), &sentCfg))
		assert.Equal(t, "shared/prod-cfg@3", sentCfg.Environment)
	})

	t.Run("removing configEnvironment calls DeleteStackConfig", func(t *testing.T) {
		t.Parallel()
		client, calls := startStackUpdateServer(t)
		olds := withConfigEnv(resource.PropertyMap{
			"project":     resource.NewStringProperty("default"),
			"environment": resource.NewStringProperty("prod-cfg"),
		})
		news := base()
		_, err := update(t, client, olds, news)
		require.NoError(t, err)
		require.Len(t, *calls, 1)
		got := (*calls)[0]
		assert.Equal(t, http.MethodDelete, got.method)
		assert.Equal(t, configPath, got.path)
	})

}
