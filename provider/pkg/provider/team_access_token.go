package provider

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
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
	err := t.deleteTeamAccessToken(ctx, req.Id)

	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsAccessToken := t.ToPulumiServiceAccessTokenInput(inputs)

	accessToken, err := t.createTeamAccessToken(ctx, inputsAccessToken)
	if err != nil {
		return nil, fmt.Errorf("error creating access token '%s': %s", inputsAccessToken.Description, err.Error())
	}

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputs)
	outputStore["value"] = resource.NewPropertyValue(accessToken.TokenValue)

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	urn := fmt.Sprintf(inputsAccessToken.OrgName + "/" + inputsAccessToken.TeamName + "/" + inputsAccessToken.Name + "/" + accessToken.ID)

	return &pulumirpc.CreateResponse{
		Id:         urn,
		Properties: outputProperties,
	}, nil

}

func (t *PulumiServiceTeamAccessTokenResource) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Update(context.Context, *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (t *PulumiServiceTeamAccessTokenResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	urn := req.GetId()

	orgName, teamName, _, tokenId, err := splitTeamAccessTokenId(urn)

	// the team access token is immutable; if we get nil it got deleted, otherwise all data is the same
	accessToken, err := GetClient[*pulumiapi.Client](ctx).GetTeamAccessToken(ctx, tokenId, orgName, teamName)
	if err != nil {
		return nil, err
	}
	if accessToken == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: req.GetProperties(),
	}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Invoke(_ *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (t *PulumiServiceTeamAccessTokenResource) createTeamAccessToken(ctx context.Context, input PulumiServiceTeamAccessTokenInput) (*pulumiapi.AccessToken, error) {

	accessToken, err := GetClient[*pulumiapi.Client](ctx).CreateTeamAccessToken(ctx, input.Name, input.OrgName, input.TeamName, input.Description)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (t *PulumiServiceTeamAccessTokenResource) deleteTeamAccessToken(ctx context.Context, id string) error {
	orgName, teamName, _, tokenId, err := splitTeamAccessTokenId(id)
	if err != nil {
		return err
	}
	return GetClient[*pulumiapi.Client](ctx).DeleteTeamAccessToken(ctx, tokenId, orgName, teamName)

}

func splitTeamAccessTokenId(id string) (string, string, string, string, error) {
	// format: organization/teamName/tokenName/tokenId
	s := strings.Split(id, "/")
	if len(s) != 4 {
		return "", "", "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], s[2], s[3], nil
}
