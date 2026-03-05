package resources

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type PulumiServiceOrgAccessTokenResource struct {
	Client *pulumiapi.Client
}

type PulumiServiceOrgAccessTokenInput struct {
	OrgName     string
	Description string
	Name        string
	Admin       bool
}

func GenerateOrgAccessTokenProperties(
	input PulumiServiceOrgAccessTokenInput,
	orgAccessToken pulumiapi.AccessToken,
) (outputs *structpb.Struct, inputs *structpb.Struct, err error) {
	inputMap := input.ToPropertyMap()

	outputMap := inputMap.Copy()
	outputMap["__inputs"] = resource.NewObjectProperty(inputMap)
	outputMap["value"] = resource.MakeSecret(resource.NewPropertyValue(orgAccessToken.TokenValue))

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

func (i *PulumiServiceOrgAccessTokenInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["description"] = resource.NewPropertyValue(i.Description)
	pm["organizationName"] = resource.NewPropertyValue(i.OrgName)
	pm["admin"] = resource.NewPropertyValue(i.Admin)
	return pm
}

func (ot *PulumiServiceOrgAccessTokenResource) ToPulumiServiceOrgAccessTokenInput(
	inputMap resource.PropertyMap,
) PulumiServiceOrgAccessTokenInput {
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

	if inputMap["admin"].HasValue() && inputMap["admin"].IsBool() {
		input.Admin = inputMap["admin"].BoolValue()
	}

	return input
}

func (ot *PulumiServiceOrgAccessTokenResource) Name() string {
	return "pulumiservice:index:OrgAccessToken"
}

func (ot *PulumiServiceOrgAccessTokenResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return diffAccessTokenProperties(req, []string{"name", "organizationName", "description", "admin"})
}

func (ot *PulumiServiceOrgAccessTokenResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	err := ot.deleteOrgAccessToken(ctx, req.Id)

	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (ot *PulumiServiceOrgAccessTokenResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	input := ot.ToPulumiServiceOrgAccessTokenInput(inputMap)

	accessToken, err := ot.createOrgAccessToken(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error creating access token '%s': %s", input.Name, err.Error())
	}

	outputs, _, err := GenerateOrgAccessTokenProperties(input, *accessToken)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         fmt.Sprintf("%s/%s/%s", input.OrgName, input.Name, accessToken.ID),
		Properties: outputs,
	}, nil

}

func (ot *PulumiServiceOrgAccessTokenResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (ot *PulumiServiceOrgAccessTokenResource) Update(_ *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (ot *PulumiServiceOrgAccessTokenResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	urn := req.GetId()

	orgName, _, tokenID, err := splitOrgAccessTokenID(urn)
	if err != nil {
		return nil, err
	}

	// the org access token is immutable; if we get nil it got deleted, otherwise all data is the same
	accessToken, err := ot.Client.GetOrgAccessToken(ctx, tokenID, orgName)
	if err != nil {
		return nil, err
	}
	if accessToken == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	var input = PulumiServiceOrgAccessTokenInput{
		Name:        accessToken.Name,
		OrgName:     orgName,
		Description: accessToken.Description,
		Admin:       accessToken.Admin,
	}

	propertyMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{})
	if err != nil {
		return nil, err
	}
	if propertyMap["value"].HasValue() {
		accessToken.TokenValue = util.GetSecretOrStringValue(propertyMap["value"])
	}

	outputs, inputs, err := GenerateOrgAccessTokenProperties(input, *accessToken)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: outputs,
		Inputs:     inputs,
	}, nil
}

func (ot *PulumiServiceOrgAccessTokenResource) createOrgAccessToken(
	ctx context.Context,
	input PulumiServiceOrgAccessTokenInput,
) (*pulumiapi.AccessToken, error) {

	accessToken, err := ot.Client.CreateOrgAccessToken(ctx, input.Name, input.OrgName, input.Description, input.Admin)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (ot *PulumiServiceOrgAccessTokenResource) deleteOrgAccessToken(ctx context.Context, id string) error {
	// we don't need the token name when we delete
	orgName, _, tokenID, err := splitOrgAccessTokenID(id)
	if err != nil {
		return err
	}
	return ot.Client.DeleteOrgAccessToken(ctx, tokenID, orgName)

}

func splitOrgAccessTokenID(id string) (string, string, string, error) {
	// format: organization/name/tokenID
	s := strings.Split(id, "/")
	if len(s) < 3 {
		return "", "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}

	org := s[0]
	tokenID := s[len(s)-1]
	// Name can contain slashes so this joins the split parts except for first and last
	name := strings.Join(s[1:len(s)-1], "/")

	return org, name, tokenID, nil
}
