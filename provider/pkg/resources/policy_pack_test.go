// Copyright 2016-2026, Pulumi Corporation.
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
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type PolicyPackClientMock struct {
	config.Client
	publishFunc func(
		ctx context.Context, org string, req pulumiapi.CreatePolicyPackRequest, archive io.Reader,
	) (int, error)
	deleteVersionFunc func(ctx context.Context, org, name, versionTag string) error
	listFunc          func(ctx context.Context, org string) ([]pulumiapi.PolicyPackWithVersions, error)
}

func (c *PolicyPackClientMock) PublishPolicyPack(
	ctx context.Context, org string, req pulumiapi.CreatePolicyPackRequest, archive io.Reader,
) (int, error) {
	if c.publishFunc == nil {
		return 0, nil
	}
	return c.publishFunc(ctx, org, req, archive)
}

func (c *PolicyPackClientMock) DeletePolicyPackVersion(ctx context.Context, org, name, versionTag string) error {
	if c.deleteVersionFunc == nil {
		return nil
	}
	return c.deleteVersionFunc(ctx, org, name, versionTag)
}

func (c *PolicyPackClientMock) ListPolicyPacks(
	ctx context.Context, org string,
) ([]pulumiapi.PolicyPackWithVersions, error) {
	if c.listFunc == nil {
		return nil, nil
	}
	return c.listFunc(ctx, org)
}

func writePolicySource(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Non-nodejs runtime so packagePolicyPackArchive uses archive.TGZ (no npm shell-out).
	require.NoError(t, os.WriteFile(filepath.Join(dir, "PulumiPolicy.yaml"), []byte("runtime: python\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "__main__.py"), []byte("# policy\n"), 0o600))
	return dir
}

func TestPolicyPackID_RoundTrip(t *testing.T) {
	id := policyPackID("acme", "guard", "1.2.3")
	assert.Equal(t, "acme/guard/1.2.3", id)

	org, name, tag, err := splitPolicyPackID(id)
	require.NoError(t, err)
	assert.Equal(t, "acme", org)
	assert.Equal(t, "guard", name)
	assert.Equal(t, "1.2.3", tag)
}

func TestSplitPolicyPackID_InvalidShape(t *testing.T) {
	for _, id := range []string{"", "only-one", "two/parts", "four/parts/here/extra"} {
		_, _, _, err := splitPolicyPackID(id)
		assert.Errorf(t, err, "expected error for id %q", id)
	}
}

func TestNormalizeConfigSchema(t *testing.T) {
	// nil and empty pass through untouched
	assert.Nil(t, normalizeConfigSchema(nil))
	empty := map[string]any{}
	assert.Equal(t, empty, normalizeConfigSchema(empty))

	// existing "type" is preserved
	withType := map[string]any{"type": "string"}
	got := normalizeConfigSchema(withType)
	assert.Equal(t, "string", got["type"])

	// missing "type" gets defaulted to object without mutating the input
	in := map[string]any{"properties": map[string]any{"x": map[string]any{"type": "number"}}}
	out := normalizeConfigSchema(in)
	assert.Equal(t, "object", out["type"])
	_, hadType := in["type"]
	assert.False(t, hadType, "input must not be mutated")
}

func TestToAPIPolicies(t *testing.T) {
	in := []PolicyPackPolicyInput{
		{
			Name:             "no-secrets",
			DisplayName:      "No Secrets",
			Description:      "block secret literals",
			EnforcementLevel: "mandatory",
			Message:          "remove the secret",
			ConfigSchema:     map[string]any{"type": "object", "required": []string{"k"}},
			Severity:         "high",
			Framework: &PolicyPackComplianceFrameworkInput{
				Name:    "PCI-DSS",
				Version: "4.0",
			},
			Tags:             []string{"secrets", "security"},
			RemediationSteps: "remove secret literals",
			URL:              "https://example.com/policies/no-secrets",
		},
		{Name: "minimal"},
	}
	got := toAPIPolicies(in)
	require.Len(t, got, 2)
	first := got[0]
	assert.Equal(t, "no-secrets", first.Name)
	assert.Equal(t, "No Secrets", first.DisplayName)
	assert.Equal(t, apitype.EnforcementLevel("mandatory"), first.EnforcementLevel)
	assert.Equal(t, apitype.PolicySeverity("high"), first.Severity)
	require.NotNil(t, first.ConfigSchema)
	assert.Equal(t, apitype.Object, first.ConfigSchema.Type)
	assert.Equal(t, []string{"k"}, first.ConfigSchema.Required)
	require.NotNil(t, first.Framework)
	assert.Equal(t, "PCI-DSS", first.Framework.Name)
	assert.Equal(t, []string{"secrets", "security"}, first.Tags)
	assert.Equal(t, "remove secret literals", first.RemediationSteps)
	assert.Equal(t, "https://example.com/policies/no-secrets", first.URL)
	assert.Equal(t, "minimal", got[1].Name)
}

func TestPoliciesNormalizedDeepEqual(t *testing.T) {
	a := []PolicyPackPolicyInput{{Name: "a", EnforcementLevel: "advisory"}}
	b := []PolicyPackPolicyInput{{Name: "a", EnforcementLevel: "advisory"}}
	assert.True(t, reflect.DeepEqual(a, b))

	b[0].EnforcementLevel = "mandatory"
	assert.False(t, reflect.DeepEqual(a, b))
}

func TestConvertAnalyzerConfigSchema(t *testing.T) {
	assert.Nil(t, convertAnalyzerConfigSchema(nil))

	got := convertAnalyzerConfigSchema(&plugin.AnalyzerPolicyConfigSchema{
		Properties: map[string]plugin.JSONSchema{
			"threshold": {"type": "number"},
		},
		Required: []string{"threshold"},
	})
	assert.Equal(t, "object", got["type"])
	assert.Equal(t, []string{"threshold"}, got["required"])

	props, ok := got["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, props, "threshold")

	// Empty schema still gets a type but no properties/required keys.
	empty := convertAnalyzerConfigSchema(&plugin.AnalyzerPolicyConfigSchema{})
	assert.Equal(t, map[string]any{"type": "object"}, empty)
}

func TestPolicyPack_Create_DryRun(t *testing.T) {
	dir := writePolicySource(t)
	mock := &PolicyPackClientMock{
		publishFunc: func(context.Context, string, pulumiapi.CreatePolicyPackRequest, io.Reader) (int, error) {
			t.Fatalf("publish should not be called during DryRun")
			return 0, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	resp, err := (&PolicyPack{}).Create(ctx, infer.CreateRequest[PolicyPackInput]{
		DryRun: true,
		Inputs: PolicyPackInput{
			Organization: "acme",
			Name:         "guard",
			VersionTag:   "1.0.0",
			SourcePath:   dir,
			Policies:     []PolicyPackPolicyInput{{Name: "rule"}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "acme", resp.Output.Organization)
	assert.Empty(t, resp.ID, "DryRun should not assign an ID")
}

func TestPolicyPack_Create_HappyPath(t *testing.T) {
	dir := writePolicySource(t)
	var capturedReq pulumiapi.CreatePolicyPackRequest
	mock := &PolicyPackClientMock{
		publishFunc: func(
			_ context.Context, org string, req pulumiapi.CreatePolicyPackRequest, archive io.Reader,
		) (int, error) {
			assert.Equal(t, "acme", org)
			capturedReq = req
			body, _ := io.ReadAll(archive)
			assert.NotEmpty(t, body, "archive should be non-empty")
			return 5, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	resp, err := (&PolicyPack{}).Create(ctx, infer.CreateRequest[PolicyPackInput]{
		Inputs: PolicyPackInput{
			Organization: "acme",
			Name:         "guard",
			DisplayName:  "Guard",
			VersionTag:   "1.0.0",
			SourcePath:   dir,
			Policies: []PolicyPackPolicyInput{
				{Name: "no-secrets", EnforcementLevel: "mandatory", ConfigSchema: map[string]any{"required": []string{"k"}}},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "acme/guard/1.0.0", resp.ID)
	assert.Equal(t, 5, resp.Output.Version)
	assert.NotEmpty(t, resp.Output.ContentHash)
	assert.Equal(t, "guard", capturedReq.Name)
	assert.Equal(t, "Guard", capturedReq.DisplayName)
	assert.Equal(t, "1.0.0", capturedReq.VersionTag)
	require.Len(t, capturedReq.Policies, 1)
	// normalizeConfigSchema should have defaulted type=object
	require.NotNil(t, capturedReq.Policies[0].ConfigSchema)
	assert.Equal(t, apitype.Object, capturedReq.Policies[0].ConfigSchema.Type)
	assert.Equal(t, []string{"k"}, capturedReq.Policies[0].ConfigSchema.Required)
}

func TestPolicyPack_Create_PublishError(t *testing.T) {
	dir := writePolicySource(t)
	mock := &PolicyPackClientMock{
		publishFunc: func(context.Context, string, pulumiapi.CreatePolicyPackRequest, io.Reader) (int, error) {
			return 0, errors.New("boom")
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	_, err := (&PolicyPack{}).Create(ctx, infer.CreateRequest[PolicyPackInput]{
		Inputs: PolicyPackInput{
			Organization: "acme",
			Name:         "guard",
			VersionTag:   "1.0.0",
			SourcePath:   dir,
			Policies:     []PolicyPackPolicyInput{{Name: "rule"}},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "publish policy pack")
}

func TestPolicyPack_Create_TarballError(t *testing.T) {
	mock := &PolicyPackClientMock{}
	ctx := config.WithMockClient(context.Background(), mock)
	_, err := (&PolicyPack{}).Create(ctx, infer.CreateRequest[PolicyPackInput]{
		Inputs: PolicyPackInput{
			Organization: "acme",
			Name:         "guard",
			VersionTag:   "1.0.0",
			SourcePath:   filepath.Join(t.TempDir(), "does-not-exist"),
			Policies:     []PolicyPackPolicyInput{{Name: "rule"}},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "package policy pack")
}

func TestPolicyPack_Diff_NoChanges(t *testing.T) {
	dir := writePolicySource(t)
	hash, err := hashPolicyPackSource(dir)
	require.NoError(t, err)

	state := PolicyPackState{
		PolicyPackInput: PolicyPackInput{
			Organization: "acme",
			Name:         "guard",
			VersionTag:   "1.0.0",
			DisplayName:  "Guard",
			SourcePath:   dir,
		},
		ContentHash: hash,
	}
	resp, err := (&PolicyPack{}).Diff(context.Background(), infer.DiffRequest[PolicyPackInput, PolicyPackState]{
		Inputs: state.PolicyPackInput,
		State:  state,
	})
	require.NoError(t, err)
	assert.False(t, resp.HasChanges)
	assert.Empty(t, resp.DetailedDiff)
}

func TestPolicyPack_Diff_ReplacesOnContentChange(t *testing.T) {
	dir := writePolicySource(t)
	state := PolicyPackState{
		PolicyPackInput: PolicyPackInput{
			Organization: "acme",
			Name:         "guard",
			VersionTag:   "1.0.0",
			SourcePath:   dir,
		},
		ContentHash: "not-the-current-hash",
	}
	resp, err := (&PolicyPack{}).Diff(context.Background(), infer.DiffRequest[PolicyPackInput, PolicyPackState]{
		Inputs: state.PolicyPackInput,
		State:  state,
	})
	require.NoError(t, err)
	assert.True(t, resp.HasChanges)
	assert.Contains(t, resp.DetailedDiff, "sourcePath")
}

func TestPolicyPack_Diff_ReplacesOnIdentityChange(t *testing.T) {
	dir := writePolicySource(t)
	hash, err := hashPolicyPackSource(dir)
	require.NoError(t, err)

	state := PolicyPackState{
		PolicyPackInput: PolicyPackInput{
			Organization: "acme",
			Name:         "guard",
			VersionTag:   "1.0.0",
			DisplayName:  "Old",
			SourcePath:   dir,
		},
		ContentHash: hash,
	}
	for _, tc := range []struct {
		field string
		patch func(*PolicyPackInput)
	}{
		{"organization", func(in *PolicyPackInput) { in.Organization = "other" }},
		{"name", func(in *PolicyPackInput) { in.Name = "renamed" }},
		{"versionTag", func(in *PolicyPackInput) { in.VersionTag = "2.0.0" }},
		{"displayName", func(in *PolicyPackInput) { in.DisplayName = "New" }},
	} {
		t.Run(tc.field, func(t *testing.T) {
			inputs := state.PolicyPackInput
			tc.patch(&inputs)
			resp, err := (&PolicyPack{}).Diff(context.Background(), infer.DiffRequest[PolicyPackInput, PolicyPackState]{
				Inputs: inputs,
				State:  state,
			})
			require.NoError(t, err)
			assert.True(t, resp.HasChanges)
			assert.Contains(t, resp.DetailedDiff, tc.field)
		})
	}
}

func TestPolicyPack_Delete(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		called := false
		mock := &PolicyPackClientMock{
			deleteVersionFunc: func(_ context.Context, org, name, tag string) error {
				called = true
				assert.Equal(t, "acme", org)
				assert.Equal(t, "guard", name)
				assert.Equal(t, "1.0.0", tag)
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		_, err := (&PolicyPack{}).Delete(ctx, infer.DeleteRequest[PolicyPackState]{
			State: PolicyPackState{
				PolicyPackInput: PolicyPackInput{Organization: "acme", Name: "guard", VersionTag: "1.0.0"},
			},
		})
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &PolicyPackClientMock{
			deleteVersionFunc: func(context.Context, string, string, string) error {
				return errors.New("upstream 500")
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		_, err := (&PolicyPack{}).Delete(ctx, infer.DeleteRequest[PolicyPackState]{
			State: PolicyPackState{
				PolicyPackInput: PolicyPackInput{Organization: "acme", Name: "guard", VersionTag: "1.0.0"},
			},
		})
		require.Error(t, err)
	})
}

func TestPolicyPack_Read(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mock := &PolicyPackClientMock{
			listFunc: func(_ context.Context, org string) ([]pulumiapi.PolicyPackWithVersions, error) {
				assert.Equal(t, "acme", org)
				return []pulumiapi.PolicyPackWithVersions{
					{Name: "other"},
					{Name: "guard", DisplayName: "Guard", Versions: []int{4, 5}, VersionTags: []string{"0.9.0", "1.0.0"}},
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		existing := PolicyPackState{
			PolicyPackInput: PolicyPackInput{SourcePath: "ignored"},
			ContentHash:     "preserved-hash",
		}
		resp, err := (&PolicyPack{}).Read(ctx, infer.ReadRequest[PolicyPackInput, PolicyPackState]{
			ID:    "acme/guard/1.0.0",
			State: existing,
		})
		require.NoError(t, err)
		assert.Equal(t, "acme/guard/1.0.0", resp.ID)
		assert.Equal(t, "Guard", resp.Inputs.DisplayName)
		assert.Equal(t, 5, resp.State.Version)
		assert.Equal(t, "preserved-hash", resp.State.ContentHash)
	})

	t.Run("missing pack returns empty response", func(t *testing.T) {
		mock := &PolicyPackClientMock{
			listFunc: func(context.Context, string) ([]pulumiapi.PolicyPackWithVersions, error) {
				return []pulumiapi.PolicyPackWithVersions{{Name: "other"}}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		resp, err := (&PolicyPack{}).Read(ctx, infer.ReadRequest[PolicyPackInput, PolicyPackState]{
			ID: "acme/guard/1.0.0",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.ID)
	})

	t.Run("missing version returns empty response", func(t *testing.T) {
		mock := &PolicyPackClientMock{
			listFunc: func(context.Context, string) ([]pulumiapi.PolicyPackWithVersions, error) {
				return []pulumiapi.PolicyPackWithVersions{
					{Name: "guard", Versions: []int{1}, VersionTags: []string{"0.9.0"}},
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		resp, err := (&PolicyPack{}).Read(ctx, infer.ReadRequest[PolicyPackInput, PolicyPackState]{
			ID: "acme/guard/1.0.0",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.ID)
	})

	t.Run("malformed id rejected", func(t *testing.T) {
		mock := &PolicyPackClientMock{}
		ctx := config.WithMockClient(context.Background(), mock)
		_, err := (&PolicyPack{}).Read(ctx, infer.ReadRequest[PolicyPackInput, PolicyPackState]{
			ID: "bogus",
		})
		require.Error(t, err)
	})
}

func TestResolvePolicies_Inline(t *testing.T) {
	in := PolicyPackInput{
		Policies: []PolicyPackPolicyInput{
			{Name: "no-secrets", ConfigSchema: map[string]any{"required": []string{"k"}}},
		},
	}
	got, err := resolvePolicies(context.Background(), in)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "object", got[0].ConfigSchema["type"])
}
