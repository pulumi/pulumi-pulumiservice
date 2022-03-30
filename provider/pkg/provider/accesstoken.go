package provider

import (
	"errors"
	"fmt"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	pulumiapi "github.com/pierskarsenbarg/pulumi-apiclient"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceAccessTokenResource struct {
	config PulumiServiceConfig
}

type PulumiServiceAccessTokenInput struct {
	Description string
}

func (i *PulumiServiceAccessTokenInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["description"] = resource.NewPropertyValue(i.Description)
	return pm
}

func (t *PulumiServiceAccessTokenResource) ToPulumiServiceAccessTokenInput(inputMap resource.PropertyMap) PulumiServiceAccessTokenInput {
	input := PulumiServiceAccessTokenInput{}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	return input
}

func (c PulumiServiceAccessTokenResource) Name() string {
	return "pulumiservice:index:AccessToken"
}

func (c *PulumiServiceAccessTokenResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
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
			Changes:             pulumirpc.DiffResponse_DIFF_NONE,
			Replaces:            []string{},
			Stables:             []string{},
			DeleteBeforeReplace: false,
		}, nil
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if diffs.Changed("description") {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            []string{},
		Stables:             []string{},
		DeleteBeforeReplace: false,
	}, nil
}

func (c *PulumiServiceAccessTokenResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	err := c.deleteAccessToken(req.Id)
	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (at *PulumiServiceAccessTokenResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsAccessToken := at.ToPulumiServiceAccessTokenInput(inputs)
	accessToken, err := at.createAccessToken(inputsAccessToken)
	if err != nil {
		return nil, fmt.Errorf("error creating access token '%s': %s", inputsAccessToken.Description, err.Error())
	}

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputs)
	outputStore["value"] = resource.NewPropertyValue(accessToken.Value)

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         accessToken.Id,
		Properties: outputProperties,
	}, nil

}

func (k *PulumiServiceAccessTokenResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (k *PulumiServiceAccessTokenResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, createUnknownResourceErrorFromRequest(req)
}

func (k *PulumiServiceAccessTokenResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, errors.New("error here read") //createUnknownResourceErrorFromRequest(req)
}

func (f *PulumiServiceAccessTokenResource) Invoke(s *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &rpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (at *PulumiServiceAccessTokenResource) Configure(config PulumiServiceConfig) {
	at.config = config
}

func (at *PulumiServiceAccessTokenResource) createAccessToken(input PulumiServiceAccessTokenInput) (*pulumiapi.AccessToken, error) {
	token, err := at.config.getPulumiAccessToken()
	if err != nil {
		return nil, err
	}

	c := pulumiapi.NewClient(*token)

	accesstoken, err := c.CreateAccessToken(input.Description)
	if err != nil {
		return nil, err
	}

	return &accesstoken, nil
}

func (at *PulumiServiceAccessTokenResource) deleteAccessToken(tokenId string) error {
	token, err := at.config.getPulumiAccessToken()
	if err != nil {
		return err
	}

	c := pulumiapi.NewClient(*token)

	err = c.DeleteAccessToken(tokenId)
	if err != nil {
		return err
	}

	return nil
}
