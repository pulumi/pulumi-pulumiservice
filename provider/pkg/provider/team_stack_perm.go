package provider

import (
	"context"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type TeamStackPermissionResource struct {
	config PulumiServiceConfig
	client *pulumiapi.Client
}

type TeamStackPermissionInput struct {
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Stack        string `pulumi:"stack"`
	Team         string `pulumi:"team"`
	Permission   int    `pulumi:"permission"`
}

func (i *TeamStackPermissionInput) ToPropertyMap() resource.PropertyMap {
	return serde.ToPropertyMap(i, structTagKey)
}

func (t *TeamStackPermissionResource) ToPulumiServiceTeamInput(inputMap resource.PropertyMap) (*TeamStackPermissionInput, error) {
	input := TeamStackPermissionInput{}
	return &input, serde.FromPropertyMap(inputMap, structTagKey, &input)
}

func (ts *TeamStackPermissionResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	var inputs TeamStackPermissionInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Stack,
	}

	err = ts.client.AddStackPermission(ctx, stackName, inputs.Team, inputs.Permission)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Properties: req.GetProperties(),
	}, nil
}

func (ts *TeamStackPermissionResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var inputs TeamStackPermissionInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackName{
		OrgName:     inputs.Organization,
		ProjectName: inputs.Project,
		StackName:   inputs.Stack,
	}
	err = ts.client.RemoveStackPermission(ctx, stackName, inputs.Team)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (ts *TeamStackPermissionResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	changedKeys, err := serde.DiffOldsAndNews(req) 
	if err != nil {
		return nil, err
	}
	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(changedKeys) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes: changes,
		Replaces: changedKeys,
	}, nil
}

func (ts *TeamStackPermissionResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}

