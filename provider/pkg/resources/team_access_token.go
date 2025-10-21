package resources

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTeamAccessTokenResource struct{}

type PulumiServiceTeamAccessTokenInput struct {
	Name        string
	OrgName     string
	TeamName    string
	Description string
}

func GenerateTeamAccessTokenProperties(input PulumiServiceTeamAccessTokenInput, teamAccessToken pulumiapi.AccessToken) (outputs *structpb.Struct, inputs *structpb.Struct, err error) {
	inputMap := input.ToPropertyMap()

	outputMap := inputMap.Copy()
	outputMap["__inputs"] = resource.NewObjectProperty(inputMap)
	outputMap["value"] = resource.MakeSecret(resource.NewPropertyValue(teamAccessToken.TokenValue))

	inputs, err = plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, nil, err
	}

	outputs, err = plugin.MarshalProperties(outputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, nil, err
	}

	return outputs, inputs, err
}

func (i *PulumiServiceTeamAccessTokenInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["description"] = resource.NewPropertyValue(i.Description)
	pm["organizationName"] = resource.NewPropertyValue(i.OrgName)
	pm["teamName"] = resource.NewPropertyValue(i.TeamName)
	return pm
}

func (t *PulumiServiceTeamAccessTokenResource) ToPulumiServiceAccessTokenInput(inputMap resource.PropertyMap) PulumiServiceTeamAccessTokenInput {
	input := PulumiServiceTeamAccessTokenInput{}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.Name = inputMap["name"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	if inputMap["teamName"].HasValue() && inputMap["teamName"].IsString() {
		input.TeamName = inputMap["teamName"].StringValue()
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrgName = inputMap["organizationName"].StringValue()
	}

	return input
}

func (t *PulumiServiceTeamAccessTokenResource) Name() string {
	return "pulumiservice:index:TeamAccessToken"
}

func (t *PulumiServiceTeamAccessTokenResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return diffAccessTokenProperties(req, []string{"name", "organizationName", "teamName", "description"})
}

func (t *PulumiServiceTeamAccessTokenResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	client := config.GetClient[*pulumiapi.Client](ctx)
	err := t.deleteTeamAccessToken(ctx, client, req.Id)

	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := t.ToPulumiServiceAccessTokenInput(inputMap)

	accessToken, err := t.createTeamAccessToken(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error creating access token '%s': %s", input.Name, err.Error())
	}

	outputs, _, err := GenerateTeamAccessTokenProperties(input, *accessToken)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         fmt.Sprintf("%s/%s/%s/%s", input.OrgName, input.TeamName, input.Name, accessToken.ID),
		Properties: outputs,
	}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Update(_ context.Context, _ *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (t *PulumiServiceTeamAccessTokenResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	client := config.GetClient[*pulumiapi.Client](ctx)
	urn := req.GetId()

	orgName, teamName, tokenName, tokenId, err := splitTeamAccessTokenId(urn)
	if err != nil {
		return nil, err
	}

	// the team access token is immutable; if we get nil it got deleted, otherwise all data is the same
	accessToken, err := client.GetTeamAccessToken(ctx, tokenId, orgName, teamName)
	if err != nil {
		return nil, err
	}
	if accessToken == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	var input = PulumiServiceTeamAccessTokenInput{
		Name:        tokenName,
		OrgName:     orgName,
		Description: accessToken.Description,
		TeamName:    teamName,
	}

	propertyMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{})
	if err != nil {
		return nil, err
	}
	if propertyMap["value"].HasValue() {
		accessToken.TokenValue = util.GetSecretOrStringValue(propertyMap["value"])
	}

	outputs, inputs, err := GenerateTeamAccessTokenProperties(input, *accessToken)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: outputs,
		Inputs:     inputs,
	}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) createTeamAccessToken(ctx context.Context, input PulumiServiceTeamAccessTokenInput) (*pulumiapi.AccessToken, error) {
	client := config.GetClient[*pulumiapi.Client](ctx)

	accessToken, err := client.CreateTeamAccessToken(ctx, input.Name, input.OrgName, input.TeamName, input.Description)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (t *PulumiServiceTeamAccessTokenResource) deleteTeamAccessToken(ctx context.Context, client *pulumiapi.Client, id string) error {
	orgName, teamName, _, tokenId, err := splitTeamAccessTokenId(id)
	if err != nil {
		return err
	}
	return client.DeleteTeamAccessToken(ctx, tokenId, orgName, teamName)

}

func splitTeamAccessTokenId(id string) (string, string, string, string, error) {
	// format: organization/teamName/tokenName/tokenId
	s := strings.Split(id, "/")
	if len(s) != 4 {
		return "", "", "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], s[2], s[3], nil
}
