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
// `pulumiservice:index:StackConfigEnvironment`. Managed=true opts the stack
// into a dedicated ESC environment that the server creates inline at stack
// creation and deletes alongside the stack.
type StackConfigEnvironment struct {
	Managed bool
}

func (e *StackConfigEnvironment) IsSet() bool {
	return e != nil && e.Managed
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
	if e.Managed {
		pm["managed"] = resource.NewPropertyValue(true)
	}
	return pm
}

// hasComputedManaged reports whether configEnvironment.managed is unknown at
// preview. Diff falls back to a conservative replace when it is, because the
// resolved value determines whether the server must create an env inline (a
// destructive transition that an in-place plan would hide).
func hasComputedManaged(v resource.PropertyValue) bool {
	if !v.HasValue() || !v.IsObject() {
		return false
	}
	obj := v.ObjectValue()
	x, ok := obj["managed"]
	return ok && x.IsComputed()
}

func parseStackConfigEnvironment(v resource.PropertyValue) *StackConfigEnvironment {
	if !v.HasValue() || !v.IsObject() {
		return nil
	}
	obj := v.ObjectValue()
	out := &StackConfigEnvironment{}
	if a, ok := obj["managed"]; ok && a.IsBool() {
		out.Managed = a.BoolValue()
	}
	return out
}

// envRefForCreate computes the wire-format env reference for the inline POST
// that creates both the stack and its managed env. The server names the env
// `<projectName>/<stackName>`.
func envRefForCreate(stack pulumiapi.StackIdentifier) string {
	return pulumiapi.FormatEnvRef(stack.ProjectName, stack.StackName, "")
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

	// Strip configEnvironment's raw entries and re-emit a single
	// configEnvironment-level diff. Every other input forces replacement.
	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)
	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	deleteBeforeReplace := false
	const cfgPrefix = "configEnvironment."
	for k, v := range dd {
		if k == "configEnvironment" || strings.HasPrefix(k, cfgPrefix) {
			continue
		}
		v.Kind = v.Kind.AsReplace()
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec
			InputDiff: v.InputDiff,
		}
		deleteBeforeReplace = true
	}

	oldEnv := parseStackConfigEnvironment(olds["configEnvironment"])
	newEnv := parseStackConfigEnvironment(news["configEnvironment"])
	// Any managed transition (on or off, or unknown at preview) is a replace:
	// the server only proves stack ownership of a managed env when it creates
	// the env inline as part of POST /stacks, and Delete uses Managed to
	// decide preserveEnvironment. Toggling in place would either silently
	// adopt a pre-existing env we don't own or fail because the env doesn't
	// exist yet.
	if hasComputedManaged(news["configEnvironment"]) || oldEnv.IsSet() != newEnv.IsSet() {
		detailedDiffs["configEnvironment"] = &pulumirpc.PropertyDiff{
			Kind: pulumirpc.PropertyDiff_Kind(plugin.DiffUpdateReplace), //nolint:gosec
		}
		deleteBeforeReplace = true
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
	// Tell the server to clean up the env alongside the stack only when this
	// stack owns it (managed mode).
	preserveEnvironment := !stack.ConfigEnvironment.IsSet()
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
	if stack.ConfigEnvironment.IsSet() {
		createConfig = &pulumiapi.StackConfig{
			Environment: envRefForCreate(stack.StackIdentifier),
		}
	}

	if err := s.Client.CreateStack(ctx, stack.StackIdentifier, createConfig); err != nil {
		return nil, err
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
	return &pulumirpc.CheckResponse{Inputs: req.News}, nil
}

func (s *PulumiServiceStackResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// Every stack input is replace-on-change, so Update is never asked to
	// reconcile anything. Re-emit the new inputs as outputs.
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

	if req.GetInputs() != nil {
		oldInputs, err := plugin.UnmarshalProperties(
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
		envProj, envName, _ := pulumiapi.ParseEnvRef(cfg.Environment)
		// Only managed envs are first-class in the schema; an env linked to
		// this stack whose name doesn't match the managed form was created
		// outside Pulumi (or under a prior schema) and isn't representable.
		if envProj == stack.ProjectName && envName == stack.StackName {
			props.ConfigEnvironment = &StackConfigEnvironment{Managed: true}
		}
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
