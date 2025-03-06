package resources

import (
	"context"
	"fmt"
	"path"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceStackResource struct {
	Client *pulumiapi.Client
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
	return util.StandardDiff(req, []string{}, true)
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
	err = s.Client.DeleteStack(ctx, stack.StackIdentifier, stack.ForceDestroy)
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

func (s *PulumiServiceStackResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

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
