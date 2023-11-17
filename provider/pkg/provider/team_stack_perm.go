package provider

import (
	"context"

	"github.com/google/uuid"
	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type TeamStackPermissionResource struct {
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
	return serde.ToPropertyMap(*i, structTagKey)
}

func (tp *TeamStackPermissionResource) ToPulumiServiceTeamInput(inputMap resource.PropertyMap) (*TeamStackPermissionInput, error) {
	input := TeamStackPermissionInput{}
	return &input, serde.FromPropertyMap(inputMap, structTagKey, &input)
}

func (tp *TeamStackPermissionResource) Name() string {
	return "pulumiservice:index:TeamStackPermission"
}

func (tp *TeamStackPermissionResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{
		Inputs: req.GetNews(),
	}, nil
}

func (tp *TeamStackPermissionResource) Configure(config PulumiServiceConfig) {

}

func (tp *TeamStackPermissionResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return &pulumirpc.ReadResponse{}, nil
}

func (tp *TeamStackPermissionResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
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

	err = tp.client.AddStackPermission(ctx, stackName, inputs.Team, inputs.Permission)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         uuid.NewString(),
		Properties: req.GetProperties(),
	}, nil
}

func (tp *TeamStackPermissionResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
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
	err = tp.client.RemoveStackPermission(ctx, stackName, inputs.Team)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (tp *TeamStackPermissionResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	changedKeys, err := serde.DiffOldsAndNews(req)
	if err != nil {
		return nil, err
	}
	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(changedKeys) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:  changes,
		Replaces: changedKeys,
	}, nil
}

// Update does nothing because we always do a replace on changes, never an update
func (tp *TeamStackPermissionResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}
