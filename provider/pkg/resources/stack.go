// Copyright 2016-2026, Pulumi Corporation.
package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type PulumiServiceStackResource struct {
	Client *pulumiapi.Client
}

// StackConfigEnvironment mirrors the schema type
// `pulumiservice:index:StackConfigEnvironment`. Either Project+Environment is
// set (link to an existing ESC env) or Auto is true (Stack manages a dedicated
// env named `<projectName>/<stackName>`); the two modes are mutually exclusive.
// Version pins a numeric revision or revision tag and is only valid with an
// existing env reference; the stack-create API rejects versioned auto refs.
type StackConfigEnvironment struct {
	Project     string
	Environment string
	Auto        bool
	Version     string
}

// IsSet reports whether the user provided any of the fields. The Stack-level
// optional input is treated as "absent" when none of the inner fields are set.
func (e *StackConfigEnvironment) IsSet() bool {
	if e == nil {
		return false
	}
	return e.Auto || e.Project != "" || e.Environment != "" || e.Version != ""
}

type PulumiServiceStack struct {
	pulumiapi.StackIdentifier
	ForceDestroy      bool
	ConfigEnvironment *StackConfigEnvironment
}

func (i *PulumiServiceStack) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organizationName"] = resource.NewPropertyValue(i.OrgName)
	pm["projectName"] = resource.NewPropertyValue(i.ProjectName)
	pm["stackName"] = resource.NewPropertyValue(i.StackName)
	if i.ForceDestroy {
		pm["forceDestroy"] = resource.NewPropertyValue(i.ForceDestroy)
	}
	if i.ConfigEnvironment.IsSet() {
		pm["configEnvironment"] = resource.NewObjectProperty(i.ConfigEnvironment.toPropertyMap())
	}
	return pm
}

func (e *StackConfigEnvironment) toPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	if e.Project != "" {
		pm["project"] = resource.NewPropertyValue(e.Project)
	}
	if e.Environment != "" {
		pm["environment"] = resource.NewPropertyValue(e.Environment)
	}
	if e.Auto {
		pm["auto"] = resource.NewPropertyValue(true)
	}
	if e.Version != "" {
		pm["version"] = resource.NewPropertyValue(e.Version)
	}
	return pm
}

// hasComputedIdentityField reports whether any field that determines the
// SBC env identity or mode (project, environment, auto) is unknown at preview
// time. Diff uses this to fall back to a conservative replace classification
// because parseStackConfigEnvironment skips computed values and would
// otherwise silently report no change.
func hasComputedIdentityField(v resource.PropertyValue) bool {
	if !v.HasValue() || !v.IsObject() {
		return false
	}
	obj := v.ObjectValue()
	for _, k := range []resource.PropertyKey{"project", "environment", "auto"} {
		if x, ok := obj[k]; ok && x.IsComputed() {
			return true
		}
	}
	return false
}

func parseStackConfigEnvironment(v resource.PropertyValue) *StackConfigEnvironment {
	if !v.HasValue() || !v.IsObject() {
		return nil
	}
	obj := v.ObjectValue()
	out := &StackConfigEnvironment{}
	if p, ok := obj["project"]; ok && p.IsString() {
		out.Project = p.StringValue()
	}
	if e, ok := obj["environment"]; ok && e.IsString() {
		out.Environment = e.StringValue()
	}
	if a, ok := obj["auto"]; ok && a.IsBool() {
		out.Auto = a.BoolValue()
	}
	if vv, ok := obj["version"]; ok && vv.IsString() {
		out.Version = vv.StringValue()
	}
	return out
}

// resolveEnvProject returns the ESC project name that should be persisted on
// the wire, defaulting to ESC's `default` project when the user omits it.
func (e *StackConfigEnvironment) resolveEnvProject() string {
	if e.Project != "" {
		return e.Project
	}
	return defaultProject
}

// envRefForCreate computes the wire-format env reference to send to the server
// at Create time. In auto mode, the server names the env `<projectName>/<stackName>`.
func (e *StackConfigEnvironment) envRefForCreate(stack pulumiapi.StackIdentifier) string {
	var project, name string
	if e.Auto {
		project = stack.ProjectName
		name = stack.StackName
	} else {
		project = e.resolveEnvProject()
		name = e.Environment
	}
	return pulumiapi.FormatEnvRef(project, name, e.Version)
}

func (s *PulumiServiceStackResource) ToPulumiServiceStackTagInput(
	inputMap resource.PropertyMap,
) (*PulumiServiceStack, error) {
	stack := PulumiServiceStack{}

	stack.OrgName = inputMap["organizationName"].StringValue()
	stack.ProjectName = inputMap["projectName"].StringValue()
	stack.StackName = inputMap["stackName"].StringValue()

	if inputMap["forceDestroy"].HasValue() && inputMap["forceDestroy"].IsBool() {
		stack.ForceDestroy = inputMap["forceDestroy"].BoolValue()
	}
	if v, ok := inputMap["configEnvironment"]; ok {
		if env := parseStackConfigEnvironment(v); env.IsSet() {
			stack.ConfigEnvironment = env
		}
	}
	return &stack, nil
}

func (s *PulumiServiceStackResource) Name() string {
	return "pulumiservice:index:Stack"
}

// configEnvDiffKind classifies how a change to configEnvironment should be
// applied: replace (any auto-mode change, or env identity change), update
// (toggle on/off in explicit-env mode, or a version-pin change), or none.
// The bool reports whether a change occurred so callers can suppress no-ops.
//
// Auto mode requires replace on every transition because (a) the server only
// proves Stack ownership of a managed env when it creates the env inline as
// part of POST /stacks, and (b) Delete uses Auto to decide preserveEnvironment.
// Toggling Auto in place would either silently adopt a pre-existing env we
// don't own (unsafe to clean up later with preserveEnvironment=false), or fail
// because the env doesn't exist yet.
func configEnvDiffKind(oldEnv, newEnv *StackConfigEnvironment) (kind plugin.DiffKind, changed bool) {
	oldAuto := oldEnv.IsSet() && oldEnv.Auto
	newAuto := newEnv.IsSet() && newEnv.Auto
	if oldAuto != newAuto {
		return plugin.DiffUpdateReplace, true
	}

	oldSet := oldEnv.IsSet()
	newSet := newEnv.IsSet()
	switch {
	case !oldSet && !newSet:
		return 0, false
	case !oldSet || !newSet:
		// Toggling explicit-env SBC on or off — link or unlink an env we don't
		// own — is a non-destructive PUT/DELETE on /config.
		return plugin.DiffUpdate, true
	}

	// Both sides set with matching auto flag: env identity is replace-only
	// because UpdateStackConfigHandler rejects re-linking to a different env.
	if oldEnv.resolveEnvProject() != newEnv.resolveEnvProject() ||
		oldEnv.Environment != newEnv.Environment {
		return plugin.DiffUpdateReplace, true
	}
	if oldEnv.Version != newEnv.Version {
		return plugin.DiffUpdate, true
	}
	return 0, false
}

func (s *PulumiServiceStackResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	// configEnvironment has subtree-aware semantics that the generic detailed
	// diff can't express, so we strip its raw entries and re-emit a single
	// configEnvironment-level diff with the right kind.
	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)
	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	deleteBeforeReplace := false
	const cfgPrefix = "configEnvironment."
	for k, v := range dd {
		if k == "configEnvironment" || strings.HasPrefix(k, cfgPrefix) {
			continue
		}
		// Existing behavior: every other input (org/project/stack/forceDestroy)
		// forces a replacement.
		v.Kind = v.Kind.AsReplace()
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec // safe conversion from plugin.DiffKind
			InputDiff: v.InputDiff,
		}
		deleteBeforeReplace = true
	}

	oldEnv := parseStackConfigEnvironment(olds["configEnvironment"])
	newEnv := parseStackConfigEnvironment(news["configEnvironment"])
	switch {
	case hasComputedIdentityField(news["configEnvironment"]):
		// Any of `project`, `environment`, `auto` arrived as an Output and won't
		// resolve until apply. parseStackConfigEnvironment can't see the value,
		// so a non-conservative classification here would let the user approve
		// a plan that says "in-place update" while the apply actually requires
		// a destructive replace (e.g. the env name resolves to a different
		// env, or auto resolves to a flipped mode). Surface the worst case at
		// preview so the plan reflects what apply may need to do.
		detailedDiffs["configEnvironment"] = &pulumirpc.PropertyDiff{
			Kind: pulumirpc.PropertyDiff_Kind(plugin.DiffUpdateReplace), //nolint:gosec
		}
		deleteBeforeReplace = true
	default:
		if kind, changed := configEnvDiffKind(oldEnv, newEnv); changed {
			detailedDiffs["configEnvironment"] = &pulumirpc.PropertyDiff{
				Kind: pulumirpc.PropertyDiff_Kind(kind), //nolint:gosec // safe conversion from plugin.DiffKind
			}
			if kind == plugin.DiffAddReplace || kind == plugin.DiffDeleteReplace || kind == plugin.DiffUpdateReplace {
				deleteBeforeReplace = true
			}
		}
	}

	if len(detailedDiffs) == 0 {
		return &pulumirpc.DiffResponse{Changes: pulumirpc.DiffResponse_DIFF_NONE}, nil
	}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		DetailedDiff:        detailedDiffs,
		DeleteBeforeReplace: deleteBeforeReplace,
		HasDetailedDiff:     true,
	}, nil
}

func (s *PulumiServiceStackResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	stack, err := s.ToPulumiServiceStackTagInput(inputs)
	if err != nil {
		return nil, err
	}
	// preserveEnvironment defaults to true (PSP doesn't own the env's
	// lifecycle), and flips to false only when this stack created the env in
	// `auto` mode — we then want the server to clean it up alongside us.
	preserveEnvironment := !stack.ConfigEnvironment.IsSet() || !stack.ConfigEnvironment.Auto
	err = s.Client.DeleteStack(ctx, stack.StackIdentifier, stack.ForceDestroy, preserveEnvironment)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (s *PulumiServiceStackResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	stack, err := s.ToPulumiServiceStackTagInput(inputs)
	if err != nil {
		return nil, err
	}

	var createConfig *pulumiapi.StackConfig
	if stack.ConfigEnvironment.IsSet() && stack.ConfigEnvironment.Auto {
		// Auto mode: the server creates the env (and optionally pre-populates
		// template YAML) inline as part of stack creation, so we send the env
		// ref in the same POST.
		createConfig = &pulumiapi.StackConfig{
			Environment: stack.ConfigEnvironment.envRefForCreate(stack.StackIdentifier),
		}
	}

	if err := s.Client.CreateStack(ctx, stack.StackIdentifier, createConfig); err != nil {
		return nil, err
	}

	if stack.ConfigEnvironment.IsSet() && !stack.ConfigEnvironment.Auto {
		// Reference-existing-env mode: link via PUT /config after the stack
		// exists. Done as a separate call because the inline POST flow tries
		// to *create* the env and would 409 against an env we don't own.
		//
		// Two-call create is non-atomic, so on link failure we best-effort
		// roll back the just-created stack with preserveEnvironment=true (we
		// never own the env in this mode). Without rollback a retry would
		// fail with already-exists, leaving an unmanaged stack behind.
		envRef := stack.ConfigEnvironment.envRefForCreate(stack.StackIdentifier)
		linkErr := s.Client.SetStackConfig(ctx, stack.StackIdentifier, pulumiapi.StackConfig{
			Environment: envRef,
		})
		if linkErr != nil {
			rollbackErr := s.Client.DeleteStack(ctx, stack.StackIdentifier, false /*forceDestroy*/, true /*preserveEnvironment*/)
			if rollbackErr != nil {
				return nil, fmt.Errorf(
					"failed to link stack %s to ESC environment %q: %w; "+
						"rollback (delete stack) also failed: %v. Manual cleanup may be required "+
						"in Pulumi Cloud before retrying",
					stack.StackIdentifier, envRef, linkErr, rollbackErr,
				)
			}
			return nil, fmt.Errorf(
				"failed to link stack %s to ESC environment %q (rolled back): %w",
				stack.StackIdentifier, envRef, linkErr,
			)
		}
	}

	outputProperties, err := plugin.MarshalProperties(
		stack.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         path.Join(stack.OrgName, stack.ProjectName, stack.StackName),
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceStackResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	// Validate configEnvironment shape, but only on concrete inputs. At preview
	// time fields can arrive as computed Output<T>; firing on those would
	// reject any program that wires `environment: env.name` or similar. We
	// flag known-bad shapes and otherwise defer until apply, when computed
	// values resolve.
	var failures []*pulumirpc.CheckFailure
	if v, ok := news["configEnvironment"]; ok && v.HasValue() && v.IsObject() {
		obj := v.ObjectValue()

		// known-true / known-false / unknown — collapse computed and missing
		// values into "unknown" so validation defers in either case.
		type tri int
		const (
			triUnknown tri = iota
			triFalse
			triTrue
		)
		// Missing keys default to false/empty (not unknown). Only an explicit
		// computed value defers the check — that's the case where the user
		// wired the field to another resource's Output<T>.
		boolField := func(key resource.PropertyKey) tri {
			x, ok := obj[key]
			if !ok || !x.HasValue() {
				return triFalse
			}
			if x.IsComputed() {
				return triUnknown
			}
			if x.IsBool() && x.BoolValue() {
				return triTrue
			}
			return triFalse
		}
		stringSet := func(key resource.PropertyKey) tri {
			x, ok := obj[key]
			if !ok || !x.HasValue() {
				return triFalse
			}
			if x.IsComputed() {
				return triUnknown
			}
			if x.IsString() && x.StringValue() != "" {
				return triTrue
			}
			return triFalse
		}

		auto := boolField("auto")
		env := stringSet("environment")
		project := stringSet("project")
		version := stringSet("version")

		// Mutex check: only fire when we know both sides positively.
		if auto == triTrue && (env == triTrue || project == triTrue) {
			failures = append(failures, &pulumirpc.CheckFailure{
				Property: "configEnvironment",
				Reason: "configEnvironment.auto cannot be combined with configEnvironment.environment " +
					"or configEnvironment.project; choose either auto-managed env or an existing env reference",
			})
		}
		if auto == triTrue && version == triTrue {
			failures = append(failures, &pulumirpc.CheckFailure{
				Property: "configEnvironment",
				Reason: "configEnvironment.version cannot be combined with configEnvironment.auto; " +
					"pin a version only when linking an existing environment",
			})
		}
		// Required check: only fire when we know auto is false AND we know
		// environment is empty/missing. If either is unknown, defer.
		if auto == triFalse && env == triFalse {
			failures = append(failures, &pulumirpc.CheckFailure{
				Property: "configEnvironment",
				Reason: "configEnvironment requires either configEnvironment.auto=true (Stack-managed env) " +
					"or configEnvironment.environment (existing env name)",
			})
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (s *PulumiServiceStackResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	olds, err := plugin.UnmarshalProperties(
		req.GetOlds(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}
	news, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	stack, err := s.ToPulumiServiceStackTagInput(news)
	if err != nil {
		return nil, err
	}
	oldStack, err := s.ToPulumiServiceStackTagInput(olds)
	if err != nil {
		return nil, err
	}

	// Update only handles configEnvironment changes that don't force a
	// replacement; Diff is responsible for funneling everything else through
	// Replace, so any other observed delta here is a programmer error.
	if stack.OrgName != oldStack.OrgName ||
		stack.ProjectName != oldStack.ProjectName ||
		stack.StackName != oldStack.StackName ||
		stack.ForceDestroy != oldStack.ForceDestroy {
		return nil, fmt.Errorf("unexpected stack identity change in update; expected replace")
	}

	oldEnv := oldStack.ConfigEnvironment
	newEnv := stack.ConfigEnvironment
	// Diff funnels every change involving Auto through Replace, so seeing it
	// here means the diff classification drifted from the create/delete logic.
	// Refusing the update is safer than silently adopting an env we don't own.
	if (oldEnv.IsSet() && oldEnv.Auto) != (newEnv.IsSet() && newEnv.Auto) {
		return nil, fmt.Errorf("unexpected configEnvironment.auto toggle in update; expected replace")
	}
	// Check rejects auto+version on the way in, but state migrations and
	// imports can bypass Check. The server's stack-create API rejects versioned
	// auto refs and SetStackConfig would either fail noisily or silently pin a
	// version on an env Pulumi expects to be unversioned-current — neither is
	// a state we want to reach.
	if newEnv.IsSet() && newEnv.Auto && newEnv.Version != "" {
		return nil, fmt.Errorf(
			"configEnvironment.version is not allowed with configEnvironment.auto=true; " +
				"remove the version pin or switch to an explicit env reference")
	}
	switch {
	case oldEnv.IsSet() && !newEnv.IsSet():
		if err := s.Client.DeleteStackConfig(ctx, stack.StackIdentifier); err != nil {
			return nil, err
		}
	case newEnv.IsSet():
		// Either toggling an explicit-env link on (old=nil, new=env) or a
		// version-pin change in either mode. The server PUT accepts both.
		if err := s.Client.SetStackConfig(ctx, stack.StackIdentifier, pulumiapi.StackConfig{
			Environment: newEnv.envRefForCreate(stack.StackIdentifier),
		}); err != nil {
			return nil, err
		}
	}

	outputProperties, err := plugin.MarshalProperties(
		stack.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{Properties: outputProperties}, nil
}

func (s *PulumiServiceStackResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	stack, err := pulumiapi.NewStackIdentifier(req.GetId())
	if err != nil {
		return nil, err
	}
	exists, err := s.Client.StackExists(ctx, stack)
	if err != nil {
		return nil, fmt.Errorf("failure while checking if stack %q exists: %w", req.Id, err)
	}
	if !exists {
		return &pulumirpc.ReadResponse{}, nil
	}

	props := PulumiServiceStack{
		StackIdentifier: stack,
	}

	// Surface old inputs so we can preserve user intent that the server
	// doesn't store: specifically `auto` vs. an explicit env reference.
	var oldInputs resource.PropertyMap
	if req.GetInputs() != nil {
		oldInputs, err = plugin.UnmarshalProperties(
			req.GetInputs(),
			plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
		)
		if err != nil {
			return nil, err
		}
		if v, ok := oldInputs["forceDestroy"]; ok && v.HasValue() && v.IsBool() {
			props.ForceDestroy = v.BoolValue()
		}
	}

	cfg, err := s.Client.GetStackConfig(ctx, stack)
	if err != nil {
		return nil, fmt.Errorf("failure while getting stack config %q: %w", req.Id, err)
	}
	if cfg != nil && cfg.Environment != "" {
		envProj, envName, envVer := pulumiapi.ParseEnvRef(cfg.Environment)
		props.ConfigEnvironment = stackConfigEnvironmentFromServer(
			envProj, envName, envVer, stack, parseStackConfigEnvironment(oldInputs["configEnvironment"]),
		)
	}

	pm := props.ToPropertyMap()
	outputs, err := plugin.MarshalProperties(pm, plugin.MarshalOptions{})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputs,
		Inputs:     outputs,
	}, nil
}

// stackConfigEnvironmentFromServer reconstructs the input shape from the
// server's env ref. When the prior input declared `auto: true` and the server
// still reports the auto-form name (`<project>/<stack>`), we keep the input as
// `auto: true`; otherwise we fall back to the explicit project/environment
// form so refresh/import paths produce a stable shape.
func stackConfigEnvironmentFromServer(
	envProj, envName, envVer string,
	stack pulumiapi.StackIdentifier,
	prior *StackConfigEnvironment,
) *StackConfigEnvironment {
	autoForm := envProj == stack.ProjectName && envName == stack.StackName
	if prior.IsSet() && prior.Auto && autoForm {
		return &StackConfigEnvironment{Auto: true, Version: envVer}
	}
	return &StackConfigEnvironment{Project: envProj, Environment: envName, Version: envVer}
}
