package resources

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/esc"
	"github.com/pulumi/esc/cmd/esc/cli/client"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/asset"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/urn"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	gcEnv = "env"
)

type getEnvironmentFunc func(
	ctx context.Context, orgName string, envName string, version string, decrypt bool,
) (yaml []byte, etag string, revision int, err error)

type getEnvironmentRevisionTagFunc func(
	ctx context.Context, orgName, envName, tagName string,
) (*client.EnvironmentRevisionTag, error)

type EscClientMock struct {
	client.Client
	getEnvironmentFunc            getEnvironmentFunc
	getEnvironmentRevisionTagFunc getEnvironmentRevisionTagFunc

	// Optional overrides used by the Environment tests.
	checkYAMLEnvironmentFunc          func(yaml []byte) ([]client.EnvironmentDiagnostic, error)
	updateEnvironmentWithRevisionFunc func(
		orgName, projectName, envName string, yaml []byte,
	) ([]client.EnvironmentDiagnostic, int, error)
	// calls records the mutating calls made, as "method org/project/env".
	calls []string
}

func (c *EscClientMock) GetEnvironment(
	ctx context.Context,
	orgName, _ /* projectName */, envName string,
	version string,
	decrypt bool,
) (yaml []byte, etag string, revision int, err error) {
	return c.getEnvironmentFunc(ctx, orgName, envName, version, decrypt)
}

func (c *EscClientMock) GetEnvironmentRevision(
	_ context.Context,
	_, _, _ string,
	_ int,
) (*client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentRevisionTag(
	ctx context.Context,
	orgName, _ /* projectName */, envName, tagName string,
) (*client.EnvironmentRevisionTag, error) {
	return c.getEnvironmentRevisionTagFunc(ctx, orgName, envName, tagName)
}

func (c *EscClientMock) GetRevisionNumber(
	_ context.Context,
	_, _, _, _ string,
) (int, error) {
	return 0, nil
}

func (c *EscClientMock) CheckYAMLEnvironment(
	_ context.Context,
	_ string,
	yaml []byte,
	_ ...client.CheckYAMLOption,
) (*esc.Environment, []client.EnvironmentDiagnostic, error) {
	if c.checkYAMLEnvironmentFunc != nil {
		diags, err := c.checkYAMLEnvironmentFunc(yaml)
		return nil, diags, err
	}
	return nil, nil, nil
}

func (c *EscClientMock) CreateEnvironment(context.Context, string, string) error {
	return nil
}

func (c *EscClientMock) CreateEnvironmentWithProject(_ context.Context, orgName, projectName, envName string) error {
	c.calls = append(c.calls, "create "+path.Join(orgName, projectName, envName))
	return nil
}

func (c *EscClientMock) CloneEnvironment(
	context.Context,
	string,
	string,
	string,
	client.CloneEnvironmentRequest,
) error {
	return nil
}

func (c *EscClientMock) DeleteEnvironment(_ context.Context, orgName, projectName, envName string) error {
	c.calls = append(c.calls, "delete "+path.Join(orgName, projectName, envName))
	return nil
}

func (c *EscClientMock) EnvironmentExists(context.Context, string, string, string) (bool, error) {
	return false, nil
}

func (c *EscClientMock) GetAnonymousOpenEnvironment(context.Context, string, string) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenEnvironment(context.Context, string, string, string) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenEnvironmentWithProject(
	context.Context,
	string,
	string,
	string,
	string,
) (*esc.Environment, error) {
	return nil, nil
}

func (c *EscClientMock) GetAnonymousOpenProperty(context.Context, string, string, string) (*esc.Value, error) {
	return nil, nil
}

func (c *EscClientMock) GetOpenProperty(context.Context, string, string, string, string, string) (*esc.Value, error) {
	return nil, nil
}

func (c *EscClientMock) GetPulumiAccountDetails(
	context.Context,
) (string, []string, *workspace.TokenInformation, error) {
	return "", nil, nil, nil
}

func (c *EscClientMock) ListEnvironments(context.Context, string) ([]client.OrgEnvironment, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) ListOrganizationEnvironments(
	context.Context,
	string,
	string,
) ([]client.OrgEnvironment, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) ListEnvironmentRevisions(
	_ context.Context,
	_, _, _ string,
	_ client.ListEnvironmentRevisionsOptions,
) ([]client.EnvironmentRevision, error) {
	return nil, nil
}

func (c *EscClientMock) ListEnvironmentRevisionTags(
	_ context.Context,
	_, _, _ string,
	_ client.ListEnvironmentRevisionTagsOptions,
) ([]client.EnvironmentRevisionTag, error) {
	return nil, nil
}

func (c *EscClientMock) OpenEnvironment(
	context.Context,
	string,
	string,
	string,
	string,
	time.Duration,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) OpenYAMLEnvironment(
	context.Context,
	string,
	[]byte,
	time.Duration,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) OpenEnvironmentDraft(
	context.Context,
	string,
	string,
	string,
	string,
	time.Duration,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) RotateEnvironment(
	context.Context,
	string,
	string,
	string,
	[]string,
) (*client.RotateEnvironmentResponse, []client.EnvironmentDiagnostic, error) {
	return nil, nil, nil
}

func (c *EscClientMock) SubmitChangeRequest(context.Context, string, string, *string) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironment(
	context.Context,
	string,
	string,
	[]byte,
	string,
) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentWithProject(
	context.Context,
	string,
	string,
	string,
	[]byte,
	string,
) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) UpdateEnvironmentDraft(
	context.Context,
	string,
	string,
	string,
	string,
	[]byte,
	string,
) ([]client.EnvironmentDiagnostic, error) {
	return nil, nil
}

func (c *EscClientMock) CreateEnvironmentDraft(
	context.Context,
	string,
	string,
	string,
	[]byte,
	string,
) (string, []client.EnvironmentDiagnostic, error) {
	return "", nil, nil
}

func (c *EscClientMock) GetDefaultOrg(context.Context) (string, error) {
	return "", nil
}

func (c *EscClientMock) GetEnvironmentDraft(context.Context, string, string, string, string) ([]byte, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) CreateEnvironmentTag(
	context.Context,
	string,
	string,
	string,
	string,
	string,
) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentTag(
	context.Context,
	string,
	string,
	string,
	string,
) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) ListEnvironmentTags(
	context.Context,
	string,
	string,
	string,
	client.ListEnvironmentTagsOptions,
) ([]*client.EnvironmentTag, string, error) {
	return nil, "", nil
}

func (c *EscClientMock) UpdateEnvironmentTag(
	context.Context,
	string,
	string,
	string,
	string,
	string,
	string,
	string,
) (*client.EnvironmentTag, error) {
	return nil, nil
}

func (c *EscClientMock) DeleteEnvironmentTag(context.Context, string, string, string, string) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironmentRevisionTag(
	_ context.Context,
	_, _, _, _ string,
	_ *int,
) error {
	return nil
}

func (c *EscClientMock) UpdateEnvironmentWithRevision(
	_ context.Context,
	orgName, projectName, envName string,
	yaml []byte,
	_ string,
) ([]client.EnvironmentDiagnostic, int, error) {
	c.calls = append(c.calls, "update "+path.Join(orgName, projectName, envName))
	if c.updateEnvironmentWithRevisionFunc != nil {
		return c.updateEnvironmentWithRevisionFunc(orgName, projectName, envName, yaml)
	}
	return nil, 0, nil
}

func (c *EscClientMock) CreateEnvironmentRevisionTag(
	_ context.Context,
	_, _, _, _ string,
	_ *int,
) error {
	return nil
}

func (c *EscClientMock) DeleteEnvironmentRevisionTag(
	_ context.Context,
	_, _, _, _ string,
) error {
	return nil
}

func (c *EscClientMock) RetractEnvironmentRevision(
	_ context.Context,
	_, _, _ string,
	_ string,
	_ *int,
	_ string,
) error {
	return nil
}

func (c *EscClientMock) CreateEnvironmentOpenRequest(
	context.Context,
	string,
	string,
	string,
	int,
	int,
) (*client.CreateEnvironmentOpenRequestResponse, error) {
	return nil, nil
}

func (c *EscClientMock) GetEnvironmentSettings(
	context.Context,
	string,
	string,
	string,
) (*client.EnvironmentSettings, error) {
	return nil, nil
}

func (c *EscClientMock) PatchEnvironmentSettings(
	context.Context,
	string,
	string,
	string,
	client.PatchEnvironmentSettingsRequest,
) error {
	return nil
}

func (c *EscClientMock) Insecure() bool {
	return false
}

func (c *EscClientMock) URL() string {
	return ""
}

func buildEscClientMock(
	getEnvironmentFunc getEnvironmentFunc,
	getEnvironmentRevisionTagFunc getEnvironmentRevisionTagFunc,
) *EscClientMock {
	return &EscClientMock{
		getEnvironmentFunc:            getEnvironmentFunc,
		getEnvironmentRevisionTagFunc: getEnvironmentRevisionTagFunc,
	}
}

type envMetadataClientMock struct {
	config.Client
	id  string
	err error
}

func (m *envMetadataClientMock) GetEnvironmentMetadata(
	context.Context, string, string, string,
) (*pulumiapi.EnvironmentMetadata, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pulumiapi.EnvironmentMetadata{ID: m.id}, nil
}

const (
	gcEnvProj   = "proj"
	gcEnvYaml   = "values:\n  foo: bar"
	gcEnvID     = "4d6a0bd2-7a5e-4bd4-9d61-0b5c7f1d8e2a"
	gcEnvProjID = gcOrg + "/" + gcEnvProj + "/" + gcEnv
)

var envURN = urn.New("test", gcEnvProj, "", environmentToken, "e")

func newEnvironmentServer(t *testing.T, escClient *EscClientMock) integration.Server {
	t.Helper()
	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&Environment{})).
		Build()
	require.NoError(t, err)
	ctx := config.WithMockEscClient(t.Context(), escClient)
	ctx = config.WithMockClient(ctx, &envMetadataClientMock{id: gcEnvID})
	server, err := integration.NewServer(ctx, "pulumiservice", semver.MustParse("0.0.0"),
		integration.WithProvider(WithEnvironmentCompat(prov)))
	require.NoError(t, err)
	return server
}

func newEnvEscClientMock(yaml string, revision int, getErr error) *EscClientMock {
	return buildEscClientMock(
		func(context.Context, string, string, string, bool) ([]byte, string, int, error) {
			if getErr != nil {
				return nil, "", 0, getErr
			}
			return []byte(yaml), "", revision, nil
		},
		func(context.Context, string, string, string) (*client.EnvironmentRevisionTag, error) {
			return nil, nil
		},
	)
}

func textAsset(t *testing.T, text string) *asset.Asset {
	t.Helper()
	a, err := asset.FromText(text)
	require.NoError(t, err)
	return a
}

// envInputs builds Environment inputs as the SDKs send them.
func envInputs(project string, yaml property.Value) property.Map {
	m := map[string]property.Value{
		gcOrganization: property.New(gcOrg),
		gcName:         property.New(gcEnv),
		gcYaml:         yaml,
	}
	if project != "" {
		m[gcProject] = property.New(project)
	}
	return property.NewMap(m)
}

// legacyEnvState builds state as the pre-infer resource wrote it: `yaml` is a
// secret string, and state from before v0.25.0 has no `project`.
func legacyEnvState(project, yaml string) property.Map {
	m := map[string]property.Value{
		gcOrganization: property.New(gcOrg),
		gcName:         property.New(gcEnv),
		gcYaml:         property.New(yaml).WithSecret(true),
		"revision":     property.New(3.0),
	}
	if project != "" {
		m[gcProject] = property.New(project)
	}
	return property.NewMap(m)
}

// legacyEnvInputs builds old inputs as the pre-infer resource's Check wrote them.
func legacyEnvInputs(project, yaml string) property.Map {
	return legacyEnvState(project, yaml).Delete("revision")
}

func checkEnv(t *testing.T, server integration.Server, inputs property.Map) p.CheckResponse {
	t.Helper()
	resp, err := server.Check(p.CheckRequest{Urn: envURN, Inputs: inputs})
	require.NoError(t, err)
	return resp
}

func TestEnvironmentCheck(t *testing.T) {
	t.Parallel()
	server := newEnvironmentServer(t, newEnvEscClientMock("", 0, nil))

	t.Run("text asset is trimmed, made secret and project defaulted", func(t *testing.T) {
		resp := checkEnv(t, server, envInputs("", property.New(textAsset(t, "\n"+gcEnvYaml+"\n\n"))))
		require.Empty(t, resp.Failures)
		assert.Equal(t, property.New(defaultProject), resp.Inputs.Get(gcProject))
		y := resp.Inputs.Get(gcYaml)
		assert.True(t, y.Secret())
		require.True(t, y.IsAsset())
		assert.Equal(t, gcEnvYaml, y.AsAsset().Text)
	})

	t.Run("file asset is converted to a text asset", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "env.yaml")
		require.NoError(t, os.WriteFile(file, []byte(gcEnvYaml+"\n"), 0o600))
		fileAsset, err := asset.FromPath(file)
		require.NoError(t, err)

		resp := checkEnv(t, server, envInputs(gcEnvProj, property.New(fileAsset)))
		require.Empty(t, resp.Failures)
		assert.Equal(t, property.New(textAsset(t, gcEnvYaml)).WithSecret(true), resp.Inputs.Get(gcYaml))
	})

	t.Run("string yaml is accepted", func(t *testing.T) {
		resp := checkEnv(t, server, envInputs(gcEnvProj, property.New(gcEnvYaml+"\n")))
		require.Empty(t, resp.Failures)
		assert.Equal(t, property.New(textAsset(t, gcEnvYaml)).WithSecret(true), resp.Inputs.Get(gcYaml))
	})

	t.Run("computed yaml", func(t *testing.T) {
		resp := checkEnv(t, server, envInputs(gcEnvProj, property.New(property.Computed)))
		require.Empty(t, resp.Failures)
		assert.True(t, resp.Inputs.Get(gcYaml).IsComputed())
	})

	t.Run("secret wrapping computed yaml", func(t *testing.T) {
		// Regression test for #606: a secret wrapping a computed value (from
		// Output.ApplyT) panicked in Check.
		resp := checkEnv(t, server, envInputs(gcEnvProj, property.New(property.Computed).WithSecret(true)))
		require.Empty(t, resp.Failures)
		assert.True(t, resp.Inputs.Get(gcYaml).IsComputed())
	})

	t.Run("identifiers containing / are rejected", func(t *testing.T) {
		inputs := envInputs("a/b", property.New(textAsset(t, gcEnvYaml))).
			Set(gcName, property.New("c/d"))
		resp := checkEnv(t, server, inputs)
		var props []string
		for _, f := range resp.Failures {
			props = append(props, f.Property)
		}
		assert.ElementsMatch(t, []string{gcProject, gcName}, props)
	})

	t.Run("missing required properties", func(t *testing.T) {
		resp := checkEnv(t, server, property.NewMap(map[string]property.Value{
			gcOrganization: property.New(gcOrg),
		}))
		var props []string
		for _, f := range resp.Failures {
			props = append(props, f.Property)
		}
		assert.ElementsMatch(t, []string{gcName, gcYaml}, props)
	})
}

func TestEnvironmentDiff(t *testing.T) {
	t.Parallel()
	server := newEnvironmentServer(t, newEnvEscClientMock("", 0, nil))

	diff := func(t *testing.T, state, news property.Map) p.DiffResponse {
		t.Helper()
		checked := checkEnv(t, server, news)
		require.Empty(t, checked.Failures)
		resp, err := server.Diff(p.DiffRequest{
			ID:     gcEnvProjID,
			Urn:    envURN,
			State:  state,
			Inputs: checked.Inputs,
		})
		require.NoError(t, err)
		return resp
	}
	newYaml := property.New(textAsset(t, gcEnvYaml+"\n"))

	t.Run("no diff against legacy state with string yaml", func(t *testing.T) {
		resp := diff(t, legacyEnvState(gcEnvProj, gcEnvYaml), envInputs(gcEnvProj, newYaml))
		assert.False(t, resp.HasChanges)
		assert.Empty(t, resp.DetailedDiff)
	})

	t.Run("no spurious replace when upgrading from pre-0.25.0 state without project", func(t *testing.T) {
		resp := diff(t, legacyEnvState("", gcEnvYaml), envInputs("", newYaml))
		assert.False(t, resp.HasChanges)
		assert.Empty(t, resp.DetailedDiff)
	})

	t.Run("yaml change updates in place", func(t *testing.T) {
		resp := diff(t, legacyEnvState(gcEnvProj, gcEnvYaml),
			envInputs(gcEnvProj, property.New(textAsset(t, "values:\n  foo: baz"))))
		assert.True(t, resp.HasChanges)
		require.Len(t, resp.DetailedDiff, 1)
		assert.Equal(t, p.Update, resp.DetailedDiff[gcYaml].Kind)
		assert.False(t, resp.DeleteBeforeReplace)
	})

	for _, field := range []string{gcOrganization, gcProject, gcName} {
		t.Run(field+" change replaces", func(t *testing.T) {
			resp := diff(t, legacyEnvState(gcEnvProj, gcEnvYaml),
				envInputs(gcEnvProj, newYaml).Set(field, property.New("other")))
			assert.True(t, resp.HasChanges)
			require.Contains(t, resp.DetailedDiff, field)
			assert.Equal(t, p.UpdateReplace, resp.DetailedDiff[field].Kind)
			assert.False(t, resp.DeleteBeforeReplace)
		})

		t.Run(field+" case-only change is ignored", func(t *testing.T) {
			// ESC looks environments up case-insensitively, so this names the
			// same environment.
			state := legacyEnvState(gcEnvProj, gcEnvYaml)
			old, ok := state.GetOk(field)
			require.True(t, ok)
			resp := diff(t, state,
				envInputs(gcEnvProj, newYaml).Set(field, property.New(strings.ToUpper(old.AsString()))))
			assert.False(t, resp.HasChanges)
			assert.Empty(t, resp.DetailedDiff)
		})
	}
}

// TestEnvironmentDiffOldInputs covers the path the engine takes in practice: it
// sends old inputs, which the default Diff compares instead of state. The
// integration server doesn't advertise old-input support, so this test drives
// the gRPC server.
func TestEnvironmentDiffOldInputs(t *testing.T) {
	t.Parallel()
	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&Environment{})).
		Build()
	require.NoError(t, err)
	server, err := p.RawServer("pulumiservice", "0.0.0", WithEnvironmentCompat(prov))(nil)
	require.NoError(t, err)
	_, err = server.Configure(t.Context(), &pulumirpc.ConfigureRequest{
		AcceptSecrets:          true,
		SendsOldInputs:         true,
		SendsOldInputsToDelete: true,
	})
	require.NoError(t, err)

	marshal := func(m resource.PropertyMap) *structpb.Struct {
		s, err := plugin.MarshalProperties(m, plugin.MarshalOptions{KeepSecrets: true, KeepUnknowns: true})
		require.NoError(t, err)
		return s
	}
	legacy := func(project string) resource.PropertyMap {
		m := resource.PropertyMap{
			gcOrganization: resource.NewProperty(gcOrg),
			gcName:         resource.NewProperty(gcEnv),
			gcYaml:         resource.MakeSecret(resource.NewProperty(gcEnvYaml)),
		}
		if project != "" {
			m[gcProject] = resource.NewProperty(project)
		}
		return m
	}
	diff := func(t *testing.T, oldInputs resource.PropertyMap, project, yaml string) *pulumirpc.DiffResponse {
		t.Helper()
		state := oldInputs.Copy()
		state["revision"] = resource.NewProperty(3.0)
		news := legacy(project)
		news[gcYaml] = resource.NewProperty(textAsset(t, yaml))
		checked, err := server.Check(t.Context(), &pulumirpc.CheckRequest{
			Urn:  string(envURN),
			News: marshal(news),
		})
		require.NoError(t, err)
		require.Empty(t, checked.GetFailures())
		resp, err := server.Diff(t.Context(), &pulumirpc.DiffRequest{
			Id:        gcEnvProjID,
			Urn:       string(envURN),
			Olds:      marshal(state),
			OldInputs: marshal(oldInputs),
			News:      checked.GetInputs(),
		})
		require.NoError(t, err)
		return resp
	}

	t.Run("no diff against legacy old inputs with string yaml", func(t *testing.T) {
		resp := diff(t, legacy(gcEnvProj), gcEnvProj, gcEnvYaml+"\n")
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.GetChanges())
		assert.Empty(t, resp.GetDetailedDiff())
	})

	t.Run("no spurious replace for pre-0.25.0 old inputs without project", func(t *testing.T) {
		resp := diff(t, legacy(""), defaultProject, gcEnvYaml)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.GetChanges())
		assert.Empty(t, resp.GetReplaces())
	})

	t.Run("no diff on refresh against legacy inputs", func(t *testing.T) {
		// After Read, the engine diffs the refreshed state against the old
		// inputs from state, which never passed through Check.
		read := legacy(gcEnvProj)
		read[gcYaml] = resource.MakeSecret(resource.NewProperty(textAsset(t, gcEnvYaml)))
		state := read.Copy()
		state["revision"] = resource.NewProperty(3.0)
		resp, err := server.Diff(t.Context(), &pulumirpc.DiffRequest{
			Id:        gcEnvProjID,
			Urn:       string(envURN),
			Olds:      marshal(state),
			OldInputs: marshal(read),
			News:      marshal(legacy(gcEnvProj)),
		})
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.GetChanges())
		assert.Empty(t, resp.GetDetailedDiff())
	})

	t.Run("case-only name change is ignored", func(t *testing.T) {
		olds := legacy(gcEnvProj)
		news := legacy(gcEnvProj)
		news[gcName] = resource.NewProperty(strings.ToUpper(gcEnv))
		news[gcYaml] = resource.NewProperty(textAsset(t, gcEnvYaml))
		checked, err := server.Check(t.Context(), &pulumirpc.CheckRequest{
			Urn:  string(envURN),
			News: marshal(news),
		})
		require.NoError(t, err)
		state := olds.Copy()
		state["revision"] = resource.NewProperty(3.0)
		resp, err := server.Diff(t.Context(), &pulumirpc.DiffRequest{
			Id:        gcEnvProjID,
			Urn:       string(envURN),
			Olds:      marshal(state),
			OldInputs: marshal(olds),
			News:      checked.GetInputs(),
		})
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.GetChanges())
		assert.Empty(t, resp.GetReplaces())
	})

	t.Run("yaml change is an update", func(t *testing.T) {
		resp := diff(t, legacy(gcEnvProj), gcEnvProj, "values:\n  foo: baz")
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.GetChanges())
		assert.Empty(t, resp.GetReplaces())
		assert.Contains(t, resp.GetDetailedDiff(), gcYaml)
	})
}

func TestEnvironmentCreate(t *testing.T) {
	t.Parallel()

	create := func(t *testing.T, escClient *EscClientMock, dryRun bool) (p.CreateResponse, error) {
		server := newEnvironmentServer(t, escClient)
		checked := checkEnv(t, server, envInputs("", property.New(textAsset(t, gcEnvYaml))))
		return server.Create(p.CreateRequest{Urn: envURN, Properties: checked.Inputs, DryRun: dryRun})
	}

	t.Run("creates the environment and records revision and id", func(t *testing.T) {
		escClient := newEnvEscClientMock("", 0, nil)
		var pushed string
		escClient.updateEnvironmentWithRevisionFunc = func(
			_, _, _ string, yaml []byte,
		) ([]client.EnvironmentDiagnostic, int, error) {
			pushed = string(yaml)
			return nil, 2, nil
		}
		resp, err := create(t, escClient, false)
		require.NoError(t, err)
		assert.Equal(t, gcOrg+"/default/"+gcEnv, resp.ID)
		assert.Equal(t, []string{"create org/default/env", "update org/default/env"}, escClient.calls)
		assert.Equal(t, gcEnvYaml, pushed)
		assert.Equal(t, property.New(2.0), resp.Properties.Get("revision"))
		assert.Equal(t, property.New(gcEnvID), resp.Properties.Get("environmentId"))
		assert.Equal(t, property.New(defaultProject), resp.Properties.Get(gcProject))
		// Programs read the output as a string, as from the pre-infer resource.
		assert.Equal(t, property.New(gcEnvYaml).WithSecret(true), resp.Properties.Get(gcYaml))
	})

	t.Run("preview makes no calls", func(t *testing.T) {
		escClient := newEnvEscClientMock("", 0, nil)
		_, err := create(t, escClient, true)
		require.NoError(t, err)
		assert.Empty(t, escClient.calls)
	})

	t.Run("invalid yaml fails before creating the environment", func(t *testing.T) {
		escClient := newEnvEscClientMock("", 0, nil)
		escClient.checkYAMLEnvironmentFunc = func([]byte) ([]client.EnvironmentDiagnostic, error) {
			return []client.EnvironmentDiagnostic{{Summary: "bad yaml"}}, nil
		}
		_, err := create(t, escClient, false)
		assert.ErrorContains(t, err, "bad yaml")
		assert.Empty(t, escClient.calls)
	})
}

func TestEnvironmentUpdate(t *testing.T) {
	t.Parallel()

	update := func(t *testing.T, escClient *EscClientMock) (p.UpdateResponse, error) {
		server := newEnvironmentServer(t, escClient)
		checked := checkEnv(t, server, envInputs(gcEnvProj, property.New(textAsset(t, "values:\n  foo: baz"))))
		return server.Update(p.UpdateRequest{
			ID:     gcEnvProjID,
			Urn:    envURN,
			State:  legacyEnvState(gcEnvProj, gcEnvYaml),
			Inputs: checked.Inputs,
		})
	}

	t.Run("updates legacy state", func(t *testing.T) {
		escClient := newEnvEscClientMock("", 0, nil)
		escClient.updateEnvironmentWithRevisionFunc = func(
			string, string, string, []byte,
		) ([]client.EnvironmentDiagnostic, int, error) {
			return nil, 4, nil
		}
		resp, err := update(t, escClient)
		require.NoError(t, err)
		assert.Equal(t, []string{"update " + gcEnvProjID}, escClient.calls)
		assert.Equal(t, property.New(4.0), resp.Properties.Get("revision"))
		assert.Equal(t, property.New("values:\n  foo: baz").WithSecret(true), resp.Properties.Get(gcYaml))
		assert.Equal(t, property.New(gcEnvID), resp.Properties.Get("environmentId"))
	})

	t.Run("API errors are returned", func(t *testing.T) {
		escClient := newEnvEscClientMock("", 0, nil)
		escClient.updateEnvironmentWithRevisionFunc = func(
			string, string, string, []byte,
		) ([]client.EnvironmentDiagnostic, int, error) {
			return nil, 0, errors.New("boom")
		}
		_, err := update(t, escClient)
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("yaml diagnostics are returned", func(t *testing.T) {
		escClient := newEnvEscClientMock("", 0, nil)
		escClient.updateEnvironmentWithRevisionFunc = func(
			string, string, string, []byte,
		) ([]client.EnvironmentDiagnostic, int, error) {
			return []client.EnvironmentDiagnostic{{Summary: "bad yaml"}}, 0, nil
		}
		_, err := update(t, escClient)
		assert.ErrorContains(t, err, "bad yaml")
	})
}

func TestEnvironmentDelete(t *testing.T) {
	t.Parallel()

	for name, state := range map[string]property.Map{
		"legacy state":                     legacyEnvState(gcEnvProj, gcEnvYaml),
		"pre-0.25.0 state without project": legacyEnvState("", gcEnvYaml),
	} {
		t.Run(name, func(t *testing.T) {
			escClient := newEnvEscClientMock("", 0, nil)
			server := newEnvironmentServer(t, escClient)
			err := server.Delete(p.DeleteRequest{ID: gcEnvProjID, Urn: envURN, Properties: state})
			require.NoError(t, err)
			project := state.Get(gcProject)
			if project.IsNull() {
				project = property.New(defaultProject)
			}
			assert.Equal(t, []string{"delete " + path.Join(gcOrg, project.AsString(), gcEnv)}, escClient.calls)
		})
	}
}

func TestEnvironmentRead(t *testing.T) {
	t.Parallel()

	read := func(t *testing.T, escClient *EscClientMock, id string) (p.ReadResponse, error) {
		server := newEnvironmentServer(t, escClient)
		return server.Read(p.ReadRequest{
			ID:         id,
			Urn:        envURN,
			Properties: legacyEnvState(gcEnvProj, gcEnvYaml),
			Inputs:     legacyEnvInputs(gcEnvProj, gcEnvYaml),
		})
	}

	t.Run("not found", func(t *testing.T) {
		resp, err := read(t, newEnvEscClientMock("", 0, &apitype.ErrorResponse{Code: http.StatusNotFound}), gcEnvProjID)
		require.NoError(t, err)
		// An empty ID tells the engine the resource is gone.
		assert.Empty(t, resp.ID)
	})

	t.Run("other errors are returned", func(t *testing.T) {
		_, err := read(t, newEnvEscClientMock("", 0, errors.New("unauthorized")), gcEnvProjID)
		assert.ErrorContains(t, err, "unauthorized")
	})

	t.Run("found, with legacy state and inputs", func(t *testing.T) {
		resp, err := read(t, newEnvEscClientMock(gcEnvYaml+"\n", 5, nil), gcEnvProjID)
		require.NoError(t, err)
		assert.Equal(t, gcEnvProjID, resp.ID)
		assert.Equal(t, property.New(textAsset(t, gcEnvYaml)).WithSecret(true), resp.Inputs.Get(gcYaml))
		// The output stays a string, as programs read it.
		assert.Equal(t, property.New(gcEnvYaml).WithSecret(true), resp.Properties.Get(gcYaml))
		assert.Equal(t, property.New(gcEnvProj), resp.Inputs.Get(gcProject))
		assert.Equal(t, property.New(5.0), resp.Properties.Get("revision"))
		assert.Equal(t, property.New(gcEnvID), resp.Properties.Get("environmentId"))
	})

	t.Run("changed yaml", func(t *testing.T) {
		resp, err := read(t, newEnvEscClientMock("values:\n  foo: baz", 5, nil), gcEnvProjID)
		require.NoError(t, err)
		assert.Equal(t, property.New(textAsset(t, "values:\n  foo: baz")).WithSecret(true), resp.Inputs.Get(gcYaml))
		assert.Equal(t, property.New("values:\n  foo: baz").WithSecret(true), resp.Properties.Get(gcYaml))
	})

	t.Run("asset state reads back as a string output", func(t *testing.T) {
		server := newEnvironmentServer(t, newEnvEscClientMock(gcEnvYaml, 5, nil))
		yaml := property.New(textAsset(t, gcEnvYaml)).WithSecret(true)
		resp, err := server.Read(p.ReadRequest{
			ID:         gcEnvProjID,
			Urn:        envURN,
			Properties: legacyEnvState(gcEnvProj, gcEnvYaml).Set(gcYaml, yaml),
			Inputs:     legacyEnvInputs(gcEnvProj, gcEnvYaml).Set(gcYaml, yaml),
		})
		require.NoError(t, err)
		assert.Equal(t, yaml, resp.Inputs.Get(gcYaml))
		assert.Equal(t, property.New(gcEnvYaml).WithSecret(true), resp.Properties.Get(gcYaml))
	})

	t.Run("legacy org/env id keeps its id and uses the default project", func(t *testing.T) {
		resp, err := read(t, newEnvEscClientMock(gcEnvYaml, 1, nil), gcOrg+"/"+gcEnv)
		require.NoError(t, err)
		assert.Equal(t, gcOrg+"/"+gcEnv, resp.ID)
		assert.Equal(t, property.New(defaultProject), resp.Properties.Get(gcProject))
	})

	t.Run("import", func(t *testing.T) {
		server := newEnvironmentServer(t, newEnvEscClientMock(gcEnvYaml, 1, nil))
		resp, err := server.Read(p.ReadRequest{ID: gcEnvProjID, Urn: envURN})
		require.NoError(t, err)
		assert.Equal(t, gcEnvProjID, resp.ID)
		// infer doesn't apply the secret tag on a Read without prior state.
		assert.True(t, resp.Inputs.Get(gcYaml).Secret())
		assert.True(t, resp.Properties.Get(gcYaml).Secret())
		assert.Equal(t, property.New(gcOrg), resp.Inputs.Get(gcOrganization))
		assert.Equal(t, property.New(gcEnvProj), resp.Inputs.Get(gcProject))
		assert.Equal(t, property.New(gcEnv), resp.Inputs.Get(gcName))
	})

	t.Run("invalid id", func(t *testing.T) {
		_, err := read(t, newEnvEscClientMock(gcEnvYaml, 1, nil), "just-one")
		assert.Error(t, err)
	})
}

func TestNormalizeLegacyEnvironment(t *testing.T) {
	t.Parallel()

	t.Run("empty map is unchanged", func(t *testing.T) {
		assert.Equal(t, property.Map{}, normalizeLegacyEnvironment(property.Map{}))
	})

	t.Run("current shape is unchanged", func(t *testing.T) {
		m := envInputs(gcEnvProj, property.New(textAsset(t, gcEnvYaml)).WithSecret(true))
		assert.Equal(t, m, normalizeLegacyEnvironment(m))
	})

	t.Run("legacy shape", func(t *testing.T) {
		got := normalizeLegacyEnvironment(legacyEnvInputs("", gcEnvYaml+"\n"))
		assert.Equal(t, envInputs(defaultProject, property.New(textAsset(t, gcEnvYaml)).WithSecret(true)), got)
	})
}

func TestEnvironmentResourceID(t *testing.T) {
	t.Parallel()
	assert.Equal(t, gcEnvProjID, environmentResourceID(gcOrg, gcEnvProj, gcEnv))
}

func TestSplitEnvironmentID(t *testing.T) {
	t.Parallel()
	for id, want := range map[string][]string{
		gcEnvProjID:         {gcOrg, gcEnvProj, gcEnv},
		gcOrg + "/" + gcEnv: {gcOrg, defaultProject, gcEnv},
	} {
		org, proj, env, err := splitEnvironmentID(id)
		require.NoError(t, err)
		assert.Equal(t, want, []string{org, proj, env})
	}
	for _, id := range []string{"just-one", "a/b/c/d"} {
		_, _, _, err := splitEnvironmentID(id)
		assert.Error(t, err, id)
	}
}

func TestProjectOrDefault(t *testing.T) {
	t.Parallel()
	assert.Equal(t, defaultProject, projectOrDefault(""))
	assert.Equal(t, "my-project", projectOrDefault("my-project"))
}
