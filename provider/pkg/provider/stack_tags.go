package provider

import (
	"context"
	"fmt"
	"path"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceStackTagResource struct {
	client *pulumiapi.Client
}

type PulumiServiceStackTagInput struct {
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Stack        string `pulumi:"stack"`
	Name         string `pulumi:"name"`
	Value        string `pulumi:"value"`
}

const structTagKey = "pulumi" // could also be "json"

func (i *PulumiServiceStackTagInput) ToPropertyMap() resource.PropertyMap {
	return serde.ToPropertyMap(*i, structTagKey)
}

func (st *PulumiServiceStackTagResource) ToPulumiServiceStackTagInput(inputMap resource.PropertyMap) PulumiServiceStackTagInput {
	input := PulumiServiceStackTagInput{}
	serde.FromPropertyMap(inputMap, structTagKey, &input)
	return input
}

func (s *PulumiServiceStackTagResource) Name() string {
	return "pulumi-service:index:StackTag"
}

func (st *PulumiServiceStackTagResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	changed, err := serde.DiffOldsAndNews(req)
	if err != nil {
		return nil, err
	}
	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(changed) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes: changes,
	}, nil
}

func (st *PulumiServiceStackTagResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var inputs PulumiServiceStackTagInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Stack,
	}
	err = st.client.DeleteStackTag(ctx, stackName, inputs.Name)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (st *PulumiServiceStackTagResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	var inputs PulumiServiceStackTagInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Stack,
	}
	stackTag := pulumiapi.StackTag{
		Name:  inputs.Name,
		Value: inputs.Value,
	}
	err = st.client.CreateTag(ctx, stackName, stackTag)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{
		Id:         path.Join(inputs.Organization, inputs.Project, inputs.Stack, inputs.Name),
		Properties: req.GetProperties(),
	}, nil
}

func (st *PulumiServiceStackTagResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (st *PulumiServiceStackTagResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}

func (st *PulumiServiceStackTagResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	var inputs PulumiServiceStackTagInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Stack,
	}
	tag, err := st.client.GetStackTag(ctx, stackName, inputs.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to read StackTag (%q): %w", req.Id, err)
	}
	if tag == nil {
		// if the tag doesn't exist, then return empty response
		return &pulumirpc.ReadResponse{}, nil
	}
	inputs.Value = tag.Value
	props, err := serde.ToProperties(inputs, structTagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
	}, nil
}

func (st *PulumiServiceStackTagResource) Invoke(s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (st *PulumiServiceStackTagResource) Configure(config PulumiServiceConfig) {
}
