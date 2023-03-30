package provider

import (
	"context"
	"fmt"
	"strings"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTeamAccessTokenResource struct {
	client *pulumiapi.Client
}

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

func (t *PulumiServiceTeamAccessTokenResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	diffs := olds["__inputs"].ObjectValue().Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if diffs.Changed("description") {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:  changes,
		Replaces: []string{"description"},
	}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	err := t.deleteTeamAccessToken(ctx, req.Id)

	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
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

func (t *PulumiServiceTeamAccessTokenResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return &pulumirpc.ReadResponse{}, nil
}

func (t *PulumiServiceTeamAccessTokenResource) Invoke(s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (t *PulumiServiceTeamAccessTokenResource) Configure(config PulumiServiceConfig) {
}

func (t *PulumiServiceTeamAccessTokenResource) createTeamAccessToken(ctx context.Context, input PulumiServiceTeamAccessTokenInput) (*pulumiapi.AccessToken, error) {

	accesstoken, err := t.client.CreateTeamAccessToken(ctx, input.Name, input.OrgName, input.TeamName, input.Description)
	if err != nil {
		return nil, err
	}

	return accesstoken, nil
}

func (t *PulumiServiceTeamAccessTokenResource) deleteTeamAccessToken(ctx context.Context, id string) error {
	orgName, teamName, _, tokenId, err := splitTeamAccessTokenId(id)
	if err != nil {
		return err
	}
	return t.client.DeleteTeamAccessToken(ctx, tokenId, orgName, teamName)

}

// FIXME: we can likely create a util that will work for all cases
func splitTeamAccessTokenId(id string) (string, string, string, string, error) {
	// format: organization/teamName/tokenName/tokenId
	s := strings.Split(id, "/")
	if len(s) != 4 {
		return "", "", "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], s[2], s[3], nil
}
