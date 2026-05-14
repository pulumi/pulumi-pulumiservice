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
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/infer/types"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/asset"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// defaultProject mirrors the ESC default project name used when an
// Environment's `project` is omitted; preserved from the legacy resource.
const defaultProject = "default"

type Environment struct{}

var (
	_ infer.CustomCreate[EnvironmentInput, EnvironmentState] = &Environment{}
	_ infer.CustomUpdate[EnvironmentInput, EnvironmentState] = &Environment{}
	_ infer.CustomDelete[EnvironmentState]                   = &Environment{}
	_ infer.CustomRead[EnvironmentInput, EnvironmentState]   = &Environment{}
	_ infer.CustomCheck[EnvironmentInput]                    = &Environment{}
)

func (*Environment) Annotate(a infer.Annotator) {
	a.Describe(&Environment{}, "An ESC Environment.")
	a.SetToken("index", "Environment")
}

type EnvironmentInput struct {
	Organization string               `pulumi:"organization"     provider:"replaceOnChanges"`
	Project      string               `pulumi:"project,optional" provider:"replaceOnChanges"`
	Name         string               `pulumi:"name"             provider:"replaceOnChanges"`
	Yaml         types.AssetOrArchive `pulumi:"yaml"`
}

func (i *EnvironmentInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.SetDefault(&i.Project, defaultProject)
	a.Describe(&i.Name, "Environment name.")
	a.Describe(&i.Yaml, "Environment's yaml file.")
}

// EnvironmentState matches the legacy schema's output shape: `project` is
// required in state (with a default of "default") so existing SDK consumers
// still see a non-nullable `Project` output even though it is optional on
// input. We can't reuse [EnvironmentInput] verbatim because the optional
// input tag would propagate to the output schema.
type EnvironmentState struct {
	Organization  string               `pulumi:"organization"`
	Project       string               `pulumi:"project"`
	Name          string               `pulumi:"name"`
	Yaml          types.AssetOrArchive `pulumi:"yaml"`
	Revision      int                  `pulumi:"revision"`
	EnvironmentID string               `pulumi:"environmentId,optional"`
}

func (s *EnvironmentState) Annotate(a infer.Annotator) {
	a.Describe(&s.Organization, "Organization name.")
	a.Describe(&s.Project, "Project name.")
	a.Describe(&s.Name, "Environment name.")
	a.Describe(&s.Yaml, "Environment's yaml file.")
	a.Describe(&s.Revision, "Revision number of the latest version.")
	a.Describe(
		&s.EnvironmentID,
		"The environment's UUID. Use this as the `identity` value when pinning a custom RBAC role to this "+
			"environment via a `PermissionLiteralExpressionEnvironment` in `OrganizationRole.permissions`, or pass "+
			"it directly to the `buildEnvironmentScopedPermissions` helper.",
	)
}

// stateFromInputs builds an EnvironmentState carrying the same field values
// as the given inputs, with `project` defaulted to "default" if unset.
func stateFromInputs(in EnvironmentInput) EnvironmentState {
	return EnvironmentState{
		Organization: in.Organization,
		Project:      projectOrDefault(in.Project),
		Name:         in.Name,
		Yaml:         in.Yaml,
	}
}

// Check validates required fields and rejects identifiers containing `/`.
// Identical semantics to the legacy resource's Check.
func (*Environment) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[EnvironmentInput], error) {
	inputs, failures, err := infer.DefaultCheck[EnvironmentInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[EnvironmentInput]{Inputs: inputs, Failures: failures}, err
	}

	// `/` is the ID separator. Reject it in identifier fields so we don't
	// have to disambiguate later. yaml is intentionally excluded — slashes
	// in environment bodies are common.
	type idField struct {
		name string
		val  string
	}
	for _, f := range []idField{
		{"organization", inputs.Organization},
		{"project", inputs.Project},
		{"name", inputs.Name},
	} {
		if strings.Contains(f.val, "/") {
			failures = append(failures, p.CheckFailure{
				Property: f.name,
				Reason:   fmt.Sprintf("'%s' property contains `/` illegal character", f.name),
			})
		}
	}

	return infer.CheckResponse[EnvironmentInput]{Inputs: inputs, Failures: failures}, nil
}

func (*Environment) Create(
	ctx context.Context,
	req infer.CreateRequest[EnvironmentInput],
) (infer.CreateResponse[EnvironmentState], error) {
	inputs := req.Inputs
	projectName := projectOrDefault(inputs.Project)

	if req.DryRun {
		return infer.CreateResponse[EnvironmentState]{
			Output: stateFromInputs(inputs),
		}, nil
	}

	yamlBytes, err := extractYAMLBytes(inputs.Yaml)
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("reading environment yaml: %w", err)
	}

	escClient := config.GetEscClient(ctx)

	// Pre-flight check before committing to creation. Mirrors legacy behaviour:
	// if the YAML is broken we surface diagnostics rather than leaving an
	// empty environment behind.
	_, diagnostics, err := escClient.CheckYAMLEnvironment(
		ctx, inputs.Organization, yamlBytes, esc_client.CheckYAMLOption{},
	)
	if diagnostics != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to check environment, yaml code failed following checks: %+v", diagnostics,
		)
	}
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("failed to check environment due to error: %w", err)
	}

	// ESC requires a two-step create (env, then push YAML).
	if err := escClient.CreateEnvironmentWithProject(ctx, inputs.Organization, projectName, inputs.Name); err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("failed to create new environment due to error: %w", err)
	}
	diagnostics, revision, err := escClient.UpdateEnvironmentWithRevision(
		ctx, inputs.Organization, projectName, inputs.Name, yamlBytes, "",
	)
	if diagnostics != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to update brand new environment with pre-checked yaml, due to failing the following checks: %+v\n"+
				"This should never happen, if you're seeing this message there's likely a bug in ESC APIs",
			diagnostics,
		)
	}
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to push yaml into environment due to error: %w", err,
		)
	}

	envID, err := fetchEnvironmentID(ctx, inputs.Organization, projectName, inputs.Name)
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("failed to resolve new environment's id: %w", err)
	}

	out := stateFromInputs(inputs)
	out.Revision = revision
	out.EnvironmentID = envID
	return infer.CreateResponse[EnvironmentState]{
		ID:     environmentResourceID(inputs.Organization, projectName, inputs.Name),
		Output: out,
	}, nil
}

func (*Environment) Update(
	ctx context.Context,
	req infer.UpdateRequest[EnvironmentInput, EnvironmentState],
) (infer.UpdateResponse[EnvironmentState], error) {
	inputs := req.Inputs
	projectName := projectOrDefault(inputs.Project)

	if req.DryRun {
		out := stateFromInputs(inputs)
		out.Revision = req.State.Revision
		out.EnvironmentID = req.State.EnvironmentID
		return infer.UpdateResponse[EnvironmentState]{Output: out}, nil
	}

	yamlBytes, err := extractYAMLBytes(inputs.Yaml)
	if err != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf("reading environment yaml: %w", err)
	}

	escClient := config.GetEscClient(ctx)
	diagnostics, revision, err := escClient.UpdateEnvironmentWithRevision(
		ctx, inputs.Organization, projectName, inputs.Name, yamlBytes, "",
	)
	if diagnostics != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to update environment, yaml code failed following checks: %+v", diagnostics,
		)
	}
	if err != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf("failed to update environment due to error: %w", err)
	}

	envID, err := fetchEnvironmentID(ctx, inputs.Organization, projectName, inputs.Name)
	if err != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf("failed to resolve environment's id: %w", err)
	}

	out := stateFromInputs(inputs)
	out.Revision = revision
	out.EnvironmentID = envID
	return infer.UpdateResponse[EnvironmentState]{Output: out}, nil
}

func (*Environment) Delete(
	ctx context.Context, req infer.DeleteRequest[EnvironmentState],
) (infer.DeleteResponse, error) {
	projectName := projectOrDefault(req.State.Project)
	return infer.DeleteResponse{}, config.GetEscClient(ctx).DeleteEnvironment(
		ctx, req.State.Organization, projectName, req.State.Name,
	)
}

func (*Environment) Read(
	ctx context.Context,
	req infer.ReadRequest[EnvironmentInput, EnvironmentState],
) (infer.ReadResponse[EnvironmentInput, EnvironmentState], error) {
	orgName, projectName, envName, err := splitEnvironmentID(req.ID)
	if err != nil {
		return infer.ReadResponse[EnvironmentInput, EnvironmentState]{}, err
	}

	escClient := config.GetEscClient(ctx)
	retrievedYaml, _, revision, err := escClient.GetEnvironment(ctx, orgName, projectName, envName, "", false)
	if err != nil {
		// Match legacy semantics: treat any error as "not found" so refresh
		// drops the resource from state rather than failing the operation.
		return infer.ReadResponse[EnvironmentInput, EnvironmentState]{}, nil
	}

	trimmedYaml := strings.TrimSpace(string(retrievedYaml))
	yamlAsset, err := asset.FromText(trimmedYaml)
	if err != nil {
		return infer.ReadResponse[EnvironmentInput, EnvironmentState]{}, fmt.Errorf(
			"failed to wrap yaml in asset: %w", err,
		)
	}

	inputs := EnvironmentInput{
		Organization: orgName,
		Project:      projectName,
		Name:         envName,
		Yaml:         types.AssetOrArchive{Asset: yamlAsset},
	}

	// Best-effort: legacy state (pre-environmentId) refreshed against an
	// older provider build can still be missing this field. Don't fail
	// refresh just because the metadata fetch errored.
	envID, err := fetchEnvironmentID(ctx, orgName, projectName, envName)
	if err != nil {
		envID = ""
	}

	state := stateFromInputs(inputs)
	state.Revision = revision
	state.EnvironmentID = envID
	return infer.ReadResponse[EnvironmentInput, EnvironmentState]{
		ID:     environmentResourceID(orgName, projectName, envName),
		Inputs: inputs,
		State:  state,
	}, nil
}

// extractYAMLBytes pulls the YAML body out of an `AssetOrArchive` input,
// trimming surrounding whitespace to match legacy Check semantics. Returns
// an error if the input is an archive rather than an asset.
func extractYAMLBytes(aoa types.AssetOrArchive) ([]byte, error) {
	if aoa.Archive != nil {
		return nil, fmt.Errorf("yaml must be an asset, not an archive")
	}
	if aoa.Asset == nil {
		return nil, nil
	}
	if aoa.Asset.Text != "" {
		return []byte(strings.TrimSpace(aoa.Asset.Text)), nil
	}
	reader, err := aoa.Asset.Read()
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return []byte(strings.TrimSpace(string(body))), nil
}

func fetchEnvironmentID(ctx context.Context, orgName, projectName, envName string) (string, error) {
	meta, err := config.GetClient(ctx).GetEnvironmentMetadata(ctx, orgName, projectName, envName)
	if err != nil {
		return "", err
	}
	if meta == nil {
		return "", nil
	}
	return meta.ID, nil
}

func projectOrDefault(p string) string {
	if p == "" {
		return defaultProject
	}
	return p
}

func environmentResourceID(orgName, projectName, envName string) string {
	return path.Join(orgName, projectName, envName)
}

// splitEnvironmentID accepts either the canonical `<org>/<project>/<env>`
// form or the legacy `<org>/<env>` form (pre-0.25.0) used by older imports.
func splitEnvironmentID(id string) (orgName, projectName, envName string, err error) {
	parts := strings.Split(id, "/")
	switch len(parts) {
	case 3:
		return parts[0], parts[1], parts[2], nil
	case 2:
		return parts[0], defaultProject, parts[1], nil
	default:
		return "", "", "", fmt.Errorf("invalid environment id: %s", id)
	}
}
