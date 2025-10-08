package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// PulumiServiceTeamEnvironmentPermissionResource manages team environment permission resources.
type PulumiServiceTeamEnvironmentPermissionResource struct {
	Client pulumiapi.TeamClient
}

type TeamEnvironmentPermissionInput struct {
	Organization    string `pulumi:"organization"`
	Team            string `pulumi:"team"`
	Environment     string `pulumi:"environment"`
	Project         string `pulumi:"project"`
	Permission      string `pulumi:"permission"`
	MaxOpenDuration string `pulumi:"maxOpenDuration"`
}

func (i *TeamEnvironmentPermissionInput) ToPropertyMap() resource.PropertyMap {
	return util.ToPropertyMap(*i, structTagKey)
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) ToPulumiServiceTeamInput(
	inputMap resource.PropertyMap,
) (*TeamEnvironmentPermissionInput, error) {
	input := TeamEnvironmentPermissionInput{}
	return &input, util.FromPropertyMap(inputMap, structTagKey, &input)
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Name() string {
	return "pulumiservice:index:TeamEnvironmentPermission"
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Check(
	req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	var input TeamEnvironmentPermissionInput
	err := util.FromProperties(req.GetNews(), structTagKey, &input)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	if input.MaxOpenDuration != "" {
		d, err := time.ParseDuration(input.MaxOpenDuration)
		if err != nil {
			failures = append(failures, &pulumirpc.CheckFailure{
				Property: "maxOpenDuration",
				Reason:   fmt.Sprintf("malformed duration: %v", err),
			})
		} else {
			// Normalize the duration to prevent spurious diffs.
			input.MaxOpenDuration = d.String()
		}
	}

	inputs, err := util.ToProperties(input, structTagKey)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{
		Inputs:   inputs,
		Failures: failures,
	}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Read(
	req *pulumirpc.ReadRequest,
) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	permID, err := splitTeamEnvironmentPermissionID(req.GetId())
	if err != nil {
		return nil, err
	}

	request := pulumiapi.TeamEnvironmentSettingsRequest{
		Organization: permID.Organization,
		Team:         permID.Team,
		Environment:  permID.Environment,
		Project:      permID.Project,
	}
	permission, maxOpenDuration, err := tp.Client.GetTeamEnvironmentSettings(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get team environment permission: %w", err)
	}
	if permission == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	var maxOpenDurationStr string
	if maxOpenDuration != nil {
		maxOpenDurationStr = (time.Duration)(*maxOpenDuration).String()
	}

	inputs := TeamEnvironmentPermissionInput{
		Organization:    permID.Organization,
		Team:            permID.Team,
		Project:         permID.Project,
		Environment:     permID.Environment,
		Permission:      *permission,
		MaxOpenDuration: maxOpenDurationStr,
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

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Create(
	req *pulumirpc.CreateRequest,
) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	var input TeamEnvironmentPermissionInput
	err := util.FromProperties(req.GetProperties(), structTagKey, &input)
	if err != nil {
		return nil, err
	}

	var maxOpenDuration *pulumiapi.Duration
	if input.MaxOpenDuration != "" {
		d, err := time.ParseDuration(input.MaxOpenDuration)
		if err != nil {
			return nil, err
		}
		maxOpenDuration = (*pulumiapi.Duration)(&d)
	}

	request := pulumiapi.CreateTeamEnvironmentSettingsRequest{
		TeamEnvironmentSettingsRequest: pulumiapi.TeamEnvironmentSettingsRequest{
			Organization: input.Organization,
			Team:         input.Team,
			Project:      input.Project,
			Environment:  input.Environment,
		},
		Permission:      input.Permission,
		MaxOpenDuration: maxOpenDuration,
	}

	err = tp.Client.AddEnvironmentSettings(ctx, request)
	if err != nil {
		return nil, err
	}

	environmentPermissionID := teamEnvironmentPermissionID{
		Organization: input.Organization,
		Team:         input.Team,
		Project:      input.Project,
		Environment:  input.Environment,
	}

	return &pulumirpc.CreateResponse{
		Id:         environmentPermissionID.String(),
		Properties: req.GetProperties(),
	}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	var input TeamEnvironmentPermissionInput
	err := util.FromProperties(req.GetProperties(), structTagKey, &input)
	if err != nil {
		return nil, err
	}
	request := pulumiapi.TeamEnvironmentSettingsRequest{
		Organization: input.Organization,
		Team:         input.Team,
		Project:      input.Project,
		Environment:  input.Environment,
	}
	err = tp.Client.RemoveEnvironmentSettings(ctx, request)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (tp *PulumiServiceTeamEnvironmentPermissionResource) Diff(
	req *pulumirpc.DiffRequest,
) (*pulumirpc.DiffResponse, error) {
	changedKeys, err := util.DiffOldsAndNews(req)
	if err != nil {
		return nil, err
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(changedKeys) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            changedKeys,
		DeleteBeforeReplace: true,
	}, nil
}

// Update does nothing because we always replace on changes, never an update
func (tp *PulumiServiceTeamEnvironmentPermissionResource) Update(
	_ *pulumirpc.UpdateRequest,
) (*pulumirpc.UpdateResponse, error) {
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

type teamEnvironmentPermissionID struct {
	Organization string
	Team         string
	Project      string
	Environment  string
}

func (s *teamEnvironmentPermissionID) String() string {
	return fmt.Sprintf("%s/%s/%s+%s", s.Organization, s.Team, s.Project, s.Environment)
}

func splitTeamEnvironmentPermissionID(id string) (teamEnvironmentPermissionID, error) {
	split := strings.Split(id, "/")
	if len(split) != 3 {
		return teamEnvironmentPermissionID{}, fmt.Errorf("invalid id %q, expected 3 parts", id)
	}

	splitProjectEnv := strings.Split(split[2], "+")
	if len(splitProjectEnv) == 1 {
		return teamEnvironmentPermissionID{
			Organization: split[0],
			Team:         split[1],
			Project:      "default",
			Environment:  splitProjectEnv[0],
		}, nil
	}
	if len(splitProjectEnv) == 2 {
		return teamEnvironmentPermissionID{
			Organization: split[0],
			Team:         split[1],
			Project:      splitProjectEnv[0],
			Environment:  splitProjectEnv[1],
		}, nil
	}

	return teamEnvironmentPermissionID{}, fmt.Errorf(
		"invalid id %q, expected environment name or project/environment in last part",
		id,
	)
}
