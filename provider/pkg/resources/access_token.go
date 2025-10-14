// Package resources implements Pulumi Service resource operations.
package resources

import (
	"context"
	"fmt"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type PulumiServiceAccessTokenResource struct {
	Client *pulumiapi.Client
}

type PulumiServiceAccessTokenInput struct {
	Description string
}

// AccessToken uses outdated way of storing input in internal __inputs property
func GenerateAcessTokenProperties(
	input PulumiServiceAccessTokenInput,
	accessToken pulumiapi.AccessToken,
) (outputs *structpb.Struct, inputs *structpb.Struct, err error) {
	inputMap := resource.PropertyMap{}
	inputMap["description"] = resource.NewPropertyValue(input.Description)
	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputMap)
	outputStore["description"] = resource.NewPropertyValue(input.Description)
	outputStore["value"] = resource.MakeSecret(resource.NewPropertyValue(accessToken.TokenValue))

	inputs, err = plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, nil, err
	}

	outputs, err = plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, nil, err
	}

	return outputs, inputs, err
}

func (at *PulumiServiceAccessTokenResource) ToPulumiServiceAccessTokenInput(
	inputMap resource.PropertyMap,
) PulumiServiceAccessTokenInput {
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
	return diffAccessTokenProperties(req, []string{"description"})
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
	inputMap, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	input := at.ToPulumiServiceAccessTokenInput(inputMap)
	accessToken, err := at.createAccessToken(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error creating access token '%s': %s", input.Description, err.Error())
	}

	outputProperties, _, err := GenerateAcessTokenProperties(input, *accessToken)
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
	accessToken, err := at.Client.GetAccessToken(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if accessToken == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	propertyMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{})
	if err != nil {
		return nil, err
	}
	if propertyMap["value"].HasValue() {
		accessToken.TokenValue = util.GetSecretOrStringValue(propertyMap["value"])
	}

	input := PulumiServiceAccessTokenInput{
		Description: accessToken.Description,
	}
	outputProperties, inputs, err := GenerateAcessTokenProperties(input, *accessToken)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         accessToken.ID,
		Properties: outputProperties,
		Inputs:     inputs,
	}, nil
}

func (at *PulumiServiceAccessTokenResource) createAccessToken(
	ctx context.Context,
	input PulumiServiceAccessTokenInput,
) (*pulumiapi.AccessToken, error) {

	accessToken, err := at.Client.CreateAccessToken(ctx, input.Description)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (at *PulumiServiceAccessTokenResource) deleteAccessToken(ctx context.Context, tokenID string) error {
	return at.Client.DeleteAccessToken(ctx, tokenID)
}

func diffAccessTokenProperties(req *pulumirpc.DiffRequest, replaceProps []string) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	inputs, ok := olds["__inputs"]
	if !ok {
		return nil, fmt.Errorf("missing __inputs property")
	}
	diffs := inputs.ObjectValue().Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	changes, replaces := pulumirpc.DiffResponse_DIFF_NONE, []string(nil)
	for _, k := range replaceProps {
		if diffs.Changed(resource.PropertyKey(k)) {
			changes = pulumirpc.DiffResponse_DIFF_SOME
			replaces = append(replaces, k)
		}
	}

	return &pulumirpc.DiffResponse{
		Changes:  changes,
		Replaces: replaces,
	}, nil
}
