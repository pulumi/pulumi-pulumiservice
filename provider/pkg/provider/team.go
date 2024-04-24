package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil/rpcerror"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTeamResource struct{}

type PulumiServiceTeamStackPermission struct {
	ProjectName string
	StackName   string
	Permission  int
}

type PulumiServiceTeamInput struct {
	Type             string
	Name             string
	DisplayName      string
	Description      string
	OrganizationName string
	Members          []string
	GitHubTeamID     int64
}

func (i *PulumiServiceTeamInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["teamType"] = resource.NewPropertyValue(i.Type)
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["members"] = resource.NewPropertyValue(i.Members)
	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)
	if i.Description != "" {
		pm["description"] = resource.NewPropertyValue(i.Description)
	}
	if i.GitHubTeamID != 0 {
		pm["githubTeamId"] = resource.NewPropertyValue(i.GitHubTeamID)
	}
	if i.DisplayName != "" {
		pm["displayName"] = resource.NewPropertyValue(i.DisplayName)
	}
	return pm
}

func (i *PulumiServiceTeamInput) ToRpc() (*structpb.Struct, error) {
	return plugin.MarshalProperties(i.ToPropertyMap(), plugin.MarshalOptions{
		KeepOutputValues: true,
	})
}

func ToPulumiServiceTeamInput(inputMap resource.PropertyMap) PulumiServiceTeamInput {
	input := PulumiServiceTeamInput{}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.Name = inputMap["name"].StringValue()
	}

	if inputMap["teamType"].HasValue() && inputMap["teamType"].IsString() {
		input.Type = inputMap["teamType"].StringValue()
	}

	if inputMap["displayName"].HasValue() && inputMap["displayName"].IsString() {
		input.DisplayName = inputMap["displayName"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	if inputMap["members"].HasValue() && inputMap["members"].IsArray() {
		for _, m := range inputMap["members"].ArrayValue() {
			if m.HasValue() && m.IsString() {
				input.Members = append(input.Members, m.StringValue())
			}
		}

		// Sort the members so the order is deterministic
		slices.Sort(input.Members)
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrganizationName = inputMap["organizationName"].StringValue()
	}

	if inputMap["githubTeamId"].HasValue() && inputMap["githubTeamId"].IsNumber() {
		input.GitHubTeamID = int64(inputMap["githubTeamId"].NumberValue())
	}

	return input
}

func (t *PulumiServiceTeamResource) Name() string {
	return "pulumiservice:index:Team"
}

func (t *PulumiServiceTeamResource) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news := req.GetNews()
	newsMap, err := plugin.UnmarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure

	var teamType string
	if newsMap["teamType"].HasValue() {
		teamType = newsMap["teamType"].StringValue()
	}

	if teamType != "github" && teamType != "pulumi" {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   fmt.Sprintf("found %q instead of 'pulumi' or 'github'", teamType),
			Property: "type",
		})
	}

	if teamType == "github" && !newsMap["githubTeamId"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "teams with teamType 'github' require a githubTeamId",
			Property: "githubTeamId",
		})
	}

	if teamType == "pulumi" && !newsMap["name"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "teams with teamType 'pulumi' require a name",
			Property: "name",
		})
	}

	return &pulumirpc.CheckResponse{Inputs: news, Failures: failures}, nil
}

func (t *PulumiServiceTeamResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	err := t.deleteTeam(ctx, req.Id)
	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (t *PulumiServiceTeamResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	oldTeam := ToPulumiServiceTeamInput(olds)
	newTeam := ToPulumiServiceTeamInput(news)

	changes := pulumirpc.DiffResponse_DIFF_NONE

	if !reflect.DeepEqual(oldTeam, newTeam) {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            []string{},
		Stables:             []string{},
		DeleteBeforeReplace: false,
	}, nil
}

func (t *PulumiServiceTeamResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	orgName, teamName, err := splitSingleSlashString(req.Id)
	if err != nil {
		return nil, err
	}

	team, err := GetClient[pulumiapi.TeamClient](ctx).GetTeam(ctx, orgName, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to read Team (%q): %w", req.Id, err)
	}
	if team == nil {
		return &pulumirpc.ReadResponse{}, nil
	}
	inputs := PulumiServiceTeamInput{
		Description:      team.Description,
		DisplayName:      team.DisplayName,
		Name:             team.Name,
		Type:             team.Type,
		OrganizationName: orgName,
	}
	for _, m := range team.Members {
		inputs.Members = append(inputs.Members, m.GithubLogin)
	}
	// Sort the members so the order is deterministic
	slices.Sort(inputs.Members)
	props, err := plugin.MarshalProperties(inputs.ToPropertyMap(), plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}

func (t *PulumiServiceTeamResource) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	inputsOld, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	teamOld := ToPulumiServiceTeamInput(inputsOld)
	teamNew := ToPulumiServiceTeamInput(inputsNew)

	inputsChanged := teamOld
	if teamOld.Description != teamNew.Description || teamOld.DisplayName != teamNew.Description {

		inputsChanged.Description = teamNew.Description
		inputsChanged.DisplayName = teamNew.DisplayName

		err = t.updateTeam(ctx, inputsChanged)
		if err != nil {
			return nil, err
		}
	}

	// github teams can't manage membership.
	if !slices.Equal(teamOld.Members, teamNew.Members) && teamNew.Type != "github" {
		inputsChanged.Members = teamNew.Members
		for _, usernameToDelete := range teamOld.Members {
			if !slices.Contains(teamNew.Members, usernameToDelete) {
				err := t.deleteFromTeam(ctx, teamNew.OrganizationName, teamNew.Name, usernameToDelete)
				if err != nil {
					return nil, err
				}
			}
		}

		for _, usernameToAdd := range teamNew.Members {
			if !slices.Contains(teamOld.Members, usernameToAdd) {
				err := t.addToTeam(ctx, teamNew.OrganizationName, teamNew.Name, usernameToAdd)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	outputProperties, err := teamNew.ToRpc()
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (t *PulumiServiceTeamResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsTeam := ToPulumiServiceTeamInput(inputs)
	team, err := t.createTeam(ctx, inputsTeam)
	if err != nil {
		return nil, fmt.Errorf("error creating team '%s': %s", inputsTeam.Name, err.Error())
	}

	// We have now created a team.  It is very important to ensure that from this point on, any other error
	// below returns the ID using the `pulumirpc.ErrorResourceInitFailed` error details annotation.  Otherwise,
	// we leak a team resource. We ensure that we wrap any errors in a partial error and return that to the RPC.

	// make copy of input so we can safely modify output without affecting input
	inProgTeam := ToPulumiServiceTeamInput(inputsTeam.ToPropertyMap())
	inProgTeam.Members = nil

	for _, memberToAdd := range inputsTeam.Members {
		err := t.addToTeam(ctx, inputsTeam.OrganizationName, inputsTeam.Name, memberToAdd)
		if err != nil {
			return nil, partialError(*team, err, inProgTeam, inputsTeam)
		}
		// if we've successfully added member to team, save them to the state we're going to return
		// so that a re-run can detect the left over members to add via Update
		inProgTeam.Members = append(inProgTeam.Members, memberToAdd)
	}

	outputProperties, err := inProgTeam.ToRpc()
	if err != nil {
		return nil, partialError(*team, err, inProgTeam, inputsTeam)
	}

	return &pulumirpc.CreateResponse{
		Id:         *team,
		Properties: outputProperties,
	}, nil
}

func (t *PulumiServiceTeamResource) updateTeam(ctx context.Context, input PulumiServiceTeamInput) error {
	err := GetClient[pulumiapi.TeamClient](ctx).UpdateTeam(ctx, input.OrganizationName, input.Name, input.DisplayName, input.Description)
	if err != nil {
		return err
	}
	return nil
}

func (t *PulumiServiceTeamResource) createTeam(ctx context.Context, input PulumiServiceTeamInput) (*string, error) {
	team, err := GetClient[pulumiapi.TeamClient](ctx).CreateTeam(ctx, input.OrganizationName, input.Name, input.Type, input.DisplayName, input.Description, input.GitHubTeamID)
	if err != nil {
		return nil, err
	}

	teamUrn := fmt.Sprintf("%s/%s", input.OrganizationName, team.Name)
	return &teamUrn, nil
}

func (t *PulumiServiceTeamResource) deleteFromTeam(ctx context.Context, orgName string, teamName string, userName string) error {
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	if len(userName) == 0 {
		return errors.New("username must not be empty")
	}

	return GetClient[pulumiapi.TeamClient](ctx).DeleteMemberFromTeam(ctx, orgName, teamName, userName)
}

func (t *PulumiServiceTeamResource) addToTeam(ctx context.Context, orgName string, teamName string, userName string) error {
	return GetClient[pulumiapi.TeamClient](ctx).AddMemberToTeam(ctx, orgName, teamName, userName)
}

func (t *PulumiServiceTeamResource) deleteTeam(ctx context.Context, id string) error {
	orgName, teamName, err := splitSingleSlashString(id)
	if err != nil {
		return err
	}
	err = GetClient[pulumiapi.TeamClient](ctx).DeleteTeam(ctx, orgName, teamName)
	if err != nil {
		return err
	}
	return nil
}

// partialError creates an error for resources that did not complete an operation in progress.
// The last known state of the object is included in the error so that it can be checkpointed.
func partialError(id string, err error, state PulumiServiceTeamInput, inputs PulumiServiceTeamInput) error {
	stateRpc, stateSerErr := state.ToRpc()
	inputRpc, inputSerErr := inputs.ToRpc()

	// combine errors if we can't serialize state or inputs for some reason
	if stateSerErr != nil {
		err = fmt.Errorf("err serializing state: %v, (src error: %v)", stateSerErr, err)
	}
	if inputSerErr != nil {
		err = fmt.Errorf("err serializing inputs: %v (src error: %v)", inputSerErr, err)
	}
	detail := pulumirpc.ErrorResourceInitFailed{
		Id:         id,
		Properties: stateRpc,
		Reasons:    []string{err.Error()},
		Inputs:     inputRpc,
	}
	return rpcerror.WithDetails(rpcerror.New(codes.Unknown, err.Error()), &detail)
}

func splitSingleSlashString(id string) (string, string, error) {
	// format: organization/webhookName
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], nil
}
