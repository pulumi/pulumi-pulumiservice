package resources

import (
	"context"
	"fmt"
	"path"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// PulumiServiceStackResource manages Pulumi Service stack resources.
type PulumiServiceStackResource struct {
	Client *pulumiapi.Client
}

// PulumiServiceStack represents a Pulumi Service stack with its properties.
type PulumiServiceStack struct {
	pulumiapi.StackIdentifier
	ForceDestroy bool
}

// ToPropertyMap converts the stack to a property map.
func (i *PulumiServiceStack) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organizationName"] = resource.NewPropertyValue(i.OrgName)
	pm["projectName"] = resource.NewPropertyValue(i.ProjectName)
	pm["stackName"] = resource.NewPropertyValue(i.StackName)
	if i.ForceDestroy {
		pm["forceDestroy"] = resource.NewPropertyValue(i.ForceDestroy)
	}
	return pm
}

// ToPulumiServiceStackTagInput converts a property map to a PulumiServiceStack input.
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
	return &stack, nil
}

// Name returns the resource type name.
func (s *PulumiServiceStackResource) Name() string {
	return "pulumiservice:index:Stack"
}

// Diff checks what impacts a hypothetical update will have on the stack resource.
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

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	for k, v := range dd {
		v.Kind = v.Kind.AsReplace()
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec // Kind values are bounded by protobuf enum
			InputDiff: v.InputDiff,
		}
	}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		DetailedDiff:        detailedDiffs,
		DeleteBeforeReplace: true,
		HasDetailedDiff:     true,
	}, nil
}

// Delete tears down an existing stack resource.
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
	err = s.Client.DeleteStack(ctx, stack.StackIdentifier, stack.ForceDestroy)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

// Create allocates a new stack instance and returns its unique ID.
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
	err = s.Client.CreateStack(ctx, stack.StackIdentifier)
	if err != nil {
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

// Check validates that the given property bag is valid for a stack resource.
func (s *PulumiServiceStackResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

// Update updates an existing stack resource with new values.
func (s *PulumiServiceStackResource) Update(_ *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
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

	outputs, err := plugin.MarshalProperties(
		props.ToPropertyMap(),
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputs,
		Inputs:     outputs,
	}, nil
}
