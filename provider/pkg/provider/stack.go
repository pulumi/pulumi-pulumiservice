package provider

import (
	"context"
	"fmt"
	"path"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceStackResource struct {
	client *pulumiapi.Client
}

type PulumiServiceStack struct {
	pulumiapi.StackIdentifier
	ForceDestroy bool
}

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

func (s *PulumiServiceStackResource) ToPulumiServiceStackTagInput(inputMap resource.PropertyMap) (*PulumiServiceStack, error) {
	stack := PulumiServiceStack{}

	stack.StackIdentifier.OrgName = inputMap["organizationName"].StringValue()
	stack.StackIdentifier.ProjectName = inputMap["projectName"].StringValue()
	stack.StackIdentifier.StackName = inputMap["stackName"].StringValue()

	if inputMap["forceDestroy"].HasValue() && inputMap["forceDestroy"].IsBool() {
		stack.ForceDestroy = inputMap["forceDestroy"].BoolValue()
	}
	return &stack, nil
}

func (s *PulumiServiceStackResource) Name() string {
	return "pulumiservice:index:Stack"
}

func (s *PulumiServiceStackResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
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
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
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

func (s *PulumiServiceStackResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	stack, err := s.ToPulumiServiceStackTagInput(inputs)
	if err != nil {
		return nil, err
	}
	err = s.client.DeleteStack(ctx, stack.StackIdentifier, stack.ForceDestroy)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (s *PulumiServiceStackResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	stack, err := s.ToPulumiServiceStackTagInput(inputs)
	if err != nil {
		return nil, err
	}
	err = s.client.CreateStack(ctx, stack.StackIdentifier)
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

func (s *PulumiServiceStackResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (s *PulumiServiceStackResource) Update(_ *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (s *PulumiServiceStackResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	var inputs PulumiServiceStack
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	exists, err := s.client.StackExists(ctx, inputs.StackIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failure while checking if stack %q exists: %w", req.Id, err)
	}
	if !exists {
		return &pulumirpc.ReadResponse{
			Id:         "",
			Properties: nil,
		}, nil
	}
	props, err := serde.ToProperties(inputs, structTagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}

func (s *PulumiServiceStackResource) Configure(_ PulumiServiceConfig) {
}
