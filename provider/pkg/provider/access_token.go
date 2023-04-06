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

func (at *PulumiServiceAccessTokenResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return &pulumirpc.UpdateResponse{}, nil
}

func (at *PulumiServiceAccessTokenResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return &pulumirpc.ReadResponse{}, nil
}

func (at *PulumiServiceAccessTokenResource) Invoke(s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (at *PulumiServiceAccessTokenResource) Configure(config PulumiServiceConfig) {
}

func (at *PulumiServiceAccessTokenResource) createAccessToken(ctx context.Context, input PulumiServiceAccessTokenInput) (*pulumiapi.AccessToken, error) {

	accesstoken, err := at.client.CreateAccessToken(ctx, input.Description)
	if err != nil {
		return nil, err
	}

	return accesstoken, nil
}

func (at *PulumiServiceAccessTokenResource) deleteAccessToken(ctx context.Context, tokenId string) error {
	return at.client.DeleteAccessToken(ctx, tokenId)
}
