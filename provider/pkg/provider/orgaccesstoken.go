package provider

import (
	"context"
	"fmt"
	"strings"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	//"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceOrgAccessTokenResource struct {
	client *pulumiapi.Client
}

type PulumiServiceOrgAccessTokenInput struct {
	OrgName     string
	Description string
	Name        string
}

func (i *PulumiServiceOrgAccessTokenInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["description"] = resource.NewPropertyValue(i.Description)
	pm["organizationName"] = resource.NewPropertyValue(i.OrgName)
	return pm
}

func (t *PulumiServiceOrgAccessTokenResource) ToPulumiServiceOrgAccessTokenInput(inputMap resource.PropertyMap) PulumiServiceOrgAccessTokenInput {
	input := PulumiServiceOrgAccessTokenInput{}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.Name = inputMap["name"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrgName = inputMap["organizationName"].StringValue()
	}

	return input
}

func (c PulumiServiceOrgAccessTokenResource) Name() string {
	return "pulumiservice:index:OrgAccessToken"
}

func (c *PulumiServiceOrgAccessTokenResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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

func (c *PulumiServiceOrgAccessTokenResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	err := c.deleteOrgAccessToken(ctx, req.Id)

	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (at *PulumiServiceOrgAccessTokenResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsAccessToken := at.ToPulumiServiceOrgAccessTokenInput(inputs)

	accessToken, err := at.createOrgAccessToken(ctx, inputsAccessToken)
	if err != nil {
		return nil, fmt.Errorf("error creating access token '%s': %s", inputsAccessToken.Name, err.Error())
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

	urn := fmt.Sprintf(inputsAccessToken.OrgName + "/" + inputsAccessToken.Name + "/" + accessToken.ID)

	return &pulumirpc.CreateResponse{
		Id:         urn,
		Properties: outputProperties,
	}, nil

}

func (k *PulumiServiceOrgAccessTokenResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (k *PulumiServiceOrgAccessTokenResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}

func (k *PulumiServiceOrgAccessTokenResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return &pulumirpc.ReadResponse{}, nil
}

func (f *PulumiServiceOrgAccessTokenResource) Invoke(s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (at *PulumiServiceOrgAccessTokenResource) Configure(config PulumiServiceConfig) {
}

func (at *PulumiServiceOrgAccessTokenResource) createOrgAccessToken(ctx context.Context, input PulumiServiceOrgAccessTokenInput) (*pulumiapi.AccessToken, error) {

	accesstoken, err := at.client.CreateOrgAccessToken(ctx, input.Name, input.OrgName, input.Description)
	if err != nil {
		return nil, err
	}

	return accesstoken, nil
}

func (at *PulumiServiceOrgAccessTokenResource) deleteOrgAccessToken(ctx context.Context, id string) error {
	// we don't need the token name when we delete
	orgName, _, tokenId, err := splitOrgAccessTokenId(id)
	if err != nil {
		return err
	}
	return at.client.DeleteOrgAccessToken(ctx, tokenId, orgName)

}

// FIXME: we can likely create a util that will work for all cases
func splitOrgAccessTokenId(id string) (string, string, string, error) {
	// format: organization/name/tokenId
	s := strings.Split(id, "/")
	if len(s) != 3 {
		return "", "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], s[2], nil
}
