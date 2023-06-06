package provider

import (
	"context"
	"fmt"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceStackResource struct {
	client *pulumiapi.Client
}

type PulumiServiceStackInput struct {
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Name         string `pulumi:"name"`
}

func (i *PulumiServiceStackInput) ToPropertyMap() resource.PropertyMap {
	return serde.ToPropertyMap(*i, structTagKey)
}

func (st *PulumiServiceStackResource) ToPulumiServiceStackInput(inputMap resource.PropertyMap) PulumiServiceStackInput {
	input := PulumiServiceStackInput{}
	serde.FromPropertyMap(inputMap, structTagKey, &input)
	return input
}

func (st *PulumiServiceStackResource) Name() string {
	return "pulumiservice:index:Stack"
}

func (st *PulumiServiceStackResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "project", "name"} {
		if !news[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (st *PulumiServiceStackResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}
	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	for k, v := range dd {
		v.Kind = v.Kind.AsReplace()
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	return &pulumirpc.DiffResponse{
		Changes:         pulumirpc.DiffResponse_DIFF_SOME,
		DetailedDiff:    detailedDiffs,
		HasDetailedDiff: true,
	}, nil
}

func (st *PulumiServiceStackResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()

	var inputs PulumiServiceStackInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Name,
	}
	err = st.client.CreateStack(ctx, stackName)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{
		Id:         stackName.String(),
		Properties: req.GetProperties(),
	}, nil
}

func (st *PulumiServiceStackResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	var stackName pulumiapi.StackName
	if err := stackName.FromID(req.GetId()); err != nil {
		return nil, err
	}

	stack, err := st.client.GetStack(ctx, stackName)
	if err != nil {
		return nil, fmt.Errorf("failed to read Stack (%q): %w", req.Id, err)
	}
	if stack == nil {
		// if the tag doesn't exist, then return empty response
		return &pulumirpc.ReadResponse{}, nil
	}

	inputs := PulumiServiceStackInput{
		Organization: stackName.OrgName,
		Project:      stackName.ProjectName,
		Name:         stackName.StackName,
	}
	props, err := serde.ToProperties(inputs, structTagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Inputs:     props,
		Properties: props,
	}, nil
}

func (st *PulumiServiceStackResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (st *PulumiServiceStackResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	var stackName pulumiapi.StackName
	if err := stackName.FromID(req.GetId()); err != nil {
		return nil, err
	}

	err := st.client.DeleteStack(ctx, stackName)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (st *PulumiServiceStackResource) Invoke(s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (st *PulumiServiceStackResource) Configure(config PulumiServiceConfig) {
}
