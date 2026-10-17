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
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

const (
	gcProject = "project"
	gcYaml    = "yaml"
)

const defaultProject = "default"

const environmentToken tokens.Type = "pulumiservice:index:Environment"

type Environment struct{}

var (
	_ infer.CustomCheck[EnvironmentInput]                    = &Environment{}
	_ infer.CustomCreate[EnvironmentInput, EnvironmentState] = &Environment{}
	_ infer.CustomUpdate[EnvironmentInput, EnvironmentState] = &Environment{}
	_ infer.CustomDelete[EnvironmentState]                   = &Environment{}
	_ infer.CustomRead[EnvironmentInput, EnvironmentState]   = &Environment{}
)

func (*Environment) Annotate(a infer.Annotator) {
	a.Describe(&Environment{}, "An ESC Environment.")
	a.SetToken("index", "Environment")
}

type EnvironmentInput struct {
	Organization string               `pulumi:"organization"     provider:"replaceOnChanges"`
	Project      string               `pulumi:"project,optional" provider:"replaceOnChanges"`
	Name         string               `pulumi:"name"             provider:"replaceOnChanges"`
	Yaml         types.AssetOrArchive `pulumi:"yaml"             provider:"secret"`
}

func (i *EnvironmentInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.SetDefault(&i.Project, defaultProject)
	a.Describe(&i.Name, "Environment name.")
	a.Describe(&i.Yaml, "Environment's yaml file.")
}

// EnvironmentState doesn't embed EnvironmentInput: `project` is optional on
// input but required in state, which keeps the SDKs' `project` output
// non-nullable.
type EnvironmentState struct {
	Organization  string               `pulumi:"organization"`
	Project       string               `pulumi:"project"`
	Name          string               `pulumi:"name"`
	Yaml          types.AssetOrArchive `pulumi:"yaml"                   provider:"secret"`
	Revision      int                  `pulumi:"revision"`
	EnvironmentID string               `pulumi:"environmentId,optional"`
}

func (s *EnvironmentState) Annotate(a infer.Annotator) {
	a.Describe(&s.Organization, "Organization name.")
	a.Describe(&s.Project, "Project name.")
	a.Describe(&s.Name, "Environment name.")
	a.Describe(&s.Yaml, "Environment's yaml file.")
	a.Describe(&s.Revision, "Revision number of the latest version.")
	a.Describe(&s.EnvironmentID,
		"The environment's UUID. Use this as the `identity` value when pinning a custom RBAC role to this "+
			"environment via a `PermissionLiteralExpressionEnvironment` in `OrganizationRole.permissions`, or pass "+
			"it directly to the `buildEnvironmentScopedPermissions` helper.")
}

func (*Environment) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[EnvironmentInput], error) {
	inputs, failures, err := infer.DefaultCheck[EnvironmentInput](ctx, normalizeLegacyEnvironment(req.NewInputs))
	if err != nil || len(failures) > 0 {
		return infer.CheckResponse[EnvironmentInput]{Inputs: inputs, Failures: failures}, err
	}

	// `/` separates the segments of the resource ID.
	for _, f := range []struct{ name, val string }{
		{gcOrganization, inputs.Organization},
		{gcProject, inputs.Project},
		{gcName, inputs.Name},
	} {
		if strings.Contains(f.val, "/") {
			failures = append(failures, p.CheckFailure{
				Property: f.name,
				Reason:   fmt.Sprintf("'%s' property contains `/` illegal character", f.name),
			})
		}
	}

	// Store the yaml as trimmed text, whatever kind of asset it came in as, so it
	// compares equal to what Read returns.
	if inputs.Yaml.Archive != nil {
		failures = append(failures, p.CheckFailure{
			Property: gcYaml,
			Reason:   "'yaml' must be an asset, not an archive",
		})
	} else if inputs.Yaml.Asset != nil {
		text, err := readYamlAsset(inputs.Yaml.Asset)
		if err != nil {
			return infer.CheckResponse[EnvironmentInput]{}, fmt.Errorf("failed to read environment yaml: %w", err)
		}
		yamlAsset, err := asset.FromText(text)
		if err != nil {
			return infer.CheckResponse[EnvironmentInput]{}, err
		}
		inputs.Yaml = types.AssetOrArchive{Asset: yamlAsset}
	}

	return infer.CheckResponse[EnvironmentInput]{Inputs: inputs, Failures: failures}, nil
}

func (*Environment) Create(
	ctx context.Context,
	req infer.CreateRequest[EnvironmentInput],
) (infer.CreateResponse[EnvironmentState], error) {
	in := req.Inputs
	state := environmentStateFromInput(in)
	if req.DryRun {
		return infer.CreateResponse[EnvironmentState]{Output: state}, nil
	}
	yaml, err := readYamlAsset(in.Yaml.Asset)
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("failed to read environment yaml: %w", err)
	}

	escClient := config.GetEscClient(ctx)

	// Check the yaml first, so invalid yaml doesn't leave an empty environment behind.
	_, diagnostics, err := escClient.CheckYAMLEnvironment(ctx, in.Organization, []byte(yaml), esc_client.CheckYAMLOption{})
	if diagnostics != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to check environment, yaml code failed following checks: %+v", diagnostics)
	}
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("failed to check environment due to error: %w", err)
	}

	// Then create environment, and update it with yaml provided. ESC API architecture doesn't let you do it in one call
	err = escClient.CreateEnvironmentWithProject(ctx, in.Organization, state.Project, in.Name)
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to create new environment due to error: %w", err)
	}
	diagnostics, revision, err := escClient.UpdateEnvironmentWithRevision(
		ctx, in.Organization, state.Project, in.Name, []byte(yaml), "")
	if diagnostics != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to update brand new environment with pre-checked yaml, due to failing the following checks: %+v \n"+
				"This should never happen, if you're seeing this message there's likely a bug in ESC APIs",
			diagnostics,
		)
	}
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to push yaml into environment due to error: %w", err)
	}

	envID, err := fetchEnvironmentID(ctx, in.Organization, state.Project, in.Name)
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("failed to resolve new environment's id: %w", err)
	}

	state.Revision = revision
	state.EnvironmentID = envID
	return infer.CreateResponse[EnvironmentState]{
		ID:     environmentResourceID(in.Organization, state.Project, in.Name),
		Output: state,
	}, nil
}

func (*Environment) Update(
	ctx context.Context,
	req infer.UpdateRequest[EnvironmentInput, EnvironmentState],
) (infer.UpdateResponse[EnvironmentState], error) {
	in := req.Inputs
	state := environmentStateFromInput(in)
	if req.DryRun {
		state.Revision = req.State.Revision
		state.EnvironmentID = req.State.EnvironmentID
		return infer.UpdateResponse[EnvironmentState]{Output: state}, nil
	}
	yaml, err := readYamlAsset(in.Yaml.Asset)
	if err != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf("failed to read environment yaml: %w", err)
	}

	diagnostics, revision, err := config.GetEscClient(ctx).UpdateEnvironmentWithRevision(
		ctx, in.Organization, state.Project, in.Name, []byte(yaml), "")
	if diagnostics != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf(
			"failed to update environment, yaml code failed following checks: %+v", diagnostics)
	}
	if err != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf("failed to update environment due to error: %w", err)
	}

	envID, err := fetchEnvironmentID(ctx, in.Organization, state.Project, in.Name)
	if err != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf("failed to resolve environment's id: %w", err)
	}

	state.Revision = revision
	state.EnvironmentID = envID
	return infer.UpdateResponse[EnvironmentState]{Output: state}, nil
}

func (*Environment) Delete(
	ctx context.Context, req infer.DeleteRequest[EnvironmentState],
) (infer.DeleteResponse, error) {
	err := config.GetEscClient(ctx).DeleteEnvironment(
		ctx, req.State.Organization, projectOrDefault(req.State.Project), req.State.Name)
	return infer.DeleteResponse{}, err
}

func (*Environment) Read(
	ctx context.Context,
	req infer.ReadRequest[EnvironmentInput, EnvironmentState],
) (infer.ReadResponse[EnvironmentInput, EnvironmentState], error) {
	orgName, projectName, envName, err := splitEnvironmentID(req.ID)
	if err != nil {
		return infer.ReadResponse[EnvironmentInput, EnvironmentState]{}, err
	}

	retrievedYaml, _, revision, err := config.GetEscClient(ctx).GetEnvironment(
		ctx, orgName, projectName, envName, "", false)
	if esc_client.IsNotFound(err) {
		return infer.ReadResponse[EnvironmentInput, EnvironmentState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[EnvironmentInput, EnvironmentState]{},
			fmt.Errorf("failed to read Environment (%q): %w", req.ID, err)
	}

	yamlAsset, err := asset.FromText(strings.TrimSpace(string(retrievedYaml)))
	if err != nil {
		return infer.ReadResponse[EnvironmentInput, EnvironmentState]{}, err
	}
	inputs := EnvironmentInput{
		Organization: orgName,
		Project:      projectName,
		Name:         envName,
		Yaml:         types.AssetOrArchive{Asset: yamlAsset},
	}

	// Best-effort: don't fail refresh just because the metadata fetch errored.
	envID, err := fetchEnvironmentID(ctx, orgName, projectName, envName)
	if err != nil {
		envID = ""
	}

	state := environmentStateFromInput(inputs)
	state.Revision = revision
	state.EnvironmentID = envID
	return infer.ReadResponse[EnvironmentInput, EnvironmentState]{
		// Keep the ID as given, so state using the legacy `org/env` form doesn't change on refresh.
		ID:     req.ID,
		Inputs: inputs,
		State:  state,
	}, nil
}

func environmentStateFromInput(in EnvironmentInput) EnvironmentState {
	return EnvironmentState{
		Organization: in.Organization,
		Project:      projectOrDefault(in.Project),
		Name:         in.Name,
		Yaml:         in.Yaml,
	}
}

// readYamlAsset returns the trimmed text of a yaml asset, reading it from its
// path or URI if it isn't a text asset.
func readYamlAsset(a *asset.Asset) (string, error) {
	if a == nil {
		return "", nil
	}
	if a.IsText() {
		return strings.TrimSpace(a.Text), nil
	}
	reader, err := a.Read()
	if err != nil {
		return "", err
	}
	defer func() { _ = reader.Close() }()
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// fetchEnvironmentID resolves an environment's UUID via the metadata endpoint.
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

func projectOrDefault(project string) string {
	if project == "" {
		return defaultProject
	}
	return project
}

func environmentResourceID(orgName, projectName, envName string) string {
	return path.Join(orgName, projectName, envName)
}

func splitEnvironmentID(id string) (orgName, projectName, envName string, err error) {
	// Format:
	//   organization/project/environment or
	//   organization/environment (legacy)
	s := strings.Split(id, "/")
	switch len(s) {
	case 3:
		return s[0], s[1], s[2], nil
	case 2:
		// Legacy pattern. Assume "default" project.
		return s[0], defaultProject, s[1], nil
	default:
		return "", "", "", fmt.Errorf("invalid environment id: %s", id)
	}
}

// normalizeLegacyEnvironment converts Environment inputs or state written by
// the pre-infer resource to the shape infer expects:
//   - the legacy resource stored `yaml` as a secret string, not an asset;
//   - state from before v0.25.0 has no `project`.
//
// Maps already in the current shape, and empty maps, are returned unchanged.
func normalizeLegacyEnvironment(m property.Map) property.Map {
	if m.Len() == 0 {
		return m
	}
	if y, ok := m.GetOk(gcYaml); ok && y.IsString() {
		// FromText can't fail on text; it only hashes it.
		if yamlAsset, err := asset.FromText(strings.TrimSpace(y.AsString())); err == nil {
			m = m.Set(gcYaml, property.New(yamlAsset).WithSecret(y.Secret()))
		}
	}
	if pr, ok := m.GetOk(gcProject); !ok || pr.IsNull() {
		m = m.Set(gcProject, property.New(defaultProject))
	}
	return m
}

// WithEnvironmentCompat keeps the infer-based Environment compatible with the
// pre-infer resource. It applies only to Environment requests.
//
//   - It normalizes, with [normalizeLegacyEnvironment], the old inputs and
//     state written by the pre-infer resource. A [infer.CustomStateMigrations]
//     isn't enough: the default Diff compares the engine's raw old inputs, so
//     legacy inputs would show a phantom `yaml` update and, without `project`,
//     a phantom replace. During refresh, the engine also calls Diff with the
//     old inputs from state as the new inputs, which never pass through Check.
//   - It deletes before replacing, as the pre-infer resource did. ESC
//     environment names are case-insensitive, so creating the replacement
//     first fails with a 409 when a rename only changes case.
//   - It returns the `yaml` output as a secret string, as the pre-infer
//     resource did, even though the schema types it as an asset. Programs in
//     dynamic languages read it as a string.
//   - It keeps `yaml` secret on Read. infer doesn't apply the `secret` tag when
//     a Read has no prior state, so `pulumi import` would store it in plaintext.
func WithEnvironmentCompat(prov p.Provider) p.Provider {
	create, diff, read, update, del := prov.Create, prov.Diff, prov.Read, prov.Update, prov.Delete
	if create != nil {
		prov.Create = func(ctx context.Context, req p.CreateRequest) (p.CreateResponse, error) {
			resp, err := create(ctx, req)
			if req.Urn.Type() == environmentToken {
				resp.Properties = legacyYamlOutput(resp.Properties)
			}
			return resp, err
		}
	}
	if diff != nil {
		prov.Diff = func(ctx context.Context, req p.DiffRequest) (p.DiffResponse, error) {
			if req.Urn.Type() != environmentToken {
				return diff(ctx, req)
			}
			req.OldInputs = normalizeLegacyEnvironment(req.OldInputs)
			req.State = normalizeLegacyEnvironment(req.State)
			req.Inputs = normalizeLegacyEnvironment(req.Inputs)
			resp, err := diff(ctx, req)
			for _, d := range resp.DetailedDiff {
				switch d.Kind {
				case p.AddReplace, p.DeleteReplace, p.UpdateReplace:
					resp.DeleteBeforeReplace = true
				}
			}
			return resp, err
		}
	}
	if read != nil {
		prov.Read = func(ctx context.Context, req p.ReadRequest) (p.ReadResponse, error) {
			if req.Urn.Type() != environmentToken {
				return read(ctx, req)
			}
			req.Inputs = normalizeLegacyEnvironment(req.Inputs)
			req.Properties = normalizeLegacyEnvironment(req.Properties)
			resp, err := read(ctx, req)
			resp.Inputs = makeYamlSecret(resp.Inputs)
			resp.Properties = legacyYamlOutput(resp.Properties)
			return resp, err
		}
	}
	if update != nil {
		prov.Update = func(ctx context.Context, req p.UpdateRequest) (p.UpdateResponse, error) {
			if req.Urn.Type() != environmentToken {
				return update(ctx, req)
			}
			req.OldInputs = normalizeLegacyEnvironment(req.OldInputs)
			req.State = normalizeLegacyEnvironment(req.State)
			resp, err := update(ctx, req)
			resp.Properties = legacyYamlOutput(resp.Properties)
			return resp, err
		}
	}
	if del != nil {
		prov.Delete = func(ctx context.Context, req p.DeleteRequest) error {
			if req.Urn.Type() == environmentToken {
				req.OldInputs = normalizeLegacyEnvironment(req.OldInputs)
				req.Properties = normalizeLegacyEnvironment(req.Properties)
			}
			return del(ctx, req)
		}
	}
	return prov
}

// legacyYamlOutput converts a `yaml` text asset in m to a secret string.
func legacyYamlOutput(m property.Map) property.Map {
	if y, ok := m.GetOk(gcYaml); ok && y.IsAsset() && y.AsAsset() != nil {
		return m.Set(gcYaml, property.New(y.AsAsset().Text).WithSecret(true))
	}
	return m
}

func makeYamlSecret(m property.Map) property.Map {
	if y, ok := m.GetOk(gcYaml); ok && !y.Secret() {
		return m.Set(gcYaml, y.WithSecret(true))
	}
	return m
}
