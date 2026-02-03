package resources

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceAgentPoolResource struct {
	Client pulumiapi.AgentPoolClient
}

type PulumiServiceAgentPoolInput struct {
	OrgName      string
	Description  string
	Name         string
	ForceDestroy bool
}

func GenerateAgentPoolProperties(
	input PulumiServiceAgentPoolInput,
	agentPool pulumiapi.AgentPool,
) (outputs *structpb.Struct, inputs *structpb.Struct, err error) {
	inputMap := resource.PropertyMap{}
	inputMap["name"] = resource.NewPropertyValue(input.Name)
	inputMap["organizationName"] = resource.NewPropertyValue(input.OrgName)
	if input.Description != "" {
		inputMap["description"] = resource.NewPropertyValue(input.Description)
	}
	if input.ForceDestroy {
		inputMap["forceDestroy"] = resource.NewPropertyValue(input.ForceDestroy)
	}

	outputMap := resource.PropertyMap{}
	outputMap["agentPoolId"] = resource.NewPropertyValue(agentPool.ID)
	outputMap["name"] = inputMap["name"]
	outputMap["organizationName"] = inputMap["organizationName"]
	outputMap["tokenValue"] = resource.NewPropertyValue(agentPool.TokenValue)
	if input.Description != "" {
		outputMap["description"] = inputMap["description"]
	}
	if input.ForceDestroy {
		outputMap["forceDestroy"] = inputMap["forceDestroy"]
	}

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

func (ap *PulumiServiceAgentPoolResource) ToPulumiServiceAgentPoolInput(
	inputMap resource.PropertyMap,
) PulumiServiceAgentPoolInput {
	input := PulumiServiceAgentPoolInput{}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.Name = inputMap["name"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrgName = inputMap["organizationName"].StringValue()
	}

	if inputMap["forceDestroy"].HasValue() && inputMap["forceDestroy"].IsBool() {
		input.ForceDestroy = inputMap["forceDestroy"].BoolValue()
	}

	return input
}

func (ap *PulumiServiceAgentPoolResource) Name() string {
	return "pulumiservice:index:AgentPool"
}

func (ap *PulumiServiceAgentPoolResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaceProperties := map[string]bool{
		"organizationName": true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(detailedDiffs) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:         changes,
		DetailedDiff:    detailedDiffs,
		HasDetailedDiff: true,
	}, nil
}

func (ap *PulumiServiceAgentPoolResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	pool := ap.ToPulumiServiceAgentPoolInput(inputs)
	if err != nil {
		return nil, err
	}

	err = ap.deleteAgentPool(ctx, req.Id, pool.ForceDestroy)

	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (ap *PulumiServiceAgentPoolResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	input := ap.ToPulumiServiceAgentPoolInput(inputMap)
	agentPool, err := ap.createAgentPool(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error creating agent pool '%s': %s", input.Name, err.Error())
	}

	outputProperties, _, err := GenerateAgentPoolProperties(input, *agentPool)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         input.OrgName + "/" + input.Name + "/" + agentPool.ID,
		Properties: outputProperties,
	}, nil

}

func (ap *PulumiServiceAgentPoolResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (ap *PulumiServiceAgentPoolResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()

	// ignore orgName because if that changed, we would have done a replace, so update would never have been called
	_, _, agentPoolId, err := splitAgentPoolId(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %v", err)
	}

	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	changedInputs := olds
	changedInputs["name"] = news["name"]
	changedInputs["description"] = news["description"]
	changedInputs["forceDestroy"] = news["forceDestroy"]

	inputsAgentPool := ap.ToPulumiServiceAgentPoolInput(changedInputs)
	err = ap.updateAgentPool(ctx, agentPoolId, inputsAgentPool)
	if err != nil {
		return nil, fmt.Errorf("error updating agent pool '%s': %s", inputsAgentPool.Name, err.Error())
	}

	outputProperties, err := plugin.MarshalProperties(
		changedInputs,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (ap *PulumiServiceAgentPoolResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	urn := req.GetId()

	orgName, _, agentPoolId, err := splitAgentPoolId(urn)
	if err != nil {
		return nil, err
	}

	// the agent id is immutable; if we get nil it got deleted, otherwise all data is the same
	agentPool, err := ap.Client.GetAgentPool(ctx, agentPoolId, orgName)
	if err != nil {
		return nil, err
	}
	if agentPool == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	propertyMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepSecrets: true})
	if err != nil {
		return nil, err
	}
	if propertyMap["tokenValue"].HasValue() {
		agentPool.TokenValue = util.GetSecretOrStringValue(propertyMap["tokenValue"])
	}

	input := PulumiServiceAgentPoolInput{
		OrgName:     orgName,
		Description: agentPool.Description,
		Name:        agentPool.Name,
	}
	outputProperties, inputs, err := GenerateAgentPoolProperties(input, *agentPool)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: outputProperties,
		Inputs:     inputs,
	}, nil
}

func (ap *PulumiServiceAgentPoolResource) createAgentPool(
	ctx context.Context,
	input PulumiServiceAgentPoolInput,
) (*pulumiapi.AgentPool, error) {
	agentPool, err := ap.Client.CreateAgentPool(ctx, input.OrgName, input.Name, input.Description)
	if err != nil {
		return nil, err
	}

	return agentPool, nil
}

func (ap *PulumiServiceAgentPoolResource) updateAgentPool(
	ctx context.Context,
	agentPoolId string,
	input PulumiServiceAgentPoolInput,
) error {
	return ap.Client.UpdateAgentPool(ctx, agentPoolId, input.OrgName, input.Name, input.Description)
}

func (ap *PulumiServiceAgentPoolResource) deleteAgentPool(ctx context.Context, id string, forceDestroy bool) error {
	// we don't need the token name when we delete
	orgName, _, agentPoolId, err := splitAgentPoolId(id)
	if err != nil {
		return err
	}
	return ap.Client.DeleteAgentPool(ctx, agentPoolId, orgName, forceDestroy)

}

func splitAgentPoolId(id string) (string, string, string, error) {
	// format: organization/name/agentPoolId
	s := strings.Split(id, "/")
	if len(s) != 3 {
		return "", "", "", fmt.Errorf("%q is invalid, must be in the format: organization/name/agentPoolId", id)
	}
	return s[0], s[1], s[2], nil
}
