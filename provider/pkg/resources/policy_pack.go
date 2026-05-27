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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag/colors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/archive"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/nodejs/npm"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type PolicyPack struct{}

var (
	_ infer.CustomCreate[PolicyPackInput, PolicyPackState] = &PolicyPack{}
	_ infer.CustomDelete[PolicyPackState]                  = &PolicyPack{}
	_ infer.CustomRead[PolicyPackInput, PolicyPackState]   = &PolicyPack{}
	_ infer.CustomDiff[PolicyPackInput, PolicyPackState]   = &PolicyPack{}
	_ infer.CustomCheck[PolicyPackInput]                   = &PolicyPack{}
)

func (*PolicyPack) Annotate(a infer.Annotator) {
	a.Describe(&PolicyPack{}, "A Policy Pack published to Pulumi Cloud. The source directory is "+
		"tarballed and uploaded on Create; changing source content publishes a new version (replace).")
	a.SetToken("index", "PolicyPack")
}

type PolicyPackComplianceFrameworkInput struct {
	Name          string `pulumi:"name"`
	Version       string `pulumi:"version,optional"`
	Reference     string `pulumi:"reference,optional"`
	Specification string `pulumi:"specification,optional"`
}

func (f *PolicyPackComplianceFrameworkInput) Annotate(a infer.Annotator) {
	a.Describe(&f.Name, "Compliance framework name (e.g. \"PCI-DSS\", \"SOC2\").")
	a.Describe(&f.Version, "Compliance framework version.")
	a.Describe(&f.Reference, "Reference to the framework (e.g. a control ID).")
	a.Describe(&f.Specification, "Free-form specification text.")
}

type PolicyPackPolicyInput struct {
	Name             string                              `pulumi:"name"`
	DisplayName      string                              `pulumi:"displayName,optional"`
	Description      string                              `pulumi:"description,optional"`
	EnforcementLevel string                              `pulumi:"enforcementLevel,optional"`
	Message          string                              `pulumi:"message,optional"`
	ConfigSchema     map[string]any                      `pulumi:"configSchema,optional"`
	Severity         string                              `pulumi:"severity,optional"`
	Framework        *PolicyPackComplianceFrameworkInput `pulumi:"framework,optional"`
	Tags             []string                            `pulumi:"tags,optional"`
	RemediationSteps string                              `pulumi:"remediationSteps,optional"`
	URL              string                              `pulumi:"url,optional"`
}

func (i *PolicyPackPolicyInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "Unique policy name within the pack.")
	a.Describe(&i.EnforcementLevel, "One of: advisory, mandatory, remediate, disabled.")
	a.Describe(&i.ConfigSchema, "JSON Schema (properties/required/type) for the policy's runtime config. "+
		"Values are supplied per-policy via the PolicyGroup's policyPacks[].config map.")
	a.Describe(&i.Severity, "Severity level: low, medium, high, or critical.")
	a.Describe(&i.Framework, "Compliance framework this policy belongs to.")
	a.Describe(&i.Tags, "Tags associated with the policy.")
	a.Describe(&i.RemediationSteps, "Description of steps to remediate a violation.")
	a.Describe(&i.URL, "URL with more information about the policy.")
}

type PolicyPackInput struct {
	Organization string                  `pulumi:"organization"`
	Name         string                  `pulumi:"name"`
	DisplayName  string                  `pulumi:"displayName,optional"`
	VersionTag   string                  `pulumi:"versionTag"`
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

// Mirrors the regex Pulumi Cloud enforces server-side; we validate up front so
// bad tags surface as actionable errors rather than opaque 400s.
var versionTagRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]{1,100}$`)

func (*PolicyPack) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[PolicyPackInput], error) {
	inputs, failures, err := infer.DefaultCheck[PolicyPackInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[PolicyPackInput]{Inputs: inputs, Failures: failures}, err
	}
	if tag := inputs.VersionTag; tag != "" && !versionTagRegex.MatchString(tag) {
		failures = append(failures, p.CheckFailure{
			Property: "versionTag",
			Reason: fmt.Sprintf(
				"%q is not a valid policy pack version tag (must match %s)",
				tag, versionTagRegex.String(),
			),
		})
	}
	return infer.CheckResponse[PolicyPackInput]{Inputs: inputs, Failures: failures}, nil
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

	tarball, err := packagePolicyPackArchive(ctx, in.SourcePath)
	if err != nil {
		return infer.CreateResponse[PolicyPackState]{}, fmt.Errorf("package policy pack: %w", err)
	}
	hash, err := hashPolicyPackSource(in.SourcePath)
	if err != nil {
		return infer.CreateResponse[PolicyPackState]{}, fmt.Errorf("hash policy pack source: %w", err)
	}

	apiReq := pulumiapi.CreatePolicyPackRequest{
		Name:        in.Name,
		DisplayName: in.DisplayName,
		VersionTag:  in.VersionTag,
		Policies:    toAPIPolicies(in.Policies),
	}
	version, err := config.GetClient(ctx).PublishPolicyPack(ctx, in.Organization, apiReq, bytes.NewReader(tarball))
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

// Versions are immutable on Pulumi Cloud — any input change requires a replace.
// PolicyPack doesn't implement CustomUpdate, so infer would already force-replace
// on every input change. CustomDiff exists solely so we can layer SourcePath
// content drift on top of the default input comparison.
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
	if len(req.Inputs.Policies) > 0 {
		inResolved := make([]PolicyPackPolicyInput, len(req.Inputs.Policies))
		copy(inResolved, req.Inputs.Policies)
		for i := range inResolved {
			inResolved[i].ConfigSchema = normalizeConfigSchema(inResolved[i].ConfigSchema)
		}
		if !reflect.DeepEqual(inResolved, req.State.Policies) {
			add("policies", p.UpdateReplace)
		}
	}

	hash, err := hashPolicyPackSource(req.Inputs.SourcePath)
	if err != nil {
		return infer.DiffResponse{}, fmt.Errorf("hash policy pack source: %w", err)
	}
	if hash != req.State.ContentHash {
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
	for i, pol := range info.Policies {
		out[i] = PolicyPackPolicyInput{
			Name:             pol.Name,
			DisplayName:      pol.DisplayName,
			Description:      pol.Description,
			EnforcementLevel: string(pol.EnforcementLevel),
			Message:          pol.Message,
			ConfigSchema:     convertAnalyzerConfigSchema(pol.ConfigSchema),
			Severity:         string(pol.Severity),
			Framework:        convertAnalyzerFramework(pol.Framework),
			Tags:             append([]string(nil), pol.Tags...),
			RemediationSteps: pol.RemediationSteps,
			URL:              pol.URL,
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

func convertAnalyzerFramework(f *plugin.AnalyzerPolicyComplianceFramework) *PolicyPackComplianceFrameworkInput {
	if f == nil {
		return nil
	}
	return &PolicyPackComplianceFrameworkInput{
		Name:          f.Name,
		Version:       f.Version,
		Reference:     f.Reference,
		Specification: f.Specification,
	}
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

func toAPIPolicies(ps []PolicyPackPolicyInput) []apitype.Policy {
	out := make([]apitype.Policy, len(ps))
	for i, pol := range ps {
		out[i] = apitype.Policy{
			Name:             pol.Name,
			DisplayName:      pol.DisplayName,
			Description:      pol.Description,
			EnforcementLevel: apitype.EnforcementLevel(pol.EnforcementLevel),
			Message:          pol.Message,
			ConfigSchema:     toAPIConfigSchema(pol.ConfigSchema),
			Severity:         apitype.PolicySeverity(pol.Severity),
			Framework:        toAPIFramework(pol.Framework),
			Tags:             append([]string(nil), pol.Tags...),
			RemediationSteps: pol.RemediationSteps,
			URL:              pol.URL,
		}
	}
	return out
}

func toAPIConfigSchema(cs map[string]any) *apitype.PolicyConfigSchema {
	if len(cs) == 0 {
		return nil
	}
	out := &apitype.PolicyConfigSchema{Type: apitype.Object}
	if t, ok := cs["type"].(string); ok && t != "" {
		out.Type = apitype.JSONSchemaType(t)
	}
	if req, ok := cs["required"].([]string); ok {
		out.Required = append([]string(nil), req...)
	} else if reqAny, ok := cs["required"].([]any); ok {
		for _, v := range reqAny {
			if s, ok := v.(string); ok {
				out.Required = append(out.Required, s)
			}
		}
	}
	if props, ok := cs["properties"].(map[string]any); ok && len(props) > 0 {
		out.Properties = make(map[string]*json.RawMessage, len(props))
		for name, v := range props {
			raw, err := json.Marshal(v)
			if err != nil {
				// A non-marshalable value would already have been rejected by
				// the engine before we get here, so this is effectively unreachable.
				continue
			}
			msg := json.RawMessage(raw)
			out.Properties[name] = &msg
		}
	}
	return out
}

func toAPIFramework(f *PolicyPackComplianceFrameworkInput) *apitype.PolicyComplianceFramework {
	if f == nil {
		return nil
	}
	return &apitype.PolicyComplianceFramework{
		Name:          f.Name,
		Version:       f.Version,
		Reference:     f.Reference,
		Specification: f.Specification,
	}
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

// packagePolicyPackArchive matches `pulumi policy publish`: shell out to the
// user's package manager for nodejs (so .npmignore / package.json:files /
// lockfiles are honored), or fall back to archive.TGZ for everything else.
// Both layouts put files under a `package/` prefix — the Cloud's policy-execution
// sandbox unpacks and reads `package/PulumiPolicy.yaml`, so the prefix is
// load-bearing.
func packagePolicyPackArchive(ctx context.Context, sourcePath string) ([]byte, error) {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", sourcePath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", sourcePath)
	}

	pulumiPolicyPath, err := workspace.DetectPolicyPackPathAt(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("detect PulumiPolicy file in %q: %w", sourcePath, err)
	}
	if pulumiPolicyPath == "" {
		return nil, fmt.Errorf("%q is missing a PulumiPolicy.yaml", sourcePath)
	}
	pack, err := workspace.LoadPolicyPack(pulumiPolicyPath)
	if err != nil {
		return nil, fmt.Errorf("load PulumiPolicy: %w", err)
	}

	if strings.EqualFold(pack.Runtime.Name(), "nodejs") {
		tarball, err := npm.Pack(ctx, npm.AutoPackageManager, sourcePath, io.Discard)
		if err != nil {
			return nil, fmt.Errorf("npm pack: %w", err)
		}
		return tarball, nil
	}
	tarball, err := archive.TGZ(sourcePath, "package", true)
	if err != nil {
		return nil, fmt.Errorf("create .tgz: %w", err)
	}
	return tarball, nil
}

// hashPolicyPackSource produces a deterministic content fingerprint of the
// source directory. We use it for drift detection in Diff, separate from the
// upload tarball — re-running `npm pack` on every preview would shell out to
// npm, which is too expensive for a hot path.
func hashPolicyPackSource(sourcePath string) (string, error) {
	hasher := sha256.New()
	err := filepath.WalkDir(sourcePath, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(sourcePath, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		base := d.Name()
		// Skip the same heavy/transient directories the canonical tarball excludes,
		// plus node_modules which isn't always in .gitignore but is never user content.
		if base == "node_modules" || base == ".git" || base == ".pulumi" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(hasher, "%s\x00%d\x00", filepath.ToSlash(rel), fi.Mode())
		switch {
		case fi.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return fmt.Errorf("read symlink %q: %w", rel, err)
			}
			hasher.Write([]byte(target))
		case fi.Mode().IsRegular():
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			_, err = io.Copy(hasher, f)
			_ = f.Close()
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
