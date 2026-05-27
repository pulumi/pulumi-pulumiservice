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
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag/colors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type PolicyPack struct{}

var (
	_ infer.CustomCreate[PolicyPackInput, PolicyPackState] = &PolicyPack{}
	_ infer.CustomDelete[PolicyPackState]                  = &PolicyPack{}
	_ infer.CustomRead[PolicyPackInput, PolicyPackState]   = &PolicyPack{}
	_ infer.CustomDiff[PolicyPackInput, PolicyPackState]   = &PolicyPack{}
)

func (*PolicyPack) Annotate(a infer.Annotator) {
	a.Describe(&PolicyPack{}, "A Policy Pack published to Pulumi Cloud. The source directory is "+
		"tarballed and uploaded on Create; changing source content publishes a new version (replace).")
	a.SetToken("index", "PolicyPack")
}

type PolicyPackPolicyInput struct {
	Name             string         `pulumi:"name"`
	DisplayName      string         `pulumi:"displayName,optional"`
	Description      string         `pulumi:"description,optional"`
	EnforcementLevel string         `pulumi:"enforcementLevel,optional"`
	Message          string         `pulumi:"message,optional"`
	ConfigSchema     map[string]any `pulumi:"configSchema,optional"`
}

func (i *PolicyPackPolicyInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "Unique policy name within the pack.")
	a.Describe(&i.EnforcementLevel, "One of: advisory, mandatory, disabled.")
	a.Describe(&i.ConfigSchema, "JSON Schema (properties/required/type) for the policy's runtime config. "+
		"Values are supplied per-policy via the PolicyGroup's policyPacks[].config map.")
}

type PolicyPackInput struct {
	Organization string                  `pulumi:"organization" provider:"replaceOnChanges"`
	Name         string                  `pulumi:"name"         provider:"replaceOnChanges"`
	DisplayName  string                  `pulumi:"displayName,optional" provider:"replaceOnChanges"`
	VersionTag   string                  `pulumi:"versionTag"   provider:"replaceOnChanges"`
	SourcePath   string                  `pulumi:"sourcePath"`
	Policies     []PolicyPackPolicyInput `pulumi:"policies,optional"`
}

func (i *PolicyPackInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Name, "Policy pack name (unique within the org).")
	a.Describe(&i.DisplayName, "Optional display name. Changing it requires a new versionTag "+
		"(policy pack versions are immutable in Pulumi Cloud).")
	a.Describe(&i.VersionTag, "Semantic version tag (e.g. \"1.0.0\"). Versions are immutable; "+
		"change to publish a new version.")
	a.Describe(&i.SourcePath, "Path to the directory containing the policy pack source. "+
		"The directory is tarballed and uploaded.")
	a.Describe(&i.Policies, "Metadata for each policy in the pack.")
}

type PolicyPackState struct {
	PolicyPackInput
	Version     int    `pulumi:"version"`
	ContentHash string `pulumi:"contentHash"`
}

func (*PolicyPack) Create(
	ctx context.Context,
	req infer.CreateRequest[PolicyPackInput],
) (infer.CreateResponse[PolicyPackState], error) {
	in := req.Inputs
	policies, err := resolvePolicies(ctx, in)
	if err != nil {
		return infer.CreateResponse[PolicyPackState]{}, err
	}
	in.Policies = policies

	if req.DryRun {
		return infer.CreateResponse[PolicyPackState]{
			Output: PolicyPackState{PolicyPackInput: in},
		}, nil
	}

	archive, hash, err := tarballDirectory(in.SourcePath)
	if err != nil {
		return infer.CreateResponse[PolicyPackState]{}, fmt.Errorf("package policy pack: %w", err)
	}

	apiReq := pulumiapi.CreatePolicyPackRequest{
		Name:        in.Name,
		DisplayName: in.DisplayName,
		VersionTag:  in.VersionTag,
		Policies:    toAPIPolicies(in.Policies),
	}
	version, err := config.GetClient(ctx).PublishPolicyPack(ctx, in.Organization, apiReq, bytes.NewReader(archive))
	if err != nil {
		return infer.CreateResponse[PolicyPackState]{}, fmt.Errorf("publish policy pack %q: %w", in.Name, err)
	}

	return infer.CreateResponse[PolicyPackState]{
		ID: policyPackID(in.Organization, in.Name, in.VersionTag),
		Output: PolicyPackState{
			PolicyPackInput: in,
			Version:         version,
			ContentHash:     hash,
		},
	}, nil
}

func (*PolicyPack) Diff(
	_ context.Context,
	req infer.DiffRequest[PolicyPackInput, PolicyPackState],
) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	add := func(key string, kind p.DiffKind) { diff[key] = p.PropertyDiff{Kind: kind, InputDiff: true} }

	if req.Inputs.Organization != req.State.Organization {
		add("organization", p.UpdateReplace)
	}
	if req.Inputs.Name != req.State.Name {
		add("name", p.UpdateReplace)
	}
	if req.Inputs.VersionTag != req.State.VersionTag {
		add("versionTag", p.UpdateReplace)
	}
	if req.Inputs.DisplayName != req.State.DisplayName {
		add("displayName", p.UpdateReplace)
	}
	// Inline policies diff explicitly; introspected ones rely on the content hash below (re-running the analyzer on every preview is too expensive).
	if len(req.Inputs.Policies) > 0 {
		inResolved := make([]PolicyPackPolicyInput, len(req.Inputs.Policies))
		copy(inResolved, req.Inputs.Policies)
		for i := range inResolved {
			inResolved[i].ConfigSchema = normalizeConfigSchema(inResolved[i].ConfigSchema)
		}
		if !policiesEqual(inResolved, req.State.Policies) {
			add("policies", p.UpdateReplace)
		}
	}

	// Content drift triggers a replace, since published versions are immutable.
	if _, hash, err := tarballDirectory(req.Inputs.SourcePath); err != nil {
		return infer.DiffResponse{}, fmt.Errorf("compute policy pack hash: %w", err)
	} else if hash != req.State.ContentHash {
		add("sourcePath", p.UpdateReplace)
	}

	return infer.DiffResponse{
		HasChanges:   len(diff) > 0,
		DetailedDiff: diff,
	}, nil
}

func (*PolicyPack) Delete(
	ctx context.Context,
	req infer.DeleteRequest[PolicyPackState],
) (infer.DeleteResponse, error) {
	// Delete only this version; the pack may have other versions managed elsewhere.
	err := config.GetClient(ctx).DeletePolicyPackVersion(
		ctx, req.State.Organization, req.State.Name, req.State.VersionTag,
	)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf(
			"delete policy pack %q version %q: %w", req.State.Name, req.State.VersionTag, err,
		)
	}
	return infer.DeleteResponse{}, nil
}

func (*PolicyPack) Read(
	ctx context.Context,
	req infer.ReadRequest[PolicyPackInput, PolicyPackState],
) (infer.ReadResponse[PolicyPackInput, PolicyPackState], error) {
	org, name, versionTag, err := splitPolicyPackID(req.ID)
	if err != nil {
		return infer.ReadResponse[PolicyPackInput, PolicyPackState]{}, err
	}
	// /latest is unreliable on some backends; list and look up by tag.
	packs, err := config.GetClient(ctx).ListPolicyPacks(ctx, org)
	if err != nil {
		return infer.ReadResponse[PolicyPackInput, PolicyPackState]{}, err
	}
	var (
		matched        *pulumiapi.PolicyPackWithVersions
		numericVersion int
	)
	for i := range packs {
		if packs[i].Name == name {
			matched = &packs[i]
			break
		}
	}
	if matched == nil {
		return infer.ReadResponse[PolicyPackInput, PolicyPackState]{}, nil
	}
	found := false
	for i, vt := range matched.VersionTags {
		if vt == versionTag && i < len(matched.Versions) {
			numericVersion = matched.Versions[i]
			found = true
			break
		}
	}
	if !found {
		return infer.ReadResponse[PolicyPackInput, PolicyPackState]{}, nil
	}
	inputs := req.Inputs
	inputs.Organization = org
	inputs.Name = name
	inputs.VersionTag = versionTag
	inputs.DisplayName = matched.DisplayName
	// Don't re-introspect on refresh; state already holds what the cloud has.
	if len(req.Inputs.Policies) == 0 {
		inputs.Policies = req.State.Policies
	}
	return infer.ReadResponse[PolicyPackInput, PolicyPackState]{
		ID:     policyPackID(org, name, versionTag),
		Inputs: inputs,
		State: PolicyPackState{
			PolicyPackInput: inputs,
			Version:         numericVersion,
			ContentHash:     req.State.ContentHash,
		},
	}, nil
}

func resolvePolicies(ctx context.Context, in PolicyPackInput) ([]PolicyPackPolicyInput, error) {
	if len(in.Policies) > 0 {
		out := make([]PolicyPackPolicyInput, len(in.Policies))
		copy(out, in.Policies)
		for i := range out {
			out[i].ConfigSchema = normalizeConfigSchema(out[i].ConfigSchema)
		}
		return out, nil
	}
	return introspectPolicyPack(ctx, in.SourcePath)
}

// Requires the pack's runtime deps installed (e.g. node_modules) — same as `pulumi policy publish`.
func introspectPolicyPack(ctx context.Context, sourcePath string) ([]PolicyPackPolicyInput, error) {
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("resolve sourcePath %q: %w", sourcePath, err)
	}
	sink := diag.DefaultSink(io.Discard, io.Discard, diag.FormatOptions{Color: colors.Never})
	pctx, err := plugin.NewContext(ctx, sink, sink, nil, nil, "", nil, false, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("init plugin context: %w", err)
	}
	defer func() { _ = pctx.Close() }()

	analyzer, err := plugin.NewPolicyAnalyzer(
		pctx.Host, pctx, tokens.QName("policypack"), absPath, nil, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("spawn policy analyzer for %s: %w", absPath, err)
	}
	defer func() { _ = analyzer.Close() }()

	info, err := analyzer.GetAnalyzerInfo()
	if err != nil {
		return nil, fmt.Errorf("get analyzer info: %w", err)
	}
	out := make([]PolicyPackPolicyInput, len(info.Policies))
	for i, p := range info.Policies {
		out[i] = PolicyPackPolicyInput{
			Name:             p.Name,
			DisplayName:      p.DisplayName,
			Description:      p.Description,
			EnforcementLevel: string(p.EnforcementLevel),
			Message:          p.Message,
			ConfigSchema:     convertAnalyzerConfigSchema(p.ConfigSchema),
		}
	}
	return out, nil
}

func convertAnalyzerConfigSchema(s *plugin.AnalyzerPolicyConfigSchema) map[string]any {
	if s == nil {
		return nil
	}
	out := map[string]any{"type": "object"}
	if len(s.Properties) > 0 {
		props := make(map[string]any, len(s.Properties))
		for k, v := range s.Properties {
			props[k] = map[string]any(v)
		}
		out["properties"] = props
	}
	if len(s.Required) > 0 {
		out["required"] = s.Required
	}
	return out
}

// Cloud accepts type:"" on publish but later 500s on config validation; default it here.
func normalizeConfigSchema(cs map[string]any) map[string]any {
	if len(cs) == 0 {
		return cs
	}
	if _, ok := cs["type"]; ok {
		return cs
	}
	out := make(map[string]any, len(cs)+1)
	for k, v := range cs {
		out[k] = v
	}
	out["type"] = "object"
	return out
}

func toAPIPolicies(ps []PolicyPackPolicyInput) []pulumiapi.Policy {
	out := make([]pulumiapi.Policy, len(ps))
	for i, p := range ps {
		out[i] = pulumiapi.Policy{
			Name:             p.Name,
			DisplayName:      p.DisplayName,
			Description:      p.Description,
			EnforcementLevel: p.EnforcementLevel,
			Message:          p.Message,
			ConfigSchema:     p.ConfigSchema,
		}
	}
	return out
}

func policiesEqual(a, b []PolicyPackPolicyInput) bool {
	return reflect.DeepEqual(a, b)
}

func policyPackID(org, name, versionTag string) string {
	return path.Join(org, name, versionTag)
}

func splitPolicyPackID(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("%q is invalid, must be organization/name/versionTag", id)
	}
	return parts[0], parts[1], parts[2], nil
}
