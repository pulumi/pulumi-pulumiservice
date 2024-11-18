package provider

import (
	"context"
	"fmt"
	"path"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
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
	_ = serde.FromPropertyMap(inputMap, structTagKey, &input)
	return input
}

func (st *PulumiServiceStackTagResource) Name() string {
	return "pulumiservice:index:StackTag"
}

func (st *PulumiServiceStackTagResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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

func (st *PulumiServiceStackTagResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var inputs PulumiServiceStackTagInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackIdentifier{
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
	stackName := pulumiapi.StackIdentifier{
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
	// all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (st *PulumiServiceStackTagResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	organization, project, stack, tagName, err := splitStackTagId(req.Id)
	if err != nil {
		return nil, err
	}

	stackName := pulumiapi.StackIdentifier{
		OrgName:     organization,
		ProjectName: project,
		StackName:   stack,
	}
	tag, err := st.client.GetStackTag(ctx, stackName, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to read StackTag (%q): %w", req.Id, err)
	}
	if tag == nil {
		// if the tag doesn't exist, then return empty response
		return &pulumirpc.ReadResponse{}, nil
	}

	input := PulumiServiceStackTagInput{
		Organization: organization,
		Project:      project,
		Stack:        stack,
		Name:         tag.Name,
		Value:        tag.Value,
	}
	props, err := serde.ToProperties(input, structTagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}

func (st *PulumiServiceStackTagResource) Configure(_ PulumiServiceConfig) {
}

func splitStackTagId(id string) (organization string, project string, stack string, tagName string, err error) {
	// format: organization/project/stack/tagName
	s := strings.Split(id, "/")
	if len(s) != 4 {
		return "", "", "", "", fmt.Errorf("%q is invalid, must be in organization/project/stack/tagName format", id)
	}
	return s[0], s[1], s[2], s[3], nil
}
