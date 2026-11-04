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
	"errors"
	"fmt"
	"path"
	"strings"

	_ "embed" // for the //go:embed example file below

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/mapper"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// DeploymentSettingsToken is the type token of DeploymentSettings.
const DeploymentSettingsToken = "pulumiservice:index:DeploymentSettings"

// DeploymentSettings is pulumiservice:index:DeploymentSettings.
//
// Its methods read the untyped request with [infer.GetRawRequest] where the
// typed values lose information Pulumi Cloud needs: which environment variables
// the program marked secret, and the old inputs that Diff compares.
type DeploymentSettings struct{}

var (
	_ infer.CustomResource[DeploymentSettingsInput, DeploymentSettingsState] = &DeploymentSettings{}
	_ infer.CustomCheck[DeploymentSettingsInput]                             = &DeploymentSettings{}
	_ infer.CustomDiff[DeploymentSettingsInput, DeploymentSettingsState]     = &DeploymentSettings{}
	_ infer.CustomRead[DeploymentSettingsInput, DeploymentSettingsState]     = &DeploymentSettings{}
	_ infer.CustomUpdate[DeploymentSettingsInput, DeploymentSettingsState]   = &DeploymentSettings{}
	_ infer.CustomDelete[DeploymentSettingsState]                            = &DeploymentSettings{}
)

// Generated from examples/deployment_settings_examples.yaml; see the
// //go:generate directive in org_access_token.go.
//
//go:embed deployment_settings_examples.md
var deploymentSettingsExamples string

func (r *DeploymentSettings) Annotate(a infer.Annotator) {
	a.Describe(r, "Deployment settings configure Pulumi Deployments for a stack.\n\n"+
		escRecommendation+" See "+cloudCredentialsGuide+" to choose between ESC and Deployments OIDC "+
		"(`operationContext.oidc`).\n\n"+
		strings.TrimSpace(deploymentSettingsExamples)+
		util.ImportDocs(DeploymentSettingsToken, "{organization}/{project}/{stack}", "my-org/my-project/my-stack",
			"Secret values can't be read back from Pulumi Cloud, so import sets each one to the placeholder "+
				"`<REPLACE WITH ACTUAL SECRET VALUE>`. Replace them with the real values in your program."))
	a.SetToken("index", "DeploymentSettings")
}

func (*DeploymentSettings) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[DeploymentSettingsInput], error) {
	in, failures, err := infer.DefaultCheck[DeploymentSettingsInput](ctx, req.NewInputs)
	if err != nil || len(failures) > 0 {
		return infer.CheckResponse[DeploymentSettingsInput]{Inputs: in, Failures: failures}, err
	}
	// DefaultCheck fills in the schema defaults, but the provider has always
	// sent an omitted flag to Pulumi Cloud as false. It also can't tell an
	// explicit false from an omitted flag, so it would replace one with a true
	// default.
	restoreTrueDefaultFlags(&in, req.NewInputs)

	if oc := in.OperationContext; oc != nil && oc.OIDC != nil && oc.OIDC.AWS != nil && oc.OIDC.AWS.Duration != "" {
		normalized, err := normalizeDurationString(oc.OIDC.AWS.Duration)
		if err != nil {
			failures = append(failures, p.CheckFailure{
				Property: "operationContext.oidc.aws.duration",
				Reason:   "Failed to normalize duration string due to error: " + err.Error(),
			})
		} else {
			oc.OIDC.AWS.Duration = *normalized
		}
	}
	failures = append(failures, checkGCPProjectNumber(req.NewInputs)...)
	return infer.CheckResponse[DeploymentSettingsInput]{Inputs: in, Failures: failures}, nil
}

// checkGCPProjectNumber rejects a GCP project ID passed where Pulumi Cloud
// expects the project number. Pulumi Cloud doesn't check the value, so without
// this the mistake only surfaces when a deployment fails to get credentials.
// It reads the raw inputs because the typed inputs can't tell an unknown value,
// which is fine during preview, from a missing one.
func checkGCPProjectNumber(news property.Map) []p.CheckFailure {
	gcp := news.Get(gcOperationContext)
	for _, key := range []string{gcOIDC, gcGCP} {
		if !gcp.IsMap() {
			return nil
		}
		gcp = gcp.AsMap().Get(key)
	}
	if !gcp.IsMap() {
		return nil
	}
	const prefix = "operationContext.oidc.gcp."
	var failures []p.CheckFailure
	set := false
	for _, key := range []string{gcProjectNumber, gcProjectID} {
		v := gcp.AsMap().Get(key)
		if v.IsComputed() {
			set = true
			continue
		}
		if !v.IsString() || v.AsString() == "" {
			continue
		}
		set = true
		if value := v.AsString(); !isDigits(value) {
			failures = append(failures, p.CheckFailure{
				Property: prefix + key,
				Reason: fmt.Sprintf("%q is not a GCP project number. %s takes the numeric project number, "+
					"not the project ID. Look it up with "+
					"`gcloud projects describe <project-id> --format='value(projectNumber)'`.", value, key),
			})
		}
	}
	if !set {
		failures = append(failures, p.CheckFailure{
			Property: prefix + gcProjectNumber,
			Reason:   "missing required property 'projectNumber'",
		})
	}
	return failures
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

// restoreTrueDefaultFlags sets the flags whose schema default is true to the
// value in news, or to false when news doesn't set them.
func restoreTrueDefaultFlags(in *DeploymentSettingsInput, news property.Map) {
	flag := func(obj, key string) bool {
		v := news.Get(obj)
		if !v.IsMap() {
			return false
		}
		f := v.AsMap().Get(key)
		return f.IsBool() && f.AsBool()
	}
	if gh := in.GitHub; gh != nil {
		gh.DeployCommits = flag("github", "deployCommits")
		gh.PreviewPullRequests = flag("github", "previewPullRequests")
	}
	if vcs := in.Vcs; vcs != nil {
		vcs.DeployCommits = flag(gcVcs, "deployCommits")
		vcs.PreviewPullRequests = flag(gcVcs, "previewPullRequests")
	}
}

func (*DeploymentSettings) Diff(
	ctx context.Context, _ infer.DiffRequest[DeploymentSettingsInput, DeploymentSettingsState],
) (infer.DiffResponse, error) {
	raw, ok := infer.GetRawRequest[p.DiffRequest](ctx)
	if !ok {
		return infer.DiffResponse{}, errors.New("DeploymentSettings.Diff: no raw request")
	}
	olds, err := plainProperties(raw.OldInputs)
	if err != nil {
		return infer.DiffResponse{}, err
	}
	news, err := plainProperties(raw.Inputs)
	if err != nil {
		return infer.DiffResponse{}, err
	}

	diff := olds.Diff(news)
	if diff == nil {
		return infer.DiffResponse{}, nil
	}
	detailed := map[string]p.PropertyDiff{}
	for k, v := range plugin.NewDetailedDiffFromObjectDiff(diff, false) {
		switch k {
		case gcOrganization, gcProject, gcStack:
			v.Kind = v.Kind.AsReplace()
		}
		detailed[k] = p.PropertyDiff{Kind: diffKind(v.Kind), InputDiff: v.InputDiff}
	}
	return infer.DiffResponse{
		HasChanges:   true,
		DetailedDiff: detailed,
		// Pulumi Cloud has one set of deployment settings per stack, so the old
		// settings have to go before the new ones can be created.
		DeleteBeforeReplace: true,
	}, nil
}

// plainProperties converts m to a property map without secrets or unknowns,
// so that Diff compares only values.
func plainProperties(m property.Map) (resource.PropertyMap, error) {
	marshaled, err := plugin.MarshalProperties(
		resource.ToResourcePropertyValue(property.New(m)).ObjectValue(),
		plugin.MarshalOptions{KeepUnknowns: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}
	return plugin.UnmarshalProperties(marshaled, util.StandardUnmarshal)
}

func diffKind(k plugin.DiffKind) p.DiffKind {
	switch k {
	case plugin.DiffAdd:
		return p.Add
	case plugin.DiffAddReplace:
		return p.AddReplace
	case plugin.DiffDelete:
		return p.Delete
	case plugin.DiffDeleteReplace:
		return p.DeleteReplace
	case plugin.DiffUpdateReplace:
		return p.UpdateReplace
	default:
		return p.Update
	}
}

func (*DeploymentSettings) Create(
	ctx context.Context, req infer.CreateRequest[DeploymentSettingsInput],
) (infer.CreateResponse[DeploymentSettingsState], error) {
	if req.DryRun {
		return infer.CreateResponse[DeploymentSettingsState]{Output: DeploymentSettingsState(req.Inputs)}, nil
	}
	raw, ok := infer.GetRawRequest[p.CreateRequest](ctx)
	if !ok {
		return infer.CreateResponse[DeploymentSettingsState]{}, errors.New("DeploymentSettings.Create: no raw request")
	}
	input := req.Inputs.toAPI(secretEnvironmentVariables(resourceMap(raw.Properties)))
	response, err := config.GetClient(ctx).CreateDeploymentSettings(ctx, input.Stack, input.DeploymentSettings)
	if err != nil {
		return infer.CreateResponse[DeploymentSettingsState]{}, err
	}
	state, err := deploymentSettingsState(req.Inputs, input, response)
	if err != nil {
		return infer.CreateResponse[DeploymentSettingsState]{}, err
	}
	return infer.CreateResponse[DeploymentSettingsState]{
		ID:     path.Join(input.Stack.OrgName, input.Stack.ProjectName, input.Stack.StackName),
		Output: state,
	}, nil
}

func (*DeploymentSettings) Update(
	ctx context.Context, req infer.UpdateRequest[DeploymentSettingsInput, DeploymentSettingsState],
) (infer.UpdateResponse[DeploymentSettingsState], error) {
	if req.DryRun {
		return infer.UpdateResponse[DeploymentSettingsState]{Output: DeploymentSettingsState(req.Inputs)}, nil
	}
	raw, ok := infer.GetRawRequest[p.UpdateRequest](ctx)
	if !ok {
		return infer.UpdateResponse[DeploymentSettingsState]{}, errors.New("DeploymentSettings.Update: no raw request")
	}
	input := req.Inputs.toAPI(secretEnvironmentVariables(resourceMap(raw.Inputs)))
	response, err := config.GetClient(ctx).UpdateDeploymentSettings(ctx, input.Stack, input.DeploymentSettings)
	if err != nil {
		return infer.UpdateResponse[DeploymentSettingsState]{}, err
	}
	state, err := deploymentSettingsState(req.Inputs, input, response)
	if err != nil {
		return infer.UpdateResponse[DeploymentSettingsState]{}, err
	}
	return infer.UpdateResponse[DeploymentSettingsState]{Output: state}, nil
}

// deploymentSettingsState builds the state from the settings Pulumi Cloud
// returned. Secret values in it are Pulumi Cloud's ciphertext.
func deploymentSettingsState(
	inputs DeploymentSettingsInput, input PulumiServiceDeploymentSettingsInput, response *pulumiapi.DeploymentSettings,
) (DeploymentSettingsState, error) {
	out := PulumiServiceDeploymentSettingsInput{DeploymentSettings: *response, Stack: input.Stack}
	state, err := fromPropertyMap[DeploymentSettingsState](out.ToPropertyMap(&input.DeploymentSettings, nil, false))
	if err != nil {
		return state, err
	}
	gcpOIDC(state.OperationContext).matchProjectKey(gcpOIDC(inputs.OperationContext))
	return state, nil
}

func gcpOIDC(oc *DeploymentSettingsOperationContext) *GCPOIDCConfiguration {
	if oc == nil || oc.OIDC == nil {
		return nil
	}
	return oc.OIDC.GCP
}

func (*DeploymentSettings) Read(
	ctx context.Context, req infer.ReadRequest[DeploymentSettingsInput, DeploymentSettingsState],
) (infer.ReadResponse[DeploymentSettingsInput, DeploymentSettingsState], error) {
	type response = infer.ReadResponse[DeploymentSettingsInput, DeploymentSettingsState]
	raw, ok := infer.GetRawRequest[p.ReadRequest](ctx)
	if !ok {
		return response{}, errors.New("DeploymentSettings.Read: no raw request")
	}
	stack, err := pulumiapi.NewStackIdentifier(req.ID)
	if err != nil {
		return response{}, err
	}
	settings, err := config.GetClient(ctx).GetDeploymentSettings(ctx, stack)
	if err != nil {
		return response{}, err
	}
	if settings == nil {
		// An empty ID removes the resource from the state.
		return response{}, nil
	}

	// On refresh, merge with the existing state: Pulumi Cloud returns only
	// ciphertext, so the plaintext of each secret comes from the old inputs.
	// On import there's no state, and secrets get a placeholder.
	var plaintext, ciphertext *pulumiapi.DeploymentSettings
	var usedGCP *GCPOIDCConfiguration
	props := resourceMap(raw.Properties)
	if props[gcStack].HasValue() {
		used, err := fromPropertyMap[DeploymentSettingsInput](resourceMap(raw.Inputs))
		if err != nil {
			return response{}, err
		}
		usedGCP = gcpOIDC(used.OperationContext)
		oldInputs, err := decodeDeploymentSettingsInput(resourceMap(raw.Inputs))
		if err != nil {
			return response{}, err
		}
		plaintext = &oldInputs.DeploymentSettings
		oldState, err := decodeDeploymentSettingsInput(props)
		if err != nil {
			return response{}, err
		}
		ciphertext = &oldState.DeploymentSettings
	}

	current := PulumiServiceDeploymentSettingsInput{Stack: stack, DeploymentSettings: *settings}
	state, err := fromPropertyMap[DeploymentSettingsState](current.ToPropertyMap(plaintext, ciphertext, false))
	if err != nil {
		return response{}, err
	}
	inputs, err := fromPropertyMap[DeploymentSettingsInput](current.ToPropertyMap(plaintext, ciphertext, true))
	if err != nil {
		return response{}, err
	}
	gcpOIDC(state.OperationContext).matchProjectKey(usedGCP)
	gcpOIDC(inputs.OperationContext).matchProjectKey(usedGCP)
	return response{ID: req.ID, Inputs: inputs, State: state}, nil
}

func (*DeploymentSettings) Delete(
	ctx context.Context, req infer.DeleteRequest[DeploymentSettingsState],
) (infer.DeleteResponse, error) {
	stack, err := pulumiapi.NewStackIdentifier(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteDeploymentSettings(ctx, stack)
}

func resourceMap(m property.Map) resource.PropertyMap {
	return resource.ToResourcePropertyValue(property.New(m)).ObjectValue()
}

// fromPropertyMap decodes a property map built by ToPropertyMap. infer
// re-applies secrets to the result where the request had them.
func fromPropertyMap[T any](pm resource.PropertyMap) (T, error) {
	var out T
	err := mapper.MapI(pm.MapRepl(nil, plainValue), &out)
	return out, err
}
