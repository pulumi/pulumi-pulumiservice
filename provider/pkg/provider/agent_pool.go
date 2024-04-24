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

type PulumiServiceAgentPoolResource struct{}

type PulumiServiceAgentPoolInput struct {
	OrgName     string
	Description string
	Name        string
}

func (i *PulumiServiceAgentPoolInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["description"] = resource.NewPropertyValue(i.Description)
	pm["organizationName"] = resource.NewPropertyValue(i.OrgName)
	return pm
}

func (ap *PulumiServiceAgentPoolResource) ToPulumiServiceAgentPoolInput(inputMap resource.PropertyMap) PulumiServiceAgentPoolInput {
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

	return input
}

func (ap *PulumiServiceAgentPoolResource) Name() string {
	return "pulumiservice:index:AgentPool"
}

func (ap *PulumiServiceAgentPoolResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	// preprocess olds to remove the `tokenValue & agentPoolId` property since it's only an output and shouldn't cause a diff
	for _, p := range []resource.PropertyKey{"tokenValue", "agentPoolId"} {
		if olds[p].HasValue() {
			delete(olds, p)
		}
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

func (ap *PulumiServiceAgentPoolResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	err := ap.deleteAgentPool(ctx, req.Id)

	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (ap *PulumiServiceAgentPoolResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsAgentPool := ap.ToPulumiServiceAgentPoolInput(inputs)

	agentPool, err := ap.createAgentPool(ctx, inputsAgentPool)
	if err != nil {
		return nil, fmt.Errorf("error creating access token '%s': %s", inputsAgentPool.Name, err.Error())
	}

	outputStore := resource.PropertyMap{}
	outputStore["agentPoolId"] = resource.NewPropertyValue(agentPool.ID)
	outputStore["name"] = inputs["name"]
	outputStore["organizationName"] = inputs["organizationName"]
	outputStore["description"] = inputs["description"]
	outputStore["tokenValue"] = resource.NewPropertyValue(agentPool.TokenValue)

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	urn := fmt.Sprintf(inputsAgentPool.OrgName + "/" + inputsAgentPool.Name + "/" + agentPool.ID)

	return &pulumirpc.CreateResponse{
		Id:         urn,
		Properties: outputProperties,
	}, nil

}

func (ap *PulumiServiceAgentPoolResource) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (ap *PulumiServiceAgentPoolResource) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {

	// ignore orgName because if that changed, we would have done a replace, so update would never have been called
	_, _, agentPoolId, err := splitAgentPoolId(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %v", err)
	}

	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
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

func (ap *PulumiServiceAgentPoolResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	urn := req.GetId()

	orgName, _, agentPoolId, err := splitAgentPoolId(urn)
	if err != nil {
		return nil, err
	}

	// the agent id is immutable; if we get nil it got deleted, otherwise all data is the same
	agentPool, err := GetClient[pulumiapi.AgentPoolClient](ctx).GetAgentPool(ctx, agentPoolId, orgName)
	if err != nil {
		return nil, err
	}
	if agentPool == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: req.GetProperties(),
	}, nil
}

func (ap *PulumiServiceAgentPoolResource) Invoke(_ *pulumiserviceProvider, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return &pulumirpc.InvokeResponse{Return: nil}, fmt.Errorf("unknown function '%s'", req.Tok)
}

func (ap *PulumiServiceAgentPoolResource) createAgentPool(ctx context.Context, input PulumiServiceAgentPoolInput) (*pulumiapi.AgentPool, error) {
	agentPool, err := GetClient[pulumiapi.AgentPoolClient](ctx).CreateAgentPool(ctx, input.OrgName, input.Name, input.Description)
	if err != nil {
		return nil, err
	}

	return agentPool, nil
}

func (ap *PulumiServiceAgentPoolResource) updateAgentPool(ctx context.Context, agentPoolId string, input PulumiServiceAgentPoolInput) error {
	return GetClient[pulumiapi.AgentPoolClient](ctx).UpdateAgentPool(ctx, agentPoolId, input.OrgName, input.Name, input.Description)
}

func (ap *PulumiServiceAgentPoolResource) deleteAgentPool(ctx context.Context, id string) error {
	// we don't need the token name when we delete
	orgName, _, agentPoolId, err := splitAgentPoolId(id)
	if err != nil {
		return err
	}
	return GetClient[pulumiapi.AgentPoolClient](ctx).DeleteAgentPool(ctx, agentPoolId, orgName)

}

func splitAgentPoolId(id string) (string, string, string, error) {
	// format: organization/name/agentPoolId
	s := strings.Split(id, "/")
	if len(s) != 3 {
		return "", "", "", fmt.Errorf("%q is invalid, must be in the format: organization/name/agentPoolId", id)
	}
	return s[0], s[1], s[2], nil
}
