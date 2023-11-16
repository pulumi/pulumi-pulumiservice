package provider

import (
	"context"
	"fmt"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceAccessTokenResource struct {
	client *pulumiapi.Client
}

type PulumiServiceAccessTokenInput struct {
	Description string
}

func (i *PulumiServiceAccessTokenInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["description"] = resource.NewPropertyValue(i.Description)
	return pm
}

func (at *PulumiServiceAccessTokenResource) ToPulumiServiceAccessTokenInput(inputMap resource.PropertyMap) PulumiServiceAccessTokenInput {
	input := PulumiServiceAccessTokenInput{}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	return input
}

func (at *PulumiServiceAccessTokenResource) Name() string {
	return "pulumiservice:index:AccessToken"
}

func (at *PulumiServiceAccessTokenResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return considerAllChangesReplaces(req)
}

func (at *PulumiServiceAccessTokenResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	err := at.deleteAccessToken(ctx, req.Id)
	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (at *PulumiServiceAccessTokenResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsAccessToken := at.ToPulumiServiceAccessTokenInput(inputs)
	accessToken, err := at.createAccessToken(ctx, inputsAccessToken)
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

	return &pulumirpc.CreateResponse{
		Id:         accessToken.ID,
		Properties: outputProperties,
	}, nil

}

func (at *PulumiServiceAccessTokenResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (at *PulumiServiceAccessTokenResource) Update(_ *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// all updates are destructive, so we just call Create.
	return nil, fmt.Errorf("unexpected call to update, expected create to be called instead")
}

func (at *PulumiServiceAccessTokenResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	// the access token is immutable; if we get nil it got deleted, otherwise all data is the same
	accessToken, err := at.client.GetAccessToken(ctx, req.GetId())
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

func (at *PulumiServiceAccessTokenResource) Invoke(_ *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (at *PulumiServiceAccessTokenResource) Configure(_ PulumiServiceConfig) {
}

func (at *PulumiServiceAccessTokenResource) createAccessToken(ctx context.Context, input PulumiServiceAccessTokenInput) (*pulumiapi.AccessToken, error) {

	accessToken, err := at.client.CreateAccessToken(ctx, input.Description)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (at *PulumiServiceAccessTokenResource) deleteAccessToken(ctx context.Context, tokenId string) error {
	return at.client.DeleteAccessToken(ctx, tokenId)
}
