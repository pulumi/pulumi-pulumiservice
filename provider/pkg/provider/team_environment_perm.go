package provider

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTeamEnvironmentPermissionResource struct {
	client pulumiapi.TeamClient
}

type TeamEnvironmentPermissionInput struct {
	Organization string `pulumi:"organization"`
	Team         string `pulumi:"team"`
	Environment  string `pulumi:"environment"`
	Permission   string `pulumi:"permission"`
}

func (i *TeamEnvironmentPermissionInput) ToPropertyMap() resource.PropertyMap {
	return serde.ToPropertyMap(*i, structTagKey)
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) ToPulumiServiceTeamInput(inputMap resource.PropertyMap) (*TeamEnvironmentPermissionInput, error) {
	input := TeamEnvironmentPermissionInput{}
	return &input, serde.FromPropertyMap(inputMap, structTagKey, &input)
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Name() string {
	return "pulumiservice:index:TeamEnvironmentPermission"
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{
		Inputs: req.GetNews(),
	}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Configure(_ PulumiServiceConfig) {

}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	permId, err := splitTeamEnvironmentPermissionId(req.GetId())
	if err != nil {
		return nil, err
	}

	request := pulumiapi.TeamEnvironmentPermissionRequest{
		Organization: permId.Organization,
		Team:         permId.Team,
		Environment:  permId.Environment,
	}
	permission, err := tp.client.GetTeamEnvironmentPermission(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get team environment permission: %w", err)
	}
	if permission == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	inputs := TeamEnvironmentPermissionInput{
		Organization: permId.Organization,
		Team:         permId.Team,
		Environment:  permId.Environment,
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

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	var inputs TeamEnvironmentPermissionInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	request := pulumiapi.CreateTeamEnvironmentPermissionRequest{
		TeamEnvironmentPermissionRequest: pulumiapi.TeamEnvironmentPermissionRequest{
			Organization: inputs.Organization,
			Team:         inputs.Team,
			Environment:  inputs.Environment,
		},
		Permission: inputs.Permission,
	}

	err = tp.client.AddEnvironmentPermission(ctx, request)
	if err != nil {
		return nil, err
	}

	environmentPermissionId := teamEnvironmentPermissionId{
		Organization: inputs.Organization,
		Team:         inputs.Team,
		Environment:  inputs.Environment,
	}

	return &pulumirpc.CreateResponse{
		Id:         environmentPermissionId.String(),
		Properties: req.GetProperties(),
	}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var inputs TeamEnvironmentPermissionInput
	err := serde.FromProperties(req.GetProperties(), structTagKey, &inputs)
	if err != nil {
		return nil, err
	}
	request := pulumiapi.TeamEnvironmentPermissionRequest{
		Organization: inputs.Organization,
		Team:         inputs.Team,
		Environment:  inputs.Environment,
	}
	err = tp.client.RemoveEnvironmentPermission(ctx, request)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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

// Update does nothing because we always replace on changes, never an update
func (tp *PulumiServiceTeamEnvironmentPermissionResource) Update(_ *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

type teamEnvironmentPermissionId struct {
	Organization string
	Team         string
	Environment  string
}

func (s *teamEnvironmentPermissionId) String() string {
	return fmt.Sprintf("%s/%s/%s", s.Organization, s.Team, s.Environment)
}

func splitTeamEnvironmentPermissionId(id string) (teamEnvironmentPermissionId, error) {
	split := strings.Split(id, "/")
	if len(split) != 3 {
		return teamEnvironmentPermissionId{}, fmt.Errorf("invalid id %q, expected 3 parts", id)
	}
	return teamEnvironmentPermissionId{
		Organization: split[0],
		Team:         split[1],
		Environment:  split[2],
	}, nil
}
