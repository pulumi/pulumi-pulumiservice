package resources

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type TeamStackPermissionResource struct {
	Client pulumiapi.TeamClient
}

type TeamStackPermissionInput struct {
	Organization string `pulumi:"organization"`
	Project      string `pulumi:"project"`
	Stack        string `pulumi:"stack"`
	Team         string `pulumi:"team"`
	Permission   int    `pulumi:"permission"`
}

func (i *TeamStackPermissionInput) ToPropertyMap() resource.PropertyMap {
	return util.ToPropertyMap(*i, structTagKey)
}

func (tp *TeamStackPermissionResource) ToPulumiServiceTeamInput(inputMap resource.PropertyMap) (*TeamStackPermissionInput, error) {
	input := TeamStackPermissionInput{}
	return &input, util.FromPropertyMap(inputMap, structTagKey, &input)
}

func (tp *TeamStackPermissionResource) Name() string {
	return "pulumiservice:index:TeamStackPermission"
}

func (tp *TeamStackPermissionResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{
		Inputs: req.GetNews(),
	}, nil
}

func (tp *TeamStackPermissionResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	id := req.GetId()

	permId, err := splitTeamStackPermissionId(id)
	if err != nil {
		if strings.Contains(err.Error(), "expected 4 parts") {
			// Return an error if attempting to refresh stack permissions created before this change.
			// We return a warning and an empty response, which will cause the resource to be deleted on refresh,
			// forcing the user to recreate it with the updated version.
			return nil, fmt.Errorf("TeamStackPermission resources created before v0.17.0 do not support refresh. " +
				"You will need to destroy and recreate this resource with >v0.17.0 to successfully refresh.")
		}
		return nil, err
	}

	permission, err := tp.Client.GetTeamStackPermission(ctx, pulumiapi.StackIdentifier{
		OrgName:     permId.Organization,
		ProjectName: permId.Project,
		StackName:   permId.Stack,
	}, permId.Team)
	if err != nil {
		return nil, fmt.Errorf("failed to get team stack permission: %w", err)
	}
	if permission == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	inputs := TeamStackPermissionInput{
		Organization: permId.Organization,
		Project:      permId.Project,
		Stack:        permId.Stack,
		Team:         permId.Team,
		Permission:   *permission,
	}

	properties, err := plugin.MarshalProperties(inputs.ToPropertyMap(), plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     properties,
	}, nil
}

func (tp *TeamStackPermissionResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	var input TeamStackPermissionInput
	err := util.FromProperties(req.GetProperties(), structTagKey, &input)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackIdentifier{
		OrgName:     input.Organization,
		ProjectName: input.Project,
		StackName:   input.Stack,
	}

	err = tp.Client.AddStackPermission(ctx, stackName, input.Team, input.Permission)
	if err != nil {
		return nil, err
	}

	stackPermissionId := fmt.Sprintf("%s/%s", stackName.String(), input.Team)

	return &pulumirpc.CreateResponse{
		Id:         stackPermissionId,
		Properties: req.GetProperties(),
	}, nil
}

func (tp *TeamStackPermissionResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var input TeamStackPermissionInput
	err := util.FromProperties(req.GetProperties(), structTagKey, &input)
	if err != nil {
		return nil, err
	}
	stackName := pulumiapi.StackIdentifier{
		OrgName:     input.Organization,
		ProjectName: input.Project,
		StackName:   input.Stack,
	}
	err = tp.Client.RemoveStackPermission(ctx, stackName, input.Team)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (tp *TeamStackPermissionResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	changedKeys, err := util.DiffOldsAndNews(req)
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

// Update does nothing because we always replace on changes, never an update
func (tp *TeamStackPermissionResource) Update(_ *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

type teamStackPermissionId struct {
	Organization string
	Project      string
	Stack        string
	Team         string
}

func splitTeamStackPermissionId(id string) (teamStackPermissionId, error) {
	split := strings.Split(id, "/")
	if len(split) != 4 {
		return teamStackPermissionId{}, fmt.Errorf("invalid id %q, expected 4 parts", id)
	}
	return teamStackPermissionId{
		Organization: split[0],
		Project:      split[1],
		Stack:        split[2],
		Team:         split[3],
	}, nil
}
